package agent

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/softbreezee/claw-os/internal/provider"
)

// SteerMessage is a single mid-run "supplemental instruction" the user
// queued while a task is executing. It gets folded into the agent's
// running message list at the next ReAct iteration boundary — never
// mid-tool, never mid-LLM-call. The contract surfaced to the UI is:
// "won't interrupt, will be picked up at the next step".
type SteerMessage struct {
	Text       string
	ReceivedAt time.Time
}

// SteerBuffer is a goroutine-safe FIFO of pending steer messages
// attached to one agent turn. The runner stores it on context once
// per task; the agent loop drains it before each model call. When the
// loop drains, every queued message is consumed atomically (we don't
// want a slow LLM call to leave half the steers sitting in the buffer
// for the next iteration where they'd appear out of order with the
// tool results that came in between).
type SteerBuffer struct {
	mu       sync.Mutex
	messages []SteerMessage
}

// NewSteerBuffer returns an empty buffer ready to be attached to a
// task context. Cheap; allocate one per task.
func NewSteerBuffer() *SteerBuffer {
	return &SteerBuffer{}
}

// Push appends a steer message. Safe for concurrent producers (the
// HTTP /api/chat/steer handler and the wsockets surface both call it).
//
// Empty / whitespace-only text is rejected by the caller — Push
// itself doesn't validate so that callers can decide on the right
// error message. We do drop trivially empty strings here as a last
// line of defence so the agent never sees an empty `<steer></steer>`
// block.
func (b *SteerBuffer) Push(text string) {
	if b == nil || text == "" {
		return
	}
	b.mu.Lock()
	b.messages = append(b.messages, SteerMessage{
		Text:       text,
		ReceivedAt: time.Now(),
	})
	b.mu.Unlock()
}

// Drain returns all queued messages and clears the buffer in one
// atomic step. Returns nil (not empty slice) when there's nothing
// pending so callers can use `if msgs := buf.Drain(); msgs != nil`
// as a fast path that avoids the wrapper-message build entirely.
func (b *SteerBuffer) Drain() []SteerMessage {
	if b == nil {
		return nil
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.messages) == 0 {
		return nil
	}
	out := b.messages
	b.messages = nil
	return out
}

// Pending reports the queue depth without draining. Useful for
// telemetry and tests; not safe to use as a precondition for Drain
// (a concurrent Push between Pending and Drain would race).
func (b *SteerBuffer) Pending() int {
	if b == nil {
		return 0
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.messages)
}

type steerCtxKey struct{}

// ContextWithSteerBuffer returns a context carrying buf so the agent
// loop can pull it during execution. Used by taskrunner to plumb the
// per-task buffer through HandleWebChatStream → HandleMessage(Stream).
func ContextWithSteerBuffer(ctx context.Context, buf *SteerBuffer) context.Context {
	if buf == nil {
		return ctx
	}
	return context.WithValue(ctx, steerCtxKey{}, buf)
}

// SteerBufferFromContext returns the buffer attached to ctx, or nil
// when no task is wired (CLI invocations, cron-fired turns, etc).
// The agent loop must treat nil as "steering disabled for this run"
// and skip the drain — never panic.
func SteerBufferFromContext(ctx context.Context) *SteerBuffer {
	if ctx == nil {
		return nil
	}
	v, _ := ctx.Value(steerCtxKey{}).(*SteerBuffer)
	return v
}

// drainSteerIntoMessages folds any pending steer messages from ctx
// into messages as a single user message, tagged so the model knows
// it's a mid-run supplemental instruction rather than a fresh turn.
//
// Called once per ReAct iteration (before the next BeforeModelCall
// hook + LLM call). The folding is one user message — not one per
// queued steer — so messages that arrived close together don't blow
// up the prompt with redundant scaffolding. Messages stay in arrival
// order inside the block.
//
// Also emits a steer_received chat event per drained message so the
// UI can show a "queued → applied" transition rather than the steer
// silently disappearing into the void. Returns the number of
// messages folded so callers can log / branch on it.
func drainSteerIntoMessages(ctx context.Context, messages []provider.Message) ([]provider.Message, int) {
	buf := SteerBufferFromContext(ctx)
	if buf == nil {
		return messages, 0
	}
	queued := buf.Drain()
	if len(queued) == 0 {
		return messages, 0
	}

	var sb strings.Builder
	sb.WriteString("[user-steer — supplemental instructions queued mid-turn; treat as user input added to the conversation, then continue toward the original objective]\n")
	for _, m := range queued {
		sb.WriteString("- ")
		sb.WriteString(m.Text)
		sb.WriteString("\n")
	}
	messages = append(messages, provider.Message{
		Role:    "user",
		Content: strings.TrimRight(sb.String(), "\n"),
	})

	// Surface to the UI so the user sees their steer was picked up
	// (vs silently lost). Each message gets its own event so the UI
	// can show one chip per steer if it wants.
	for _, m := range queued {
		emitEvent(ctx, ChatEvent{
			Type: "steer_received",
			Data: map[string]any{"text": m.Text},
		})
	}
	return messages, len(queued)
}
