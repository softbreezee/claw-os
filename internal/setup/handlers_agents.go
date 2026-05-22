package setup

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/softbreezee/claw-os/internal/config"
	"github.com/softbreezee/claw-os/internal/provider"
	"github.com/softbreezee/claw-os/internal/store/pg"
)

// --- Agent Management ---

// agentWorkspaceFiles is the ordered list of workspace files exposed
// through the /api/agents GET and PUT endpoints. The UI renders them
// as tabs so users can inspect and edit each file individually.
// HISTORY.md is included as read-only (the frontend should disable
// editing for it, and the PUT handler skips it).
var agentWorkspaceFiles = []string{
	"SOUL.md",
	"MEMORY.md",
	"USER.md",
	"AGENTS.md",
	"HISTORY.md",
}

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
		// Read all workspace files into a map so the UI can display and
		// edit each one individually.
		files := map[string]string{}
		for _, name := range agentWorkspaceFiles {
			if data, readErr := os.ReadFile(filepath.Join(ra.Workspace, name)); readErr == nil {
				files[name] = string(data)
			}
		}
		// Keep the legacy top-level `soul` field for backwards compat.
		soul := files["SOUL.md"]
		agents = append(agents, map[string]any{
			"id":                ra.ID,
			"model":             normalizeModel(ra.Model, cfg),
			"embedModel":        ra.EmbedModel,
			"workspace":         ra.Workspace,
			"maxTokens":         ra.MaxTokens,
			"temperature":       ra.Temperature,
			"maxToolIterations": ra.MaxToolIterations,
			"thinking":          ra.Thinking,
			"soul":              soul,
			"files":             files,
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

func (s *Server) handleRebuildEmbeddings(w http.ResponseWriter, r *http.Request) {
	agentID := r.PathValue("id")
	cfg, err := config.Load()
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}

	resolved := config.ResolveAgents(cfg)
	var workspace, embedModel string
	for _, ra := range resolved {
		if ra.ID == agentID {
			workspace = ra.Workspace
			embedModel = ra.EmbedModel
			break
		}
	}
	if workspace == "" {
		jsonResponse(w, http.StatusNotFound, map[string]any{"ok": false, "error": "agent not found"})
		return
	}
	if embedModel == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "embedModel not configured for this agent"})
		return
	}

	memoryPath := filepath.Join(workspace, "MEMORY.md")
	data, err := os.ReadFile(memoryPath)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "read MEMORY.md: " + err.Error()})
		return
	}

	content := strings.TrimSpace(string(data))
	if content == "" {
		jsonResponse(w, http.StatusOK, map[string]any{"ok": true, "facts": 0, "message": "MEMORY.md is empty"})
		return
	}

	var facts []string
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Skip markdown table separators (|---|...)
		stripped := strings.TrimRight(strings.TrimLeft(line, "|"), "|")
		if stripped != "" && isOnly(stripped, '-', ' ', ':') {
			continue
		}
		// Skip lines that are only formatting (list markers, dashes)
		if line == "-" || line == "---" || line == "***" {
			continue
		}
		facts = append(facts, line)
	}

	idx := strings.Index(embedModel, "/")
	if idx < 0 {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "embedModel must be 'provider/model'"})
		return
	}
	providerName := embedModel[:idx]
	provCfg, ok := cfg.Providers[providerName]
	if !ok {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "provider not found: " + providerName})
		return
	}
	prov := provider.NewProvider(provCfg.APIKey, provCfg.APIBase, provCfg.APIType, provCfg.EmbedPath)

	if cfg.Storage.Type != "postgres" || cfg.Storage.DSN == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "storage must be postgres"})
		return
	}
	db, err := pg.Open(context.Background(), cfg.Storage.DSN)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "connect pg: " + err.Error()})
		return
	}
	defer db.Pool.Close()

	memStore := pg.NewMemoryStore(db)
	if _, err := db.Pool.Exec(context.Background(), `DELETE FROM memories WHERE agent_id = $1`, agentID); err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "delete old: " + err.Error()})
		return
	}

	inserted := 0
	for _, fact := range facts {
		emb, e := prov.Embed(context.Background(), fact, embedModel)
		if e != nil {
			continue
		}
		if _, e := memStore.Insert(context.Background(), agentID, "fact", fact, emb, nil); e != nil {
			continue
		}
		inserted++
	}

	jsonResponse(w, http.StatusOK, map[string]any{"ok": true, "facts": len(facts), "inserted": inserted, "message": fmt.Sprintf("Rebuilt: %d/%d facts", inserted, len(facts))})
}

func isOnly(s string, chars ...rune) bool {
	outer:
	for _, r := range s {
		for _, c := range chars {
			if r == c {
				continue outer
			}
		}
		return false
	}
	return true
}

func (s *Server) handleUpdateAgent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Model      string            `json:"model"`
		EmbedModel string            `json:"embedModel"`
		Soul       string            `json:"soul"`
		Files      map[string]string `json:"files"`
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
			cfg.Agents.List[i].EmbedModel = req.EmbedModel
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

	// Update workspace files.
	// Build allowlist for safety — only files in agentWorkspaceFiles may be
	// written, and HISTORY.md is always skipped (it is append-only by the
	// agent runtime).
	allowedFiles := map[string]bool{}
	for _, name := range agentWorkspaceFiles {
		if name != "HISTORY.md" {
			allowedFiles[name] = true
		}
	}

	homeDir, _ := config.HomeDir()
	agentDir := filepath.Join(homeDir, "agents", id, "agent")

	// Write files submitted via the new `files` map.
	for name, content := range req.Files {
		if !allowedFiles[name] {
			continue // silently ignore unknown / protected files
		}
		os.WriteFile(filepath.Join(agentDir, name), []byte(content), 0o644)
	}

	// Legacy single-field paths — still honoured so the Create dialog
	// (which only sends `soul`) keeps working without changes.
	if req.Soul != "" {
		if _, alreadyWritten := req.Files["SOUL.md"]; !alreadyWritten {
			os.WriteFile(filepath.Join(agentDir, "SOUL.md"), []byte(req.Soul), 0o644)
		}
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
