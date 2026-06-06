package goal

import (
	"context"
	"errors"
)

// ErrAlreadyExists is returned by Store.CreateGoal when the (agent,
// session_key) pair already has a goal. Slash handlers translate
// this into a "clear first" reply.
var ErrAlreadyExists = errors.New("goal: already exists for this session")

// ErrNotFound is returned by Store.GetGoalBySession / DeleteGoal
// when no goal matches the lookup key.
var ErrNotFound = errors.New("goal: not found")

// Store is the narrow persistence interface the slash handlers,
// continuation helper, and accounting hook depend on. Production
// is satisfied by an adapter over internal/store/pg.GoalStore;
// tests use the in-memory MemStore (see store_mem.go).
//
// Errors returned by implementations MUST satisfy errors.Is against
// ErrAlreadyExists / ErrNotFound — the adapter is responsible for
// translating its native errors before they bubble up.
type Store interface {
	CreateGoal(ctx context.Context, g *Goal) error
	GetGoalBySession(ctx context.Context, agentID, sessionKey string) (*Goal, error)
	UpdateGoal(ctx context.Context, g *Goal) error
	DeleteGoal(ctx context.Context, goalID string) error
}
