package setup

// HTTP handlers for the async web chat task subsystem.

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/softbreezee/claw-os/internal/bus"
	"github.com/softbreezee/claw-os/internal/eventbus"
	"github.com/softbreezee/claw-os/internal/store"
	"github.com/softbreezee/claw-os/internal/taskrunner"
)

// maxAttachmentBytes caps a single uploaded file. 25 MiB is comfortably
// above the typical mobile photo size (5–8 MiB) but well under the
// inline-base64 limits of OpenAI/Anthropic (which are roughly 20 MiB
// per image after base64 expansion). When we add a config block for
// attachments this becomes a settable knob.
const maxAttachmentBytes = 25 * 1024 * 1024

// maxMultipartMemory bounds in-memory buffering during ParseMultipartForm.
// Anything larger spills to /tmp; 32 MiB is a sane Go default.
const maxMultipartMemory = 32 * 1024 * 1024

// chatSubmitRequest is the wire format of POST /api/chat/submit.
//
// Model is optional ("" = use the agent's configured default). When set,
// it overrides the agent's model for this single message only — the
// agent's persistent config is unchanged. The override is read by
// agent.effectiveModel at every primary LLM call site, so it covers
// both the ReAct loop and the final streaming response.
//
// DeliverToChannel / DeliverToChatID specify an additional delivery target
// for the agent's reply. When non-empty, the task runner sends the reply to
// this channel/chatID after the web stream completes. Used for cross-channel
// mirroring (e.g. reply in Web UI → also appear in Discord).
type chatSubmitRequest struct {
	AgentID          string `json:"agentId"`
	SessionID        string `json:"sessionId"`
	Message          string `json:"message"`
	Model            string `json:"model,omitempty"`
	DeliverToChannel string `json:"deliverToChannel,omitempty"`
	DeliverToChatID  string `json:"deliverToChatId,omitempty"`
}

// chatTaskWired returns true iff the async runtime is fully configured.
// Handlers short-circuit with 503 when not wired.
func (s *Server) chatTaskWired() bool {
	return s.chatRunner != nil && s.eventBus != nil && s.taskStore != nil
}

// POST /api/chat/submit – enqueue a new task, return its ID.
//
// Accepts two content types:
//   - application/json (legacy text-only path)
//   - multipart/form-data (text + file attachments)
//
// The multipart form fields are: agentId, sessionId, message, model,
// and one or more "files". Files are saved into the upload store
// before being attached to the task.
func (s *Server) handleChatSubmit(w http.ResponseWriter, r *http.Request) {
	if !s.chatTaskWired() {
		jsonResponse(w, http.StatusServiceUnavailable, map[string]any{"error": "chat task subsystem not initialised"})
		return
	}

	var (
		req         chatSubmitRequest
		attachments []bus.Attachment
	)

	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "multipart/form-data") {
		if err := r.ParseMultipartForm(maxMultipartMemory); err != nil {
			jsonResponse(w, http.StatusBadRequest, map[string]any{"error": "invalid multipart form: " + err.Error()})
			return
		}
		req.AgentID = r.FormValue("agentId")
		req.SessionID = r.FormValue("sessionId")
		req.Message = r.FormValue("message")
		req.Model = r.FormValue("model")

		// Save each uploaded file to the content-addressed store.
		// Failures are reported as 4xx (rather than silently skipped)
		// so the user sees what happened to their attachment.
		if r.MultipartForm != nil {
			files := r.MultipartForm.File["files"]
			if len(files) > 0 {
				atts, err := s.persistMultipartFiles(files)
				if err != nil {
					jsonResponse(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
					return
				}
				attachments = atts
			}
		}
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonResponse(w, http.StatusBadRequest, map[string]any{"error": "invalid request"})
			return
		}
	}

	// Allow message to be empty when there's at least one attachment —
	// "[user uploaded an image without a caption]" is a valid prompt.
	if req.AgentID == "" || req.SessionID == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"error": "agentId and sessionId required"})
		return
	}
	if req.Message == "" && len(attachments) == 0 {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"error": "message or attachment required"})
		return
	}

	taskID, err := s.chatRunner.SubmitWithOptions(r.Context(), req.AgentID, req.SessionID, req.Message, taskrunner.SubmitOptions{
		ModelOverride:    req.Model,
		Attachments:      attachments,
		DeliverToChannel: req.DeliverToChannel,
		DeliverToChatID:  req.DeliverToChatID,
	})
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]any{
		"taskId": taskID,
		"status": string(store.ChatTaskPending),
	})
}

// persistMultipartFiles streams each uploaded file into the upload
// store and returns the corresponding bus.Attachment slice. Any file
// over maxAttachmentBytes aborts the whole batch (we'd rather fail
// loudly than partially attach).
func (s *Server) persistMultipartFiles(headers []*multipartFileHeader) ([]bus.Attachment, error) {
	store, err := s.uploadStore()
	if err != nil {
		return nil, fmt.Errorf("upload store unavailable: %w", err)
	}

	out := make([]bus.Attachment, 0, len(headers))
	for _, fh := range headers {
		if fh.Size > maxAttachmentBytes {
			return nil, fmt.Errorf("attachment %q is %.1f MiB, exceeds limit of %d MiB",
				fh.Filename, float64(fh.Size)/1024/1024, maxAttachmentBytes/1024/1024)
		}
		f, err := fh.Open()
		if err != nil {
			return nil, fmt.Errorf("open %s: %w", fh.Filename, err)
		}
		mime := fh.Header.Get("Content-Type")
		path, err := store.Save(f, mime, fh.Filename)
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("save %s: %w", fh.Filename, err)
		}
		out = append(out, bus.Attachment{
			Path:     path,
			MimeType: mime,
			Name:     fh.Filename,
			Size:     fh.Size,
		})
	}
	return out, nil
}

// multipartFileHeader is just an alias to keep the helper signature
// independent of net/textproto's pointer-heavy nesting in tests.
type multipartFileHeader = multipart.FileHeader

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

// POST /api/chat/tasks/{id}/steer
//
// Body: {"text": "..."}
//
// Folds a mid-run user instruction into the in-flight task. The
// agent loop drains the steer buffer at the next ReAct iteration
// boundary so the message lands at the next tool-call decision
// point — never mid-tool. Returns 409 when the target task is no
// longer running (terminal or pending), so the client can fall back
// to a regular /submit. See docs/v0.3-plan.md § Week 2.
func (s *Server) handleChatTaskSteer(w http.ResponseWriter, r *http.Request) {
	if !s.chatTaskWired() {
		jsonResponse(w, http.StatusServiceUnavailable, map[string]any{"error": "chat task subsystem not initialised"})
		return
	}
	taskID := r.PathValue("id")
	if taskID == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"error": "missing task id"})
		return
	}
	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"error": "invalid body: " + err.Error()})
		return
	}
	text := strings.TrimSpace(req.Text)
	if text == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"error": "text is required"})
		return
	}
	if err := s.chatRunner.Steer(taskID, text); err != nil {
		if errors.Is(err, taskrunner.ErrTaskNotRunning) {
			jsonResponse(w, http.StatusConflict, map[string]any{
				"error": "task is not running; submit a new message instead",
			})
			return
		}
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]any{
		"ok":     true,
		"taskId": taskID,
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
