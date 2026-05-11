package setup

// HTTP handlers for the Apps page — a simple registry of external web
// apps the user wants one-click access to (e.g. surge.sh-hosted
// dashboards, internal tools). fastclaw doesn't fetch or proxy these;
// the SPA opens them in a new tab. The whole feature is just shared,
// persistent bookmarks scoped to this fastclaw instance.

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/fastclaw-ai/fastclaw/internal/config"
)

// GET /api/apps – list configured apps.
func (s *Server) handleListApps(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.Load()
	if err != nil {
		jsonResponse(w, http.StatusOK, []any{})
		return
	}
	if cfg.Apps == nil {
		jsonResponse(w, http.StatusOK, []any{})
		return
	}
	jsonResponse(w, http.StatusOK, cfg.Apps)
}

// POST /api/apps – add a new app entry.
//
// Validation:
//   - name and url are required
//   - url must be http(s)://
//   - duplicate names rejected (the UI list keys on name)
func (s *Server) handleCreateApp(w http.ResponseWriter, r *http.Request) {
	var req config.AppEntry
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid request"})
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.URL = strings.TrimSpace(req.URL)
	if req.Name == "" || req.URL == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "name and url are required"})
		return
	}
	if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "url must start with http:// or https://"})
		return
	}

	cfg, err := config.Load()
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	for _, a := range cfg.Apps {
		if a.Name == req.Name {
			jsonResponse(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "an app with this name already exists"})
			return
		}
	}
	cfg.Apps = append(cfg.Apps, req)
	if err := saveConfigFile(cfg); err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]any{"ok": true})
}

// PUT /api/apps/{name} – replace an existing app entry.
//
// The path param is the OLD name (URL-encoded). The body carries the
// new field values, which may include a renamed name.
func (s *Server) handleUpdateApp(w http.ResponseWriter, r *http.Request) {
	oldName := r.PathValue("name")
	var req config.AppEntry
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid request"})
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.URL = strings.TrimSpace(req.URL)
	if req.Name == "" || req.URL == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "name and url are required"})
		return
	}

	cfg, err := config.Load()
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	idx := -1
	for i, a := range cfg.Apps {
		if a.Name == oldName {
			idx = i
			break
		}
	}
	if idx < 0 {
		jsonResponse(w, http.StatusNotFound, map[string]any{"ok": false, "error": "app not found"})
		return
	}
	// Reject collisions when renaming.
	if req.Name != oldName {
		for _, a := range cfg.Apps {
			if a.Name == req.Name {
				jsonResponse(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "an app with this name already exists"})
				return
			}
		}
	}
	cfg.Apps[idx] = req
	if err := saveConfigFile(cfg); err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]any{"ok": true})
}

// DELETE /api/apps/{name} – remove an app entry.
func (s *Server) handleDeleteApp(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	cfg, err := config.Load()
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	out := make([]config.AppEntry, 0, len(cfg.Apps))
	for _, a := range cfg.Apps {
		if a.Name != name {
			out = append(out, a)
		}
	}
	cfg.Apps = out
	if err := saveConfigFile(cfg); err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]any{"ok": true})
}
