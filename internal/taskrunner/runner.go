// Package taskrunner runs web chat messages as asynchronous tasks.
//
// Web chat tasks differ from IM tasks (internal/taskqueue):
//
//   * They are HTTP-driven but their lifetime exceeds the HTTP request:
//     a task keeps running even if the SSE client disconnects, and a
//     reconnecting client can resume by re-subscribing to the EventBus
//     topic for that task.
//
//   * They have an independent context that can be cancelled by an
//     explicit POST /api/chat/tasks/:id/cancel call. The HTTP request
//     context is not used for cancellation – disconnect != cancel.
//
//   * Per-(agent, session) FIFO serialisation keeps ordering and avoids
//     concurrent mutation of session history.
//
// The Runner emits two kinds of events on the bus topic "task:{taskID}":
//   - Lifecycle: task_pending, task_running, task_cancelled, task_done,
//     task_error (these are also persisted to the chat_tasks table)
//   - Streaming: content, tool_call, tool_result, done (forwarded
//     verbatim from agent.HandleWebChatStream's events channel)
package taskrunner

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/softbreezee/claw-os/internal/agent"
	// busmsg alias avoids shadowing in `func New(.. bus eventbus.Bus ..)`
	// where `bus` is a parameter name, not the package.
	busmsg "github.com/softbreezee/claw-os/internal/bus"
	"github.com/softbreezee/claw-os/internal/eventbus"
	"github.com/softbreezee/claw-os/internal/store"
)

// AgentResolver finds an agent by ID. The taskrunner package does not
// depend on internal/agent.Manager directly – callers inject any type
// that satisfies this minimal contract, which avoids an import cycle
// (Manager already imports many internal/* packages).
type AgentResolver interface {
	AgentByID(id string) AgentHandle
}

// AgentHandle is the slice of *agent.Agent that the runner needs.
// HandleWebChatStream blocks until the conversation turn is complete
// and returns the final assistant content; events are streamed via the
// channel argument while it runs.
type AgentHandle interface {
	HandleWebChatStream(ctx context.Context, sessionID, text string, events chan<- agent.ChatEvent) string
	SendToChat(ctx context.Context, channel, chatID, text string) error
}

// Runner is the package's main type. Construct with New, drive with
// Submit, shut down cleanly with Stop.
type Runner struct {
	store    store.Store
	bus      eventbus.Bus
	resolver AgentResolver
	tenantID string
	timeout  time.Duration

	mu       sync.Mutex
	queues   map[string]*sessionQueue // sessionKey(agentID:sessionID) -> queue
	inflight map[string]*runningTask  // taskID -> currently executing task
	seq      uint64

	// pendingModels caches per-task model overrides between Submit and run.
	// We don't persist this on ChatTaskRecord — chat tasks aren't resumed
	// across process restarts, so an in-memory map is enough and keeps the
	// store schema unchanged.
	pendingModels map[string]string // taskID -> model

	// pendingAttachments is the same idea as pendingModels: per-task
	// per-process state that doesn't need to outlive a restart. The
	// attachment files themselves live in internal/upload's directory
	// and survive restarts; this map is just the (taskID -> file paths)
	// binding that the runner pops on its way to executing the agent.
	pendingAttachments map[string][]busmsg.Attachment
	pendingDeliveries  map[string]deliveryTarget // taskID -> cross-channel delivery
	// PR3: per-task ring buffer of recent events, keyed by taskID. Used
	// by handlers to replay events to clients that missed them (e.g.
	// after a network blip). historySize caps memory per task.
	historyMu   sync.RWMutex
	history     map[string]*eventHistory
	historySize int

	rootCtx    context.Context
	rootCancel context.CancelFunc
}

// eventHistory is a bounded ring buffer of recent events for one task.
// We could use a slice + pruning but a ring is O(1) append and easier
// to reason about under concurrent access.
type eventHistory struct {
	events  []eventbus.Event // oldest at index 0 after wrap
	nextSeq int64            // next seq to assign (1-based, never reused)
}

// runningTask holds the cancel func of a task currently being executed,
// so POST /api/chat/tasks/:id/cancel can interrupt it. Pending tasks
// (waiting in the per-session queue) don't need this; they're cancelled
// by setting their status before the worker picks them up.
type runningTask struct {
	cancel context.CancelFunc
}

// Options tweaks Runner behaviour. Zero values are sensible defaults.
type Options struct {
	// TenantID for all tasks the runner creates. Default "default".
	TenantID string
	// Timeout caps a single task's execution. Default 5 minutes (matches
	// existing internal/taskqueue.Queue and avoids lingering ghost tasks).
	Timeout time.Duration
}

