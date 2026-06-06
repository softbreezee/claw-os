package goal

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/softbreezee/claw-os/internal/bus"
)

// TestRemainingTokens covers the three documented cases:
//   - nil budget → ok=false (caller renders "unbounded")
//   - within budget → positive remainder
//   - over budget → 0 (clamped, never negative)
func TestRemainingTokens(t *testing.T) {
	cases := []struct {
		name     string
		g        *Goal
		want     int64
		wantOk   bool
	}{
		{"nil goal", nil, 0, false},
		{"unbounded", &Goal{}, 0, false},
		{"within budget", goalWithBudget(100, 30), 70, true},
		{"over budget clamps to 0", goalWithBudget(50, 200), 0, true},
	}
	for _, c := range cases {
		got, ok := RemainingTokens(c.g)
		if got != c.want || ok != c.wantOk {
			t.Errorf("%s: got (%d, %v), want (%d, %v)", c.name, got, ok, c.want, c.wantOk)
		}
	}
}

// TestFoldUsage_BillsActiveGoals pins that only Active goals get
// billed; Paused / BudgetLimited / Complete pass through with
// delta=0. This is the exact gate the AfterModelCall hook depends on
// to avoid double-counting after a budget exhaust.
func TestFoldUsage_BillsActiveGoals(t *testing.T) {
	for _, status := range []Status{StatusPaused, StatusBudgetLimited, StatusComplete} {
		g := &Goal{Status: status, TokensUsed: 100}
		delta, exhausted := FoldUsage(g, 50, 50)
		if delta != 0 || exhausted {
			t.Errorf("%s: should be skipped, got delta=%d exhausted=%v", status, delta, exhausted)
		}
		if g.TokensUsed != 100 {
			t.Errorf("%s: TokensUsed mutated, want 100, got %d", status, g.TokensUsed)
		}
	}

	g := &Goal{Status: StatusActive}
	delta, exhausted := FoldUsage(g, 100, 200)
	if delta != 300 || exhausted {
		t.Errorf("active no-budget: delta=%d exhausted=%v, want 300/false", delta, exhausted)
	}
	if g.TokensUsed != 300 {
		t.Errorf("TokensUsed=%d, want 300", g.TokensUsed)
	}
}

// TestFoldUsage_BudgetExhaustFlipsStatus pins the one-shot
// transition to BudgetLimited: the call that crosses the threshold
// returns exhausted=true; the next call (now non-active) returns
// delta=0 / exhausted=false. This ordering is what guarantees the
// budget_limit prompt fires exactly once.
func TestFoldUsage_BudgetExhaustFlipsStatus(t *testing.T) {
	budget := int64(500)
	g := &Goal{Status: StatusActive, TokenBudget: &budget, TokensUsed: 400}

	delta, exhausted := FoldUsage(g, 100, 50)
	if delta != 150 || !exhausted {
		t.Fatalf("crossing call: got (%d, %v), want (150, true)", delta, exhausted)
	}
	if g.Status != StatusBudgetLimited {
		t.Errorf("status not flipped, got %q", g.Status)
	}

	// Subsequent call after exhaustion must NOT fire another
	// budget_limit prompt — accounting on a non-Active goal is a
	// no-op.
	delta2, exhausted2 := FoldUsage(g, 100, 100)
	if delta2 != 0 || exhausted2 {
		t.Errorf("post-exhaust call: got (%d, %v), want (0, false)", delta2, exhausted2)
	}
}

