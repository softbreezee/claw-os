package setup

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fastclaw-ai/fastclaw/internal/config"
)

// --- Agent Management ---

// normalizeModel ensures every model name carries an explicit
// "<provider>/" prefix, matching the multi-provider routing convention.
// Models stored from before the registry refactor (or typed without
// a prefix in the UI) get the prefix injected so list / save round-trip
// produces stable, displayable values.
//
// If the model already contains "/" we trust it. If no providers are
// configured we leave it alone (callers fall back to the bare name).
//
// Provider selection order, matching gateway.buildProviderRegistry:
//   1. Prefix of cfg.Agents.Defaults.Model
//   2. "default" / "openai" / "openrouter" (in that order)
//   3. Lexicographically first registered provider
func normalizeModel(model string, cfg *config.Config) string {
	if model == "" || strings.Contains(model, "/") {
		return model
	}
	if cfg == nil || len(cfg.Providers) == 0 {
		return model
	}

	// Pick provider (mirrors gateway.buildProviderRegistry's default rules)
	chosen := ""
	if def := cfg.Agents.Defaults.Model; strings.Contains(def, "/") {
		candidate := def[:strings.Index(def, "/")]
		if _, ok := cfg.Providers[candidate]; ok {
			chosen = candidate
		}
	}
	if chosen == "" {
		for _, key := range []string{"default", "openai", "openrouter"} {
			if _, ok := cfg.Providers[key]; ok {
				chosen = key
				break
			}
		}
	}
	if chosen == "" {
		names := make([]string, 0, len(cfg.Providers))
		for n := range cfg.Providers {
			names = append(names, n)
		}
		sort.Strings(names)
		if len(names) > 0 {
			chosen = names[0]
		}
	}
	if chosen == "" {
		return model
	}
	return chosen + "/" + model
}

func (s *Server) handleListAgents(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.Load()
	if err != nil {
		jsonResponse(w, http.StatusOK, []any{})
		return
	}
	resolved := config.ResolveAgents(cfg)
	var agents []map[string]any
	for _, ra := range resolved {
		soul := ""
		soulPath := filepath.Join(ra.Workspace, "SOUL.md")
		if data, readErr := os.ReadFile(soulPath); readErr == nil {
			soul = string(data)
		}
		agents = append(agents, map[string]any{
			"id":                ra.ID,
			"model":             normalizeModel(ra.Model, cfg),
			"workspace":         ra.Workspace,
			"maxTokens":         ra.MaxTokens,
			"temperature":       ra.Temperature,
			"maxToolIterations": ra.MaxToolIterations,
			"thinking":          ra.Thinking,
			"soul":              soul,
		})
	}
	if agents == nil {
		agents = []map[string]any{}
	}
	jsonResponse(w, http.StatusOK, agents)
}

func (s *Server) handleCreateAgent(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID    string `json:"id"`
		Model string `json:"model"`
		Soul  string `json:"soul"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid request"})
		return
	}
	if req.ID == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "id is required"})
		return
	}

	cfg, err := config.Load()
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}

	// Normalise into "<provider>/<id>" form so saved config matches
	// what the UI later displays — avoids the "looks different but
	// works" inconsistency old configs exhibited.
	req.Model = normalizeModel(req.Model, cfg)

	// Add agent to config
	cfg.Agents.List = append(cfg.Agents.List, config.AgentEntry{
		ID:    req.ID,
		Model: req.Model,
	})

	if err := saveConfigFile(cfg); err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}

	// Create workspace
	homeDir, _ := config.HomeDir()
	agentDir := filepath.Join(homeDir, "agents", req.ID, "agent")
	for _, dir := range []string{agentDir, filepath.Join(agentDir, "memory"), filepath.Join(agentDir, "sessions"), filepath.Join(agentDir, "skills")} {
		os.MkdirAll(dir, 0o755)
	}
	if req.Soul != "" {
		os.WriteFile(filepath.Join(agentDir, "SOUL.md"), []byte(req.Soul), 0o644)
	}
	agentCfg := config.AgentFileConfig{Model: req.Model}
	agentData, _ := json.MarshalIndent(agentCfg, "", "  ")
	os.WriteFile(filepath.Join(agentDir, "agent.json"), agentData, 0o644)

	jsonResponse(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleUpdateAgent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Model string `json:"model"`
		Soul  string `json:"soul"`
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

	// Normalise BEFORE writing so the next list call sees the
	// prefixed form (and so the file on disk stays consistent
	// with the rest of the multi-provider conventions).
	if req.Model != "" {
		req.Model = normalizeModel(req.Model, cfg)
	}

	found := false
	for i, entry := range cfg.Agents.List {
		if entry.ID == id {
			if req.Model != "" {
				cfg.Agents.List[i].Model = req.Model
			}
			found = true
			break
		}
	}
	if !found {
		jsonResponse(w, http.StatusNotFound, map[string]any{"ok": false, "error": "agent not found"})
		return
	}

	if err := saveConfigFile(cfg); err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}

	// Update workspace files
	homeDir, _ := config.HomeDir()
	agentDir := filepath.Join(homeDir, "agents", id, "agent")
	if req.Soul != "" {
		os.WriteFile(filepath.Join(agentDir, "SOUL.md"), []byte(req.Soul), 0o644)
	}
	if req.Model != "" {
		agentCfg := config.AgentFileConfig{Model: req.Model}
		agentData, _ := json.MarshalIndent(agentCfg, "", "  ")
		os.WriteFile(filepath.Join(agentDir, "agent.json"), agentData, 0o644)
	}

	jsonResponse(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleDeleteAgent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	cfg, err := config.Load()
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}

	newList := make([]config.AgentEntry, 0, len(cfg.Agents.List))
	for _, entry := range cfg.Agents.List {
		if entry.ID != id {
			newList = append(newList, entry)
		}
	}
	cfg.Agents.List = newList

	if err := saveConfigFile(cfg); err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}

	jsonResponse(w, http.StatusOK, map[string]any{"ok": true})
}
