package agent

import (
	"strings"
	"testing"
)

// TestTrimSkillSummary covers the rune-safe truncation used by
// BuildSkillsSummary. Important properties: collapse internal
// whitespace, leave short strings untouched, never slice mid-rune,
// add an ellipsis when the cut happens.
func TestTrimSkillSummary(t *testing.T) {
	cases := []struct {
		name string
		in   string
		max  int
		want string
	}{
		{
			name: "short stays whole",
			in:   "  hello   world  ",
			max:  100,
			want: "hello world",
		},
		{
			name: "ascii cut adds ellipsis",
			in:   "abcdefghij",
			max:  4,
			want: "abcd…",
		},
		{
			name: "cjk cut on rune boundary",
			// 6 CJK runes (each 3 bytes in UTF-8); cap at 3 runes
			in:   "盘前简报作战计划",
			max:  3,
			want: "盘前简…",
		},
		{
			name: "zero max means no truncation",
			in:   "anything",
			max:  0,
			want: "anything",
		},
	}
	for _, c := range cases {
		got := trimSkillSummary(c.in, c.max)
		if got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

// TestBuildSkillsSummary_BudgetEnforced builds a synthetic loader and
// 100 skills with long descriptions. Asserts:
//   - the rendered section is under SkillsSummaryBudgetTokens (× a
//     small slack to absorb XML scaffolding)
//   - a "<more dropped=N" hint surfaces so the model knows to call
//     load_skill for skills beyond the listed prefix
//   - always-loaded skills survive truncation (they appear in the
//     output regardless of where they fall in input order)
func TestBuildSkillsSummary_BudgetEnforced(t *testing.T) {
	sl := &SkillsLoader{}

	// Long-ish description — enough that ~100 skills would otherwise
	// blow well past the budget.
	longDesc := strings.Repeat("trading strategy and analysis pipeline ", 5) // ~190 chars

	skills := make([]Skill, 0, 100)
	for i := 0; i < 100; i++ {
		skills = append(skills, Skill{
			Name:        skillName(i),
			Layer:       "user",
			Description: longDesc,
			Content:     "ignored — non-always-loaded skills only render their description",
		})
	}
	// Sentinel: an always-loaded skill in the middle. We use the
	// metadata.always path to avoid wiring config.SkillsConfig in
	// the test.
	always := Skill{
		Name:        "trend-trader-must-load",
		Layer:       "agent",
		Description: "always-loaded sentinel",
		Content:     "ALWAYS-BODY-MARKER",
		Metadata: &SkillMetadata{
			Pawnix: &OpenClawMeta{Always: true},
		},
	}
	skills = append(skills[:50], append([]Skill{always}, skills[50:]...)...)

	out := sl.BuildSkillsSummary(skills)

	if !strings.Contains(out, "ALWAYS-BODY-MARKER") {
		t.Errorf("always-loaded skill body missing from output")
	}
	if !strings.Contains(out, "<more dropped=") {
		preview := out
		if len(preview) > 400 {
			preview = preview[:400]
		}
		t.Errorf("expected truncation hint when 100 skills overflow budget; got:\n%s", preview)
	}
	tokens := estimateStringTokens(out)
	// Allow modest slack over the per-section budget for XML
	// scaffolding + the always-loaded body that is exempt.
	if tokens > SkillsSummaryBudgetTokens*2 {
		t.Errorf("section tokens=%d, budget=%d — budget enforcement is broken",
			tokens, SkillsSummaryBudgetTokens)
	}
}

// TestBuildSkillsSummary_EmptyReturnsEmpty pins the contract that an
// empty skills set produces "" (not "<skills></skills>"), so the
// system prompt builder can skip the section entirely.
func TestBuildSkillsSummary_EmptyReturnsEmpty(t *testing.T) {
	sl := &SkillsLoader{}
	if got := sl.BuildSkillsSummary(nil); got != "" {
		t.Errorf("empty skills should produce empty string, got %q", got)
	}
}

func skillName(i int) string {
	return "skill-" + string(rune('a'+i%26)) + "-" + intStr(i)
}

func intStr(i int) string {
	if i == 0 {
		return "0"
	}
	var digits []byte
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}
	return string(digits)
}
