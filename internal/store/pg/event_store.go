package pg

import (
	"context"
	"fmt"
	"time"
)

// MCPEvent is one telemetry row: a single call to a memory MCP tool by
// one client subprocess. It is deliberately append-only and free of
// business meaning — the observability dashboard aggregates these into
// per-source and per-session views. Writing an event must NEVER affect
// the tool result, so callers treat Insert as best-effort.
type MCPEvent struct {
	ConnectionID string // per-subprocess UUID; one client "session"
	Source       string // origin tool tag: claude-code / codex / hermes / codewiz
	AgentID      string // memory pool id (usually "shared")
	Tool         string // memory_search / memory_write / memory_stats
	Query        string // search query text (empty for write/stats)
	Kind         string // write kind: fact / user_note / report (empty otherwise)
	ResultCount  int    // rows returned (search) or written (1 for write)
	Hit          bool   // search returned >0 rows
	Error        string // non-empty when the tool call failed
	DurationMs   int    // wall-clock time of the tool call
}

// EventStore persists and aggregates MCP tool-call telemetry.
type EventStore struct {
	db *DB
}

// NewEventStore creates an EventStore backed by the given DB.
func NewEventStore(db *DB) *EventStore {
	return &EventStore{db: db}
}

// EnsureSchema creates the mcp_events table if it does not exist. The
// MCP subprocess calls Open (not Migrate), so the table must be able to
// self-provision here; the web server also calls this before reading so
// the dashboard works before any subprocess has run.
func (s *EventStore) EnsureSchema(ctx context.Context) error {
	stmts := []string{
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
		if _, err := s.db.Pool.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("pg: ensure mcp_events schema: %w", err)
		}
	}
	return nil
}

