package setup

import (
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
	Type        string `json:"type"`
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

	// Layer 1: per-agent skills.
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

	if out == nil {
		jsonResponse(w, http.StatusOK, []any{})
		return
	}
	jsonResponse(w, http.StatusOK, out)
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
// of a markdown file – used as a one-line description.
func firstNonHeadingLine(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		return line
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
	// Try every layer + both forms (dir / .md). Stop at the first hit.
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
