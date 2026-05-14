package tools

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"

	"github.com/softbreezee/claw-os/internal/store"
)

// ChatOriginGetter pulls the current chat's delivery info out of ctx.
// We accept it as an injected function rather than importing
// internal/agent here to avoid the agent → tools → agent import cycle.
// Gateway/agent code wires the real getter at registration time.
type ChatOriginGetter func(ctx context.Context) (channel, accountID, chatID string)

type createCronJobArgs struct {
	Name     string `json:"name"`
	Schedule string `json:"schedule"`
	Message  string `json:"message"`
	Type     string `json:"type"`
	// Channel optionally targets a specific delivery channel
	// ("telegram", "slack", …). When blank, the scheduler treats the
	// trigger as an in-process web-chat reminder routed to the agent
	// directly. Most LLM-initiated cron jobs leave this blank.
	Channel string `json:"channel"`
	ChatID  string `json:"chatId"`
}

type deleteCronJobArgs struct {
	ID string `json:"id"`
}

// RegisterCronTools registers cron job management tools.
//
// channel / chatID are the FALLBACK delivery target — used when
// originGetter is nil or returns empty values. In practice the
// gateway always passes channel="" / chatID="" + a non-nil
// originGetter, which means cron jobs created from a Telegram chat
// default to delivering back through Telegram, while web-chat ones
// default to the in-app Inbox. Per-call args.Channel still wins on
// top of both.
//
// recipientResolver is optional. When set, it lets create_cron_job
// auto-fill the chatID for cross-channel jobs: "in the web chat, ask
// the agent to send the cron reminder to telegram" works without the
// LLM having to know the user's telegram chat ID — the resolver pulls
// it from cfg.Channels[telegram].MyChatID. Pass nil if no resolver
// (cron tool will then require an explicit chatID arg for non-web
// channels).
func RegisterCronTools(r *Registry, st store.Store, tenantID, agentID, channel, chatID string, originGetter ChatOriginGetter, recipientResolver RecipientResolver) {
	r.Register("create_cron_job",
		"Create a scheduled task that fires the given prompt back at this agent on a schedule. "+
			"Prefer this over writing shell scripts to the system crontab — jobs created here "+
			"are visible in the Pawnix UI's Cron Jobs page and persist across restarts.",
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Short human-readable task name (shown in the UI). Example: 'morning-summary'.",
				},
				"schedule": map[string]interface{}{
					"type":        "string",
					"description": "Schedule string, format depends on `type`: cron expression like '0 9 * * *' (type=cron), Go duration like '30m'/'2h' (type=interval), HH:MM 24h (type=exact), or RFC3339 datetime (type=once).",
				},
				"message": map[string]interface{}{
					"type":        "string",
					"description": "The prompt that will be re-sent to this agent when the schedule fires.",
				},
				"type": map[string]interface{}{
					"type":        "string",
					"description": "Schedule type. One of: 'cron' (default), 'interval', 'exact', 'once'.",
				},
			},
			"required": []string{"name", "schedule", "message"},
		},
		makeCreateCronJob(st, tenantID, agentID, channel, chatID, originGetter, recipientResolver),
	)

	r.Register("list_cron_jobs",
		"List all scheduled tasks for this agent.",
		map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		makeListCronJobs(st, tenantID, agentID),
	)

	r.Register("delete_cron_job",
		"Delete a scheduled task by ID.",
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type":        "string",
					"description": "The cron job ID to delete",
				},
			},
			"required": []string{"id"},
		},
		makeDeleteCronJob(st, tenantID),
	)
}

