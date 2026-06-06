package pg

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
)

// MemoryRecord is a single persisted memory fact or user note.
type MemoryRecord struct {
	ID        string
	AgentID   string
	Kind      string // "fact" | "user_note" | "report"
	Content   string
	Tags      []string
	CreatedAt time.Time
}

// MemoryStore provides PostgreSQL+pgvector-backed memory persistence.
type MemoryStore struct {
	db *DB
}

// NewMemoryStore creates a MemoryStore backed by the given DB.
func NewMemoryStore(db *DB) *MemoryStore {
	return &MemoryStore{db: db}
}

// Insert adds a new memory record. embedding may be nil when no embedding
// provider is configured (falls back to keyword search only).
func (s *MemoryStore) Insert(ctx context.Context, agentID, kind, content string, embedding []float32, tags []string) (string, error) {
	if tags == nil {
		tags = []string{}
	}

	var id string
	var err error
	if len(embedding) > 0 {
		err = s.db.Pool.QueryRow(ctx,
			`INSERT INTO memories (agent_id, kind, content, embedding, tags)
			 VALUES ($1, $2, $3, $4, $5)
			 RETURNING id`,
			agentID, kind, content, NewVec(embedding), tags,
		).Scan(&id)
	} else {
		err = s.db.Pool.QueryRow(ctx,
			`INSERT INTO memories (agent_id, kind, content, tags)
			 VALUES ($1, $2, $3, $4)
			 RETURNING id`,
			agentID, kind, content, tags,
		).Scan(&id)
	}
	if err != nil {
		return "", fmt.Errorf("pg: memory insert: %w", err)
	}
	return id, nil
}

// SearchSemantic returns the top-k memories closest to the query embedding.
// Falls back to keyword search if embedding is nil.
func (s *MemoryStore) SearchSemantic(ctx context.Context, agentID string, queryEmbedding []float32, limit int) ([]MemoryRecord, error) {
	if limit <= 0 {
		limit = 10
	}
	if len(queryEmbedding) == 0 {
		return s.SearchKeyword(ctx, agentID, "", limit)
	}

	rows, err := s.db.Pool.Query(ctx,
		`SELECT id, agent_id, kind, content, tags, created_at
		 FROM memories
		 WHERE agent_id = $1 AND embedding IS NOT NULL
		 ORDER BY embedding <=> $2
		 LIMIT $3`,
		agentID, NewVec(queryEmbedding), limit,
	)
	if err != nil {
		return nil, fmt.Errorf("pg: memory semantic search: %w", err)
	}
	defer rows.Close()
	return scanMemories(rows)
}

