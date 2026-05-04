package taskrunner

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/fastclaw-ai/fastclaw/internal/agent"
	"github.com/fastclaw-ai/fastclaw/internal/eventbus"
	"github.com/fastclaw-ai/fastclaw/internal/store"
)

// fakeAgent simulates an Agent for tests. The behaviour func receives
// the same args as Agent.HandleWebChatStream and is expected to drive
// the events channel just like the real implementation.
type fakeAgent struct {
	behaviour func(ctx context.Context, sessionID, text string, events chan<- agent.ChatEvent) string
}

func (f *fakeAgent) HandleWebChatStream(ctx context.Context, sessionID, text string, events chan<- agent.ChatEvent) string {
	return f.behaviour(ctx, sessionID, text, events)
}

type fakeResolver struct{ ag AgentHandle }

func (r *fakeResolver) AgentByID(string) AgentHandle { return r.ag }

// memStore is the smallest possible Store implementation – satisfies
// only the chat-task methods used by the runner. All other methods are
// stubs that return zero values; tests that need them would fail loudly.
type memStore struct {
	mu    sync.Mutex
	tasks map[string]*store.ChatTaskRecord
}

func newMemStore() *memStore { return &memStore{tasks: map[string]*store.ChatTaskRecord{}} }

func (m *memStore) CreateChatTask(_ context.Context, _ string, t *store.ChatTaskRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *t
	m.tasks[t.ID] = &cp
	return nil
}

func (m *memStore) UpdateChatTask(_ context.Context, _ string, t *store.ChatTaskRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *t
	m.tasks[t.ID] = &cp
	return nil
}

func (m *memStore) GetChatTask(_ context.Context, _, id string) (*store.ChatTaskRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tasks[id]
	if !ok {
		return nil, store.ErrNotSupported
	}
	cp := *t
	return &cp, nil
}

func (m *memStore) ListChatTasks(context.Context, string, store.ChatTaskFilters) ([]store.ChatTaskRecord, error) {
	return nil, store.ErrNotSupported
}
func (m *memStore) DeleteChatTask(context.Context, string, string) error { return nil }

// Stubs for the rest of the Store interface – none of these are reached
// by the taskrunner code under test, but Go needs them to satisfy the
// interface. Returning zero values is fine; if a test ever actually
// calls one of these, the assertion that follows will catch the misuse.
func (m *memStore) GetConfig(context.Context, string) (*store.TenantConfig, error)        { return nil, nil }
func (m *memStore) SaveConfig(context.Context, string, *store.TenantConfig) error         { return nil }
func (m *memStore) DeleteConfig(context.Context, string) error                            { return nil }
func (m *memStore) ListAgents(context.Context, string) ([]store.AgentRecord, error)       { return nil, nil }
func (m *memStore) GetAgent(context.Context, string, string) (*store.AgentRecord, error)  { return nil, nil }
func (m *memStore) SaveAgent(context.Context, string, *store.AgentRecord) error           { return nil }
func (m *memStore) DeleteAgent(context.Context, string, string) error                     { return nil }
func (m *memStore) GetSession(context.Context, string, string, string) (*store.SessionRecord, error) {
	return nil, nil
}
func (m *memStore) SaveSession(context.Context, string, string, string, *store.SessionRecord) error {
	return nil
}
func (m *memStore) ListSessions(context.Context, string, string) ([]store.SessionMeta, error) {
	return nil, nil
}
func (m *memStore) DeleteSession(context.Context, string, string, string) error    { return nil }
func (m *memStore) GetMemory(context.Context, string, string) (string, error)      { return "", nil }
func (m *memStore) SaveMemory(context.Context, string, string, string) error       { return nil }
func (m *memStore) SearchMemory(context.Context, string, string, string, int) ([]store.MemoryEntry, error) {
	return nil, nil
}
func (m *memStore) AppendMemoryLog(context.Context, string, string, store.MemoryEntry) error {
	return nil
}
func (m *memStore) GetWorkspaceFile(context.Context, string, string, string) ([]byte, error) {
	return nil, nil
}
func (m *memStore) SaveWorkspaceFile(context.Context, string, string, string, []byte) error {
	return nil
}
func (m *memStore) ListWorkspaceFiles(context.Context, string, string) ([]string, error) {
	return nil, nil
}
func (m *memStore) ListCronJobs(context.Context, string) ([]store.CronJobRecord, error) {
	return nil, nil
}
func (m *memStore) GetCronJob(context.Context, string, string) (*store.CronJobRecord, error) {
	return nil, nil
}
func (m *memStore) SaveCronJob(context.Context, string, *store.CronJobRecord) error    { return nil }
func (m *memStore) DeleteCronJob(context.Context, string, string) error                { return nil }
func (m *memStore) GetDueCronJobs(context.Context, time.Time) ([]store.CronJobRecord, error) {
	return nil, nil
}
func (m *memStore) LockCronJob(context.Context, string, string) (bool, error)         { return false, nil }
func (m *memStore) UpdateCronJobRun(context.Context, string, time.Time, time.Time) error {
	return nil
}
func (m *memStore) Close() error { return nil }