// New constructs a Runner. The caller must call Stop when shutting down.
func New(s store.Store, bus eventbus.Bus, resolver AgentResolver, opts Options) *Runner {
	if opts.TenantID == "" {
		opts.TenantID = store.DefaultTenantID
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 5 * time.Minute
	}
	rootCtx, cancel := context.WithCancel(context.Background())
	return &Runner{
		store:              s,
		bus:                bus,
		resolver:           resolver,
		tenantID:           opts.TenantID,
		timeout:            opts.Timeout,
		queues:             make(map[string]*sessionQueue),
		inflight:           make(map[string]*runningTask),
		pendingModels:      make(map[string]string),
		pendingAttachments: make(map[string][]busmsg.Attachment),
		pendingDeliveries:  make(map[string]deliveryTarget),
		history:            make(map[string]*eventHistory),
		historySize:        256, // ~256 events covers a full agent turn comfortably
		rootCtx:            rootCtx,
		rootCancel:         cancel,
	}
}

// publishEvent assigns a monotonic per-task sequence, appends to the
// history ring, then broadcasts on the bus. ALL task event publishes
// must go through this function (do not call r.bus.Publish directly)
// so that resumable subscribers can replay reliably.
func (r *Runner) publishEvent(ctx context.Context, taskID string, evt eventbus.Event) {
	r.historyMu.Lock()
	h, ok := r.history[taskID]
	if !ok {
		h = &eventHistory{nextSeq: 1}
		r.history[taskID] = h
	}
	evt.Seq = h.nextSeq
	h.nextSeq++
	if len(h.events) >= r.historySize {
		// Drop the oldest event. This means very-late subscribers may
		// miss the start of the task, but they get a snapshot from the
		// store (terminal-state replay in the handler) so we never lose
		// "what happened" – only intermediate streaming chunks.
		h.events = append(h.events[1:], evt)
	} else {
		h.events = append(h.events, evt)
	}
	r.historyMu.Unlock()

	r.bus.Publish(ctx, TopicFor(taskID), evt)
}

// EventsAfter returns buffered events for a task with Seq strictly
// greater than after. Returns nil if no buffered events match (caller
// should still subscribe for live updates).
func (r *Runner) EventsAfter(taskID string, after int64) []eventbus.Event {
	r.historyMu.RLock()
	defer r.historyMu.RUnlock()
	h, ok := r.history[taskID]
	if !ok {
		return nil
	}
	out := make([]eventbus.Event, 0, len(h.events))
	for _, e := range h.events {
		if e.Seq > after {
			out = append(out, e)
		}
	}
	return out
}

// dropHistory releases the per-task event buffer. Called automatically
// after task completion + a grace period; can also be called explicitly
// when the chat task record is deleted.
func (r *Runner) dropHistory(taskID string) {
	r.historyMu.Lock()
	delete(r.history, taskID)
	r.historyMu.Unlock()
}

// Stop signals all per-session worker goroutines to exit and cancels
// any in-flight tasks. Safe to call multiple times.
func (r *Runner) Stop() {
	r.rootCancel()
}

// TopicFor returns the eventbus topic name used for a given task.
// Exposed so HTTP handlers can subscribe to the right topic.
func TopicFor(taskID string) string {
	return "task:" + taskID
}

// Submit enqueues a new chat task and returns its ID. The task is
// processed asynchronously – Submit never blocks waiting for the agent.
//
// agentID and sessionID identify the agent and session; message is the
// user's input. The returned taskID can be used to subscribe via
// eventbus, query state via store, or cancel via Cancel.
// Submit enqueues a new chat task and returns its ID. The task is
// processed asynchronously – Submit never blocks waiting for the agent.
//
// modelOverride is optional ("" = use the agent's configured default).
// When non-empty, the override is attached to the task ctx and read by
// agent.effectiveModel at every primary LLM call site, so the user can
// pick a different model per message without editing the agent.
// SubmitOptions carries optional knobs for Submit. Required scalars
// (agentID, sessionID, message) stay positional for callers that don't
// need the extras; this struct is only used when something beyond the
// vanilla text-only path is in play.
type deliveryTarget struct {
	channel string
	chatID  string
}

type SubmitOptions struct {
	ModelOverride     string
	Attachments       []busmsg.Attachment
	DeliverToChannel  string // cross-channel reply delivery target
	DeliverToChatID   string
}

func (r *Runner) Submit(ctx context.Context, agentID, sessionID, message, modelOverride string) (string, error) {
	return r.SubmitWithOptions(ctx, agentID, sessionID, message, SubmitOptions{ModelOverride: modelOverride})
}

