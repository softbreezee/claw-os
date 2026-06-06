package agent

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/softbreezee/claw-os/internal/bus"
	"github.com/softbreezee/claw-os/internal/provider"
)

// planModeNudge is the system message we prepend on plan-mode turns.
// It pins the contract for the model:
//   - tools are DISABLED this turn (we pass tools=nil to the provider)
//   - tools WILL be available on the next turn — reference them by name
//     so the execution turn knows what to invoke
//   - the FIRST execution action should be write_file('todo.md', ...)
//     so the per-session todo panel renders progress
//   - 3-7 numbered steps, end with the literal "Reply with 'go' …" line
//
// Adapted from fastclaw 531eb07. Differences from upstream:
//   - claw-os has no PromptMode / per-agent tool allowlist yet, so we
//     skip the "filter by allowlist" branch
//   - we don't ship delegate_task in v0.3 — the nudge mentions it as
//     a forward-looking pattern but doesn't require it
func planModeNudge() string {
	return "# PLAN MODE — output a plan only\n\n" +
		"The user has switched on plan mode for this message. They want " +
		"to see what you intend to do BEFORE any real work happens.\n\n" +
		"Tools are DISABLED for this response only — do not attempt to call " +
		"any tool, it will fail. They WILL be available on the next turn " +
		"when the user replies (the available set is listed in the tool " +
		"catalog system message). Reference tool names by name in the " +
		"plan so the execution turn knows what you intend to invoke at " +
		"each step.\n\n" +
		"Your VERY FIRST execution action (next turn) should be " +
		"`write_file('todo.md', <plan as `- [ ] ` items>)` so the user " +
		"sees a live progress panel as you work, then update_todo (or " +
		"another `write_file('todo.md', ...)` rewrite) as steps complete. " +
		"Mention this in the plan as an explicit Step 0 (or fold it into " +
		"Step 1) — the UI panel keys off this file.\n\n" +
		"Output a numbered plan with 3-7 steps. Each step is one or two " +
		"sentences describing the action plus the tool you'll use, e.g. " +
		"\"Step 3: Use `web_search` to confirm the latest filing date for " +
		"恩捷股份, then `write_file('todo.md', ...)` to mark Step 2 done.\". " +
		"Group related micro-actions into a single step — a plan is a " +
		"roadmap, not a transcript.\n\n" +
		"End with exactly one line: \"Reply with 'go' to execute, or " +
		"tell me what to change.\"\n\n" +
		"Do not start the work. Do not apologize for needing a plan. " +
		"Just the plan."
}

// buildToolCatalogForPlan renders a compact "what tools are available
// next turn" reference, injected as its own system message during plan
// mode. We pass tools=nil to the LLM in plan mode so the model can't
// accidentally call anything — but that also hides the registry, which
// historically caused the model to write plans omitting tools that
// were the whole point of plan mode (delegate_task, db_query, etc).
//
// Format: name + first sentence of description, one per line.
func buildToolCatalogForPlan(toolDefs []provider.Tool) string {
	if len(toolDefs) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("# Tool catalog (reference only — disabled THIS turn, available next turn)\n\n")
	b.WriteString("When your plan needs one of these, name it explicitly in the relevant step.\n\n")
	for _, t := range toolDefs {
		name := t.Function.Name
		desc := strings.TrimSpace(t.Function.Description)
		// First sentence, capped at 200 chars. Some tool descriptions
		// are run-on paragraphs and would otherwise dominate the
		// catalog.
		if idx := strings.IndexAny(desc, ".\n"); idx > 0 && idx < 200 {
			desc = desc[:idx]
		} else if len(desc) > 200 {
			desc = desc[:200] + "…"
		}
		fmt.Fprintf(&b, "- `%s` — %s\n", name, desc)
	}
	return b.String()
}

