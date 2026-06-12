package agent

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
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
		"Just the plan.\n\n" +
		"CRITICAL OUTPUT RULE: Respond with PLAIN TEXT / Markdown only. " +
		"Do NOT emit any tool-call syntax — no <tool_call>, <invoke>, " +
		"<function_calls>, <parameter>, no DSML tokens, no JSON tool " +
		"arguments. Tools are disabled this turn; emitting their syntax " +
		"produces garbage the user can't read. Name tools in prose " +
		"(e.g. \"use web_fetch on the Sina page\"), never as a call."
}

// Plan-mode tool-leak scrubbing.
//
// In plan mode we pass tools=nil to the provider so the model CANNOT
// emit a structured tool_call. Weaker / non-OpenAI models ignore the
// "don't call tools" instruction anyway and dump a tool invocation as
// raw text — which then renders as garbage in the plan bubble
// ("< | DSML | invoke name=\"web_fetch\">…"). These patterns strip
// that leaked text back out so the user sees either a clean plan or a
// clear "model misbehaved, retry" message instead of mangled markup.
//
// Three leak shapes are covered:
//   - OpenAI/Anthropic XML:  <tool_call>…</tool_call>, <invoke …>…</invoke>,
//     <function_calls>…</function_calls>, <parameter …>…</parameter>
//   - Codex-style DSML tokens that some providers stream literally,
//     including the fullwidth-pipe ("｜") mangled variant seen in the
//     wild (e.g. "<｜tool▁calls▁begin｜>"), plus the space-separated
//     "< | DSML | …>" form a markdown renderer produces from them
//   - bare leftover tags after the above (a lone </invoke> etc.)
var (
	// Block-level tool-call wrappers — remove the whole element incl.
	// inner content. (?s) so . matches newlines; non-greedy so two
	// adjacent blocks don't get merged into one giant match.
	planToolBlockRe = regexp.MustCompile(`(?s)<\s*(tool_calls?|function_calls?|invoke|antml:invoke)\b.*?</\s*(tool_calls?|function_calls?|invoke|antml:invoke)\s*>`)

	// DSML / special-token sentinels, including fullwidth-pipe mangled
	// forms and the "< | DSML | … >" renderer artifact. We delete the
	// whole angle-bracket span greedily per line.
	planDSMLRe = regexp.MustCompile(`(?s)<\s*[|｜][^>]*>|<[^>]*\bDSML\b[^>]*>|[<>]?\s*[|｜]\s*(DSML|tool_calls?|invoke|parameter)\s*[|｜]?`)

	// Leftover single tags (orphan opens/closes that escaped the
	// block matcher because the model never produced a matching pair).
	planOrphanTagRe = regexp.MustCompile(`</?\s*(tool_calls?|function_calls?|invoke|antml:invoke|parameter)\b[^>]*>`)

	// A dangling half-open tag fragment at the END of the kept prefix,
	// left when truncation cut at a leak marker that sat a few chars
	// past the opening bracket (e.g. "…实际业务。\n<｜" — the "<｜" has
	// no closing ">"). Anchored to end-of-string so we only trim a
	// trailing fragment, never an inner "<" the user legitimately
	// wrote mid-sentence.
	planDanglingOpenRe = regexp.MustCompile(`(?s)\s*<[|｜<\s]*$`)

	// Strong-signal markers that say "this response is (partly) a
	// leaked tool call" regardless of how mangled the surrounding
	// markup is. When any of these appear we don't trust the
	// post-scrub leftovers (they're typically orphaned parameter
	// VALUES like a bare URL or "6000" stranded between deleted tags),
	// so the caller routes to the retry fallback instead of rendering
	// half-garbage. See sanitizePlanText's leaked return.
	planLeakSignalRe = regexp.MustCompile(`(?i)\bDSML\b|<\s*(tool_calls?|function_calls?|invoke)\b|name\s*=\s*"(web_fetch|exec|web_search|read_file|write_file|db_query)"|[|｜]\s*(tool_calls?|invoke|parameter)`)
)

// sanitizePlanText strips leaked tool-call markup from a plan-mode
// response and tidies whitespace.
//
// Returns (clean, leaked). When leaked is true the model emitted a
// tool call (in any of the OpenAI / Anthropic / DSML shapes).
//
// Strategy: TRUNCATE at the first leak marker, don't try to scrub
// in place. Once the model starts emitting tool-call syntax in plan
// mode, everything after it is an execution attempt — including big
// inline payloads (curl commands, embedded python, regex) that live
// BETWEEN the tags as parameter values and can't be removed by
// tag-only regexes. Cutting at the first marker keeps any genuine
// preamble the model wrote before it lost the plot, and discards the
// whole tool-call dump. We then run the tag scrubbers on the kept
// prefix too, in case a stray marker slipped in earlier.
//
// Idempotent — safe on already-clean text (no marker found, nothing
// truncated or stripped, leaked=false).
func sanitizePlanText(s string) (clean string, leaked bool) {
	if s == "" {
		return "", false
	}
	if loc := planLeakSignalRe.FindStringIndex(s); loc != nil {
		leaked = true
		s = s[:loc[0]] // keep only the prose before the first leak
	}
	// Defence in depth: scrub any tag fragments that survived in the
	// kept prefix (rare, but a model can interleave a short stray tag
	// mid-sentence before the big dump).
	s = planToolBlockRe.ReplaceAllString(s, "")
	s = planDSMLRe.ReplaceAllString(s, "")
	s = planOrphanTagRe.ReplaceAllString(s, "")
	// The truncation cut at the leak MARKER, which may sit a couple
	// chars after the opening bracket (e.g. cutting at "DSML" leaves a
	// dangling "<｜" before it). Strip any half-open tag fragment left
	// at the very end of the kept prefix.
	s = planDanglingOpenRe.ReplaceAllString(s, "")
	s = regexp.MustCompile(`\n{3,}`).ReplaceAllString(s, "\n\n")
	return strings.TrimSpace(s), leaked
}

