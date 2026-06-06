package gateway

import (
	"context"
	"errors"

	"github.com/softbreezee/claw-os/internal/agent/goal"
	pgstore "github.com/softbreezee/claw-os/internal/store/pg"
)

// goalStoreAdapter bridges *pg.GoalStore (which speaks pg.GoalRecord
// with native pg errors) to agent/goal.Store (which speaks
// agent/goal.Goal with package-local sentinels). Same dependency-
// direction pattern as memstore_adapter.go: gateway imports both
// agent and store/pg, but agent never imports store/pg.
type goalStoreAdapter struct {
	inner *pgstore.GoalStore
}

func (a *goalStoreAdapter) CreateGoal(ctx context.Context, g *goal.Goal) error {
	rec := toRecord(g)
	if err := a.inner.CreateGoal(ctx, rec); err != nil {
		if errors.Is(err, pgstore.ErrGoalAlreadyExists) {
			return goal.ErrAlreadyExists
		}
		return err
	}
	g.CreatedAt = rec.CreatedAt
	g.UpdatedAt = rec.UpdatedAt
	return nil
}

func (a *goalStoreAdapter) GetGoalBySession(ctx context.Context, agentID, sessionKey string) (*goal.Goal, error) {
	rec, err := a.inner.GetGoalBySession(ctx, agentID, sessionKey)
	if err != nil {
		if errors.Is(err, pgstore.ErrGoalNotFound) {
			return nil, goal.ErrNotFound
		}
		return nil, err
	}
	g := fromRecord(rec)
	return &g, nil
}

func (a *goalStoreAdapter) UpdateGoal(ctx context.Context, g *goal.Goal) error {
	rec := toRecord(g)
	if err := a.inner.UpdateGoal(ctx, rec); err != nil {
		if errors.Is(err, pgstore.ErrGoalNotFound) {
			return goal.ErrNotFound
		}
		return err
	}
	g.UpdatedAt = rec.UpdatedAt
	return nil
}

func (a *goalStoreAdapter) DeleteGoal(ctx context.Context, goalID string) error {
	if err := a.inner.DeleteGoal(ctx, goalID); err != nil {
		if errors.Is(err, pgstore.ErrGoalNotFound) {
			return goal.ErrNotFound
		}
		return err
	}
	return nil
}

func toRecord(g *goal.Goal) *pgstore.GoalRecord {
	return &pgstore.GoalRecord{
		ID:          g.ID,
		AgentID:     g.AgentID,
		SessionKey:  g.SessionKey,
		Channel:     g.Channel,
		ChatID:      g.ChatID,
		Objective:   g.Objective,
		Status:      g.Status,
		TokenBudget: g.TokenBudget,
		TokensUsed:  g.TokensUsed,
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
	}
}

func fromRecord(rec *pgstore.GoalRecord) goal.Goal {
	return goal.Goal{
		ID:          rec.ID,
		AgentID:     rec.AgentID,
		SessionKey:  rec.SessionKey,
		Channel:     rec.Channel,
		ChatID:      rec.ChatID,
		Objective:   rec.Objective,
		Status:      rec.Status,
		TokenBudget: rec.TokenBudget,
		TokensUsed:  rec.TokensUsed,
		CreatedAt:   rec.CreatedAt,
		UpdatedAt:   rec.UpdatedAt,
	}
}
