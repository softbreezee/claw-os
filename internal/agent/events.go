package agent

import (
	"context"

	"github.com/softbreezee/claw-os/internal/bus"
)

// ChatEvent represents a real-time event emitted during the agent ReAct loop.
type ChatEvent struct {
	Type string         `json:"type"` // "content", "tool_call", "tool_result", "done"
	Data map[string]any `json:"data,omitempty"`
}

type chatEventsKey struct{}

// ChatEventsFromContext retrieves the events channel from context, if present.
func ChatEventsFromContext(ctx context.Context) chan<- ChatEvent {
	ch, _ := ctx.Value(chatEventsKey{}).(chan<- ChatEvent)
	return ch
}

// ContextWithChatEvents returns a new context with the events channel attached.
func ContextWithChatEvents(ctx context.Context, ch chan<- ChatEvent) context.Context {
	return context.WithValue(ctx, chatEventsKey{}, ch)
}

// emitEvent sends an event to the channel in context, if present. Non-blocking.
func emitEvent(ctx context.Context, evt ChatEvent) {
	ch := ChatEventsFromContext(ctx)
	if ch == nil {
		return
	}
	select {
	case ch <- evt:
	case <-ctx.Done():
	}
}

// modelOverrideKey is the ctx-key for per-call model overrides. Uses the
// same pattern as chatEventsKey – a typed empty struct so it can't collide
// with arbitrary string keys other packages might attach.
type modelOverrideKey struct{}

// ContextWithModel attaches a per-call model override that takes precedence
// over Agent.model for the duration of the request. Used by the web chat
// handler to let users pick a different model without editing the agent.
// Empty model is treated as "no override" – returns ctx unchanged.
func ContextWithModel(ctx context.Context, model string) context.Context {
	if model == "" {
		return ctx
	}
	return context.WithValue(ctx, modelOverrideKey{}, model)
}

// ModelFromContext returns the per-call model override if present, "" otherwise.
func ModelFromContext(ctx context.Context) string {
	m, _ := ctx.Value(modelOverrideKey{}).(string)
	return m
}

// effectiveModel returns the per-call model override if set, otherwise the
// agent's default. Use at every primary LLM call site instead of a.model
// so the chat UI can offer a per-message model picker.
func effectiveModel(ctx context.Context, defaultModel string) string {
	if m := ModelFromContext(ctx); m != "" {
		return m
	}
	return defaultModel
}

// attachmentsKey is the ctx-key for per-call inbound attachments.
// Mirrors modelOverrideKey: a typed empty struct so it can't collide
// with arbitrary string keys.
type attachmentsKey struct{}

// ContextWithAttachments attaches a list of inbound attachments to the
// context. The Web chat path uses this to plumb files through
// HandleWebChatStream without changing its signature (which would
// ripple through the AgentHandle interface in taskrunner).
//
// Empty slice is treated as "no attachments" — returns ctx unchanged.
func ContextWithAttachments(ctx context.Context, atts []bus.Attachment) context.Context {
	if len(atts) == 0 {
		return ctx
	}
	return context.WithValue(ctx, attachmentsKey{}, atts)
}

// AttachmentsFromContext returns the attachments stashed by
// ContextWithAttachments, or nil.
func AttachmentsFromContext(ctx context.Context) []bus.Attachment {
	a, _ := ctx.Value(attachmentsKey{}).([]bus.Attachment)
	return a
}
