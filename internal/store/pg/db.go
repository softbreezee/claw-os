// Package pg implements PostgreSQL-backed storage for sessions, memory, and
// research data. It uses pgx/v5 for the connection pool and pgvector for
// semantic (embedding-based) memory search.
package pg

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgvector "github.com/pgvector/pgvector-go"
	pgxvector "github.com/pgvector/pgvector-go/pgx"
)

// DB wraps a pgxpool and provides shared helpers.
type DB struct {
	Pool *pgxpool.Pool
}

// Open creates a connection pool and registers pgvector types.
func Open(ctx context.Context, dsn string) (*DB, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("pg: parse dsn: %w", err)
	}

	// Register pgvector codec for every new connection.
	cfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		return pgxvector.RegisterTypes(ctx, conn)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("pg: connect: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pg: ping: %w", err)
	}

	slog.Info("pg: connected", "host", cfg.ConnConfig.Host, "db", cfg.ConnConfig.Database)
	return &DB{Pool: pool}, nil
}

// Migrate creates all required tables and indexes if they do not exist.
func (db *DB) Migrate(ctx context.Context) error {
	stmts := []string{
		// pgvector extension (idempotent)
		`CREATE EXTENSION IF NOT EXISTS vector`,

		// Sessions table — replaces per-agent JSONL files.
		`CREATE TABLE IF NOT EXISTS sessions (
			agent_id   TEXT        NOT NULL,
			channel    TEXT        NOT NULL,
			session_id TEXT        NOT NULL,
			messages   JSONB       NOT NULL DEFAULT '[]',
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			PRIMARY KEY (agent_id, channel, session_id)
		)`,
		`CREATE INDEX IF NOT EXISTS sessions_agent_channel
			ON sessions (agent_id, channel)`,

		// Memories table — replaces MEMORY.md / USER.md flat files.
		`CREATE TABLE IF NOT EXISTS memories (
			id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
			agent_id   TEXT        NOT NULL,
			kind       TEXT        NOT NULL DEFAULT 'fact',
			content    TEXT        NOT NULL,
			embedding  vector(1536),
			tags       TEXT[]      NOT NULL DEFAULT '{}',
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
		// HNSW index for fast approximate nearest-neighbour search.
		`CREATE INDEX IF NOT EXISTS memories_embedding_hnsw
			ON memories USING hnsw (embedding vector_cosine_ops)`,
		`CREATE INDEX IF NOT EXISTS memories_agent_kind
			ON memories (agent_id, kind)`,

		// Research data table — stores arbitrary structured data collected by agents.
		`CREATE TABLE IF NOT EXISTS research_data (
			id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
			agent_id   TEXT        NOT NULL,
			topic      TEXT        NOT NULL,
			content    TEXT        NOT NULL,
			data       JSONB       NOT NULL DEFAULT '{}',
			embedding  vector(1536),
			source_url TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
		`CREATE INDEX IF NOT EXISTS research_data_agent_topic
			ON research_data (agent_id, topic)`,
		`CREATE INDEX IF NOT EXISTS research_data_embedding_hnsw
			ON research_data USING hnsw (embedding vector_cosine_ops)`,

		// Schema registry — records every table created by agents via db_create_table.
		`CREATE TABLE IF NOT EXISTS schema_registry (
			table_name  TEXT        PRIMARY KEY,
			agent_id    TEXT        NOT NULL,
			purpose     TEXT        NOT NULL,
			ddl         TEXT        NOT NULL,
			created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
	}

	for _, stmt := range stmts {
		if _, err := db.Pool.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("pg: migrate: %w\nSQL: %s", err, stmt)
		}
	}

	slog.Info("pg: migrations applied")
	return nil
}

// Close shuts down the connection pool.
func (db *DB) Close() {
	db.Pool.Close()
}

// QueryRows executes a SELECT and returns results as a slice of maps.
// Implements tools.DBQuerier.
func (db *DB) QueryRows(ctx context.Context, sql string, args ...any) ([]map[string]any, error) {
	rows, err := db.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("pg: query: %w", err)
	}
	defer rows.Close()

	fields := rows.FieldDescriptions()
	var result []map[string]any
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("pg: scan row: %w", err)
		}
		row := make(map[string]any, len(fields))
		for i, f := range fields {
			row[string(f.Name)] = vals[i]
		}
		result = append(result, row)
	}
	return result, nil
}

// Exec executes a non-SELECT statement and returns affected row count.
// Implements tools.DBQuerier.
func (db *DB) Exec(ctx context.Context, sql string, args ...any) (int64, error) {
	tag, err := db.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return 0, fmt.Errorf("pg: exec: %w", err)
	}
	return tag.RowsAffected(), nil
}

// RegisterSchema records a CREATE TABLE statement in schema_registry.
// Implements tools.SchemaRegistrar.
func (db *DB) RegisterSchema(ctx context.Context, tableName, agentID, purpose, ddl string) error {
	_, err := db.Pool.Exec(ctx,
		`INSERT INTO schema_registry (table_name, agent_id, purpose, ddl)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (table_name) DO UPDATE
		   SET agent_id = EXCLUDED.agent_id,
		       purpose  = EXCLUDED.purpose,
		       ddl      = EXCLUDED.ddl`,
		tableName, agentID, purpose, ddl,
	)
	return err
}

// Vec is a convenience alias so callers don't need to import pgvector-go directly.
type Vec = pgvector.Vector

// NewVec wraps a float32 slice into a pgvector.Vector.
func NewVec(v []float32) Vec {
	return pgvector.NewVector(v)
}
