package setup

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/softbreezee/claw-os/internal/config"
	"gopkg.in/yaml.v3"
)

// --- Skills ---
//
// Skills can live in several layers (mirrors agent.SkillsLoader):
//   * agent workspace:    ~/.pawnix/agents/{agentId}/agent/skills/
//   * user-installed:     ~/.pawnix/skills/
// In each layer a skill is either a subdirectory containing SKILL.md or
// a single *.md file (the latter is the convention for short, one-shot
// skills like tradingagents-ashare.md).
//
// handleListSkills aggregates across both layers and de-duplicates by
// name; first-seen wins (so agent-layer overrides user-layer, matching
// the runtime precedence).

type skillInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Location    string `json:"location"`
	Type        string `json:"type"`            // layer: "builtin" | "user" | "agent"
	Builtin     bool   `json:"builtin"`         // convenience flag for the UI
	Owner       string `json:"owner,omitempty"` // agent id when scoped to one
	// Kind comes from the SKILL.md frontmatter `type:` field. Common
	// values seen in the wild: "skill" (atomic), "protocol" (multi-step
	// procedure), "suite" (orchestrator that loads other skills).
	// The UI uses this to badge orchestrator-style skills differently
	// from atomic capability skills.
	Kind string `json:"kind,omitempty"`
}

func (s *Server) handleListSkills(w http.ResponseWriter, r *http.Request) {
	homeDir, err := config.HomeDir()
	if err != nil {
		jsonResponse(w, http.StatusOK, []any{})
		return
	}

	seen := make(map[string]bool)
	var out []skillInfo

	// Layer 1: per-agent skills (highest precedence at runtime).
	agentsDir := filepath.Join(homeDir, "agents")
	if entries, rerr := os.ReadDir(agentsDir); rerr == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			agentID := e.Name()
			dir := filepath.Join(agentsDir, agentID, "agent", "skills")
			for _, sk := range scanSkillsDir(dir, "agent") {
				if seen[sk.Name] {
					continue
				}
				sk.Owner = agentID
				seen[sk.Name] = true
				out = append(out, sk)
			}
		}
	}

	// Layer 2: user-installed (shared across all agents).
	for _, sk := range scanSkillsDir(filepath.Join(homeDir, "skills"), "user") {
		if seen[sk.Name] {
			continue
		}
		seen[sk.Name] = true
		out = append(out, sk)
	}

	// Layer 3: builtin (shipped with the binary, project-root skills/ dir).
	for _, sk := range scanSkillsDir(builtinSkillsDir(), "builtin") {
		if seen[sk.Name] {
			continue
		}
		sk.Builtin = true
		seen[sk.Name] = true
		out = append(out, sk)
	}

	if out == nil {
		jsonResponse(w, http.StatusOK, []any{})
		return
	}
	jsonResponse(w, http.StatusOK, out)
}

// GET /api/skills/{name} – returns the full SKILL.md content of one
// skill, located via the same scan-all-layers logic as the list
// endpoint. Used by the Skills page detail dialog.
func (s *Server) handleGetSkill(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"error": "skill name required"})
		return
	}
	homeDir, err := config.HomeDir()
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	// Probe the same layer roots as handleListSkills, in the same
	// precedence order (per-agent > user > builtin). At each layer
	// accept either <name>/SKILL.md or <name>.md.
	roots := []string{}
	agentsDir := filepath.Join(homeDir, "agents")
	if entries, rerr := os.ReadDir(agentsDir); rerr == nil {
		for _, e := range entries {
			if e.IsDir() {
				roots = append(roots, filepath.Join(agentsDir, e.Name(), "agent", "skills"))
			}
		}
	}
	roots = append(roots, filepath.Join(homeDir, "skills"))
	if b := builtinSkillsDir(); b != "" {
		roots = append(roots, b)
	}

	for _, root := range roots {
		for _, candidate := range []string{
			filepath.Join(root, name, "SKILL.md"),
			filepath.Join(root, name+".md"),
		} {
			data, err := os.ReadFile(candidate)
			if err != nil {
				continue
			}
			jsonResponse(w, http.StatusOK, map[string]any{
				"name":     name,
				"content":  string(data),
				"location": candidate,
			})
			return
		}
	}

	jsonResponse(w, http.StatusNotFound, map[string]any{"error": "skill not found"})
}

// builtinSkillsDir locates the project-shipped skills/ directory by
// probing common positions relative to the running binary. Empty string
// if we can't find it (e.g. binary moved standalone).
//
// Probed in order:
//   1. {exeDir}/skills/         – installed layout
//   2. {exeDir}/../skills/      – go run / dev (bin/pawnix → repo root)
//   3. ./skills/                – CWD fallback
func builtinSkillsDir() string {
	candidates := []string{}
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "skills"),
			filepath.Join(exeDir, "..", "skills"),
		)
	}
	candidates = append(candidates, "skills")
	for _, p := range candidates {
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			return p
		}
	}
	return ""
}

