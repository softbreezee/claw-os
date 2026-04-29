package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// DBQuerier executes SQL queries and returns results as a string.
// Implemented by *pg.DB; defined here to avoid import cycles.
type DBQuerier interface {
	QueryRows(ctx context.Context, sql string, args ...any) ([]map[string]any, error)
	Exec(ctx context.Context, sql string, args ...any) (int64, error)
}

type dbQueryArgs struct {
	SQL    string `json:"sql"`
	Params []any  `json:"params,omitempty"`
}

// RegisterDBQuery registers the db_query tool so agents can read/write the
// PostgreSQL research database.
func RegisterDBQuery(r *Registry, querier DBQuerier) {
	r.Register(
		"db_query",
		`Execute a SQL query on the research database.
For SELECT queries returns rows as JSON. For INSERT/UPDATE/DELETE returns affected row count.
Schema:
  memories(id uuid, agent_id text, kind text, content text, tags text[], created_at timestamptz)
  research_data(id uuid, agent_id text, topic text, content text, data jsonb, source_url text, created_at timestamptz)
  sessions(agent_id text, channel text, session_id text, messages jsonb, updated_at timestamptz)
Always use parameterised queries: pass values in "params" array and reference them as $1, $2, … in SQL.`,
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"sql": map[string]any{
					"type":        "string",
					"description": "SQL statement to execute",
				},
				"params": map[string]any{
					"type":        "array",
					"description": "Optional query parameters ($1, $2, …)",
					"items":       map[string]any{"type": "string"},
				},
			},
			"required": []string{"sql"},
		},
		makeDBQueryTool(querier),
	)
}

func makeDBQueryTool(querier DBQuerier) ToolFunc {
	return func(ctx context.Context, rawArgs json.RawMessage) (string, error) {
		var args dbQueryArgs
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return "", fmt.Errorf("db_query: parse args: %w", err)
		}
		if strings.TrimSpace(args.SQL) == "" {
			return "", fmt.Errorf("db_query: sql is required")
		}

	upper := strings.ToUpper(strings.TrimSpace(args.SQL))

	// Block DDL — use db_create_table for schema changes.
	ddlKeywords := []string{"CREATE ", "DROP ", "ALTER ", "TRUNCATE ", "RENAME "}
	for _, kw := range ddlKeywords {
		if strings.HasPrefix(upper, kw) {
			return "", fmt.Errorf("db_query: DDL is not allowed here — use the db_create_table tool to create tables")
		}
	}

	// Route SELECT / EXPLAIN to QueryRows; everything else to Exec.
	if strings.HasPrefix(upper, "SELECT") || strings.HasPrefix(upper, "EXPLAIN") || strings.HasPrefix(upper, "WITH") {
			rows, err := querier.QueryRows(ctx, args.SQL, args.Params...)
			if err != nil {
				return "", fmt.Errorf("db_query: %w", err)
			}
			if len(rows) == 0 {
				return "No rows returned.", nil
			}
			data, _ := json.MarshalIndent(rows, "", "  ")
			result := string(data)
			if len(result) > maxExecOutputBytes {
				result = result[:maxExecOutputBytes] + "\n\n[Output truncated]"
			}
			return fmt.Sprintf("%d row(s):\n%s", len(rows), result), nil
		}

		affected, err := querier.Exec(ctx, args.SQL, args.Params...)
		if err != nil {
			return "", fmt.Errorf("db_query: %w", err)
		}
		return fmt.Sprintf("OK — %d row(s) affected.", affected), nil
	}
}
