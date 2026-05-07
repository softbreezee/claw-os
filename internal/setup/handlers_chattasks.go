package setup

// HTTP handlers for the async web chat task subsystem.

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/fastclaw-ai/fastclaw/internal/eventbus"
	"github.com/fastclaw-ai/fastclaw/internal/store"
	"github.com/fastclaw-ai/fastclaw/internal/taskrunner"
)

// chatSubmitRequest is the wire format of POST /api/chat/submit.
type chatSubmitRequest struct {
	AgentID   string `json:"agentId"`
	SessionID string `json:"sessionId"`
	Message   string `json:"message"`
}

// chatTaskWired returns true iff the async runtime is fully configured.
// Handlers short-circuit with 503 when not wired.
func (s *Server) chatTaskWired() bool {
	return s.chatRunner != nil && s.eventBus != nil && s.taskStore != nil
}

// POST /api/chat/submit – enqueue a new task, return its ID.
func (s *Server) handleChatSubmit(w http.ResponseWriter, r *http.Request) {
	if !s.chatTaskWired() {
		jsonResponse(w, http.StatusServiceUnavailable, map[string]any{"error": "chat task subsystem not initialised"})
		return
	}

	var req chatSubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"error": "invalid request"})
		return
	}
	if req.AgentID == "" || req.SessionID == "" || req.Message == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"error": "agentId, sessionId, message required"})
		return
	}

	taskID, err := s.chatRunner.Submit(r.Context(), req.AgentID, req.SessionID, req.Message)
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]any{
		"taskId": taskID,
		"status": string(store.ChatTaskPending),
	})
}

// GET /api/chat/tasks/{id}/events?after=N – SSE subscription to a task's
// event stream. The connection stays open until either:
//   * the client disconnects (handled by ctx cancel),
//   * a terminal event (task_done / task_error / task_cancelled) is received, or
//   * the EventBus closes the topic (process shutdown).
//
// PR3: ?after=N replays buffered events with seq > N before subscribing
// to live updates, enabling resumable streams across reconnects.
func (s *Server) handleChatTaskEvents(w http.ResponseWriter, r *http.Request) {
	if !s.chatTaskWired() {
		http.Error(w, "chat task subsystem not initialised", http.StatusServiceUnavailable)
		return
	}
	taskID := r.PathValue("id")
	if taskID == "" {
		http.Error(w, "missing task id", http.StatusBadRequest)
		return
	}

	// Reject early if the task doesn't exist – avoids leaving a phantom
	// SSE connection hanging until the client times out.
	rec, err := s.taskStore.GetChatTask(r.Context(), s.tenantID, taskID)
	if err != nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Parse resume cursor. Invalid/missing → 0 (replay everything in buffer).
	var afterSeq int64
	if v := r.URL.Query().Get("after"); v != "" {
		if n, perr := strconv.ParseInt(v, 10, 64); perr == nil && n >= 0 {
			afterSeq = n
		}
	}

	// CRITICAL ORDERING (avoid both gaps and duplicates):
	//   1. Subscribe to live first (locks in our position in the bus).
	//   2. Snapshot the buffered history.
	//   3. Replay buffered events with seq > afterSeq.
	//   4. Drain live events, but skip any whose seq <= the highest
	//      seq we replayed (because the bus may emit events that were
	//      already in the snapshot).
	ch, cancelSub := s.eventBus.Subscribe(taskrunner.TopicFor(taskID))
	defer cancelSub()

	buffered := s.chatRunner.EventsAfter(taskID, afterSeq)
	var maxReplayedSeq int64 = afterSeq
	for _, evt := range buffered {
		writeSSEEvent(w, flusher, evt)
		if evt.Seq > maxReplayedSeq {
			maxReplayedSeq = evt.Seq
		}
		if isTerminalEventType(evt.Type) {
			return
		}
	}

	// If the persisted record is already terminal AND we've replayed
	// everything in the buffer, no live events will arrive – close out
	// with a snapshot if the buffer didn't include the terminal event.
	if isTerminalChatTaskStatus(rec.Status) && len(buffered) == 0 {
		writeSSEEvent(w, flusher, snapshotEvent(rec))
		return
	}

	for {
		select {
		case <-r.Context().Done():
			return
		case evt, ok := <-ch:
			if !ok {
				return
			}
			// De-dup: bus may deliver an event we already replayed
			// from history. publishEvent always assigns Seq>=1, so a
			// straight comparison is enough.
			if evt.Seq <= maxReplayedSeq {
				continue
			}
			writeSSEEvent(w, flusher, evt)
			if isTerminalEventType(evt.Type) {
				return
			}
		}
	}
}

