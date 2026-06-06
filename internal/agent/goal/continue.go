package goal

import (
	"context"
	"errors"
	"log/slog"

	"github.com/softbreezee/claw-os/internal/bus"
)

// TryFireContinuation runs the gate cascade for one (agent, session)
// and, on success, publishes a continuation prompt back onto the
// inbound bus. Safe to call synchronously from PostTurn / accounting
// hooks — gates are cheap (one indexed read) and any failure is a
// silent no-op.
//
// Gates (in order):
//   - goal exists for this (agent, session)
//   - goal status is Active (Paused / BudgetLimited / Complete are
//     terminal-for-continuation; the user must /goal resume to re-arm)
//   - goal carries routing info (channel + chat_id) — without them
//     the continuation message would be unroutable on the bus
//
// Errors land at warn level rather than blocking the caller —
// continuation is best-effort, and the next PostTurn will retry if
// the failure was transient (DB blip, full bus channel).
func TryFireContinuation(ctx context.Context, st Store, mb *bus.MessageBus, agentID, sessionKey string) {
	g, err := st.GetGoalBySession(ctx, agentID, sessionKey)
	if errors.Is(err, ErrNotFound) {
		return
	}
	if err != nil {
		slog.Warn("goal continue: load goal failed",
			"agent_id", agentID, "session_key", sessionKey, "error", err)
		return
	}
	if g.Status != StatusActive {
		return
	}
	if g.Channel == "" && g.ChatID == "" {
		slog.Warn("goal continue: skipping — goal has no routing info",
			"agent_id", agentID, "session_key", sessionKey, "goal_id", g.ID)
		return
	}
	if !Publish(mb, g, ContinuationPrompt(g)) {
		slog.Warn("goal continue: bus full, dropped continuation",
			"agent_id", agentID, "session_key", sessionKey)
	}
}

// Publish pushes a goal-context prompt (continuation or budget-limit
// wrap-up) onto the bus. Tagged with bus.OriginGoalContext so the
// agent loop and reply router can distinguish runtime-injected goal
// prompts from real user input. Returns true when queued, false when
// the bus channel is full.
//
// PeerKind="dm" because all goal-fired turns are 1:1 — even if the
// original session lived in a group chat, the continuation should
// behave like the agent answering the user directly, not addressing
// the room.
func Publish(mb *bus.MessageBus, g *Goal, prompt string) bool {
	if mb == nil || g == nil {
		return false
	}
	msg := bus.InboundMessage{
		Channel:  g.Channel,
		ChatID:   g.ChatID,
		UserID:   "goal",
		AgentID:  g.AgentID,
		Text:     prompt,
		PeerKind: "dm",
		Origin:   bus.OriginGoalContext,
	}
	select {
	case mb.Inbound <- msg:
		return true
	default:
		return false
	}
}
