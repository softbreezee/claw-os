package setup

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/fastclaw-ai/fastclaw/internal/config"
)

// --- Skills ---
//
// Skills can live in several layers (mirrors agent.SkillsLoader):
//   * agent workspace:    ~/.fastclaw/agents/{agentId}/agent/skills/
//   * user-installed:     ~/.fastclaw/skills/
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
	Type        string `json:"type"`            // "builtin" | "user" | "agent"
	Builtin     bool   `json:"builtin"`         // convenience flag for the UI
	Owner       string `json:"owner,omitempty"` // agent id when scoped to one
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

// builtinSkillsDir locates the project-shipped skills/ directory by
// probing common positions relative to the running binary. Empty string
// if we can't find it (e.g. binary moved standalone).
//
// Probed in order:
//   1. {exeDir}/skills/         – installed layout
//   2. {exeDir}/../skills/      – go run / dev (bin/fastclaw → repo root)
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
			out = append(out, skillInfo{
				Name:        name,
				Description: firstNonHeadingLine(filepath.Join(path, "SKILL.md")),
				Location:    path,
				Type:        layer,
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
		})
	}
	return out
}

// firstNonHeadingLine returns the first non-empty, non-#-prefixed line
// of a markdown file – used as a one-line description. Skips a leading
// YAML frontmatter block (`---` ... `---`) and prefers a `description:`
// frontmatter key when present.
func firstNonHeadingLine(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")

	// Detect leading frontmatter and harvest description if it's there.
	body := lines
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		for i := 1; i < len(lines); i++ {
			line := lines[i]
			if strings.TrimSpace(line) == "---" {
				body = lines[i+1:]
				break
			}
			if strings.HasPrefix(line, "description:") {
				return strings.TrimSpace(strings.TrimPrefix(line, "description:"))
			}
		}
	}

	for _, line := range body {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		return trimmed
	}
	return ""
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
					"error": "cannot delete a builtin skill (override it under ~/.fastclaw/skills/ or an agent workspace instead)",
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
//   "user"        – move to ~/.fastclaw/skills/ (shared across agents)
//   "agent:<id>"  – move to ~/.fastclaw/agents/<id>/agent/skills/
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
	// and directories. ~/.fastclaw is on one volume in normal installs.
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