// collectEvents drains a subscription into a slice. Returns when the
// channel is closed OR maxWait elapses (whichever first).
func collectEvents(ch <-chan eventbus.Event, maxWait time.Duration) []eventbus.Event {
	deadline := time.After(maxWait)
	var out []eventbus.Event
	for {
		select {
		case evt, ok := <-ch:
			if !ok {
				return out
			}
			out = append(out, evt)
			if isTerminalEvt(evt.Type) {
				return out
			}
		case <-deadline:
			return out
		}
	}
}

func isTerminalEvt(t string) bool {
	return t == "task_done" || t == "task_error" || t == "task_cancelled"
}

// happyPath: agent emits content, runner forwards to bus and persists done.
func TestRunner_HappyPath(t *testing.T) {
	st := newMemStore()
	bus := eventbus.NewMemoryBus()
	defer bus.Close()
	ag := &fakeAgent{
		behaviour: func(ctx context.Context, sid, text string, events chan<- agent.ChatEvent) string {
			events <- agent.ChatEvent{Type: "content", Data: map[string]any{"content": "hi"}}
			return "hi"
		},
	}
	r := New(st, bus, &fakeResolver{ag: ag}, Options{Timeout: 2 * time.Second})
	defer r.Stop()

	taskID, err := r.Submit(context.Background(), "agent-1", "sess-1", "hello")
	if err != nil {
		t.Fatalf("submit: %v", err)
	}

	// Subscribe AFTER submit – we should still receive task_running/done
	// because the per-session worker hasn't processed yet (submit is fast,
	// but the worker goroutine is async).
	ch, cancel := bus.Subscribe(TopicFor(taskID))
	defer cancel()
	evts := collectEvents(ch, 2*time.Second)

	// Find the terminal event.
	var terminal *eventbus.Event
	for i := range evts {
		if isTerminalEvt(evts[i].Type) {
			terminal = &evts[i]
			break
		}
	}
	if terminal == nil {
		t.Fatalf("no terminal event in %d events", len(evts))
	}
	if terminal.Type != "task_done" {
		t.Fatalf("terminal=%q, want task_done; events=%+v", terminal.Type, evts)
	}
	if got := terminal.Data["result"]; got != "hi" {
		t.Fatalf("result=%v, want hi", got)
	}

	// Verify persistence.
	rec, err := st.GetChatTask(context.Background(), store.DefaultTenantID, taskID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if rec.Status != store.ChatTaskDone || rec.Result != "hi" {
		t.Fatalf("rec=%+v", rec)
	}
}

// agentBlocksUntilCancel: agent's HandleWebChatStream blocks on ctx.Done().
// External Cancel() must interrupt it and produce task_cancelled.
func TestRunner_Cancel(t *testing.T) {
	st := newMemStore()
	bus := eventbus.NewMemoryBus()
	defer bus.Close()

	started := make(chan struct{})
	ag := &fakeAgent{
		behaviour: func(ctx context.Context, sid, text string, events chan<- agent.ChatEvent) string {
			close(started)
			<-ctx.Done() // wait for cancel
			return ""
		},
	}
	r := New(st, bus, &fakeResolver{ag: ag}, Options{Timeout: 5 * time.Second})
	defer r.Stop()

	taskID, err := r.Submit(context.Background(), "agent-1", "sess-1", "go")
	if err != nil {
		t.Fatalf("submit: %v", err)
	}

	ch, cancel := bus.Subscribe(TopicFor(taskID))
	defer cancel()

	// Wait for the agent to actually start, then cancel.
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("agent never started")
	}
	if err := r.Cancel(context.Background(), taskID); err != nil {
		t.Fatalf("cancel: %v", err)
	}

	evts := collectEvents(ch, 2*time.Second)
	var sawCancel bool
	for _, e := range evts {
		if e.Type == "task_cancelled" {
			sawCancel = true
			break
		}
	}
	if !sawCancel {
		t.Fatalf("no task_cancelled event in %+v", evts)
	}

	rec, _ := st.GetChatTask(context.Background(), store.DefaultTenantID, taskID)
	if rec.Status != store.ChatTaskCancelled {
		t.Fatalf("status=%q, want cancelled", rec.Status)
	}
}

