package goal

import (
	"context"
	"sync"
	"time"
)

// MemStore is an in-memory Store for tests and CLI dry-runs. Not
// concurrency-safe across processes, but the per-process mutex makes
// it safe for parallel TaskRunner / heartbeat goroutines that share
// it in tests.
type MemStore struct {
	mu sync.Mutex
	// Keyed by goal id; lookup-by-session iterates because there's
	// only ever one active goal per session anyway (UNIQUE in PG).
	goals map[string]*Goal
}

// NewMemStore returns an empty MemStore.
func NewMemStore() *MemStore {
	return &MemStore{goals: map[string]*Goal{}}
}

// CreateGoal returns ErrAlreadyExists when the (agent, session_key)
// pair already has a row, mirroring the PG UNIQUE-violation path.
func (m *MemStore) CreateGoal(_ context.Context, g *Goal) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, existing := range m.goals {
		if existing.AgentID == g.AgentID && existing.SessionKey == g.SessionKey {
			return ErrAlreadyExists
		}
	}
	now := time.Now()
	g.CreatedAt = now
	g.UpdatedAt = now
	// Clone defensively so a caller that mutates the input later
	// doesn't accidentally mutate our stored row.
	stored := *g
	m.goals[g.ID] = &stored
	return nil
}

// GetGoalBySession returns a defensive copy so the caller's mutations
// (e.g. FoldUsage) don't leak into the store until UpdateGoal.
func (m *MemStore) GetGoalBySession(_ context.Context, agentID, sessionKey string) (*Goal, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, g := range m.goals {
		if g.AgentID == agentID && g.SessionKey == sessionKey {
			out := *g
			return &out, nil
		}
	}
	return nil, ErrNotFound
}

// UpdateGoal writes the caller's modified record back. Identified by
// id, not (agent, session_key), so a record that just flipped status
// (e.g. cleared and re-created concurrently) doesn't accidentally
// overwrite the new row.
func (m *MemStore) UpdateGoal(_ context.Context, g *Goal) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.goals[g.ID]; !ok {
		return ErrNotFound
	}
	g.UpdatedAt = time.Now()
	stored := *g
	m.goals[g.ID] = &stored
	return nil
}

// DeleteGoal returns ErrNotFound when the id is unknown, matching the
// PG store's contract.
func (m *MemStore) DeleteGoal(_ context.Context, goalID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.goals[goalID]; !ok {
		return ErrNotFound
	}
	delete(m.goals, goalID)
	return nil
}
