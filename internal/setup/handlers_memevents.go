package setup

import (
	"context"
	"net/http"
	"time"

	"github.com/softbreezee/claw-os/internal/config"
	"github.com/softbreezee/claw-os/internal/store/pg"
)

// handleMemoryUsage powers the Memory observability dashboard. It rolls
// up the mcp_events telemetry (written by every `pawnix mcp` subprocess)
// into per-source and per-session views: which tool searched/wrote, how
// many turns, what topics, and whether reads hit.
//
// The setup Server holds no memory DB handle (memory lives in the agent
// daemon's Postgres, addressed by the shared DSN), so — like
// handleRebuildEmbeddings — this opens its own short-lived pool per
// request. EnsureSchema runs first so the page renders even before any
// MCP subprocess has created the table.
//
// When storage isn't Postgres the endpoint returns 200 with
// {available:false} rather than an error, so the frontend can show a
// friendly "telemetry needs Postgres" state instead of a broken page.
func (s *Server) handleMemoryUsage(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.Load()
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}

	if cfg.Storage.Type != "postgres" || cfg.Storage.DSN == "" {
		jsonResponse(w, http.StatusOK, map[string]any{"ok": true, "available": false})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	db, err := pg.Open(ctx, cfg.Storage.DSN)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "connect pg: " + err.Error()})
		return
	}
	defer db.Close()

	events := pg.NewEventStore(db)
	if err := events.EnsureSchema(ctx); err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "ensure schema: " + err.Error()})
		return
	}

	overview, err := events.Overview(ctx, 50, 50)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "usage overview: " + err.Error()})
		return
	}

	jsonResponse(w, http.StatusOK, overview)
}
