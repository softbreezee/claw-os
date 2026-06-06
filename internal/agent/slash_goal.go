package agent

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/softbreezee/claw-os/internal/agent/goal"
	"github.com/softbreezee/claw-os/internal/bus"
)

// slashGoal dispatches `/goal …` to one of the sub-handlers.
//
// Grammar:
//
//	/goal <objective>             → create
//	/goal                         → show current
//	/goal pause | resume | clear  → state transitions
//
// `/goal budget <N>` is deliberately absent — fastclaw doesn't ship
// it either and mid-flight budget changes have ambiguous semantics
// (do tokens already spent count?). Set the budget at create time;
// for v0.3 we hard-code an unbounded default and let users `/goal
// clear` + recreate if they want to bound a fresh attempt.
func (a *Agent) slashGoal(msg bus.InboundMessage, args []string) slashResult {
	if a.goalStore == nil {
		return slashResult{
			handled: true,
			reply:   "The /goal feature requires the PostgreSQL backend. Set FASTCLAW_STORAGE_DSN (or pawnix.json storage) to enable.",
		}
	}

	sub := ""
	if len(args) > 0 {
		sub = strings.ToLower(args[0])
	}
	switch sub {
	case "":
		return a.slashGoalShow(msg)
	case "pause":
		return a.slashGoalPause(msg)
	case "resume":
		return a.slashGoalResume(msg)
	case "clear":
		return a.slashGoalClear(msg)
	}
	// Default: treat the entire arg list as objective text. /goal
	// pause / resume / clear are short enough keywords that the
	// ambiguity ("did the user mean a goal called 'pause'?") isn't
	// real in practice.
	objective := strings.Join(args, " ")
	return a.slashGoalCreate(msg, objective)
}

func (a *Agent) slashGoalShow(msg bus.InboundMessage) slashResult {
	g, err := a.loadGoal(msg)
	if errors.Is(err, goal.ErrNotFound) || g == nil {
		return slashResult{handled: true, reply: "No goal set. Use `/goal <objective>` to start one."}
	}
	if err != nil {
		return slashResult{handled: true, reply: fmt.Sprintf("Error reading goal: %v", err)}
	}
	return slashResult{
		handled: true,
		reply:   fmt.Sprintf("**%s**: %s\n_tokens used: %d_", g.Status, g.Objective, g.TokensUsed),
	}
}

func (a *Agent) slashGoalCreate(msg bus.InboundMessage, objective string) slashResult {
	objective = strings.TrimSpace(objective)
	if objective == "" {
		return slashResult{handled: true, reply: "Usage: `/goal <objective>` — describe what you want to accomplish across multiple turns."}
	}

	g := &goal.Goal{
		ID:         goal.NewID(),
		AgentID:    a.name,
		SessionKey: a.sessionKeyFor(msg),
		Channel:    msg.Channel,
		ChatID:     msg.ChatID,
		Objective:  objective,
		Status:     goal.StatusActive,
	}
	if err := a.goalStore.CreateGoal(context.Background(), g); err != nil {
		if errors.Is(err, goal.ErrAlreadyExists) {
			return slashResult{
				handled: true,
				reply:   "Goal already exists for this session. `/goal clear` first if you want to redefine it.",
			}
		}
		return slashResult{handled: true, reply: fmt.Sprintf("Error creating goal: %v", err)}
	}

	// Fire the first continuation immediately off the user's own
	// /goal turn so the chat shows a substantive reply (the model
	// reading the objective and starting work) instead of a silent
	// ack. The continuation is non-blocking — bus full just drops
	// silently and the next user turn / heartbeat will refire.
	goal.TryFireContinuation(context.Background(), a.goalStore, a.messageBus, a.name, g.SessionKey)

	return slashResult{
		handled: true,
		reply:   fmt.Sprintf("✅ Goal set: %s\n\n_The agent will keep this in mind across turns. Use `/goal` to check status, `/goal pause` to suspend, `/goal clear` to drop._", objective),
	}
}

