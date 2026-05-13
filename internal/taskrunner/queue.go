package taskrunner

import (
	"log/slog"

	"github.com/softbreezee/claw-os/internal/store"
)

// sessionQueue is a per-(agent, session) FIFO with its own worker
// goroutine. Mirrors the design of internal/taskqueue.chatQueue but
// stores chat task records (so we can persist + emit lifecycle events).
//
// The buffer of 100 is generous: web chats rarely backlog more than a
// handful of tasks per session. If a single session somehow hits 100
// pending tasks the user is doing something pathological and Submit
// will block on send – which is the right pressure signal.
type sessionQueue struct {
	ch chan *store.ChatTaskRecord
}

// queueKey identifies a per-session queue. Different agents addressing
// the same logical sessionID still get separate queues, because session
// state is keyed by (agent, session) downstream.
func queueKey(agentID, sessionID string) string {
	return agentID + ":" + sessionID
}

// enqueue places a task on its session's queue, lazily creating the
// queue + worker goroutine on first use.
func (r *Runner) enqueue(rec *store.ChatTaskRecord) {
	key := queueKey(rec.AgentID, rec.SessionKey)

	r.mu.Lock()
	q, ok := r.queues[key]
	if !ok {
		q = &sessionQueue{ch: make(chan *store.ChatTaskRecord, 100)}
		r.queues[key] = q
		go r.runQueue(key, q)
	}
	depth := len(q.ch)
	r.mu.Unlock()

	if depth > 50 {
		slog.Warn("taskrunner: session queue depth high",
			"session", key, "depth", depth+1, "task", rec.ID)
	}

	// Best-effort, non-blocking-ish send. The channel is buffered (100),
	// so this only blocks under sustained pressure – which is the right
	// behaviour (apply backpressure to Submit caller).
	q.ch <- rec
}

// runQueue is the worker goroutine for one session: pull tasks off the
// channel and execute them serially. Exits when rootCtx is cancelled.
func (r *Runner) runQueue(key string, q *sessionQueue) {
	defer slog.Debug("taskrunner: session worker exit", "session", key)
	for {
		select {
		case <-r.rootCtx.Done():
			return
		case rec, ok := <-q.ch:
			if !ok {
				return
			}
			r.run(rec)
		}
	}
}
