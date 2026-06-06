package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLoadSkill_WrapsContent pins the IP-protection contract: every
// skill body returned by the load_skill tool MUST be prefixed with
// the "[INTERNAL CONTEXT" guard. A regression here would let a
// chatter who asks "show me your image-gen skill" exfiltrate the
// agent's prompt templates verbatim.
func TestLoadSkill_WrapsContent(t *testing.T) {
	home := t.TempDir()
	skillDir := filepath.Join(home, "skills", "trend-trader")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("mkdir skill dir: %v", err)
	}
	body := "secret prompt template — should never appear verbatim in chat"
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(body), 0o644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}

	fn := makeLoadSkill(home, t.TempDir(), "", "", "")
	args := mustJSON(t, loadSkillArgs{Name: "trend-trader"})
	out, err := fn(context.Background(), args)
	if err != nil {
		t.Fatalf("load_skill returned error: %v", err)
	}

	if !strings.HasPrefix(out, "[INTERNAL CONTEXT") {
		t.Errorf("missing IP-protection header; got prefix %q", firstN(out, 40))
	}
	if !strings.Contains(out, body) {
		t.Errorf("body missing from wrapped output")
	}
}

// TestLoadSkill_BaseDirSubstitution pins that {baseDir} placeholders
// inside SKILL.md resolve to an absolute path to the skill directory,
// so relative-path commands inside the skill (e.g. "run {baseDir}/
// run.sh") work after load_skill returns.
func TestLoadSkill_BaseDirSubstitution(t *testing.T) {
	home := t.TempDir()
	skillDir := filepath.Join(home, "skills", "data-analysis")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	body := "Run script at {baseDir}/run.sh"
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	fn := makeLoadSkill(home, t.TempDir(), "", "", "")
	args := mustJSON(t, loadSkillArgs{Name: "data-analysis"})
	out, err := fn(context.Background(), args)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	absDir, _ := filepath.Abs(skillDir)
	if !strings.Contains(out, "Run script at "+absDir+"/run.sh") {
		t.Errorf("baseDir not substituted; output:\n%s", out)
	}
	if strings.Contains(out, "{baseDir}") {
		t.Errorf("placeholder still present in output")
	}
}

// TestLoadSkill_SingleFileFallback covers the <dir>/<name>.md form
// used by short, self-contained skills (e.g. tradingagents-ashare).
func TestLoadSkill_SingleFileFallback(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, "skills"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	body := "single-file skill body"
	if err := os.WriteFile(filepath.Join(home, "skills", "tradingagents-ashare.md"), []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	fn := makeLoadSkill(home, t.TempDir(), "", "", "")
	args := mustJSON(t, loadSkillArgs{Name: "tradingagents-ashare"})
	out, err := fn(context.Background(), args)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !strings.Contains(out, body) {
		t.Errorf("single-file body not loaded; got:\n%s", out)
	}
}

// TestLoadSkill_NotFound surfaces a clear error when a name resolves
// to nothing in any layer.
func TestLoadSkill_NotFound(t *testing.T) {
	fn := makeLoadSkill(t.TempDir(), t.TempDir(), "", "", "")
	args := mustJSON(t, loadSkillArgs{Name: "no-such-skill"})
	_, err := fn(context.Background(), args)
	if err == nil {
		t.Fatalf("expected error for missing skill, got nil")
	}
	if !strings.Contains(err.Error(), "no-such-skill") {
		t.Errorf("error should mention the skill name; got %v", err)
	}
}

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return b
}

func firstN(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