// SearchKeyword does a full-text ILIKE search when no embedding is available.
func (s *MemoryStore) SearchKeyword(ctx context.Context, agentID, query string, limit int) ([]MemoryRecord, error) {
	if limit <= 0 {
		limit = 10
	}

	var (
		rawRows interface {
			Next() bool
			Scan(dest ...any) error
			Close()
		}
		err error
	)

	if query != "" {
		rawRows, err = s.db.Pool.Query(ctx,
			`SELECT id, agent_id, kind, content, tags, created_at
			 FROM memories
			 WHERE agent_id = $1 AND content ILIKE $2
			 ORDER BY created_at DESC
			 LIMIT $3`,
			agentID, "%"+query+"%", limit,
		)
	} else {
		rawRows, err = s.db.Pool.Query(ctx,
			`SELECT id, agent_id, kind, content, tags, created_at
			 FROM memories
			 WHERE agent_id = $1
			 ORDER BY created_at DESC
			 LIMIT $2`,
			agentID, limit,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("pg: memory keyword search: %w", err)
	}
	defer rawRows.Close()
	return scanMemories(rawRows)
}

// LoadAll returns all memory facts for an agent (for system prompt injection).
func (s *MemoryStore) LoadAll(ctx context.Context, agentID string) ([]MemoryRecord, error) {
	rows, err := s.db.Pool.Query(ctx,
		`SELECT id, agent_id, kind, content, tags, created_at
		 FROM memories
		 WHERE agent_id = $1
		 ORDER BY created_at ASC`,
		agentID,
	)
	if err != nil {
		return nil, fmt.Errorf("pg: memory load all: %w", err)
	}
	defer rows.Close()
	return scanMemories(rows)
}

type rowScanner interface {
	Next() bool
	Scan(dest ...any) error
	Close()
}

func scanMemories(rows rowScanner) ([]MemoryRecord, error) {
	var results []MemoryRecord
	for rows.Next() {
		var r MemoryRecord
		if err := rows.Scan(&r.ID, &r.AgentID, &r.Kind, &r.Content, &r.Tags, &r.CreatedAt); err != nil {
			slog.Warn("pg: scan memory row", "error", err)
			continue
		}
		results = append(results, r)
	}
	return results, nil
}

// MemoryHealth summarises the embedding-coverage health of an agent's
// memory rows. Surfaced via Heartbeat so silent failures (no embed
// model, embed API broken, agent_id renamed leaving rows orphaned)
// stop being invisible — see docs/memory-verification.md for the
// failure modes this guards against.
type MemoryHealth struct {
	Total          int64 // total memory rows for the agent
	WithEmbedding  int64 // rows where embedding IS NOT NULL
	RecentTotal    int64 // rows created in the last 24h
	RecentEmbedded int64 // recent rows with non-null embedding
}

// CoveragePct returns WithEmbedding / Total as a 0..100 integer, or
// -1 when Total is 0 (no data, "not applicable" rather than "0%").
func (h MemoryHealth) CoveragePct() int {
	if h.Total == 0 {
		return -1
	}
	return int((h.WithEmbedding * 100) / h.Total)
}

// RecentCoveragePct is the same ratio but bounded to the last 24h
// window. This is the more useful number for "is the embedding
// pipeline working RIGHT NOW" — old rows from before embedding was
// configured drag down the all-time stat indefinitely.
func (h MemoryHealth) RecentCoveragePct() int {
	if h.RecentTotal == 0 {
		return -1
	}
	return int((h.RecentEmbedded * 100) / h.RecentTotal)
}

// HealthStats reports embedding coverage for the agent's memory rows.
// One round-trip; returns the zero value with an error on query
// failure rather than partial data.
func (s *MemoryStore) HealthStats(ctx context.Context, agentID string) (MemoryHealth, error) {
	var h MemoryHealth
	err := s.db.Pool.QueryRow(ctx,
		`SELECT
		   COUNT(*) AS total,
		   COUNT(*) FILTER (WHERE embedding IS NOT NULL) AS with_embed,
		   COUNT(*) FILTER (WHERE created_at > now() - interval '24 hours') AS recent_total,
		   COUNT(*) FILTER (WHERE created_at > now() - interval '24 hours' AND embedding IS NOT NULL) AS recent_embed
		 FROM memories
		 WHERE agent_id = $1`,
		agentID,
	).Scan(&h.Total, &h.WithEmbedding, &h.RecentTotal, &h.RecentEmbedded)
	if err != nil {
		return MemoryHealth{}, fmt.Errorf("pg: memory health: %w", err)
	}
	return h, nil
}

// VerifyVectorReady performs a one-shot health check that the
// pgvector extension is loaded and the codec is wired correctly.
// Cheap (single round-trip, no schema touch) so it can run on every
// startup — silent failures of pgvector ("CREATE TABLE memories"
// succeeds without the extension on some managed Postgres flavors,
// but ORDER BY embedding <=> ... fails at query time) are exactly
// the failure mode docs/memory-verification.md was written for.
//
// Returns nil when the extension answers a trivial vector op; any
// error indicates the read path will not work and the caller should
// degrade to MEMORY.md-only mode rather than letting the agent
// silently lose access to its DB-backed memory.
func (s *MemoryStore) VerifyVectorReady(ctx context.Context) error {
	var ok int
	if err := s.db.Pool.QueryRow(ctx,
		`SELECT 1 WHERE '[1]'::vector <=> '[1]'::vector = 0`,
	).Scan(&ok); err != nil {
		return fmt.Errorf("pgvector readiness probe failed: %w", err)
	}
	if ok != 1 {
		return fmt.Errorf("pgvector readiness probe returned unexpected result: %d", ok)
	}
	return nil
}

// InsertResearch stores a research data record collected by an agent.
func (s *MemoryStore) InsertResearch(ctx context.Context, agentID, topic, content string, data map[string]any, embedding []float32, sourceURL string) (string, error) {
	dataJSON, err := json.Marshal(data)
	if err != nil || data == nil {
		dataJSON = []byte("{}")
	}

	var id string
	if len(embedding) > 0 {
		err = s.db.Pool.QueryRow(ctx,
			`INSERT INTO research_data (agent_id, topic, content, data, embedding, source_url)
			 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
			agentID, topic, content, dataJSON, NewVec(embedding), sourceURL,
		).Scan(&id)
	} else {
		err = s.db.Pool.QueryRow(ctx,
			`INSERT INTO research_data (agent_id, topic, content, data, source_url)
			 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			agentID, topic, content, dataJSON, sourceURL,
		).Scan(&id)
	}
	if err != nil {
		return "", fmt.Errorf("pg: research insert: %w", err)
	}
	slog.Info("pg: research data stored", "agent", agentID, "topic", topic, "id", id)
	return id, nil
}
