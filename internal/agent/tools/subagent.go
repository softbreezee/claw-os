package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fastclaw-ai/fastclaw/internal/bus"
)

// SubAgentSpawner is the interface for spawning sub-agents.
type SubAgentSpawner interface {
	// SpawnSubAgent sends a task to another agent and returns its response.
	SpawnSubAgent(ctx context.Context, agentID string, msg bus.InboundMessage) string
}

// AttachmentsGetter pulls the current turn's attachments out of the
// context. Provided as a function rather than via a direct call into
// the agent package to avoid an import cycle (agent imports tools,
// not the other way around).
//
// Returns nil/empty when no attachments are present, which is the
// common case.
type AttachmentsGetter func(ctx context.Context) []bus.Attachment

type spawnSubagentArgs struct {
	AgentID            string `json:"agentId"`
	Task               string `json:"task"`
	ForwardAttachments bool   `json:"forward_attachments,omitempty"`
}

// RegisterSubAgent registers the spawn_subagent tool.
//
// `getAttachments` may be nil for legacy callers that don't need to
// surface attachment forwarding; the tool will then act as if the
// caller had no attachments (forward_attachments=true is silently a
// no-op).
func RegisterSubAgent(r *Registry, spawner SubAgentSpawner, callerAgentID string, getAttachments AttachmentsGetter) {
	r.Register("spawn_subagent",
		"Spawn another agent as a sub-task and return its response. "+
			"Use this to delegate work to specialised agents (e.g. a vision-capable "+
			"agent for image analysis, or a reasoning model for hard decisions). "+
			"Set forward_attachments=true to pass the user's attached files (images "+
			"etc.) along to the sub-agent — useful when the sub-agent is the one "+
			"that can actually process them.",
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"agentId": map[string]interface{}{
					"type":        "string",
					"description": "The ID of the agent to spawn (e.g. \"zz\", \"dp-v4-pro\")",
				},
				"task": map[string]interface{}{
					"type":        "string",
					"description": "The message/prompt to send to the sub-agent",
				},
				"forward_attachments": map[string]interface{}{
					"type":        "boolean",
					"description": "When true, forward the user's currently-attached files (e.g. uploaded images) to the sub-agent so a vision model can see them. Default: false.",
				},
			},
			"required": []string{"agentId", "task"},
		},
		makeSubAgentTool(spawner, callerAgentID, getAttachments),
	)
}

func makeSubAgentTool(spawner SubAgentSpawner, callerAgentID string, getAttachments AttachmentsGetter) ToolFunc {
	return func(ctx context.Context, rawArgs json.RawMessage) (string, error) {
		var args spawnSubagentArgs
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return "", fmt.Errorf("parse args: %w", err)
		}

		if args.AgentID == "" {
			return "", fmt.Errorf("agentId is required")
		}
		if args.Task == "" {
			return "", fmt.Errorf("task is required")
		}
		if args.AgentID == callerAgentID {
			return "", fmt.Errorf("cannot spawn yourself as a sub-agent")
		}

		// Each delegation gets a unique session so history doesn't accumulate
		// across repeated calls and the sub-agent starts with a clean slate.
		sessionID := fmt.Sprintf("subagent-%s-%s-%d", callerAgentID, args.AgentID, time.Now().UnixNano())
		msg := bus.InboundMessage{
			Channel:  "subagent",
			ChatID:   sessionID,
			UserID:   callerAgentID,
			Text:     args.Task,
			PeerKind: "dm",
		}

		// Attachment forwarding: lift the current caller's attachments
		// from ctx and re-attach them to the sub-agent's inbound message.
		// Note we forward ALL of the caller's attachments — the LLM has
		// no per-file granularity in the tool schema today (intentional;
		// keeps the tool dead-simple). If the caller wants to forward a
		// subset, it can spawn the sub-agent multiple times.
		if args.ForwardAttachments && getAttachments != nil {
			if atts := getAttachments(ctx); len(atts) > 0 {
				msg.Attachments = atts
			}
		}

		result := spawner.SpawnSubAgent(ctx, args.AgentID, msg)
		return result, nil
	}
}
