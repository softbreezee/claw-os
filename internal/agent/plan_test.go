package agent

import (
	"strings"
	"testing"

	"github.com/softbreezee/claw-os/internal/bus"
	"github.com/softbreezee/claw-os/internal/provider"
)

// busInboundDummy returns a zero-value InboundMessage suitable for
// testing slashPlan, which only reads args (msg is unused). Kept as a
// helper so future plan-mode tests that DO inspect msg fields have a
// single place to extend.
func busInboundDummy() bus.InboundMessage {
	return bus.InboundMessage{Channel: "web", ChatID: "test-chat"}
}

// TestPlanModeNudge_Contract pins the load-bearing claims in the
// nudge: tools are disabled this turn, todo.md should be the first
// execution action, the model must end with the literal "Reply with
// 'go' …" line. If any of these wording cues drift the UI's
// downstream expectations (Continue/Adjust buttons, todo panel)
// silently break.
func TestPlanModeNudge_Contract(t *testing.T) {
	body := planModeNudge()
	mustContain := []string{
		"PLAN MODE",
		"DISABLED",
		"todo.md",
		"3-7 steps",
		"Reply with 'go' to execute",
	}
	for _, want := range mustContain {
		if !strings.Contains(body, want) {
			t.Errorf("planModeNudge missing required phrase %q", want)
		}
	}
}

// TestBuildToolCatalogForPlan_TruncatesAndLists locks two important
// behaviours: empty input returns "" (so the system message isn't
// injected as an empty block), and long descriptions get cut on a
// sentence boundary so the catalog stays scannable.
func TestBuildToolCatalogForPlan_TruncatesAndLists(t *testing.T) {
	if got := buildToolCatalogForPlan(nil); got != "" {
		t.Errorf("nil tools should produce empty catalog, got %q", got)
	}

	tools := []provider.Tool{
		{Function: provider.ToolFunction{
			Name:        "web_search",
			Description: "Search the web. Returns top results with snippets.",
		}},
		{Function: provider.ToolFunction{
			Name: "exec",
			Description: strings.Repeat(
				"Run a shell command in the agent workspace and return stdout/stderr/exit code; ", 6),
		}},
	}
	out := buildToolCatalogForPlan(tools)

	if !strings.Contains(out, "`web_search`") {
		t.Errorf("output missing web_search entry:\n%s", out)
	}
	if !strings.Contains(out, "Search the web") {
		t.Errorf("output missing first-sentence summary:\n%s", out)
	}
	// First sentence ends at the first period; the catalog must NOT
	// carry the rest of the sentence after the cut.
	if strings.Contains(out, "Returns top results") {
		t.Errorf("first-sentence cut not applied; output kept post-period text:\n%s", out)
	}
	// Run-on description must get a hard 200-char cut + ellipsis.
	if !strings.Contains(out, "…") {
		t.Errorf("long description should be truncated with ellipsis:\n%s", out)
	}
}

// TestSlashPlan_EmptyArgs returns a usage hint with no planTask flag
// — slashResult.planTask is the trigger that reroutes the loop into
// handlePlanMode, and we must NOT enter plan mode without a task.
func TestSlashPlan_EmptyArgs(t *testing.T) {
	a := &Agent{}
	res := a.slashPlan(busInboundDummy(), nil)
	if !res.handled {
		t.Fatal("/plan with no args should be handled")
	}
	if res.planTask != "" {
		t.Errorf("empty args produced planTask=%q, want empty", res.planTask)
	}
	if !strings.Contains(res.reply, "Usage:") {
		t.Errorf("expected usage hint in reply, got %q", res.reply)
	}
}

// TestSlashPlan_ParsesTask joins multi-word args and sets planTask so
// the loop knows to call handlePlanMode. reply stays empty — the
// plan stream IS the visible response.
func TestSlashPlan_ParsesTask(t *testing.T) {
	a := &Agent{}
	res := a.slashPlan(busInboundDummy(), []string{"复盘", "今天的", "两个", "账户"})
	if !res.handled {
		t.Fatal("/plan should be handled")
	}
	if res.planTask != "复盘 今天的 两个 账户" {
		t.Errorf("planTask=%q, want joined args", res.planTask)
	}
	if res.reply != "" {
		t.Errorf("reply should be empty in plan-trigger path, got %q", res.reply)
	}
}