// scanSkillsDir lists skills under one directory. A skill is either:
//   * a subdirectory whose root contains SKILL.md, or
//   * a top-level *.md file (single-file skill).
// Returns nothing on any IO error – the caller may pass paths that
// don't exist (e.g. a tenant has no per-agent skills dir).
func scanSkillsDir(dir, layer string) []skillInfo {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []skillInfo
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			path := filepath.Join(dir, name)
			skillFile := filepath.Join(path, "SKILL.md")
			out = append(out, skillInfo{
				Name:        name,
				Description: firstNonHeadingLine(skillFile),
				Location:    path,
				Type:        layer,
				Kind:        skillKind(skillFile),
			})
			continue
		}
		// Single-file skills: anything ending in .md (excluding common
		// non-skill markdown like README.md).
		lower := strings.ToLower(name)
		if !strings.HasSuffix(lower, ".md") || lower == "readme.md" {
			continue
		}
		path := filepath.Join(dir, name)
		out = append(out, skillInfo{
			Name:        strings.TrimSuffix(name, filepath.Ext(name)),
			Description: firstNonHeadingLine(path),
			Location:    path,
			Type:        layer,
			Kind:        skillKind(path),
		})
	}
	return out
}

// skillKind reads the YAML frontmatter `type:` field. Returns "" when
// the file has no frontmatter, no type field, or any parse error.
//
// Convention seen in the codebase:
//   - "skill" or empty → atomic capability (single-step "do X")
//   - "protocol"       → multi-step procedure
//   - "suite"          → orchestrator that loads other skills
func skillKind(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	frontmatter, _ := splitFrontmatter(data)
	if frontmatter == "" {
		return ""
	}
	var fm struct {
		Type string `yaml:"type"`
	}
	if err := yaml.Unmarshal([]byte(frontmatter), &fm); err != nil {
		return ""
	}
	return strings.TrimSpace(fm.Type)
}

// firstNonHeadingLine returns a one-line description for the skill.
//
// Strategy:
//  1. Parse the YAML frontmatter via gopkg.in/yaml.v3, which correctly
//     handles multi-line scalars like `description: |\n  ...`. The
//     previous line-by-line scanner returned literally `"|"` for these
//     skills, which is what the user saw on the Skills page.
//  2. If that yields a description, normalise to a single line: collapse
//     whitespace and clip overly long blocks at ~200 chars so the card
//     stays a card.
//  3. If no frontmatter description, fall back to the first non-blank,
//     non-# markdown line of the body.
func firstNonHeadingLine(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	frontmatter, body := splitFrontmatter(data)
	if frontmatter != "" {
		var fm struct {
			Description string `yaml:"description"`
		}
		if err := yaml.Unmarshal([]byte(frontmatter), &fm); err == nil {
			if d := normaliseDescription(fm.Description); d != "" {
				return d
			}
		}
	}

	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		return trimmed
	}
	return ""
}

// splitFrontmatter pulls the leading "---\n…\n---\n" block off a markdown
// file. Returns ("", body) when no frontmatter is present.
func splitFrontmatter(data []byte) (frontmatter, body string) {
	content := string(data)
	if !strings.HasPrefix(strings.TrimLeft(content, " \t\r\n"), "---") {
		return "", content
	}
	// Strip optional leading whitespace + the opening "---".
	trimmed := strings.TrimLeft(content, " \t\r\n")
	rest := trimmed[3:]
	endIdx := strings.Index(rest, "\n---")
	if endIdx < 0 {
		return "", content
	}
	return rest[:endIdx], rest[endIdx+4:]
}

// normaliseDescription collapses whitespace in a multi-line description
// to a single readable line and clips it so the UI card doesn't grow
// unbounded.
func normaliseDescription(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	// Collapse runs of whitespace (including newlines) into single spaces.
	s = strings.Join(strings.Fields(s), " ")
	const maxLen = 240
	if len(s) > maxLen {
		// Trim at the last space before maxLen so we don't cut a word in half.
		head := s[:maxLen]
		if i := strings.LastIndex(head, " "); i > 0 {
			head = head[:i]
		}
		s = head + "…"
	}
	return s
}

func (s *Server) handleDeleteSkill(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	homeDir, err := config.HomeDir()
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}

	// Builtin skills are part of the binary distribution; refuse to
	// delete them so users can't accidentally remove ships-by-default
	// behaviour. They can still hide a builtin by overriding the same
	// name with a user/agent layer.
	if builtin := builtinSkillsDir(); builtin != "" {
		for _, p := range []string{
			filepath.Join(builtin, name),
			filepath.Join(builtin, name+".md"),
		} {
			if _, sterr := os.Stat(p); sterr == nil {
				jsonResponse(w, http.StatusForbidden, map[string]any{
					"ok":    false,
					"error": "cannot delete a builtin skill (override it under ~/.pawnix/skills/ or an agent workspace instead)",
				})
				return
			}
		}
	}

	// Try every writable layer + both forms (dir / .md). Stop at first hit.
	candidates := []string{
		filepath.Join(homeDir, "skills", name),
		filepath.Join(homeDir, "skills", name+".md"),
	}
	if entries, rerr := os.ReadDir(filepath.Join(homeDir, "agents")); rerr == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			base := filepath.Join(homeDir, "agents", e.Name(), "agent", "skills")
			candidates = append(candidates,
				filepath.Join(base, name),
				filepath.Join(base, name+".md"),
			)
		}
	}
	for _, p := range candidates {
		if _, statErr := os.Stat(p); statErr != nil {
			continue
		}
		if rmErr := os.RemoveAll(p); rmErr != nil {
			jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": rmErr.Error()})
			return
		}
		jsonResponse(w, http.StatusOK, map[string]any{"ok": true})
		return
	}
	jsonResponse(w, http.StatusNotFound, map[string]any{"ok": false, "error": "skill not found"})
}