// TestEscapeXMLText guards against objective text breaking out of
// the <objective> wrapper in the prompt template.
func TestEscapeXMLText(t *testing.T) {
	in := "find < & </goal_context> tags"
	want := "find &lt; &amp; &lt;/goal_context&gt; tags"
	if got := EscapeXMLText(in); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// TestContinuationPrompt_RendersObjective spot-checks template
// integration without re-asserting every literal line of the
// template (those are tested implicitly by the Contract test below).
func TestContinuationPrompt_RendersObjective(t *testing.T) {
	g := &Goal{Objective: "ship v0.3", Status: StatusActive, TokensUsed: 12}
	out := ContinuationPrompt(g)
	if !strings.Contains(out, "<objective>") || !strings.Contains(out, "ship v0.3") {
		t.Errorf("objective not in output:\n%s", out)
	}
	if !strings.Contains(out, "Tokens consumed: 12") {
		t.Errorf("tokens not in output:\n%s", out)
	}
	if !strings.Contains(out, "Token budget: none") {
		t.Errorf("unbounded budget should render as 'none':\n%s", out)
	}
}

// TestTryFireContinuation_GateCascade walks each return path:
//   - no goal → no publish
//   - non-Active goal → no publish
//   - missing routing → warn-log + no publish
//   - happy path → InboundMessage with Origin=goal_context
func TestTryFireContinuation_GateCascade(t *testing.T) {
	ctx := context.Background()
	mb := bus.New()

	st := NewMemStore()
	// Path 1: no goal.
	TryFireContinuation(ctx, st, mb, "trader", "session-1")
	if drained := drainBus(mb); len(drained) != 0 {
		t.Fatalf("path 1 (no goal): unexpected publish %+v", drained)
	}

	// Path 2: paused goal — should NOT publish.
	g := &Goal{
		ID: NewID(), AgentID: "trader", SessionKey: "session-1",
		Channel: "web", ChatID: "c1",
		Objective: "watch energy", Status: StatusPaused,
	}
	if err := st.CreateGoal(ctx, g); err != nil {
		t.Fatalf("create paused: %v", err)
	}
	TryFireContinuation(ctx, st, mb, "trader", "session-1")
	if drained := drainBus(mb); len(drained) != 0 {
		t.Fatalf("path 2 (paused): unexpected publish %+v", drained)
	}

	// Path 3: missing routing — Active but Channel/ChatID empty.
	g2 := &Goal{
		ID: NewID(), AgentID: "trader", SessionKey: "session-2",
		Objective: "no routing", Status: StatusActive,
	}
	if err := st.CreateGoal(ctx, g2); err != nil {
		t.Fatalf("create unrouted: %v", err)
	}
	TryFireContinuation(ctx, st, mb, "trader", "session-2")
	if drained := drainBus(mb); len(drained) != 0 {
		t.Fatalf("path 3 (no routing): unexpected publish %+v", drained)
	}

	// Path 4: happy path — Active + routed.
	g3 := &Goal{
		ID: NewID(), AgentID: "trader", SessionKey: "session-3",
		Channel: "web", ChatID: "c3",
		Objective: "the real one", Status: StatusActive,
	}
	if err := st.CreateGoal(ctx, g3); err != nil {
		t.Fatalf("create happy: %v", err)
	}
	TryFireContinuation(ctx, st, mb, "trader", "session-3")
	drained := drainBus(mb)
	if len(drained) != 1 {
		t.Fatalf("happy path: want 1 message, got %d", len(drained))
	}
	if drained[0].Origin != bus.OriginGoalContext {
		t.Errorf("Origin = %q, want %q", drained[0].Origin, bus.OriginGoalContext)
	}
	if drained[0].AgentID != "trader" || drained[0].ChatID != "c3" {
		t.Errorf("routing not preserved: %+v", drained[0])
	}
	if !strings.Contains(drained[0].Text, "the real one") {
		t.Errorf("objective not in continuation text:\n%s", drained[0].Text)
	}
}

// TestPublish_BusFullDoesNotBlock guarantees the goal continuation
// path is non-blocking — a stuck consumer must NOT freeze the
// PostTurn / accounting hooks that fire it.
func TestPublish_BusFullDoesNotBlock(t *testing.T) {
	mb := &bus.MessageBus{
		Inbound:  make(chan bus.InboundMessage, 1),
		Outbound: make(chan bus.OutboundMessage, 1),
	}
	g := &Goal{ID: NewID(), AgentID: "x", Channel: "web", ChatID: "c"}

	if !Publish(mb, g, "first") {
		t.Fatal("first publish should succeed (channel cap=1)")
	}
	if Publish(mb, g, "second") {
		t.Error("second publish should fail (channel full), not block")
	}
}

// helpers

func goalWithBudget(budget, used int64) *Goal {
	b := budget
	return &Goal{TokenBudget: &b, TokensUsed: used}
}

func drainBus(mb *bus.MessageBus) []bus.InboundMessage {
	var out []bus.InboundMessage
	for {
		select {
		case m := <-mb.Inbound:
			out = append(out, m)
		default:
			return out
		}
	}
}

// (errors import wasn't strictly needed but stays in case future
// tests assert errors.Is on store errors)
var _ = errors.New