// Insert appends one telemetry row. Errors are returned so the caller
// can log them, but the caller must not fail the tool call on error.
func (s *EventStore) Insert(ctx context.Context, e MCPEvent) error {
	agentID := e.AgentID
	if agentID == "" {
		agentID = "shared"
	}
	_, err := s.db.Pool.Exec(ctx,
		`INSERT INTO mcp_events
		   (connection_id, source, agent_id, tool, query, kind, result_count, hit, error, duration_ms)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		e.ConnectionID, e.Source, agentID, e.Tool, e.Query, e.Kind,
		e.ResultCount, e.Hit, e.Error, e.DurationMs,
	)
	if err != nil {
		return fmt.Errorf("pg: mcp event insert: %w", err)
	}
	return nil
}

// ── Aggregation result types (camelCase JSON for the dashboard) ──────

// SourceUsage rolls up all activity for one origin tool.
type SourceUsage struct {
	Source     string    `json:"source"`
	Sessions   int64     `json:"sessions"`
	Searches   int64     `json:"searches"`
	Writes     int64     `json:"writes"`
	Stats      int64     `json:"stats"`
	Hits       int64     `json:"hits"`
	Misses     int64     `json:"misses"`
	LastActive time.Time `json:"lastActive"`
}

// SessionUsage rolls up one client subprocess (connection_id). Turns is
// approximated as the memory_search count (the design fires ~1 search
// per conversation turn); Topics are the most recent search queries.
type SessionUsage struct {
	ConnectionID string    `json:"connectionId"`
	Source       string    `json:"source"`
	FirstSeen    time.Time `json:"firstSeen"`
	LastSeen     time.Time `json:"lastSeen"`
	Turns        int64     `json:"turns"`
	Writes       int64     `json:"writes"`
	Hits         int64     `json:"hits"`
	Topics       []string  `json:"topics"`
}

// RecentEvent is one raw tool call for the live activity feed.
type RecentEvent struct {
	Source      string    `json:"source"`
	Tool        string    `json:"tool"`
	Query       string    `json:"query"`
	Kind        string    `json:"kind"`
	ResultCount int       `json:"resultCount"`
	Hit         bool      `json:"hit"`
	CreatedAt   time.Time `json:"createdAt"`
}

// UsageOverview is the full dashboard payload.
type UsageOverview struct {
	Available     bool           `json:"available"`
	TotalEvents   int64          `json:"totalEvents"`
	TotalSessions int64          `json:"totalSessions"`
	TotalSearches int64          `json:"totalSearches"`
	TotalWrites   int64          `json:"totalWrites"`
	TotalHits     int64          `json:"totalHits"`
	Sources       []SourceUsage  `json:"sources"`
	Sessions      []SessionUsage `json:"sessions"`
	Recent        []RecentEvent  `json:"recent"`
}

// Overview assembles the dashboard payload in four queries: totals,
// per-source rollup, per-session rollup (most recent sessionLimit), and
// the recent activity feed (most recent recentLimit events).
func (s *EventStore) Overview(ctx context.Context, sessionLimit, recentLimit int) (UsageOverview, error) {
	if sessionLimit <= 0 {
		sessionLimit = 50
	}
	if recentLimit <= 0 {
		recentLimit = 50
	}
	var o UsageOverview
	o.Available = true
	o.Sources = []SourceUsage{}
	o.Sessions = []SessionUsage{}
	o.Recent = []RecentEvent{}

	// Totals.
	if err := s.db.Pool.QueryRow(ctx,
		`SELECT
		   COUNT(*),
		   COUNT(DISTINCT connection_id),
		   COUNT(*) FILTER (WHERE tool = 'memory_search'),
		   COUNT(*) FILTER (WHERE tool = 'memory_write'),
		   COUNT(*) FILTER (WHERE tool = 'memory_search' AND hit)
		 FROM mcp_events`,
	).Scan(&o.TotalEvents, &o.TotalSessions, &o.TotalSearches, &o.TotalWrites, &o.TotalHits); err != nil {
		return UsageOverview{}, fmt.Errorf("pg: usage totals: %w", err)
	}

	// Per-source rollup.
	srcRows, err := s.db.Pool.Query(ctx,
		`SELECT
		   source,
		   COUNT(DISTINCT connection_id)                                       AS sessions,
		   COUNT(*) FILTER (WHERE tool = 'memory_search')                      AS searches,
		   COUNT(*) FILTER (WHERE tool = 'memory_write')                       AS writes,
		   COUNT(*) FILTER (WHERE tool = 'memory_stats')                       AS stats,
		   COUNT(*) FILTER (WHERE tool = 'memory_search' AND hit)              AS hits,
		   COUNT(*) FILTER (WHERE tool = 'memory_search' AND NOT hit)          AS misses,
		   MAX(created_at)                                                     AS last_active
		 FROM mcp_events
		 GROUP BY source
		 ORDER BY last_active DESC`,
	)
	if err != nil {
		return UsageOverview{}, fmt.Errorf("pg: usage by source: %w", err)
	}
	defer srcRows.Close()
	for srcRows.Next() {
		var u SourceUsage
		if err := srcRows.Scan(&u.Source, &u.Sessions, &u.Searches, &u.Writes,
			&u.Stats, &u.Hits, &u.Misses, &u.LastActive); err != nil {
			return UsageOverview{}, fmt.Errorf("pg: scan source row: %w", err)
		}
		if u.Source == "" {
			u.Source = "(unknown)"
		}
		o.Sources = append(o.Sources, u)
	}
	srcRows.Close()

	// Per-session rollup. topics = most recent 5 non-empty queries.
	sessRows, err := s.db.Pool.Query(ctx,
		`SELECT
		   connection_id,
		   MAX(source)                                                         AS source,
		   MIN(created_at)                                                     AS first_seen,
		   MAX(created_at)                                                     AS last_seen,
		   COUNT(*) FILTER (WHERE tool = 'memory_search')                      AS turns,
		   COUNT(*) FILTER (WHERE tool = 'memory_write')                       AS writes,
		   COUNT(*) FILTER (WHERE hit)                                         AS hits,
		   COALESCE((array_agg(query ORDER BY created_at DESC)
		             FILTER (WHERE query <> ''))[1:5], '{}'::text[])           AS topics
		 FROM mcp_events
		 GROUP BY connection_id
		 ORDER BY last_seen DESC
		 LIMIT $1`,
		sessionLimit,
	)
	if err != nil {
		return UsageOverview{}, fmt.Errorf("pg: usage by session: %w", err)
	}
	defer sessRows.Close()
	for sessRows.Next() {
		var u SessionUsage
		if err := sessRows.Scan(&u.ConnectionID, &u.Source, &u.FirstSeen, &u.LastSeen,
			&u.Turns, &u.Writes, &u.Hits, &u.Topics); err != nil {
			return UsageOverview{}, fmt.Errorf("pg: scan session row: %w", err)
		}
		if u.Source == "" {
			u.Source = "(unknown)"
		}
		if u.Topics == nil {
			u.Topics = []string{}
		}
		o.Sessions = append(o.Sessions, u)
	}
	sessRows.Close()

	// Recent activity feed.
	recRows, err := s.db.Pool.Query(ctx,
		`SELECT source, tool, query, kind, result_count, hit, created_at
		 FROM mcp_events
		 ORDER BY created_at DESC
		 LIMIT $1`,
		recentLimit,
	)
	if err != nil {
		return UsageOverview{}, fmt.Errorf("pg: usage recent: %w", err)
	}
	defer recRows.Close()
	for recRows.Next() {
		var e RecentEvent
		if err := recRows.Scan(&e.Source, &e.Tool, &e.Query, &e.Kind,
			&e.ResultCount, &e.Hit, &e.CreatedAt); err != nil {
			return UsageOverview{}, fmt.Errorf("pg: scan recent row: %w", err)
		}
		if e.Source == "" {
			e.Source = "(unknown)"
		}
		o.Recent = append(o.Recent, e)
	}
	recRows.Close()

	return o, nil
}
