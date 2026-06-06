package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// fakeRunner records the task it was asked to run and returns a
// fixed reply / error. Just enough surface for the registration
// path tests below.
type fakeRunner struct {
	gotTask string
	gotMax  int
	reply   string
	err     error
}

func (f *fakeRunner) RunDelegated(_ context.Context, task string, maxIterations int) (string, error) {
	f.gotTask = task
	f.gotMax = maxIterations
	return f.reply, f.err
}

// TestRegisterDelegateTask_NilRunnerNoOp pins the documented opt-out:
// passing nil should not register anything (so tests / sandboxed
// flavors stay clean).
func TestRegisterDelegateTask_NilRunnerNoOp(t *testing.T) {
	r := NewRegistry(t.TempDir())
	RegisterDelegateTask(r, nil)
	for _, def := range r.Definitions() {
		if def.Function.Name == "delegate_task" {
			t.Fatalf("nil runner should not register the tool, got %+v", def)
		}
	}
}

// TestDelegateTask_ForwardsArgs verifies the JSON arg parsing +
// expected_output suffixing. The runner sees the task with the
// format spec appended verbatim — this is what lets the parent agent
// pin the output shape (table / JSON / etc.) for splicing.
func TestDelegateTask_ForwardsArgs(t *testing.T) {
	r := NewRegistry(t.TempDir())
	runner := &fakeRunner{reply: "ok"}
	RegisterDelegateTask(r, runner)

	args := map[string]any{
		"task":            "find 5 lithium battery makers in China",
		"expected_output": "markdown table: ticker, name, market_cap",
		"max_iterations":  12,
	}
	raw, _ := json.Marshal(args)
	out, err := r.Execute(context.Background(), "delegate_task", string(raw))
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if out != "ok" {
		t.Errorf("output = %q, want %q", out, "ok")
	}
	if !strings.Contains(runner.gotTask, "find 5 lithium battery makers") {
		t.Errorf("runner did not see task body: %q", runner.gotTask)
	}
	if !strings.Contains(runner.gotTask, "markdown table") {
		t.Errorf("runner did not see expected_output suffix: %q", runner.gotTask)
	}
	if runner.gotMax != 12 {
		t.Errorf("runner.gotMax = %d, want 12", runner.gotMax)
	}
}

// TestDelegateTask_EmptyTaskRejected guards against the model
// emitting `{"task": ""}` which would otherwise produce an empty
// LLM call inside the sub-task and burn an iteration for nothing.
func TestDelegateTask_EmptyTaskRejected(t *testing.T) {
	r := NewRegistry(t.TempDir())
	RegisterDelegateTask(r, &fakeRunner{reply: "ok"})

	_, err := r.Execute(context.Background(), "delegate_task", `{"task": ""}`)
	if err == nil {
		t.Fatal("empty task should error")
	}
	if !strings.Contains(err.Error(), "task is required") {
		t.Errorf("error message should mention required, got %v", err)
	}
}

// TestDelegateTask_RunnerErrorSurfaces pins the contract that a
// runner error becomes a tool_result envelope (the parent sees a
// failed tool call and can react) rather than crashing the call.
func TestDelegateTask_RunnerErrorSurfaces(t *testing.T) {
	r := NewRegistry(t.TempDir())
	runner := &fakeRunner{err: errors.New("provider timeout")}
	RegisterDelegateTask(r, runner)

	out, err := r.Execute(context.Background(), "delegate_task", `{"task": "summarize today"}`)
	if err == nil {
		t.Fatal("runner error should propagate up")
	}
	// Registry.Execute appends the standard "[Analyze the error...]"
	// envelope on error; the runner's own message must still be in the
	// composite output so the parent agent can read what failed.
	if !strings.Contains(out, "provider timeout") {
		t.Errorf("error message should be inside tool result envelope, got %q", out)
	}
}
