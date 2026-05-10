package setup

// HTTP handlers for serving local files into the chat UI.
//
// Two needs share this code:
//
//   1. Render images the user just uploaded (live in ~/.fastclaw/uploads).
//   2. Render images / documents the agent itself produced inside its
//      workspace directory.
//
// Browsers can't `<img src="file:///…">` for security reasons, so the
// only way to surface local content in the SPA is to proxy it through
// an HTTP endpoint. The same endpoint also serves the "click to open
// in Finder" affordance via a sibling POST handler.

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fastclaw-ai/fastclaw/internal/config"
)

// GET /api/files?kind=upload&path=<sha>.png
// GET /api/files?kind=workspace&agentId=<id>&path=relative/under/workspace
//
// Path-traversal hardening:
//   - We compute the absolute path then verify it remains under the
//     allowed root. If `path` contains "../foo" that escapes the root,
//     we reject with 403.
//   - The endpoint refuses absolute paths in the query so callers can't
//     short-circuit the root check.
func (s *Server) handleServeFile(w http.ResponseWriter, r *http.Request) {
	kind := r.URL.Query().Get("kind")
	rel := r.URL.Query().Get("path")
	if rel == "" {
		http.Error(w, "missing path", http.StatusBadRequest)
		return
	}
	if filepath.IsAbs(rel) {
		http.Error(w, "absolute paths not allowed", http.StatusBadRequest)
		return
	}

	var root string
	switch kind {
	case "upload":
		homeDir, err := config.HomeDir()
		if err != nil {
			http.Error(w, "home dir unavailable", http.StatusInternalServerError)
			return
		}
		root = filepath.Join(homeDir, "uploads")
	case "workspace":
		agentID := r.URL.Query().Get("agentId")
		if agentID == "" || s.agentProvider == nil {
			http.Error(w, "agentId required", http.StatusBadRequest)
			return
		}
		ag := s.agentProvider.AgentByID(agentID)
		if ag == nil {
			http.Error(w, "agent not found", http.StatusNotFound)
			return
		}
		root = ag.WorkspacePath()
		if root == "" {
			http.Error(w, "agent has no workspace", http.StatusNotFound)
			return
		}
	default:
		http.Error(w, "kind must be 'upload' or 'workspace'", http.StatusBadRequest)
		return
	}

	rootAbs, err := filepath.Abs(root)
	if err != nil {
		http.Error(w, "bad root", http.StatusInternalServerError)
		return
	}
	full := filepath.Join(rootAbs, rel)
	fullClean := filepath.Clean(full)
	// `fullClean == rootAbs` allows directory listing via 404; we only
	// permit serving regular files inside the root.
	if !strings.HasPrefix(fullClean, rootAbs+string(filepath.Separator)) {
		http.Error(w, "path escapes root", http.StatusForbidden)
		return
	}

	// http.ServeFile takes care of Content-Type sniffing, Range requests,
	// 304 If-Modified-Since, etc. If the path is a directory it would
	// emit a redirect to add a trailing slash and then 403/listing —
	// we don't want that, so reject explicitly.
	info, err := os.Stat(fullClean)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if info.IsDir() {
		http.Error(w, "is a directory", http.StatusForbidden)
		return
	}

	// Allow the chat UI to inline these (helps with ~25 MiB images
	// the user re-views while scrolling).
	w.Header().Set("Cache-Control", "private, max-age=300")
	http.ServeFile(w, r, fullClean)
}

// POST /api/workspace/open?agentId=<id>
//
// Pops open the agent's workspace directory in the host OS file
// explorer. macOS uses `open`, Linux uses `xdg-open`, Windows uses
// `explorer`. The handler returns immediately; we don't wait for the
// child process because the user only cares that the GUI launched.
//
// Why a server-side handler at all? Web clients can't shell out, and
// `file://` URLs are blocked by every modern browser. The user is
// already trusting fastclaw to run arbitrary tools, so launching a
// file explorer on the same host is well within scope.
func (s *Server) handleOpenWorkspace(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agentId")
	if agentID == "" || s.agentProvider == nil {
		http.Error(w, "agentId required", http.StatusBadRequest)
		return
	}
	ag := s.agentProvider.AgentByID(agentID)
	if ag == nil {
		http.Error(w, "agent not found", http.StatusNotFound)
		return
	}
	path := ag.WorkspacePath()
	if path == "" {
		http.Error(w, "agent has no workspace", http.StatusNotFound)
		return
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", path)
	default:
		// Linux + the BSDs all use freedesktop.org's xdg-open.
		cmd = exec.Command("xdg-open", path)
	}

	if err := cmd.Start(); err != nil {
		slog.Warn("open workspace failed", "agent", agentID, "path", path, "error", err)
		jsonResponse(w, http.StatusInternalServerError, map[string]any{
			"ok":    false,
			"error": fmt.Sprintf("failed to open: %v", err),
		})
		return
	}
	// Don't Wait — let the GUI app live its own life. Reaping is fine
	// because the Start() call already cleaned up the launcher's pipes.
	go cmd.Wait()

	jsonResponse(w, http.StatusOK, map[string]any{
		"ok":   true,
		"path": path,
	})
}
