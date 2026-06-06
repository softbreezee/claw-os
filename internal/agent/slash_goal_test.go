package agent

import (
	"strings"
	"testing"

	"github.com/softbreezee/claw-os/internal/agent/goal"
	"github.com/softbreezee/claw-os/internal/bus"
)

// agentForGoalTest builds a minimal Agent wired with the in-memory
// goal store, no provider, no sessions. Just enough to exercise the
// slash dispatch + persistence state machine without spinning up
// the LLM stack.
func agentForGoalTest() *Agent {
	mb := bus.New()
	return &Agent{
		name:       "trader",
		messageBus: mb,
		goalStore:  goal.NewMemStore(),
	}
}

func goalMsg() bus.InboundMessage {
	return bus.InboundMessage{Channel: "web", ChatID: "c-1"}
}

// TestSlashGoal_Disabled confirms the "PG not wired" path returns a
// helpful message instead of panicking. Important: file-only deploys
// must keep working; /goal just stays unavailable.
func TestSlashGoal_Disabled(t *testing.T) {
	a := &Agent{name: "trader"} // no goalStore
	res := a.slashGoal(goalMsg(), nil)
	if !res.handled {
		t.Fatal("disabled /goal should still be handled (with explanation)")
	}
	if !strings.Contains(res.reply, "PostgreSQL backend") {
		t.Errorf("reply should mention PG requirement, got %q", res.reply)
	}
}

// TestSlashGoal_CreateShowClear walks the happy path:
//
//	/goal <task>            → create + ✅
//	/goal                   → show with current status
//	/goal pause             → ▶ paused
//	/goal resume            → ▶ resumed
//	/goal clear             → 🗑️
//	/goal                   → "no goal set"
func TestSlashGoal_CreateShowClear(t *testing.T) {
	a := agentForGoalTest()
	msg := goalMsg()

	res := a.slashGoal(msg, []string{"watch", "energy", "stocks"})
	if !res.handled || !strings.Contains(res.reply, "Goal set") {
		t.Fatalf("create reply unexpected: %+v", res)
	}

	res = a.slashGoal(msg, nil) // show
	if !strings.Contains(res.reply, "active") || !strings.Contains(res.reply, "watch energy stocks") {
		t.Errorf("show reply unexpected: %q", res.reply)
	}

	res = a.slashGoal(msg, []string{"pause"})
	if !strings.HasPrefix(res.reply, "▶") || !strings.Contains(res.reply, "paused") {
		t.Errorf("pause reply unexpected: %q", res.reply)
	}

	// pause again → wrong state
	res = a.slashGoal(msg, []string{"pause"})
	if !strings.Contains(res.reply, "isn't active") {
		t.Errorf("double-pause should report wrong state, got %q", res.reply)
	}

	res = a.slashGoal(msg, []string{"resume"})
	if !strings.HasPrefix(res.reply, "▶") || !strings.Contains(res.reply, "resumed") {
		t.Errorf("resume reply unexpected: %q", res.reply)
	}

	res = a.slashGoal(msg, []string{"clear"})
	if !strings.Contains(res.reply, "cleared") {
		t.Errorf("clear reply unexpected: %q", res.reply)
	}

	res = a.slashGoal(msg, nil)
	if !strings.Contains(res.reply, "No goal set") {
		t.Errorf("show after clear should report empty, got %q", res.reply)
	}
}

// TestSlashGoal_DuplicateCreate enforces the UNIQUE (agent, session)
// guarantee from the slash layer's perspective: create twice without
// clearing returns the "clear first" hint, not a generic error or
// silent overwrite.
func TestSlashGoal_DuplicateCreate(t *testing.T) {
	a := agentForGoalTest()
	msg := goalMsg()

	a.slashGoal(msg, []string{"first goal"})
	res := a.slashGoal(msg, []string{"second goal"})
	// Reply contains backticks (`/goal clear` first) so we match
	// the structural cue "already exists" rather than a literal
	// substring that would be brittle against copy edits.
	if !strings.Contains(res.reply, "already exists") || !strings.Contains(res.reply, "/goal clear") {
		t.Errorf("duplicate create should hint /goal clear, got %q", res.reply)
	}
}

// TestSlashGoal_EmptyObjective rejects /goal with no args (after the
// "no sub-command" branch falls through to create with empty
// objective). Without the guard, we'd insert a row with
// Objective="" and the model would get an empty <objective> wrapper
// next turn.
func TestSlashGoal_EmptyObjective(t *testing.T) {
	a := agentForGoalTest()
	res := a.slashGoalCreate(goalMsg(), "   ")
	if !strings.Contains(res.reply, "Usage:") {
		t.Errorf("empty objective should produce usage hint, got %q", res.reply)
	}
}

// TestSlashGoal_PerSessionIsolation: same agent, two different
// (channel, chatID) tuples → two independent goals. Important
// because the SessionKey derivation is "channel:chatID" and a regression
// would silently merge unrelated chats.
func TestSlashGoal_PerSessionIsolation(t *testing.T) {
	a := agentForGoalTest()
	msgA := bus.InboundMessage{Channel: "web", ChatID: "chat-A"}
	msgB := bus.InboundMessage{Channel: "web", ChatID: "chat-B"}

	a.slashGoal(msgA, []string{"goal A"})
	a.slashGoal(msgB, []string{"goal B"})

	resA := a.slashGoal(msgA, nil)
	resB := a.slashGoal(msgB, nil)
	if !strings.Contains(resA.reply, "goal A") {
		t.Errorf("session A reply mismatched: %q", resA.reply)
	}
	if !strings.Contains(resB.reply, "goal B") {
		t.Errorf("session B reply mismatched: %q", resB.reply)
	}
}

// TestMaybeFireGoalContinuation_BreaksLoop pins the chain-prevention
// gate: a turn whose own Origin is OriginGoalContext must NOT
// trigger another continuation, otherwise an unbounded-budget goal
// would tight-loop the LLM forever.
func TestMaybeFireGoalContinuation_BreaksLoop(t *testing.T) {
	a := agentForGoalTest()
	a.slashGoalCreate(goalMsg(), "watch energy stocks")
	// Drain the create-time continuation so the queue is empty for
	// the assertion below.
	for {
		select {
		case <-a.messageBus.Inbound:
			continue
		default:
			goto drained
		}
	}
drained:

	// Simulate finishing a goal-fired turn.
	contextMsg := bus.InboundMessage{
		Channel: "web", ChatID: "c-1",
		Origin: bus.OriginGoalContext,
	}
	a.maybeFireGoalContinuation(nil, contextMsg) // ctx unused in TryFireContinuation reads

	select {
	case got := <-a.messageBus.Inbound:
		t.Fatalf("goal-fired turn should NOT chain another continuation, got %+v", got)
	default:
	}

	// User-originated turn DOES fire one.
	a.maybeFireGoalContinuation(nil, goalMsg())
	select {
	case got := <-a.messageBus.Inbound:
		if got.Origin != bus.OriginGoalContext {
			t.Errorf("expected Origin=goal_context, got %q", got.Origin)
		}
	default:
		t.Fatal("user-originated turn should fire a continuation")
	}
}
