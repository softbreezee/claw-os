package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/codeany-ai/open-agent-sdk-go/costtracker"
	sdktools "github.com/codeany-ai/open-agent-sdk-go/tools"
	sdktypes "github.com/codeany-ai/open-agent-sdk-go/types"

	"github.com/softbreezee/claw-os/internal/agent/tools"
	"github.com/softbreezee/claw-os/internal/provider"
)

// readOnlyTools lists tools that are safe to run concurrently.
var readOnlyTools = map[string]bool{
	"read_file":     true,
	"list_dir":      true,
	"web_fetch":     true,
	"web_search":    true,
	"memory_search": true,
	"load_skill":    true,
}

// toolAdapter wraps a Pawnix tool as an SDK Tool interface.
type toolAdapter struct {
	name        string
	description string
	params      interface{}
	fn          tools.ToolFunc
}

func (t *toolAdapter) Name() string        { return t.name }
func (t *toolAdapter) Description() string  { return t.description }

func (t *toolAdapter) InputSchema() sdktypes.ToolInputSchema {
	// Convert Pawnix params (interface{}) to SDK ToolInputSchema
	if t.params == nil {
		return sdktypes.ToolInputSchema{Type: "object"}
	}
	data, err := json.Marshal(t.params)
	if err != nil {
		return sdktypes.ToolInputSchema{Type: "object"}
	}
	var schema sdktypes.ToolInputSchema
	if err := json.Unmarshal(data, &schema); err != nil {
		return sdktypes.ToolInputSchema{Type: "object"}
	}
	return schema
}

func (t *toolAdapter) Call(ctx context.Context, input map[string]interface{}, tCtx *sdktypes.ToolUseContext) (*sdktypes.ToolResult, error) {
	// Convert input map to JSON for Pawnix's ToolFunc
	argsJSON, err := json.Marshal(input)
	if err != nil {
		return &sdktypes.ToolResult{IsError: true, Error: err.Error()}, nil
	}

	result, err := t.fn(ctx, json.RawMessage(argsJSON))
	if err != nil {
		errText := result
		if errText != "" {
			errText += "\n"
		}
		errText += err.Error()
		return &sdktypes.ToolResult{
			IsError: true,
			Error:   errText,
			Content: []sdktypes.ContentBlock{{
				Type: sdktypes.ContentBlockText,
				Text: errText,
			}},
		}, nil
	}

	return &sdktypes.ToolResult{
		Content: []sdktypes.ContentBlock{{
			Type: sdktypes.ContentBlockText,
			Text: result,
		}},
	}, nil
}

func (t *toolAdapter) IsConcurrencySafe(input map[string]interface{}) bool {
	return readOnlyTools[t.name]
}

func (t *toolAdapter) IsReadOnly(input map[string]interface{}) bool {
	return readOnlyTools[t.name]
}

// sdkEngine wraps SDK components for concurrent tool execution and cost tracking.
type sdkEngine struct {
	costTracker *costtracker.Tracker
}

// newSDKEngine creates a new SDK engine with cost tracking.
func newSDKEngine(sessionID string) *sdkEngine {
	return &sdkEngine{
		costTracker: costtracker.NewTracker(sessionID),
	}
}

// buildSDKRegistry converts Pawnix's tool registry into an SDK registry.
func buildSDKRegistry(fcRegistry *tools.Registry) *sdktools.Registry {
	sdkReg := sdktools.NewRegistry()
	for _, def := range fcRegistry.Definitions() {
		fn := fcRegistry.GetFunc(def.Function.Name)
		if fn == nil {
			continue
		}
		sdkReg.Register(&toolAdapter{
			name:        def.Function.Name,
			description: def.Function.Description,
			params:      def.Function.Parameters,
			fn:          fn,
		})
	}
	return sdkReg
}

// toolCallResult holds the result of a single tool call with metadata.
type toolCallResult struct {
	toolCallID string
	toolName   string
	result     string
	err        error
}

