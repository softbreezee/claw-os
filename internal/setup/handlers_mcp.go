package setup

import (
	"encoding/json"
	"net/http"

	"github.com/softbreezee/claw-os/internal/config"
)

// --- MCP Servers ---
//
// MCP (Model Context Protocol) servers extend agent toolset with
// external tools. Servers are configured at the global level
// (shared across all agents) and can also be overridden per-agent
// via agent.json. This handler manages the global list; per-agent
// overrides are shown inline on the Agents page.

// mcpServerInfo is the public shape returned by the API.
type mcpServerInfo struct {
	Name    string `json:"name"`
	Type    string `json:"type"`    // "stdio" or "http"
	URL     string `json:"url,omitempty"`
	Command string `json:"command,omitempty"`
	Args    []string `json:"args,omitempty"`
	Status  string `json:"status"`  // "connected" | "disconnected" | "unknown"
	ToolCount int  `json:"toolCount,omitempty"` // -1 = unknown
}

func (s *Server) handleListMCPServers(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.Load()
	if err != nil {
		jsonResponse(w, http.StatusOK, []any{})
		return
	}

	servers := make([]mcpServerInfo, 0, len(cfg.MCPServers))
	for name, sc := range cfg.MCPServers {
		info := mcpServerInfo{
			Name:    name,
			Type:    sc.Type,
			URL:     sc.URL,
			Command: sc.Command,
			Args:    sc.Args,
			Status:  "unknown",
		}
		if info.Type == "" {
			info.Type = "stdio"
		}
		servers = append(servers, info)
	}
	if servers == nil {
		servers = []mcpServerInfo{}
	}
	jsonResponse(w, http.StatusOK, servers)
}

func (s *Server) handleCreateMCPServer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string            `json:"name"`
		Type    string            `json:"type"`    // "stdio" or "http"
		URL     string            `json:"url,omitempty"`
		Command string            `json:"command,omitempty"`
		Args    []string          `json:"args,omitempty"`
		Env     map[string]string `json:"env,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid request"})
		return
	}
	if req.Name == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "name is required"})
		return
	}
	if req.Type == "" {
		req.Type = "stdio"
	}

	cfg, err := config.Load()
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}

	if cfg.MCPServers == nil {
		cfg.MCPServers = make(map[string]config.MCPServerConfig)
	}
	if _, exists := cfg.MCPServers[req.Name]; exists {
		jsonResponse(w, http.StatusConflict, map[string]any{"ok": false, "error": "server with this name already exists"})
		return
	}

	cfg.MCPServers[req.Name] = config.MCPServerConfig{
		Type:    req.Type,
		URL:     req.URL,
		Command: req.Command,
		Args:    req.Args,
		Env:     req.Env,
	}

	if err := saveConfigFile(cfg); err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "save config: " + err.Error()})
		return
	}

	jsonResponse(w, http.StatusCreated, map[string]any{"ok": true, "name": req.Name})
}

func (s *Server) handleUpdateMCPServer(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "name is required"})
		return
	}

	var req struct {
		Type    string            `json:"type,omitempty"`
		URL     string            `json:"url,omitempty"`
		Command string            `json:"command,omitempty"`
		Args    []string          `json:"args,omitempty"`
		Env     map[string]string `json:"env,omitempty"`
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

	sc, ok := cfg.MCPServers[name]
	if !ok {
		jsonResponse(w, http.StatusNotFound, map[string]any{"ok": false, "error": "server not found"})
		return
	}

	if req.Type != "" {
		sc.Type = req.Type
	}
	if req.URL != "" {
		sc.URL = req.URL
	}
	if req.Command != "" {
		sc.Command = req.Command
	}
	if req.Args != nil {
		sc.Args = req.Args
	}
	if req.Env != nil {
		sc.Env = req.Env
	}
	cfg.MCPServers[name] = sc

	if err := saveConfigFile(cfg); err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "save config: " + err.Error()})
		return
	}

	jsonResponse(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleDeleteMCPServer(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "name is required"})
		return
	}

	cfg, err := config.Load()
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}

	if _, ok := cfg.MCPServers[name]; !ok {
		jsonResponse(w, http.StatusNotFound, map[string]any{"ok": false, "error": "server not found"})
		return
	}

	delete(cfg.MCPServers, name)

	if err := saveConfigFile(cfg); err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "save config: " + err.Error()})
		return
	}

	jsonResponse(w, http.StatusOK, map[string]any{"ok": true})
}
