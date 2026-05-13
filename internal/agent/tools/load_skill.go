package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type loadSkillArgs struct {
	Name string `json:"name"`
}

// RegisterLoadSkill registers the load_skill tool that reads full SKILL.md content.
//
// agentID enables search of the per-agent scoped directory at
// ~/.pawnix/agents/<agentID>/agent/skills/, which is where the
// Web UI's "move skill to agent" flow places skill files. Pass ""
// to skip that lookup.
//
// builtinDir is the project-shipped skills/ directory (typically
// computed via agent.builtinSkillsDir()). Pass "" to skip — but the
// resulting tool then can't load shipped skills like docx/pdf, which
// is almost certainly not what you want.
func RegisterLoadSkill(r *Registry, homeDir, agentDir, teamDir, agentID, builtinDir string) {
	r.Register("load_skill", "Load the full content of a skill by name. Use this when you need detailed instructions for a specific skill.", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "The skill name to load",
			},
		},
		"required": []string{"name"},
	}, makeLoadSkill(homeDir, agentDir, teamDir, agentID, builtinDir))
}

func makeLoadSkill(homeDir, agentDir, teamDir, agentID, builtinDir string) ToolFunc {
	// Directories to search in priority order — must mirror the
	// SkillsLoader's layer order so the tool resolves the same skill
	// the system prompt advertised.
	searchDirs := []string{
		filepath.Join(agentDir, "skills"),
	}
	if agentID != "" {
		searchDirs = append(searchDirs, filepath.Join(homeDir, "agents", agentID, "agent", "skills"))
	}
	if teamDir != "" {
		searchDirs = append(searchDirs, filepath.Join(teamDir, "skills"))
	}
	searchDirs = append(searchDirs, filepath.Join(homeDir, "skills"))
	if builtinDir != "" {
		searchDirs = append(searchDirs, builtinDir)
	}

	return func(ctx context.Context, rawArgs json.RawMessage) (string, error) {
		var args loadSkillArgs
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return "", fmt.Errorf("parse args: %w", err)
		}

		if args.Name == "" {
			return "", fmt.Errorf("skill name is required")
		}

		// Search through directories in priority order. A skill name
		// can refer to either:
		//   * <dir>/<name>/SKILL.md   — canonical multi-file layout
		//   * <dir>/<name>.md         — single-file skill
		// Both forms must be supported because the Web UI's Skills page
		// already lists single-file skills, and users (reasonably)
		// expect to be able to load_skill them by name.
		for _, dir := range searchDirs {
			candidates := []string{
				filepath.Join(dir, args.Name, "SKILL.md"),
				filepath.Join(dir, args.Name+".md"),
			}
			for _, p := range candidates {
				data, err := os.ReadFile(p)
				if err == nil {
					return string(data), nil
				}
			}
		}

		return "", fmt.Errorf("skill %q not found", args.Name)
	}
}
