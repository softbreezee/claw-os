package agent

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/softbreezee/claw-os/internal/provider"
)

// delegateDefaultTimeout caps the wall time one delegated task can
// spend on its loop, independent of how long the parent's overall
// turn has left. Big enough for ~15-20 web/DB-driven iterations.
// Smaller than the parent's turn timeout so a slow delegate doesn't
// take the parent down with it; parent ctx cancel still propagates
// (we wrap, not detach).
const delegateDefaultTimeout = 15 * time.Minute

// delegateMu serialises delegate_task calls within a single agent.
// Sub-tasks share the parent's workspace + sandbox + tool registry —
// running them in parallel caused races (one writing todo.md while
// another read it). Trading wall-time for correctness; the tool
// description tells the model to plan accordingly.
//
// Per-agent scope (not global): the lock is on Agent so two different
// agents can still run concurrent delegations of their own. Within
// one agent, all RunDelegated invocations queue up serially.
type delegateLock struct {
	mu sync.Mutex
}

// delegate gates concurrent RunDelegated calls per Agent. Stored on
// the Agent via SetDelegateRunner registration time so we don't grow
// Agent's struct surface for a feature most calls won't exercise.
var delegateLocks sync.Map // map[agentName]*delegateLock

func (a *Agent) delegateLock() *sync.Mutex {
	v, _ := delegateLocks.LoadOrStore(a.name, &delegateLock{})
	return &v.(*delegateLock).mu
}

// RunDelegated implements tools.DelegateRunner so the delegate_task
// tool can call back into the Agent without creating an import
// cycle.
//
// Always emits a `subagent_progress` event with phase="done" on exit
// (success, error, or panic) so the frontend's "currently delegating"
// indicator can clear cleanly between serial sub-task runs — even
// when the next task doesn't start immediately or the parent decides
// to handle the result before issuing another delegate_task.
func (a *Agent) RunDelegated(ctx context.Context, task string, maxIterations int) (out string, err error) {
	mu := a.delegateLock()
	mu.Lock()
	defer mu.Unlock()

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("delegated task panic: %v", r)
		}
		emitEvent(ctx, ChatEvent{Type: "subagent_progress", Data: map[string]any{
			"phase": "done",
		}})
	}()
	return a.runDelegatedLoop(ctx, task, maxIterations)
}

