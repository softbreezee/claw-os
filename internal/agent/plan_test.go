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

// TestSanitizePlanText_StripsToolLeaks pins the TRUNCATE-at-first-leak
// behaviour — the dogfood bug where the model ignored "plan only, no
// tools" and dumped a raw tool call (sometimes with big inline
// curl/python payloads between the tags). We cut at the first leak
// marker: anything BEFORE it (genuine preamble) survives, everything
// from the marker onward (the whole tool dump) is discarded.
func TestSanitizePlanText_StripsToolLeaks(t *testing.T) {
	cases := []struct {
		name string
		in   string
		// substrings that MUST be gone after scrubbing
		gone []string
		// substrings that MUST survive (prose before the first leak)
		keep []string
	}{
		{
			name: "anthropic invoke block truncates tail",
			in:   "Step 1: check the filing.\n<invoke name=\"web_fetch\"><parameter name=\"url\">https://x.com</parameter></invoke>\nStep 2: summarize.",
			gone: []string{"invoke", "parameter", "web_fetch", "https://x.com", "Step 2: summarize."},
			keep: []string{"Step 1: check the filing."},
		},
		{
			name: "openai tool_call block truncates tail",
			in:   "Plan:\n<tool_call>{\"name\":\"exec\"}</tool_call>\nDone.",
			gone: []string{"tool_call", "exec", "Done."},
			keep: []string{"Plan:"},
		},
		{
			name: "DSML fullwidth-pipe leak",
			in:   "Preamble.\n<｜tool▁calls▁begin｜>web_fetch<｜tool▁calls▁end｜>\nexecution attempt here.",
			gone: []string{"tool▁calls", "｜", "web_fetch", "execution attempt"},
			keep: []string{"Preamble."},
		},
		{
			name: "mixed leak with inline curl/python (screenshot 2)",
			in:   "好。我先拉一下它的实际业务。\n<｜DSML｜tool_calls><｜DSML｜invoke name=\"exec\"><｜DSML｜parameter name=\"command\">curl -sL \"https://xueqiu.com/S/SH600283\" | python3 -c 'import re; print(re.findall(...))'</｜DSML｜parameter></｜DSML｜invoke>",
			gone: []string{"DSML", "exec", "curl", "python3", "re.findall", "xueqiu"},
			keep: []string{"好。我先拉一下它的实际业务。"},
		},
		{
			name: "clean plan untouched (tool named in prose)",
			in:   "Step 1: use web_fetch on the Sina page.\nStep 2: summarize.",
			gone: []string{},
			keep: []string{"Step 1: use web_fetch on the Sina page.", "Step 2: summarize."},
		},
	}
	for _, c := range cases {
		got, _ := sanitizePlanText(c.in)
		for _, g := range c.gone {
			if strings.Contains(got, g) {
				t.Errorf("%s: %q should have been removed, still in:\n%s", c.name, g, got)
			}
		}
		for _, k := range c.keep {
			if !strings.Contains(got, k) {
				t.Errorf("%s: %q should have survived, missing from:\n%s", c.name, k, got)
			}
		}
	}
}

// TestSanitizePlanText_LeakFlag pins the leaked-bool: a clean plan
// reports leaked=false (so resolvePlan keeps it), any tool-call shape
// reports leaked=true (so resolvePlan can fall back when the
// remainder is thin).
func TestSanitizePlanText_LeakFlag(t *testing.T) {
	if _, leaked := sanitizePlanText("Step 1: use web_fetch on the Sina page.\nStep 2: summarize."); leaked {
		t.Error("clean plan that merely NAMES a tool in prose must not be flagged as leaked")
	}
	if _, leaked := sanitizePlanText(`<invoke name="web_fetch"><parameter name="url">x</parameter></invoke>`); !leaked {
		t.Error("anthropic invoke block must be flagged leaked")
	}
	if _, leaked := sanitizePlanText("< | DSML | tool_calls>"); !leaked {
		t.Error("DSML token must be flagged leaked")
	}
}

// TestResolvePlan covers the four branches the renderer depends on:
//   - clean plan → passes through unchanged
//   - full leak (pure tool dump) → retry hint, no garbage
//   - mixed leak (real preamble + tool dump) → keep preamble + notice
//   - empty → empty-plan hint
func TestResolvePlan(t *testing.T) {
	real := "Step 1: pull both accounts' holdings.\nStep 2: assess each vs cost basis.\nStep 3: write todo.md.\nReply with 'go' to execute."
	if out := resolvePlan(real); out != real {
		t.Errorf("real plan should pass through, got %q", out)
	}

	// Pure leak: whole response is a mangled web_fetch dump. Truncating
	// at the first marker leaves nothing → retry hint.
	screenshot := `< | DSML | tool_calls> < | DSML | invoke name="web_fetch">
< | DSML | parameter name="max_length" string="false">6000</ | DSML | parameter> < | DSML | parameter name="url" string="true">https://stock.finance.sina.com.cn/corp/go.php/vCI_CorpInfo/stockid/600283.phtml</ | DSML | parameter> </ | DSML | invoke> </ | DSML | tool_calls>`
	out := resolvePlan(screenshot)
	if strings.Contains(out, "DSML") || strings.Contains(out, "web_fetch") || strings.Contains(out, "sina.com") {
		t.Errorf("pure leak must not survive into rendered plan:\n%s", out)
	}
	if !strings.Contains(out, "/plan") {
		t.Errorf("pure leak should yield retry hint, got %q", out)
	}

	// Mixed leak (screenshot 2 shape): a real Chinese preamble, then a
	// tool dump with inline curl/python. Preamble survives; the dump
	// (and its code) is gone; a notice is appended.
	mixed := "好。我之前没拆它,因为它的逻辑和氦气/氪气/六氟化钨完全不是一个维度,但名字里的「燃气」确实容易让人联想。我先拉一下它的实际业务。\n" +
		"<｜DSML｜tool_calls><｜DSML｜invoke name=\"exec\"><｜DSML｜parameter name=\"command\">curl -sL \"https://xueqiu.com/S/SH600283\" | python3 -c 'import re; print(re.findall(r\"stockName\", content))'</｜DSML｜parameter></｜DSML｜invoke>"
	mixedOut := resolvePlan(mixed)
	if strings.Contains(mixedOut, "DSML") || strings.Contains(mixedOut, "curl") ||
		strings.Contains(mixedOut, "python3") || strings.Contains(mixedOut, "xueqiu") ||
		strings.Contains(mixedOut, "re.findall") {
		t.Errorf("mixed leak: tool dump / code must not survive:\n%s", mixedOut)
	}
	if !strings.Contains(mixedOut, "六氟化钨") {
		t.Errorf("mixed leak: real preamble should survive:\n%s", mixedOut)
	}
	if !strings.Contains(mixedOut, "已拦截") {
		t.Errorf("mixed leak: should append an interception notice:\n%s", mixedOut)
	}

	if out := resolvePlan(""); !strings.Contains(out, "/plan") {
		t.Errorf("empty input should yield a hint, got %q", out)
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
