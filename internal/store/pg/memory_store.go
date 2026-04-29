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
