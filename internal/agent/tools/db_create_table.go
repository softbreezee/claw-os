package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// SchemaRegistrar records and executes CREATE TABLE statements.
// Implemented by *pg.DB.
type SchemaRegistrar interface {
	DBQuerier
	RegisterSchema(ctx context.Context, tableName, agentID, purpose, ddl string) error
}

type createTableArgs struct {
	TableName string `json:"tableName"`
	SQL       string `json:"sql"`
	Purpose   string `json:"purpose"`
}

var (
	// Allow CREATE TABLE or CREATE INDEX
	reAllowedDDL = regexp.MustCompile(`(?i)^\s*CREATE\s+(UNIQUE\s+)?(TABLE|INDEX)`)
	// Extract table name from CREATE TABLE
	reCreateTable = regexp.MustCompile(`(?i)^\s*CREATE\s+TABLE\s+(IF\s+NOT\s+EXISTS\s+)?(\w+)\s*\(`)
	// Reject anything that touches system or reserved tables
	reservedTables = map[string]bool{
		"sessions": true, "memories": true, "research_data": true,
		"schema_registry": true,
	}
)

// RegisterDBCreateTable registers the db_create_table tool.
func RegisterDBCreateTable(r *Registry, reg SchemaRegistrar) {
	r.Register(
		"db_create_table",
		`Create a new PostgreSQL table for storing research or analysis data.
Only CREATE TABLE is allowed — use db_query for INSERT/SELECT.
Always follow the db-schema-designer skill conventions:
  - Include standard fields: id UUID PK, agent_id TEXT, created_at TIMESTAMPTZ
  - Add embedding vector(1536) for semantic search
  - Add extra JSONB DEFAULT '{}' for flexible attributes
  - Name tables: research_{topic}_{type}
The table definition is recorded in schema_registry for future reference.`,
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"tableName": map[string]any{
					"type":        "string",
					"description": "Table name, e.g. research_nvidia_companies",
				},
				"sql": map[string]any{
					"type":        "string",
					"description": "Full CREATE TABLE SQL statement",
				},
				"purpose": map[string]any{
					"type":        "string",
					"description": "One-line description of what this table stores",
				},
			},
			"required": []string{"tableName", "sql", "purpose"},
		},
		makeCreateTableTool(reg),
	)
}

func makeCreateTableTool(reg SchemaRegistrar) ToolFunc {
	return func(ctx context.Context, rawArgs json.RawMessage) (string, error) {
		var args createTableArgs
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return "", fmt.Errorf("db_create_table: parse args: %w", err)
		}

		name := strings.TrimSpace(strings.ToLower(args.TableName))
		sql := strings.TrimSpace(args.SQL)

		if name == "" || sql == "" || args.Purpose == "" {
			return "", fmt.Errorf("db_create_table: tableName, sql, and purpose are all required")
		}

		// Safety: only CREATE TABLE / CREATE INDEX allowed
		if !reAllowedDDL.MatchString(sql) {
			return "", fmt.Errorf("db_create_table: only CREATE TABLE and CREATE INDEX statements are allowed")
		}

		// Protect system tables
		if reservedTables[name] {
			return "", fmt.Errorf("db_create_table: table %q is reserved and cannot be modified", name)
		}

		isIndex := regexp.MustCompile(`(?i)^\s*CREATE\s+(UNIQUE\s+)?INDEX`).MatchString(sql)

		if !isIndex {
			// Validate naming convention for tables only
			if !strings.HasPrefix(name, "research_") {
				return "", fmt.Errorf("db_create_table: table name must start with 'research_' (got %q)", name)
			}
		}

		// Execute DDL
		if _, err := reg.Exec(ctx, sql); err != nil {
			return "", fmt.Errorf("db_create_table: execute DDL: %w", err)
		}

		if isIndex {
			return fmt.Sprintf("Index on %q created successfully.", name), nil
		}

		// Record table in schema_registry
		if err := reg.RegisterSchema(ctx, name, "agent", args.Purpose, sql); err != nil {
			return fmt.Sprintf("Table %q created successfully.\nWarning: failed to record in schema_registry: %v", name, err), nil
		}

		return fmt.Sprintf(
			"Table %q created and registered.\nPurpose: %s\n\nVerify with:\n  SELECT column_name, data_type FROM information_schema.columns WHERE table_name = '%s' ORDER BY ordinal_position;",
			name, args.Purpose, name,
		), nil
	}
}