// executeToolsConcurrently runs tool calls using the SDK's concurrent executor.
func (e *sdkEngine) executeToolsConcurrently(ctx context.Context, fcRegistry *tools.Registry, toolCalls []provider.ToolCall, workspace string) []toolCallResult {
	sdkReg := buildSDKRegistry(fcRegistry)
	executor := sdktools.NewExecutor(sdkReg, nil, &sdktypes.ToolUseContext{
		WorkingDir: workspace,
		AbortCtx:   ctx,
	})

	// Convert Pawnix tool calls to SDK format.
	//
	// Historically a JSON parse failure here fell back to wrapping
	// the whole raw string under {"_raw": "..."} and shipping it on
	// to the executor. That mask was a footgun: the tool would then
	// see an empty `path` (or whatever the real schema expects) and
	// emit a confusing downstream error like "is a directory", while
	// the model never learned that its JSON was malformed and would
	// retry the exact same broken payload forever. Observed in
	// production with kimi-k2.6 emitting ~38 KB write_file content
	// with escape errors.
	//
	// New behaviour: surface the parse error directly back to the
	// model as an isError tool result, so the next turn can fix
	// the JSON instead of repeating the mistake. We also slog.Warn
	// with a length-capped preview so operators can spot
	// systematically misbehaving providers.
	// results stays aligned 1:1 with toolCalls so the caller can
	// pair each entry with its tool_call_id without needing a
	// secondary index. We track sdkToOrig to map back from the
	// (possibly shorter) calls slice we hand to the executor.
	results := make([]toolCallResult, len(toolCalls))
	calls := make([]sdktools.ToolCallRequest, 0, len(toolCalls))
	sdkToOrig := make([]int, 0, len(toolCalls))

	for i, tc := range toolCalls {
		var input map[string]interface{}
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &input); err != nil {
			argsPreview := tc.Function.Arguments
			const maxPreview = 256
			if len(argsPreview) > maxPreview {
				argsPreview = argsPreview[:maxPreview] + fmt.Sprintf("…(%d bytes total)", len(tc.Function.Arguments))
			}
			slog.Warn("tool call arguments are not valid JSON",
				"tool", tc.Function.Name,
				"tool_call_id", tc.ID,
				"parse_error", err.Error(),
				"raw_args_bytes", len(tc.Function.Arguments),
				"raw_args_preview", argsPreview,
			)
			msg := fmt.Sprintf(
				"tool arguments are not valid JSON: %v. The arguments string was %d bytes; common causes for long inputs are unescaped quotes, raw newlines inside string values, or truncation. Regenerate the tool call with strict JSON formatting.",
				err, len(tc.Function.Arguments),
			)
			results[i] = toolCallResult{
				toolCallID: tc.ID,
				toolName:   tc.Function.Name,
				result:     msg + "\n[Analyze the error above and try a different approach.]",
				err:        fmt.Errorf("%s", msg),
			}
			continue
		}
		calls = append(calls, sdktools.ToolCallRequest{
			ToolUseID: tc.ID,
			ToolName:  tc.Function.Name,
			Input:     input,
		})
		sdkToOrig = append(sdkToOrig, i)
	}

	start := time.Now()
	responses := executor.RunTools(ctx, calls)
	e.costTracker.AddToolDuration(time.Since(start))

	// Convert SDK responses back to Pawnix format. We use
	// sdkToOrig to write each response into its original slot in
	// `results`, so JSON-parse failures earlier in the loop don't
	// shift indices and corrupt the response/tool_call pairing.
	for sdkIdx, resp := range responses {
		if sdkIdx >= len(sdkToOrig) {
			// Defensive: SDK returned more responses than we sent.
			// Should never happen, but guard against panics.
			break
		}
		origIdx := sdkToOrig[sdkIdx]
		tc := toolCalls[origIdx]

		var resultText string
		if resp.Result != nil {
			if resp.Result.IsError {
				resultText = resp.Result.Error
				if resultText == "" && len(resp.Result.Content) > 0 {
					resultText = resp.Result.Content[0].Text
				}
				results[origIdx] = toolCallResult{
					toolCallID: resp.ToolUseID,
					toolName:   tc.Function.Name,
					result:     resultText + "\n[Analyze the error above and try a different approach.]",
					err:        fmt.Errorf("%s", resultText),
				}
				continue
			}
			// Extract text from content blocks
			var parts []string
			for _, cb := range resp.Result.Content {
				if cb.Text != "" {
					parts = append(parts, cb.Text)
				}
			}
			resultText = strings.Join(parts, "\n")
		}
		if resp.Error != nil {
			results[origIdx] = toolCallResult{
				toolCallID: resp.ToolUseID,
				toolName:   tc.Function.Name,
				result:     resultText,
				err:        resp.Error,
			}
		} else {
			results[origIdx] = toolCallResult{
				toolCallID: resp.ToolUseID,
				toolName:   tc.Function.Name,
				result:     resultText,
			}
		}
	}
	return results
}
