package setup

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/fastclaw-ai/fastclaw/internal/modelcatalog"
)

// GET /api/model-catalog — return the current catalog (as editable JSON).
func (s *Server) handleGetModelCatalog(w http.ResponseWriter, r *http.Request) {
	cat := modelcatalog.Get()
	jsonResponse(w, http.StatusOK, cat)
}

// PUT /api/model-catalog — save the catalog to disk and update in-memory.
func (s *Server) handleSaveModelCatalog(w http.ResponseWriter, r *http.Request) {
	var cat modelcatalog.Catalog
	if err := json.NewDecoder(r.Body).Decode(&cat); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid catalog JSON"})
		return
	}

	if cat.Models == nil {
		cat.Models = map[string]modelcatalog.ModelInfo{}
	}

	// Merge built-in defaults so new models from upgrades are added.
	modelcatalog.MergeBuiltins(&cat)

	if err := modelcatalog.Save(&cat); err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}

	slog.Info("model catalog saved via API")
	jsonResponse(w, http.StatusOK, map[string]any{"ok": true, "message": "catalog saved"})
}

// POST /api/model-catalog/reload — force reload from disk.
func (s *Server) handleReloadModelCatalog(w http.ResponseWriter, r *http.Request) {
	if err := modelcatalog.Reload(); err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	slog.Info("model catalog reloaded via API")
	jsonResponse(w, http.StatusOK, map[string]any{"ok": true, "message": "catalog reloaded"})
}