// SubmitWithOptions is the explicit form of Submit. Use it whenever
// you need to attach files (or any other future per-task knob) — the
// shorter Submit stays for backward compatibility with existing tests
// and call sites.
func (r *Runner) SubmitWithOptions(ctx context.Context, agentID, sessionID, message string, opts SubmitOptions) (string, error) {
	if r.resolver.AgentByID(agentID) == nil {
		return "", fmt.Errorf("unknown agent: %s", agentID)
	}

	r.mu.Lock()
	r.seq++
	taskID := fmt.Sprintf("ct-%d-%d", time.Now().UnixMilli(), r.seq)
	if opts.ModelOverride != "" {
		r.pendingModels[taskID] = opts.ModelOverride
	}
	if len(opts.Attachments) > 0 {
		r.pendingAttachments[taskID] = opts.Attachments
	}
	if opts.DeliverToChannel != "" {
		r.pendingDeliveries[taskID] = deliveryTarget{
			channel: opts.DeliverToChannel,
			chatID:  opts.DeliverToChatID,
		}
	}
	r.mu.Unlock()

	now := time.Now().UTC()
	rec := &store.ChatTaskRecord{
		ID:         taskID,
		TenantID:   r.tenantID,
		AgentID:    agentID,
		SessionKey: sessionID,
		Status:     store.ChatTaskPending,
		Message:    message,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := r.store.CreateChatTask(ctx, r.tenantID, rec); err != nil {
		return "", fmt.Errorf("persist task: %w", err)
	}

	// Pending event for any subscriber that's already listening (rare but
	// possible when callers Subscribe immediately after Submit).
	r.publishEvent(ctx, taskID, eventbus.Event{
		Type:      "task_pending",
		Data:      map[string]any{"taskId": taskID},
		Timestamp: now,
	})

	r.enqueue(rec)
	return taskID, nil
}

// Cancel attempts to cancel a task. If the task is currently running,
// its context is cancelled (which propagates through to in-flight LLM
// calls and exec subprocesses via exec.CommandContext). If pending in
// the queue, it'll be skipped when the worker picks it up.
//
// Returns nil even if the task is already terminal – cancel is idempotent.
func (r *Runner) Cancel(ctx context.Context, taskID string) error {
	rec, err := r.store.GetChatTask(ctx, r.tenantID, taskID)
	if err != nil {
		return fmt.Errorf("get task: %w", err)
	}
	if isTerminal(rec.Status) {
		return nil // idempotent: already done/failed/cancelled
	}

	// Mark cancelled so a queued worker will skip it.
	rec.Status = store.ChatTaskCancelled
	now := time.Now().UTC()
	rec.DoneAt = &now
	if err := r.store.UpdateChatTask(ctx, r.tenantID, rec); err != nil {
		slog.Warn("taskrunner: persist cancel failed", "task", taskID, "err", err)
	}

	// Cancel in-flight context if any. Also drop any pending model
	// override / attachments — if the task was cancelled before run()
	// consumed them, the entries would otherwise sit in the maps until
	// process exit.
	r.mu.Lock()
	rt, running := r.inflight[taskID]
	delete(r.pendingModels, taskID)
	delete(r.pendingAttachments, taskID)
	r.mu.Unlock()
	if running && rt.cancel != nil {
		rt.cancel()
	}

	r.publishEvent(ctx, taskID, eventbus.Event{
		Type:      "task_cancelled",
		Data:      map[string]any{"taskId": taskID},
		Timestamp: now,
	})
	return nil
}

func isTerminal(s store.ChatTaskStatus) bool {
	switch s {
	case store.ChatTaskDone, store.ChatTaskFailed, store.ChatTaskCancelled:
		return true
	}
	return false
}

// run executes a single task. Called by the per-session worker.
//
// Lifecycle:
//   1. Skip immediately if the task was already cancelled while pending.
//   2. Mark running, persist, emit task_running.
//   3. Set up an independent (ctx, cancel) so an external Cancel() call
//      can interrupt the agent. Also enforce r.timeout.
//   4. Spawn agent.HandleWebChatStream in a goroutine, forward its
//      events to the bus.
//   5. On completion: persist done/failed/cancelled, emit terminal event.
func (r *Runner) run(rec *store.ChatTaskRecord) {
	ctx0 := r.rootCtx
	// Re-fetch to honour any cancel that arrived while the task was
	// pending in the queue.
	cur, err := r.store.GetChatTask(ctx0, r.tenantID, rec.ID)
	if err == nil && cur.Status == store.ChatTaskCancelled {
		// Already cancelled before we got a chance to start – terminal
		// event has been emitted by Cancel(); nothing more to do.
		return
	}

	taskCtx, taskCancel := context.WithTimeout(ctx0, r.timeout)
	defer taskCancel()

	// Pop and apply the per-task model override (set by Submit). Done
	// here rather than at enqueue time so a CancelBeforeRun task doesn't
	// leak into the map. effectiveModel inside agent.HandleMessage will
	// pick this up via ContextWithModel.
	r.mu.Lock()
	r.inflight[rec.ID] = &runningTask{cancel: taskCancel}
	if model, ok := r.pendingModels[rec.ID]; ok {
		taskCtx = agent.ContextWithModel(taskCtx, model)
		delete(r.pendingModels, rec.ID)
	}
	if atts, ok := r.pendingAttachments[rec.ID]; ok {
		taskCtx = agent.ContextWithAttachments(taskCtx, atts)
		delete(r.pendingAttachments, rec.ID)
	}
	r.mu.Unlock()
	defer func() {
		r.mu.Lock()
		delete(r.inflight, rec.ID)
		r.mu.Unlock()
	}()

	startedAt := time.Now().UTC()
	rec.Status = store.ChatTaskRunning
	rec.StartedAt = &startedAt
	if uerr := r.store.UpdateChatTask(taskCtx, r.tenantID, rec); uerr != nil {
		slog.Warn("taskrunner: mark running failed", "task", rec.ID, "err", uerr)
	}
	r.publishEvent(taskCtx, rec.ID, eventbus.Event{
		Type:      "task_running",
		Data:      map[string]any{"taskId": rec.ID},
		Timestamp: startedAt,
	})

	ah := r.resolver.AgentByID(rec.AgentID)
	if ah == nil {
		r.finishWithError(rec, fmt.Errorf("agent disappeared: %s", rec.AgentID))
		return
	}

	// Bridge: agent emits ChatEvent on a channel; we forward each to the
	// bus as an eventbus.Event. Buffer of 32 covers typical SSE jitter.
	events := make(chan agent.ChatEvent, 32)
	var (
		result    string
		streamErr error
	)
	go func() {
		defer close(events)
		defer func() {
			if rec := recover(); rec != nil {
				streamErr = fmt.Errorf("agent panic: %v", rec)
			}
		}()
		result = ah.HandleWebChatStream(taskCtx, rec.SessionKey, rec.Message, events)
	}()

	for evt := range events {
		r.publishEvent(taskCtx, rec.ID, eventbus.Event{
			Type:      evt.Type,
			Data:      evt.Data,
			Timestamp: time.Now().UTC(),
		})
	}

	// Determine terminal state. Prefer ctx error (cancel/timeout) over
	// stream error so users see "cancelled" rather than a confusing
	// downstream "context deadline exceeded".
	if streamErr == nil {
		if cerr := taskCtx.Err(); cerr != nil {
			streamErr = cerr
		}
	}

	switch {
	case errors.Is(streamErr, context.Canceled):
		// Cancel() already wrote the cancelled state and emitted the event.
		// Avoid double-publishing.
		return
	case errors.Is(streamErr, context.DeadlineExceeded):
		r.finishWithError(rec, fmt.Errorf("timed out after %s", r.timeout))
	case streamErr != nil:
		r.finishWithError(rec, streamErr)
	default:
		r.finishOK(rec, result)
	}
}

func (r *Runner) finishOK(rec *store.ChatTaskRecord, result string) {
	// Cross-channel delivery: if this web chat task had a Discord/Telegram
	// delivery target, forward the reply there so both sides stay in sync.
	r.mu.Lock()
	dt, hasDelivery := r.pendingDeliveries[rec.ID]
	delete(r.pendingDeliveries, rec.ID)
	r.mu.Unlock()
	if hasDelivery && result != "" {
		agent := r.resolver.AgentByID(rec.AgentID)
		if agent != nil {
			if err := agent.SendToChat(context.Background(), dt.channel, dt.chatID, result); err != nil {
				slog.Warn("taskrunner: cross-channel delivery failed", "channel", dt.channel, "chatID", dt.chatID, "err", err)
			}
		}
	}

	now := time.Now().UTC()
	rec.Status = store.ChatTaskDone
	rec.Result = result
	rec.DoneAt = &now
	if uerr := r.store.UpdateChatTask(r.rootCtx, r.tenantID, rec); uerr != nil {
		slog.Warn("taskrunner: persist done failed", "task", rec.ID, "err", uerr)
	}
	r.publishEvent(r.rootCtx, rec.ID, eventbus.Event{
		Type:      "task_done",
		Data:      map[string]any{"taskId": rec.ID, "result": result},
		Timestamp: now,
	})
}

func (r *Runner) finishWithError(rec *store.ChatTaskRecord, err error) {
	now := time.Now().UTC()
	rec.Status = store.ChatTaskFailed
	rec.Error = err.Error()
	rec.DoneAt = &now
	if uerr := r.store.UpdateChatTask(r.rootCtx, r.tenantID, rec); uerr != nil {
		slog.Warn("taskrunner: persist failed-state failed", "task", rec.ID, "err", uerr)
	}
	r.publishEvent(r.rootCtx, rec.ID, eventbus.Event{
		Type:      "task_error",
		Data:      map[string]any{"taskId": rec.ID, "error": err.Error()},
		Timestamp: now,
	})
}
