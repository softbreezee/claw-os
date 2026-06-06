// Package goal implements persisted thread goals — a `/goal <objective>`
// becomes a long-running, audit-driven loop where the runtime keeps
// injecting continuation prompts until the model marks the goal
// complete, the token budget runs out, or the user pauses or clears
// it.
//
// The design is modeled on OpenAI Codex CLI's /goal (codex-rs/core/src/
// goals.rs) and the fastclaw upstream port. See docs/v0.3-plan.md
// § Week 4 for the claw-os scope (slash-only, single-agent dogfood).
//
// This package owns the domain types; persistence lives in
// internal/store/pg.GoalStore, and is reached through the Store
// interface defined here so internal/agent doesn't import internal/
// store/pg directly (the gateway wires an adapter in
// internal/gateway/goalstore_adapter.go).
package goal

import "time"

// Status is the lifecycle state of a goal. Four values; "unmet" is
// intentionally absent — a goal that cannot complete simply stays
// Active until the user pauses or clears it.
type Status = string

const (
	StatusActive        Status = "active"
	StatusPaused        Status = "paused"
	StatusBudgetLimited Status = "budget_limited"
	StatusComplete      Status = "complete"
)

// Goal is the agent-package-local view of a persisted goal row. The
// fields mirror internal/store/pg.GoalRecord so the gateway adapter
// can copy across with no derived state. We don't alias the pg type
// here because that would force every package importing internal/
// agent/goal to link in pgx; this domain type stays pure stdlib.
type Goal struct {
	ID         string
	AgentID    string
	SessionKey string

	// Routing tuple — stamped at create time so a continuation can
	// publish onto the same bus address the original turn arrived
	// on. claw-os routes on (Channel, ChatID); we don't carry an
	// AccountID / OwnerUserID like fastclaw because the runtime is
	// single-tenant in v0.3.
	Channel string
	ChatID  string

	Objective string
	Status    Status

	// TokenBudget is nil for unbounded goals.
	TokenBudget *int64
	TokensUsed  int64

	CreatedAt time.Time
	UpdatedAt time.Time
}

// RemainingTokens returns budget − used (≥0). When the goal has no
// budget, ok is false. Used by the prompt renderer.
func RemainingTokens(g *Goal) (remaining int64, ok bool) {
	if g == nil || g.TokenBudget == nil {
		return 0, false
	}
	r := *g.TokenBudget - g.TokensUsed
	if r < 0 {
		r = 0
	}
	return r, true
}
