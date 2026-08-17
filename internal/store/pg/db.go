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
		// Set session timezone so TIMESTAMPTZ reads return local wall-clock.
		// Without this, pgx maps UTC-zero times back as UTC, and cron
		// next-runs show 8h off in Asia/Shanghai.
		if _, err := conn.Exec(ctx, "SET timezone = 'Asia/Shanghai'"); err != nil {
			return fmt.Errorf("pg: set timezone: %w", err)
		}
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

		// agent_goals backs the /goal feature: one persistent objective
		// per (agent, session). The UNIQUE (agent_id, session_key)
		// constraint is the source of truth for "this session already
		// has a goal" — CreateGoal translates the conflict into
		// ErrGoalAlreadyExists. Routing fields (channel, chat_id) are
		// stamped at create time so a continuation can publish onto
		// the same bus address the original turn arrived on.
		`CREATE TABLE IF NOT EXISTS agent_goals (
			id           TEXT        PRIMARY KEY,
			agent_id     TEXT        NOT NULL,
			session_key  TEXT        NOT NULL,
			channel      TEXT        NOT NULL DEFAULT '',
			chat_id      TEXT        NOT NULL DEFAULT '',
			objective    TEXT        NOT NULL,
			status       TEXT        NOT NULL DEFAULT 'active',
			token_budget BIGINT,
			tokens_used  BIGINT      NOT NULL DEFAULT 0,
			created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS agent_goals_session
			ON agent_goals (agent_id, session_key)`,

		// mcp_events backs the memory observability dashboard: one row
		// per memory MCP tool call (search/write/stats), tagged with the
		// origin tool (source) and the client subprocess (connection_id
		// ≈ one session). The `pawnix mcp` subprocess self-provisions
		// this via EventStore.EnsureSchema too, since runMCP calls Open
		// without Migrate — but keeping it here means the main server
		// creates it up-front. See internal/store/pg/event_store.go.
		`CREATE TABLE IF NOT EXISTS mcp_events (
			id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
			connection_id TEXT        NOT NULL,
			source        TEXT        NOT NULL DEFAULT '',
			agent_id      TEXT        NOT NULL DEFAULT 'shared',
			tool          TEXT        NOT NULL,
			query         TEXT        NOT NULL DEFAULT '',
			kind          TEXT        NOT NULL DEFAULT '',
			result_count  INT         NOT NULL DEFAULT 0,
			hit           BOOLEAN     NOT NULL DEFAULT false,
			error         TEXT        NOT NULL DEFAULT '',
			duration_ms   INT         NOT NULL DEFAULT 0,
			created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
		`CREATE INDEX IF NOT EXISTS mcp_events_source_time
			ON mcp_events (source, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS mcp_events_conn
			ON mcp_events (connection_id)`,
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