func (a *Agent) slashGoalPause(msg bus.InboundMessage) slashResult {
	return a.transitionGoal(msg, goal.StatusActive, goal.StatusPaused, "Goal isn't active.")
}

func (a *Agent) slashGoalResume(msg bus.InboundMessage) slashResult {
	res := a.transitionGoal(msg, goal.StatusPaused, goal.StatusActive, "Goal isn't paused.")
	if res.handled && strings.HasPrefix(res.reply, "▶") {
		// Re-arm the continuation loop. Same non-blocking semantics
		// as create — best-effort; PostTurn / heartbeat will retry.
		goal.TryFireContinuation(context.Background(), a.goalStore, a.messageBus, a.name, a.sessionKeyFor(msg))
	}
	return res
}

func (a *Agent) slashGoalClear(msg bus.InboundMessage) slashResult {
	g, err := a.loadGoal(msg)
	if errors.Is(err, goal.ErrNotFound) || g == nil {
		return slashResult{handled: true, reply: "No goal set."}
	}
	if err != nil {
		return slashResult{handled: true, reply: fmt.Sprintf("Error reading goal: %v", err)}
	}
	if err := a.goalStore.DeleteGoal(context.Background(), g.ID); err != nil {
		return slashResult{handled: true, reply: fmt.Sprintf("Error clearing goal: %v", err)}
	}
	return slashResult{handled: true, reply: "🗑️ Goal cleared."}
}

// transitionGoal centralizes load → state-check → flip → persist.
// Reply prefixed with ▶ on success so slashGoalResume can detect
// and fire a continuation; non-success replies surface the exact
// reason (no goal / wrong state / store error).
func (a *Agent) transitionGoal(msg bus.InboundMessage, from, to goal.Status, wrongStateMsg string) slashResult {
	g, err := a.loadGoal(msg)
	if errors.Is(err, goal.ErrNotFound) || g == nil {
		return slashResult{handled: true, reply: "No goal set."}
	}
	if err != nil {
		return slashResult{handled: true, reply: fmt.Sprintf("Error reading goal: %v", err)}
	}
	if g.Status != from {
		return slashResult{handled: true, reply: wrongStateMsg + fmt.Sprintf(" (current: %s)", g.Status)}
	}
	g.Status = to
	if err := a.goalStore.UpdateGoal(context.Background(), g); err != nil {
		return slashResult{handled: true, reply: fmt.Sprintf("Error updating goal: %v", err)}
	}
	verb := "paused"
	if to == goal.StatusActive {
		verb = "resumed"
	}
	return slashResult{handled: true, reply: fmt.Sprintf("▶ Goal %s.", verb)}
}

func (a *Agent) loadGoal(msg bus.InboundMessage) (*goal.Goal, error) {
	return a.goalStore.GetGoalBySession(context.Background(), a.name, a.sessionKeyFor(msg))
}

// maybeFireGoalContinuation runs after each completed turn — when an
// active goal exists for the (agent, session) pair, it publishes a
// new continuation prompt onto the bus tagged with OriginGoalContext.
//
// Skipped when:
//   - goalStore not wired (file-only deploy)
//   - the just-completed turn was itself a goal continuation
//     (msg.Origin == OriginGoalContext) — without this gate, every
//     continuation would chain a fresh one and a goal with an
//     unbounded budget would burn the LLM in a tight loop forever.
//     Real continuations are paced by user input or by the
//     accounting hook flipping status to BudgetLimited.
//
// The "fire on each user turn" pacing matches fastclaw's PostTurn
// hook design but keeps the v0.3 implementation simple — we don't
// need the full Hook abstraction yet.
func (a *Agent) maybeFireGoalContinuation(ctx context.Context, msg bus.InboundMessage) {
	if a.goalStore == nil {
		return
	}
	if msg.Origin == bus.OriginGoalContext {
		return
	}
	goal.TryFireContinuation(ctx, a.goalStore, a.messageBus, a.name, a.sessionKeyFor(msg))
}