// POST /api/chat/tasks/{id}/cancel
func (s *Server) handleChatTaskCancel(w http.ResponseWriter, r *http.Request) {
	if !s.chatTaskWired() {
		jsonResponse(w, http.StatusServiceUnavailable, map[string]any{"error": "chat task subsystem not initialised"})
		return
	}
	taskID := r.PathValue("id")
	if taskID == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"error": "missing task id"})
		return
	}
	if err := s.chatRunner.Cancel(r.Context(), taskID); err != nil {
		jsonResponse(w, http.StatusNotFound, map[string]any{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]any{
		"ok":     true,
		"taskId": taskID,
		"status": string(store.ChatTaskCancelled),
	})
}

// GET /api/chat/tasks/{id}
func (s *Server) handleGetChatTask(w http.ResponseWriter, r *http.Request) {
	if !s.chatTaskWired() {
		jsonResponse(w, http.StatusServiceUnavailable, map[string]any{"error": "chat task subsystem not initialised"})
		return
	}
	taskID := r.PathValue("id")
	if taskID == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"error": "missing task id"})
		return
	}
	rec, err := s.taskStore.GetChatTask(r.Context(), s.tenantID, taskID)
	if err != nil {
		jsonResponse(w, http.StatusNotFound, map[string]any{"error": "task not found"})
		return
	}
	jsonResponse(w, http.StatusOK, rec)
}

// GET /api/chat/tasks?agentId=...&sessionKey=...&status=...&limit=N&offset=N
func (s *Server) handleListChatTasks(w http.ResponseWriter, r *http.Request) {
	if !s.chatTaskWired() {
		jsonResponse(w, http.StatusServiceUnavailable, map[string]any{"error": "chat task subsystem not initialised"})
		return
	}
	q := r.URL.Query()
	filters := store.ChatTaskFilters{
		AgentID:    q.Get("agentId"),
		SessionKey: q.Get("sessionKey"),
		Status:     store.ChatTaskStatus(q.Get("status")),
	}
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			filters.Limit = n
		}
	}
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			filters.Offset = n
		}
	}

	tasks, err := s.taskStore.ListChatTasks(r.Context(), s.tenantID, filters)
	if err != nil {
		// File backend doesn't support listing – communicate that clearly
		// rather than returning an opaque 500.
		if errors.Is(err, store.ErrNotSupported) {
			jsonResponse(w, http.StatusNotImplemented, map[string]any{
				"error": "list not supported by file storage backend; switch to database storage",
			})
			return
		}
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	if tasks == nil {
		tasks = []store.ChatTaskRecord{}
	}
	jsonResponse(w, http.StatusOK, map[string]any{
		"tasks": tasks,
		"total": len(tasks),
	})
}

// --- helpers ---

func isTerminalChatTaskStatus(s store.ChatTaskStatus) bool {
	switch s {
	case store.ChatTaskDone, store.ChatTaskFailed, store.ChatTaskCancelled:
		return true
	}
	return false
}

func isTerminalEventType(t string) bool {
	switch t {
	case "task_done", "task_error", "task_cancelled":
		return true
	}
	return false
}

// snapshotEvent fabricates a single terminal event from a persisted task,
// used when a client subscribes after the task has already finished.
func snapshotEvent(rec *store.ChatTaskRecord) eventbus.Event {
	switch rec.Status {
	case store.ChatTaskDone:
		return eventbus.Event{
			Type: "task_done",
			Data: map[string]any{"taskId": rec.ID, "result": rec.Result},
		}
	case store.ChatTaskFailed:
		return eventbus.Event{
			Type: "task_error",
			Data: map[string]any{"taskId": rec.ID, "error": rec.Error},
		}
	case store.ChatTaskCancelled:
		return eventbus.Event{
			Type: "task_cancelled",
			Data: map[string]any{"taskId": rec.ID},
		}
	default:
		return eventbus.Event{
			Type: "task_running",
			Data: map[string]any{"taskId": rec.ID},
		}
	}
}

// writeSSEEvent emits one event in the standard SSE wire format
// (single "data:" line per event, terminated by an empty line).
func writeSSEEvent(w http.ResponseWriter, flusher http.Flusher, evt eventbus.Event) {
	data, err := json.Marshal(evt)
	if err != nil {
		slog.Warn("chat task events: marshal failed", "err", err)
		return
	}
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
}