// --- Move ---

// moveSkillRequest specifies where to move a skill. Scope formats:
//   "user"        – move to ~/.pawnix/skills/ (shared across agents)
//   "agent:<id>"  – move to ~/.pawnix/agents/<id>/agent/skills/
type moveSkillRequest struct {
	Scope string `json:"scope"`
}

// POST /api/skills/{name}/move
//
// Moves a skill between writable layers (user ↔ agent). Builtin skills
// are immutable and cannot be moved out of the binary distribution; to
// "fork" a builtin into an editable layer the user should create a new
// skill with the same name in user/ or an agent/ workspace (override
// works because the runtime SkillsLoader picks higher-precedence layers
// first).
func (s *Server) handleMoveSkill(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "missing skill name"})
		return
	}

	var req moveSkillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid request"})
		return
	}

	homeDir, err := config.HomeDir()
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}

	// Resolve destination directory from scope.
	destDir, err := resolveScopeDir(homeDir, req.Scope)
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
		return
	}

	// Find the source: search user + every agent layer for either
	// {name}/ subdirectory or {name}.md single-file form. Refuse to
	// touch the builtin layer.
	srcPath, err := findWritableSkillPath(homeDir, name)
	if err != nil {
		jsonResponse(w, http.StatusNotFound, map[string]any{"ok": false, "error": err.Error()})
		return
	}

	// No-op if already in the destination layer.
	if filepath.Dir(srcPath) == destDir {
		jsonResponse(w, http.StatusOK, map[string]any{
			"ok": true, "noop": true, "location": srcPath,
		})
		return
	}

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}

	destPath := filepath.Join(destDir, filepath.Base(srcPath))

	// Refuse to overwrite an existing skill at the destination – avoids
	// silent data loss when the same name exists in two writable layers.
	if _, statErr := os.Stat(destPath); statErr == nil {
		jsonResponse(w, http.StatusConflict, map[string]any{
			"ok": false, "error": "destination already contains a skill with this name",
		})
		return
	}

	// os.Rename is atomic on the same filesystem and works for both files
	// and directories. ~/.pawnix is on one volume in normal installs.
	if err := os.Rename(srcPath, destPath); err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}

	jsonResponse(w, http.StatusOK, map[string]any{
		"ok":       true,
		"location": destPath,
	})
}

// resolveScopeDir maps a scope string to a writable destination directory.
func resolveScopeDir(homeDir, scope string) (string, error) {
	if scope == "user" {
		return filepath.Join(homeDir, "skills"), nil
	}
	if strings.HasPrefix(scope, "agent:") {
		agentID := strings.TrimPrefix(scope, "agent:")
		if agentID == "" || strings.ContainsAny(agentID, "/\\") {
			return "", fmt.Errorf("invalid agent id in scope: %q", scope)
		}
		return filepath.Join(homeDir, "agents", agentID, "agent", "skills"), nil
	}
	return "", fmt.Errorf("unsupported scope %q (want \"user\" or \"agent:<id>\")", scope)
}

// findWritableSkillPath returns the absolute path of a skill in any
// writable layer (user or per-agent), as either a directory or a .md
// file. Returns error if the skill only exists in builtin or doesn't
// exist at all – callers must not silently create a new file.
func findWritableSkillPath(homeDir, name string) (string, error) {
	// Search per-agent first (matches runtime precedence).
	if entries, err := os.ReadDir(filepath.Join(homeDir, "agents")); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			base := filepath.Join(homeDir, "agents", e.Name(), "agent", "skills")
			if p := matchSkillFile(base, name); p != "" {
				return p, nil
			}
		}
	}
	// Then user layer.
	if p := matchSkillFile(filepath.Join(homeDir, "skills"), name); p != "" {
		return p, nil
	}
	return "", fmt.Errorf("skill not found in any writable layer: %s", name)
}

// matchSkillFile returns the path of {dir}/{name} or {dir}/{name}.md if
// either exists, otherwise empty string.
func matchSkillFile(dir, name string) string {
	for _, candidate := range []string{
		filepath.Join(dir, name),
		filepath.Join(dir, name+".md"),
	} {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}
