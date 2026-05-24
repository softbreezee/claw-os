package setup

import (
	"encoding/json"
	"net/http"

	"github.com/softbreezee/claw-os/internal/config"
)

// handleUpdateSkill updates a skill's config entry (scope etc).
// PUT /api/skills/{name}
func (s *Server) handleUpdateSkill(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	var req struct {
		Agents []string `json:"agents,omitempty"`
		Tags   []string `json:"tags,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid request"})
		return
	}

	cfg, err := config.Load()
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}

	if cfg.Skills.Entries == nil {
		cfg.Skills.Entries = make(map[string]config.SkillEntryCfg)
	}

	entry := cfg.Skills.Entries[name]
	entry.Enabled = true
	if req.Agents != nil {
		entry.Agents = req.Agents
	}
	if req.Tags != nil {
		entry.Tags = req.Tags
	}
	cfg.Skills.Entries[name] = entry

	if err := saveConfigFile(cfg); err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "save: " + err.Error()})
		return
	}

	jsonResponse(w, http.StatusOK, map[string]any{"ok": true})
}
