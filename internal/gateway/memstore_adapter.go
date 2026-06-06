package gateway

import (
	"context"

	"github.com/softbreezee/claw-os/internal/agent"
	pgstore "github.com/softbreezee/claw-os/internal/store/pg"
)

// memStoreAdapter bridges *pg.MemoryStore (which returns pg.MemoryRecord)
// to agent.PGMemoryStore (which expects agent.MemoryHit). Without this
// thin shim the agent layer would have to import internal/store/pg,
// creating a cycle (gateway → agent → store/pg → ...).
//
// Adapter pattern keeps the dependency direction clean:
//
//	gateway → agent (calls SetPGBackend with adapter)
//	gateway → store/pg (calls NewMemoryStore directly)
//	agent   ↛ store/pg
type memStoreAdapter struct {
	inner *pgstore.MemoryStore
}

// Insert is a straight pass-through; the signature matches both sides.
func (a *memStoreAdapter) Insert(ctx context.Context, agentID, kind, content string, embedding []float32, tags []string) (string, error) {
	return a.inner.Insert(ctx, agentID, kind, content, embedding, tags)
}

// SearchSemantic delegates to the underlying store and projects the
// rich MemoryRecord rows down to the minimal MemoryHit struct that the
// agent layer consumes (Kind + Content only — Tags/CreatedAt etc.
// aren't surfaced in the system prompt).
func (a *memStoreAdapter) SearchSemantic(ctx context.Context, agentID string, queryEmbedding []float32, limit int) ([]agent.MemoryHit, error) {
	rows, err := a.inner.SearchSemantic(ctx, agentID, queryEmbedding, limit)
	if err != nil {
		return nil, err
	}
	hits := make([]agent.MemoryHit, 0, len(rows))
	for _, r := range rows {
		hits = append(hits, agent.MemoryHit{Kind: r.Kind, Content: r.Content})
	}
	return hits, nil
}

// HealthStats projects pg.MemoryHealth into the agent-local snapshot
// type. Heartbeat consumes this to detect silent embedding pipeline
// failures (recent NULL rate spiking) without internal/agent learning
// about pg internals. See docs/memory-verification.md.
func (a *memStoreAdapter) HealthStats(ctx context.Context, agentID string) (agent.MemoryHealthSnapshot, error) {
	h, err := a.inner.HealthStats(ctx, agentID)
	if err != nil {
		return agent.MemoryHealthSnapshot{}, err
	}
	return agent.MemoryHealthSnapshot{
		Total:          h.Total,
		WithEmbedding:  h.WithEmbedding,
		RecentTotal:    h.RecentTotal,
		RecentEmbedded: h.RecentEmbedded,
	}, nil
}