// planProseRe estimates how much of a string is actual prose vs.
// orphaned parameter values (URLs, bare numbers, punctuation). We
// count "word-like" runs of letters (incl. CJK) — a real plan has
// many; a scrubbed tool-call leak leaves mostly a stranded URL and a
// number like "6000", which score near zero.
var planProseRe = regexp.MustCompile(`[\p{L}]{2,}`)

// planURLRe strips URLs before the prose count. A scrubbed tool-call
// leak typically strands a parameter URL whose host/path fragments
// (stock, finance, sina, corp, php, …) would otherwise count as many
// prose "words" and defeat the leak check. Removing URLs first means
// a stranded "6000 https://…" scores ~0 prose runs and falls back.
var planURLRe = regexp.MustCompile(`https?://\S+`)

// planLeakNotice is appended when a leaked response kept a genuine
// preamble but the model never produced an actual plan (it started
// executing instead). Tells the user the truncation happened and what
// to do, without throwing away the preamble they might want to read.
const planLeakNotice = "\n\n---\n\n_⚠️ 模型在这里开始直接调用工具(已拦截)而没有写出完整计划。" +
	"想要计划就重发 `/plan <任务>`;想让它直接执行,去掉 `/plan` 直接发任务即可。_"

// resolvePlan combines sanitize + fallback into the single value the
// caller renders. Three outcomes:
//   - not leaked → return the clean plan as-is
//   - leaked, preamble too thin to be useful → full retry hint
//   - leaked, real preamble survived → keep it + append a notice so
//     the user knows the plan was cut short (the model jumped to
//     execution mid-response)
//
// "Thin preamble" is measured by letter-run count AFTER stripping
// URLs — a stranded parameter URL's path fragments (stock/finance/
// sina/…) would otherwise inflate the count and let a pure leak slip
// through as if it were prose.
func resolvePlan(raw string) string {
	clean, leaked := sanitizePlanText(raw)
	trimmed := strings.TrimSpace(clean)

	if trimmed == "" {
		if leaked {
			return "_(模型试图直接执行而不是规划。重发 `/plan <任务>`,或去掉 `/plan` 直接发任务让它执行。)_"
		}
		return "_(模型返回了空计划。重发 `/plan <任务>`。)_"
	}

	if leaked {
		proseOnly := planURLRe.ReplaceAllString(trimmed, "")
		wordRuns := len(planProseRe.FindAllString(proseOnly, -1))
		if wordRuns < 6 {
			slog.Warn("plan-mode produced no usable plan after scrubbing leaked tool calls",
				"scrubbed_len", len(trimmed), "word_runs", wordRuns)
			return "_(模型试图直接执行而不是规划。重发 `/plan <任务>`,或去掉 `/plan` 直接发任务让它执行。)_"
		}
		// Real preamble survived but the model never finished a plan —
		// keep what it wrote, flag that the rest was an intercepted
		// tool dump.
		slog.Info("plan-mode kept preamble, truncated leaked tool dump",
			"kept_len", len(trimmed), "word_runs", wordRuns)
		return clean + planLeakNotice
	}
	return clean
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
		final := resolvePlan(resp.Content)
		sess.Append(provider.Message{
			Role:    "assistant",
			Content: final,
		})
		// One content_delta carries the whole cleaned plan into the
		// plan bubble the frontend opened on plan_started; plan_proposed
		// then finalizes it and renders the Continue/Adjust footer.
		emitEvent(ctx, ChatEvent{Type: "content_delta", Data: map[string]any{"delta": final}})
		emitEvent(ctx, ChatEvent{Type: "plan_proposed", Data: map[string]any{"plan": final}})
		emitEvent(ctx, ChatEvent{Type: "done"})
		return final
	}

	// We accumulate the full stream and sanitize ONCE at the end
	// rather than emitting raw content_delta chunks. Plan mode is the
	// one path where the model frequently leaks a raw tool-call token
	// stream (it wants to just DO the work) — and a half-sanitized
	// delta would flash garbage like "< | DSML | invoke name=..." in
	// the UI before the final scrub. Plans are short; buffering them
	// trades a little streaming feel for never showing the user a
	// mangled tool-call dump. We emit plan_started (spinner) up front
	// so the bubble still appears immediately.
	var planContent strings.Builder
	for {
		chunk, ok := sr.Next()
		if !ok {
			break
		}
		if chunk.Content != "" {
			planContent.WriteString(chunk.Content)
		}
		if chunk.Done {
			break
		}
	}
	if err := sr.Err(); err != nil {
		slog.Warn("plan-mode stream error", "agent", a.name, "error", err)
	}

	final := resolvePlan(planContent.String())
	sess.Append(provider.Message{
		Role:    "assistant",
		Content: final,
	})

	// One content_delta carries the cleaned plan into the plan bubble
	// the frontend opened on plan_started; plan_proposed then finalizes
	// it (renders the Continue / Adjust footer). Using content_delta —
	// not content — keeps the plan text inside the plan bubble instead
	// of spawning a separate agent bubble.
	emitEvent(ctx, ChatEvent{Type: "content_delta", Data: map[string]any{"delta": final}})
	emitEvent(ctx, ChatEvent{
		Type: "plan_proposed",
		Data: map[string]any{"plan": final},
	})
	emitEvent(ctx, ChatEvent{Type: "done"})
	return final
}