// timeout: short Timeout option triggers task_error.
func TestRunner_Timeout(t *testing.T) {
	st := newMemStore()
	bus := eventbus.NewMemoryBus()
	defer bus.Close()

	ag := &fakeAgent{
		behaviour: func(ctx context.Context, sid, text string, events chan<- agent.ChatEvent) string {
			<-ctx.Done()
			return ""
		},
	}
	r := New(st, bus, &fakeResolver{ag: ag}, Options{Timeout: 100 * time.Millisecond})
	defer r.Stop()

	taskID, err := r.Submit(context.Background(), "agent-1", "sess-1", "go")
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	ch, cancel := bus.Subscribe(TopicFor(taskID))
	defer cancel()

	evts := collectEvents(ch, 2*time.Second)
	var terminal string
	for _, e := range evts {
		if isTerminalEvt(e.Type) {
			terminal = e.Type
			break
		}
	}
	if terminal != "task_error" {
		t.Fatalf("terminal=%q, want task_error", terminal)
	}
}

// perSessionSerialisation: two tasks for the same session must run
// serially (second starts only after first finishes). Different sessions
// run in parallel.
func TestRunner_PerSessionSerial(t *testing.T) {
	st := newMemStore()
	bus := eventbus.NewMemoryBus()
	defer bus.Close()

	var (
		mu       sync.Mutex
		started  []time.Time
		finished []time.Time
	)
	ag := &fakeAgent{
		behaviour: func(ctx context.Context, sid, text string, events chan<- agent.ChatEvent) string {
			mu.Lock()
			started = append(started, time.Now())
			mu.Unlock()
			time.Sleep(150 * time.Millisecond)
			mu.Lock()
			finished = append(finished, time.Now())
			mu.Unlock()
			return text
		},
	}
	r := New(st, bus, &fakeResolver{ag: ag}, Options{Timeout: 5 * time.Second})
	defer r.Stop()

	t1, _ := r.Submit(context.Background(), "agent-1", "session-A", "1")
	t2, _ := r.Submit(context.Background(), "agent-1", "session-A", "2")

	ch1, c1 := bus.Subscribe(TopicFor(t1))
	defer c1()
	ch2, c2 := bus.Subscribe(TopicFor(t2))
	defer c2()
	collectEvents(ch1, 3*time.Second)
	collectEvents(ch2, 3*time.Second)

	mu.Lock()
	defer mu.Unlock()
	if len(started) != 2 || len(finished) != 2 {
		t.Fatalf("started=%d finished=%d", len(started), len(finished))
	}
	// Strict serialisation: task 2 starts AFTER task 1 finishes.
	if !started[1].After(finished[0]) {
		t.Fatalf("tasks ran in parallel: started=%v finished=%v", started, finished)
	}
}

// terminalSnapshotForResume: the bus emits task_done; a late subscriber
// should not get it (already drained). The handler-level snapshot logic
// is what saves us in production – we test that GetChatTask reflects the
// correct terminal state, which is the data the snapshot reads.
func TestRunner_TerminalStatePersisted(t *testing.T) {
	st := newMemStore()
	bus := eventbus.NewMemoryBus()
	defer bus.Close()
	ag := &fakeAgent{
		behaviour: func(ctx context.Context, sid, text string, events chan<- agent.ChatEvent) string {
			return "result"
		},
	}
	r := New(st, bus, &fakeResolver{ag: ag}, Options{Timeout: time.Second})
	defer r.Stop()

	taskID, _ := r.Submit(context.Background(), "agent-1", "sess-1", "x")
	// Wait for completion via subscription.
	ch, cancel := bus.Subscribe(TopicFor(taskID))
	defer cancel()
	collectEvents(ch, 2*time.Second)

	rec, err := st.GetChatTask(context.Background(), store.DefaultTenantID, taskID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if rec.Status != store.ChatTaskDone {
		t.Fatalf("status=%q, want done", rec.Status)
	}
	if rec.DoneAt == nil {
		t.Fatal("DoneAt not set")
	}
}
