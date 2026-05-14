package setup

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/softbreezee/claw-os/internal/store"
)

// --- Notifications ---
//
// The OS-level inbox. Endpoints back onto s.cronStore (which is also
// the unified Store). When the store isn't wired (setup wizard mode)
// list returns empty + count returns 0; mutations return 503 so the
// UI can show a clean message instead of crashing.

func notificationToJSON(n store.NotificationRecord) map[string]any {
	return map[string]any{
		"id":        n.ID,
		"agentId":   n.AgentID,
		"source":    n.Source,
		"sourceId":  n.SourceID,
		"title":     n.Title,
		"body":      n.Body,
		"link":      n.Link,
		"read":      n.Read,
		"createdAt": n.CreatedAt.Format(time.RFC3339),
	}
}

func (s *Server) handleListNotifications(w http.ResponseWriter, r *http.Request) {
	if s.cronStore == nil {
		jsonResponse(w, http.StatusOK, []any{})
		return
	}
	q := r.URL.Query()
	filters := store.NotificationFilters{
		AgentID:    q.Get("agentId"),
		Source:     q.Get("source"),
		UnreadOnly: q.Get("unreadOnly") == "true" || q.Get("unreadOnly") == "1",
	}
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filters.Limit = n
		}
	}
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filters.Offset = n
		}
	}
	notifs, err := s.cronStore.ListNotifications(r.Context(), s.cronTenantID, filters)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError,
			map[string]any{"ok": false, "error": err.Error()})
		return
	}
	out := make([]map[string]any, 0, len(notifs))
	for _, n := range notifs {
		out = append(out, notificationToJSON(n))
	}
	jsonResponse(w, http.StatusOK, out)
}

// handleUnreadCount is the hot endpoint: the sidebar polls this every
// few seconds to decide whether to render the red dot. Kept separate
// from List to avoid pulling 50 rows just to count.
func (s *Server) handleUnreadNotificationCount(w http.ResponseWriter, r *http.Request) {
	if s.cronStore == nil {
		jsonResponse(w, http.StatusOK, map[string]any{"count": 0})
		return
	}
	n, err := s.cronStore.CountUnreadNotifications(r.Context(), s.cronTenantID)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError,
			map[string]any{"ok": false, "error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]any{"count": n})
}

func (s *Server) handleMarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	if s.cronStore == nil {
		jsonResponse(w, http.StatusServiceUnavailable,
			map[string]any{"ok": false, "error": "notifications store not initialised"})
		return
	}
	id := r.PathValue("id")
	if id == "" {
		jsonResponse(w, http.StatusBadRequest,
			map[string]any{"ok": false, "error": "id is required"})
		return
	}
	var req struct {
		Read *bool `json:"read,omitempty"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	read := true
	if req.Read != nil {
		read = *req.Read
	}
	if err := s.cronStore.MarkNotificationRead(r.Context(), s.cronTenantID, id, read); err != nil {
		jsonResponse(w, http.StatusInternalServerError,
			map[string]any{"ok": false, "error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleMarkAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	if s.cronStore == nil {
		jsonResponse(w, http.StatusServiceUnavailable,
			map[string]any{"ok": false, "error": "notifications store not initialised"})
		return
	}
	if err := s.cronStore.MarkAllNotificationsRead(r.Context(), s.cronTenantID); err != nil {
		jsonResponse(w, http.StatusInternalServerError,
			map[string]any{"ok": false, "error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleDeleteNotification(w http.ResponseWriter, r *http.Request) {
	if s.cronStore == nil {
		jsonResponse(w, http.StatusServiceUnavailable,
			map[string]any{"ok": false, "error": "notifications store not initialised"})
		return
	}
	id := r.PathValue("id")
	if id == "" {
		jsonResponse(w, http.StatusBadRequest,
			map[string]any{"ok": false, "error": "id is required"})
		return
	}
	if err := s.cronStore.DeleteNotification(r.Context(), s.cronTenantID, id); err != nil {
		jsonResponse(w, http.StatusInternalServerError,
			map[string]any{"ok": false, "error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]any{"ok": true})
}