func makeCreateCronJob(st store.Store, tenantID, agentID, defaultChannel, defaultChatID string, originGetter ChatOriginGetter, recipientResolver RecipientResolver) ToolFunc {
	return func(ctx context.Context, rawArgs json.RawMessage) (string, error) {
		var args createCronJobArgs
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return "", fmt.Errorf("parse args: %w", err)
		}
		if args.Name == "" || args.Schedule == "" || args.Message == "" {
			return "", fmt.Errorf("name, schedule, and message are required")
		}
		jobType := args.Type
		if jobType == "" {
			jobType = "cron"
		}

		// Resolution order for channel/accountID/chatID:
		//   1. Explicit args from the tool call (LLM said "send to
		//      telegram chat 999")
		//   2. Current chat origin from ctx (the user is talking to
		//      the agent in Telegram → reminder also goes to Telegram)
		//   3. Defaults bound at register time (legacy fallback)
		channel := args.Channel
		chatID := args.ChatID
		var accountID string
		if originGetter != nil {
			oc, oa, oid := originGetter(ctx)
			if channel == "" {
				channel = oc
			}
			if chatID == "" {
				chatID = oid
			}
			accountID = oa
		}
		if channel == "" {
			channel = defaultChannel
		}
		if chatID == "" {
			chatID = defaultChatID
		}

		// Cross-channel fallback: when the LLM said "send to telegram"
		// from a web conversation, originGetter returns ChatID from
		// the web session (useless for telegram delivery). Fall back
		// to the user's "MyChatID" config for the target channel so
		// cron-on-IM works without the LLM needing to know addresses.
		// Skip for "" / "web" — those go to the Inbox (no chat ID needed).
		if (chatID == "" || (channel != "" && channel != "web" && chatID == defaultChatID)) &&
			channel != "" && channel != "web" && recipientResolver != nil {
			if resolved, ok := recipientResolver(channel, accountID); ok {
				chatID = resolved
			}
		}

		// Inbox normalisation: cron jobs delivered to the in-app
		// Inbox don't need a chat/account address. Clearing these
		// avoids two real problems:
		//   1. UI bug — the Cron Jobs page shows "main → s-…" text
		//      under Delivery for inbox jobs, which is misleading.
		//   2. Migration headache — if we ever change Inbox storage,
		//      we don't want to find junk session IDs in the chatId
		//      column from cron jobs created in a web chat session.
		// The originGetter happily fills chatID with the web session
		// key when the user creates an inbox cron from /chat, which
		// is wrong for inbox semantics (the inbox is keyed by
		// notification id, not chat id).
		if channel == "" || channel == "web" {
			chatID = ""
			accountID = ""
		}

		id := generateUUID()
		now := time.Now()
		// NextRun=now means "fire on the scheduler's next poll cycle";
		// after first fire the scheduler computes the proper next
		// tick from the schedule. This means newly-created jobs may
		// fire up to ~1 minute late, which we accept in exchange for
		// not duplicating cron-expression parsing inside the tool.
		job := &store.CronJobRecord{
			ID:        id,
			TenantID:  tenantID,
			AgentID:   agentID,
			Name:      args.Name,
			Type:      jobType,
			Schedule:  args.Schedule,
			Message:   args.Message,
			Channel:   channel,
			ChatID:    chatID,
			AccountID: accountID,
			Timezone:  "Local",
			Enabled:   true,
			NextRun:   &now,
			CreatedAt: now,
		}

		if err := st.SaveCronJob(ctx, tenantID, job); err != nil {
			return "", fmt.Errorf("save cron job: %w", err)
		}

		// Tell the LLM where the reminder will land so it can echo
		// the right thing back to the user ("I'll ping you on
		// Telegram every morning"). The "via" line is omitted for
		// web-Inbox jobs since the user is already in the app.
		via := ""
		switch channel {
		case "", "web":
			via = "Inbox"
		default:
			via = channel
			if accountID != "" {
				via = via + "/" + accountID
			}
		}
		return fmt.Sprintf("Cron job created successfully.\nID: %s\nName: %s\nSchedule: %s\nType: %s\nDelivery: %s\nIt will fire on the next scheduler poll (within 60s).", id, args.Name, args.Schedule, jobType, via), nil
	}
}

func makeListCronJobs(st store.Store, tenantID, agentID string) ToolFunc {
	return func(ctx context.Context, rawArgs json.RawMessage) (string, error) {
		jobs, err := st.ListCronJobs(ctx, tenantID)
		if err != nil {
			return "", fmt.Errorf("list cron jobs: %w", err)
		}

		var filtered []store.CronJobRecord
		for _, j := range jobs {
			if j.AgentID == agentID {
				filtered = append(filtered, j)
			}
		}

		if len(filtered) == 0 {
			return "No cron jobs found for this agent.", nil
		}

		data, _ := json.MarshalIndent(filtered, "", "  ")
		return string(data), nil
	}
}

func makeDeleteCronJob(st store.Store, tenantID string) ToolFunc {
	return func(ctx context.Context, rawArgs json.RawMessage) (string, error) {
		var args deleteCronJobArgs
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return "", fmt.Errorf("parse args: %w", err)
		}
		if args.ID == "" {
			return "", fmt.Errorf("id is required")
		}
		if err := st.DeleteCronJob(ctx, tenantID, args.ID); err != nil {
			return "", fmt.Errorf("delete cron job: %w", err)
		}
		return fmt.Sprintf("Cron job %s deleted.", args.ID), nil
	}
}

func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