// handlePlanMode is the single-shot plan-only path: store the user
// message, ask the model for a plan with tools=nil, persist + emit
// the response with a planMode metadata flag so the UI can badge the
// bubble and offer "Continue / Adjust" affordances. No iteration
// loop, no tool execution. The next turn (sent without /plan) goes
// through the regular HandleMessage path and executes against the
// full session including the plan we just wrote.
//
// Streams content as the model produces it so the user sees the plan
// taking shape rather than waiting for a complete response. Uses the
// same provider + model + maxTokens / temperature as a regular turn.
func (a *Agent) handlePlanMode(ctx context.Context, msg bus.InboundMessage, task string) string {
	sess := a.sessions.Get(msg.Channel, msg.ChatID)

	// Persist the user's plan request so reload / cross-tab views see
	// the original objective, not just the assistant's plan body.
	userMsg := provider.Message{Role: "user", Content: task}
	sess.Append(userMsg)

	systemPrompt := a.ctxBuilder.BuildSystemPrompt()

	toolDefs := a.registry.Definitions()
	catalog := buildToolCatalogForPlan(toolDefs)

	messages := []provider.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "system", Content: planModeNudge()},
	}
	if catalog != "" {
		messages = append(messages, provider.Message{Role: "system", Content: catalog})
	}
	messages = append(messages, sess.GetMessages()...)

	emitEvent(ctx, ChatEvent{
		Type: "plan_started",
		Data: map[string]any{"task": task},
	})

	// Pass tools=nil — the model cannot emit tool_calls in plan mode.
	// Use ChatStream so the user sees plan tokens as they're written.
	sr, err := a.effectiveProvider(ctx).ChatStream(ctx, messages, nil, effectiveModel(ctx, a.model), a.maxTokens, a.temperature)
	if err != nil {
		slog.Warn("plan-mode chat failed, falling back to non-stream", "agent", a.name, "error", err)
		resp, ferr := a.effectiveProvider(ctx).Chat(ctx, messages, nil, effectiveModel(ctx, a.model), a.maxTokens, a.temperature)
		if ferr != nil {
			errMsg := fmt.Sprintf("⚠️ Plan generation failed (model=%s): %s", effectiveModel(ctx, a.model), ferr.Error())
			emitEvent(ctx, ChatEvent{Type: "content", Data: map[string]any{"content": errMsg}})
			emitEvent(ctx, ChatEvent{Type: "done"})
			return errMsg
		}
		sess.Append(provider.Message{
			Role:    "assistant",
			Content: resp.Content,
		})
		emitEvent(ctx, ChatEvent{
			Type: "content",
			Data: map[string]any{"content": resp.Content, "metadata": map[string]any{"planMode": true}},
		})
		emitEvent(ctx, ChatEvent{Type: "plan_proposed", Data: map[string]any{"plan": resp.Content}})
		emitEvent(ctx, ChatEvent{Type: "done"})
		return resp.Content
	}

	var planContent strings.Builder
	for {
		chunk, ok := sr.Next()
		if !ok {
			break
		}
		if chunk.Content != "" {
			planContent.WriteString(chunk.Content)
			emitEvent(ctx, ChatEvent{
				Type: "content_delta",
				Data: map[string]any{"delta": chunk.Content},
			})
		}
		if chunk.Done {
			break
		}
	}
	if err := sr.Err(); err != nil {
		slog.Warn("plan-mode stream error", "agent", a.name, "error", err)
	}

	final := planContent.String()
	sess.Append(provider.Message{
		Role:    "assistant",
		Content: final,
	})

	// plan_proposed is the durable event the UI hangs the
	// "Continue / Adjust" buttons off — it carries the full plan
	// text in one event so a late subscriber can render the buttons
	// without replaying every content_delta.
	emitEvent(ctx, ChatEvent{
		Type: "plan_proposed",
		Data: map[string]any{"plan": final},
	})
	emitEvent(ctx, ChatEvent{Type: "done"})
	return final
}
