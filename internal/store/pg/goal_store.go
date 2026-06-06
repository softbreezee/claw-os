package pg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// GoalRecord is the persisted shape of a /goal target. One per
// (agent, session_key) — UNIQUE index enforces it. See
// internal/agent/goal for the domain layer that wraps it.
//
// Routing fields (Channel, ChatID) are stamped at create time so
// continuation prompts published from PostTurn / accounting hooks
// can address the same chat the goal was created in. Without them
// a goal restored after a restart would have nowhere to deliver
// its continuation turn.
type GoalRecord struct {
	ID         string `json:"id"`
	AgentID    string `json:"agentId"`
	SessionKey string `json:"sessionKey"`

	Channel string `json:"channel,omitempty"`
	ChatID  string `json:"chatId,omitempty"`

	Objective string `json:"objective"`
	Status    string `json:"status"` // active | paused | budget_limited | complete

	// TokenBudget is nil for unbounded goals. Stored as a nullable
	// BIGINT column.
	TokenBudget *int64 `json:"tokenBudget,omitempty"`
	TokensUsed  int64  `json:"tokensUsed"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ErrGoalAlreadyExists is returned by CreateGoal when a goal already
// exists for the (agent, session_key) pair. The /goal slash handler
// translates this into a "clear first" reply.
var ErrGoalAlreadyExists = errors.New("pg: goal already exists for this session")

// ErrGoalNotFound is returned by GetGoalBySession / DeleteGoal when no
// goal matches the lookup key.
var ErrGoalNotFound = errors.New("pg: goal not found")

// GoalStore provides PostgreSQL-backed persistence for goal records.
// Mirrors MemoryStore's shape — one type, narrow surface, no caching.
type GoalStore struct {
	db *DB
}

// NewGoalStore constructs a GoalStore against the given pool.
func NewGoalStore(db *DB) *GoalStore {
	return &GoalStore{db: db}
}

// CreateGoal inserts a new goal row. Returns ErrGoalAlreadyExists when
// the (agent_id, session_key) UNIQUE index is violated — which is
// exactly the "this session already has an active goal" case the
// slash handler wants to surface as "clear first".
//
// Stamps CreatedAt / UpdatedAt server-side via DEFAULT now() so we
// don't have to pass them through the pgx codec; we read them back
// in the same RETURNING clause.
func (s *GoalStore) CreateGoal(ctx context.Context, g *GoalRecord) error {
	if g == nil {
		return errors.New("pg: nil goal")
	}
	err := s.db.Pool.QueryRow(ctx,
		`INSERT INTO agent_goals
		   (id, agent_id, session_key, channel, chat_id, objective,
		    status, token_budget, tokens_used)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING created_at, updated_at`,
		g.ID, g.AgentID, g.SessionKey, g.Channel, g.ChatID, g.Objective,
		g.Status, g.TokenBudget, g.TokensUsed,
	).Scan(&g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// 23505 = unique_violation. The only UNIQUE on this table
			// is (agent_id, session_key); no need to inspect the
			// constraint name to disambiguate.
			return ErrGoalAlreadyExists
		}
		return fmt.Errorf("pg: create goal: %w", err)
	}
	return nil
}

// GetGoalBySession returns the row matching (agent_id, session_key),
// or ErrGoalNotFound when no row exists. The hot path for the
// continuation hook — must stay one indexed lookup.
func (s *GoalStore) GetGoalBySession(ctx context.Context, agentID, sessionKey string) (*GoalRecord, error) {
	var g GoalRecord
	err := s.db.Pool.QueryRow(ctx,
		`SELECT id, agent_id, session_key, channel, chat_id, objective,
		        status, token_budget, tokens_used, created_at, updated_at
		 FROM agent_goals
		 WHERE agent_id = $1 AND session_key = $2`,
		agentID, sessionKey,
	).Scan(
		&g.ID, &g.AgentID, &g.SessionKey, &g.Channel, &g.ChatID, &g.Objective,
		&g.Status, &g.TokenBudget, &g.TokensUsed, &g.CreatedAt, &g.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrGoalNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("pg: get goal: %w", err)
	}
	return &g, nil
}

// UpdateGoal persists status / objective / token-counter changes for
// an existing row. updated_at is bumped server-side; we read it back
// so the in-memory record stays in sync.
//
// Identified by primary key (id), not by (agent, session_key), so a
// caller that already holds a *GoalRecord doesn't need to re-fetch.
func (s *GoalStore) UpdateGoal(ctx context.Context, g *GoalRecord) error {
	if g == nil {
		return errors.New("pg: nil goal")
	}
	err := s.db.Pool.QueryRow(ctx,
		`UPDATE agent_goals
		 SET objective    = $2,
		     status       = $3,
		     token_budget = $4,
		     tokens_used  = $5,
		     updated_at   = now()
		 WHERE id = $1
		 RETURNING updated_at`,
		g.ID, g.Objective, g.Status, g.TokenBudget, g.TokensUsed,
	).Scan(&g.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrGoalNotFound
	}
	if err != nil {
		return fmt.Errorf("pg: update goal: %w", err)
	}
	return nil
}

// DeleteGoal removes the row by id. Idempotent at the SQL level — a
// missing row is reported as ErrGoalNotFound so callers can decide
// whether to surface that to the user (the /goal clear handler
// treats it as "already cleared, no-op").
func (s *GoalStore) DeleteGoal(ctx context.Context, goalID string) error {
	tag, err := s.db.Pool.Exec(ctx,
		`DELETE FROM agent_goals WHERE id = $1`, goalID)
	if err != nil {
		return fmt.Errorf("pg: delete goal: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrGoalNotFound
	}
	return nil
}
