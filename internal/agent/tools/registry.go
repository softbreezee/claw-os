package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/softbreezee/claw-os/internal/provider"
)

// ToolFunc is a function that executes a tool with JSON arguments and returns a result string.
type ToolFunc func(ctx context.Context, args json.RawMessage) (string, error)

// ToolSource indicates where a tool was registered from.
type ToolSource int

const (
	SourceBuiltin ToolSource = iota // built-in tool
	SourceMCP                       // MCP server tool
	SourcePlugin                    // plugin-provided tool
)

// Registry holds all registered tools.
type Registry struct {
	tools map[string]registeredTool

	// workspace is the agent's root directory. Stored on the registry so
	// later registrations (Skill env injection, sandbox toggling, etc)
	// can reuse the value instead of having every caller pass it again.
	// All file/exec tool product paths are anchored here.
	workspace string
}

// Workspace returns the agent's root directory captured at construction.
// Used by exec to set cmd.Dir, and by callers that need to anchor paths
// relative to the agent rather than the daemon's startup cwd.
func (r *Registry) Workspace() string {
	return r.workspace
}

type registeredTool struct {
	def    provider.Tool
	fn     ToolFunc
	source ToolSource
}

// NewRegistry creates a new tool registry with built-in tools.
func NewRegistry(workspace string) *Registry {
	r := &Registry{
		tools:     make(map[string]registeredTool),
		workspace: workspace,
	}
	r.registerBuiltins(workspace)
	return r
}

// Register adds a tool to the registry (as a built-in tool).
func (r *Registry) Register(name, description string, parameters interface{}, fn ToolFunc) {
	r.RegisterFrom(name, description, parameters, fn, SourceBuiltin)
}

// RegisterFrom adds a tool to the registry with an explicit source.
// Plugin-sourced tools can override built-in tools with the same name.
func (r *Registry) RegisterFrom(name, description string, parameters interface{}, fn ToolFunc, source ToolSource) {
	r.tools[name] = registeredTool{
		def: provider.Tool{
			Type: "function",
			Function: provider.ToolFunction{
				Name:        name,
				Description: description,
				Parameters:  parameters,
			},
		},
		fn:     fn,
		source: source,
	}
}

// HasBuiltin returns true if a built-in tool with the given name exists.
func (r *Registry) HasBuiltin(name string) bool {
	t, ok := r.tools[name]
	return ok && t.source == SourceBuiltin
}

// GetFunc returns the ToolFunc for a tool by name, or nil if not found.
func (r *Registry) GetFunc(name string) ToolFunc {
	t, ok := r.tools[name]
	if !ok {
		return nil
	}
	return t.fn
}

// Definitions returns all tool definitions for the LLM in stable alphabetical order.
// A consistent ordering prevents non-deterministic tool selection in the model.
func (r *Registry) Definitions() []provider.Tool {
	defs := make([]provider.Tool, 0, len(r.tools))
	for _, t := range r.tools {
		defs = append(defs, t.def)
	}
	sort.Slice(defs, func(i, j int) bool {
		return defs[i].Function.Name < defs[j].Function.Name
	})
	return defs
}

// Execute runs a tool by name with the given arguments.
func (r *Registry) Execute(ctx context.Context, name string, args string) (string, error) {
	tool, ok := r.tools[name]
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", name)
	}

	result, err := tool.fn(ctx, json.RawMessage(args))
	if err != nil {
		return result + "\n[Analyze the error above and try a different approach.]", err
	}
	return result, nil
}

// SetSandboxConfig updates the exec tool to use sandbox mode.
func (r *Registry) SetSandboxConfig(sbCfg *SandboxConfig) {
	registerExecWithSandbox(r, sbCfg, r.workspace)
}

func (r *Registry) registerBuiltins(workspace string) {
	registerExec(r, workspace)
	registerFile(r, workspace)
	registerMessage(r)
}