// runDelegatedLoop is a self-contained ReAct loop used by
// delegate_task.
//
// What it shares with HandleMessage:
//   - the parent's provider, model, tool registry, and SDK engine
//   - the same loop-detection pattern (3× same tool+args → break)
//
// What it deliberately does NOT do:
//   - no session persistence — the sub-task's working messages live
//     in a private slice and never touch session_messages
//   - no chat-event emission for content / tool_call / tool_result;
//     the parent's chat UI sees the delegate_task tool card + final
//     tool_result only, not the sub-task's intermediate steps
//   - no hooks, no skill-store refresh, no compaction, no runPostTurn
//   - no slash-command / plan-mode / steer short-circuit (caller is
//     the parent model via the delegate_task tool, not a human)
//
// delegate_task itself is filtered out of the sub-task's toolset so
// sub-tasks can't recursively delegate (v0.4 nesting limit; we may
// relax in a future version).
func (a *Agent) runDelegatedLoop(ctx context.Context, task string, maxIterations int) (string, error) {
	if a.getProvider() == nil {
		return "", errors.New("agent has no provider configured")
	}
	if maxIterations <= 0 {
		maxIterations = a.maxToolIterations
	}
	if maxIterations <= 0 {
		maxIterations = 20
	}

	// Each sub-task gets its own bounded ctx so a slow one can't
	// drain the rest of a fan-out. Parent cancel still wins — we're
	// wrapping, not detaching.
	subCtx, cancel := context.WithTimeout(ctx, delegateDefaultTimeout)
	defer cancel()
	ctx = subCtx

	systemPrompt := a.ctxBuilder.BuildSystemPrompt() + delegateSystemSuffix()
	messages := []provider.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: task},
	}

	// Filter delegate_task out of the sub-task's toolset — no
	// recursion in v0.4. Other tools (web_fetch, exec, file ops,
	// MCP, db_query, …) flow through unchanged.
	allTools := a.registry.Definitions()
	toolDefs := make([]provider.Tool, 0, len(allTools))
	for _, t := range allTools {
		if t.Function.Name == "delegate_task" {
			continue
		}
		toolDefs = append(toolDefs, t)
	}

	type sig struct {
		name string
		hash [32]byte
	}
	var lastSig sig
	consecutiveCount := 0

	for i := 0; i < maxIterations; i++ {
		emitEvent(ctx, ChatEvent{Type: "subagent_progress", Data: map[string]any{
			"iteration": i + 1,
			"max":       maxIterations,
			"phase":     "thinking",
		}})

		resp, err := a.effectiveProvider(ctx).Chat(ctx, messages, toolDefs, effectiveModel(ctx, a.model), a.maxTokens, a.temperature)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
				return "", fmt.Errorf(
					"delegated task ran out of its %s wall-time budget at iteration %d — task was too large; the parent should retry with a tighter scope or lower max_iterations",
					delegateDefaultTimeout, i+1)
			}
			return "", fmt.Errorf("delegated task chat failed at iteration %d: %w", i+1, err)
		}

		if !resp.HasToolCalls() {
			return resp.Content, nil
		}

		messages = append(messages, provider.Message{
			Role:             "assistant",
			Content:          resp.Content,
			ToolCalls:        resp.ToolCalls,
			ReasoningContent: resp.ReasoningContent,
		})

		// Loop detection — same as HandleMessage but on private state.
		loopDetected := false
		for _, tc := range resp.ToolCalls {
			s := sig{name: tc.Function.Name, hash: sha256.Sum256([]byte(tc.Function.Arguments))}
			if s.name == lastSig.name && s.hash == lastSig.hash {
				consecutiveCount++
			} else {
				consecutiveCount = 1
				lastSig = s
			}
			if consecutiveCount >= 3 {
				slog.Warn("delegated task tool-loop detected",
					"agent", a.name, "tool", tc.Function.Name)
				messages = append(messages, provider.Message{
					Role:    "system",
					Content: "Loop detected: same tool with same arguments 3 times. Stop and produce the deliverable from what you have.",
				})
				loopDetected = true
				break
			}
		}
		if loopDetected {
			break
		}

		toolNames := make([]string, 0, len(resp.ToolCalls))
		for _, tc := range resp.ToolCalls {
			toolNames = append(toolNames, tc.Function.Name)
		}
		emitEvent(ctx, ChatEvent{Type: "subagent_progress", Data: map[string]any{
			"iteration": i + 1,
			"max":       maxIterations,
			"phase":     "running",
			"tools":     toolNames,
		}})

		results := a.engine.executeToolsConcurrently(ctx, a.registry, resp.ToolCalls, a.workspacePath)
		for idx, r := range results {
			tc := resp.ToolCalls[idx]
			messages = append(messages, provider.Message{
				Role:       "tool",
				Content:    r.result,
				ToolCallID: tc.ID,
				Name:       r.toolName,
			})
		}
	}

	// Cap reached — forced-delivery turn with tools off so the
	// model produces *something* instead of timing out silently.
	slog.Warn("delegated task max iterations reached — forcing final delivery",
		"agent", a.name, "max", maxIterations)
	emitEvent(ctx, ChatEvent{Type: "subagent_progress", Data: map[string]any{
		"iteration": maxIterations,
		"max":       maxIterations,
		"phase":     "final-delivery",
	}})
	finalMessages := append(messages, provider.Message{
		Role: "system",
		Content: fmt.Sprintf(
			"You've reached the %d-iteration cap. Stop calling tools and produce the deliverable from what you have, with explicit gaps marked.",
			maxIterations),
	})
	finalResp, err := a.effectiveProvider(ctx).Chat(ctx, finalMessages, nil, effectiveModel(ctx, a.model), a.maxTokens, a.temperature)
	if err != nil {
		return "", fmt.Errorf("delegated task forced final delivery failed: %w", err)
	}
	if finalResp.Content == "" {
		return fmt.Sprintf("[delegated task reached %d-iteration limit without producing a final answer]", maxIterations), nil
	}
	return finalResp.Content, nil
}

// delegateSystemSuffix is appended to the agent's normal system
// prompt when running under runDelegatedLoop. Spells out the
// contract: the reply is a tool result for the parent, not chat
// with a human. Without this, sub-tasks emit chatty "Sure, I'll help
// you find …" preambles that the parent then has to strip before
// splicing into the final answer.
func delegateSystemSuffix() string {
	return "\n\n# Delegated-task mode\n\n" +
		"You are running as a delegated sub-task invoked by your parent " +
		"turn via the `delegate_task` tool. Your reply is consumed as a " +
		"tool result, not displayed to a human as chat. Follow these " +
		"rules strictly:\n\n" +
		"- Output **only** the deliverable the task asks for. No " +
		"preamble (\"Sure, I'll help…\"), no reassurance, no follow-up " +
		"questions, no offers to continue.\n" +
		"- If the task specifies an output format (table, JSON, " +
		"markdown rows), produce exactly that format — the parent " +
		"splices your output into a larger result.\n" +
		"- If you can't complete the task, return a brief note " +
		"explaining what you got and what blocked you. Partial " +
		"structured output beats no output.\n" +
		"- You have the parent's full tool set except `delegate_task` " +
		"itself (no nesting). Use them as normal.\n" +
		"- You don't see the parent's prior conversation. Everything " +
		"you need to do this task is in the user message below."
}
