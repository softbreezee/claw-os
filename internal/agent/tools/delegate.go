package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

// DelegateRunner is what the delegate_task tool calls to spawn an
// independent ReAct loop on the parent agent. The agent package
// implements this on Agent so we avoid pulling agent into tools
// (would form an import cycle).
type DelegateRunner interface {
	RunDelegated(ctx context.Context, task string, maxIterations int) (string, error)
}

type delegateTaskArgs struct {
	Task           string `json:"task"`
	ExpectedOutput string `json:"expected_output,omitempty"`
	MaxIterations  int    `json:"max_iterations,omitempty"`
}

// RegisterDelegateTask wires the delegate_task tool. No-op when runner
// is nil so callers can opt out by simply not constructing one (e.g.
// in tests, or for agent flavors where context-isolation fan-out
// doesn't apply).
//
// SERIAL registration: two delegate_task calls cannot run concurrently
// within one parent turn. Sub-agents share the parent's single
// workspace + sandbox, and parallel execution caused tool result
// races (one sub-agent writing todo.md while another reads it). The
// serial constraint trades fan-out wall time (5×N min instead of
// N min) for correctness.
//
// The tool description carries enough WHY for the model:
//   - clean parent context
//   - independent iteration budget
//   - serial — plan accordingly
//   - no nesting (sub-agents don't get this tool)
//
// Without the "no nesting" line, weaker models try to recursively
// delegate and burn through budgets exponentially.
func RegisterDelegateTask(r *Registry, runner DelegateRunner) {
	if runner == nil {
		return
	}
	r.Register("delegate_task",
		"Spawn a sub-task with its OWN context and OWN iteration budget. "+
			"Use this when the user's request decomposes into several large independent chunks "+
			"(e.g. \"复盘 trend-trader 持仓\" + \"复盘 dragon-hunter 持仓\" + \"汇总两个账户对比\"). "+
			"Each delegated task gets a fresh tool-iteration budget so you don't burn yours exploring, "+
			"and your own context stays clean of the dozens of intermediate tool results the sub-task goes through.\n\n"+
			"**Sub-tasks run SERIALLY, not in parallel.** Even if you emit 5 delegate_task calls in one round, "+
			"they execute one at a time. Plan accordingly: smaller per-task scope is better than fewer, larger calls.\n\n"+
			"The sub-task runs against the same provider, model, and tools you have (minus delegate_task itself — no nesting). "+
			"It cannot see your prior conversation, so pass everything it needs in the `task` arg: criteria, search hints, "+
			"earlier findings to build on, output format. Sub-tasks are best for tasks that produce a self-contained "+
			"artifact (a table, a draft email, a structured summary).\n\n"+
			"Return: the sub-task's final text exactly as it produced it. You then assemble multiple sub-task "+
			"results into the final deliverable for the user.",
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"task": map[string]interface{}{
					"type":        "string",
					"description": "Self-contained task description. The sub-task does NOT see your prior conversation — include all the context it needs to act: criteria, search hints, prior findings it should build on, region / language constraints, anything the sub-task must respect.",
				},
				"expected_output": map[string]interface{}{
					"type":        "string",
					"description": "Optional concrete format the sub-task should produce — e.g. \"markdown table with columns: ticker, position, cost; one row per holding; no preamble\". Appended to the task verbatim so the format spec is unambiguous.",
				},
				"max_iterations": map[string]interface{}{
					"type":        "integer",
					"description": "Optional override for the sub-task's tool-iteration budget. Default is the same cap as your turn (typically 20). For pure synthesis (no tools), 3-5 is enough; for web/db research, 10-20 is realistic.",
				},
			},
			"required": []string{"task"},
		},
		func(ctx context.Context, raw json.RawMessage) (string, error) {
			var args delegateTaskArgs
			if err := json.Unmarshal(raw, &args); err != nil {
				return "", fmt.Errorf("parse args: %w", err)
			}
			if args.Task == "" {
				return "", fmt.Errorf("task is required")
			}
			taskPrompt := args.Task
			if args.ExpectedOutput != "" {
				taskPrompt += "\n\n## Expected output format\n\n" + args.ExpectedOutput
			}
			out, err := runner.RunDelegated(ctx, taskPrompt, args.MaxIterations)
			if err != nil {
				// Surface the error inside the tool_result so the parent
				// sees it as a normal tool failure (gets the "analyze
				// the error and try a different approach" envelope from
				// the registry) rather than a hard tool-execution error.
				return fmt.Sprintf("[delegated task failed: %s]", err.Error()), err
			}
			return out, nil
		})
}
