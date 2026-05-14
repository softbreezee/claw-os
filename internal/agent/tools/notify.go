package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/softbreezee/claw-os/internal/bus"
)

// RecipientResolver returns the user's "send to me" address (chat ID,
// user ID, email, openid, ...) for a given channel + accountID.
//
// Resolution order is implementation-defined but typically:
//   1. AccountConfig.MyChatID for that specific (channel, accountID)
//   2. ChannelConfig.MyChatID (channel-level fallback)
//   3. Empty string (unconfigured)
//
// Injected from the agent layer via RegisterNotifyTool so the tools
// package doesn't need to import internal/config and we keep the
// agent → tools dependency one-way.
type RecipientResolver func(channel, accountID string) (chatID string, ok bool)

// OutboundSender pushes an outbound message onto the bus. The agent
// layer injects this rather than letting tools talk to the bus
// directly so tools stay testable in isolation (notify_test can pass
// a capturing fake instead of a real bus.MessageBus).
type OutboundSender func(msg bus.OutboundMessage)

// NotificationWriter persists a notification record. Used by the
// notify tool when the resolved channel is the in-app inbox (web /
// empty) — sending an outbound message there would be silently
// dropped because no IM handler is registered.
//
// Returning the ID lets future iterations build deep links back to
// the notification in the UI (today we don't expose it but the hook
// is here for free).
type NotificationWriter func(ctx context.Context, agentID, source, title, body string) (id string, err error)

type notifyArgs struct {
	Text    string `json:"text"`
	Channel string `json:"channel"`
	Title   string `json:"title"`
}

// RegisterNotifyTool wires the notify(text, channel?, title?) tool
// into the registry. The agent uses this to push an unsolicited
// message back to the user — the OS-level "any agent can ping me"
// primitive on top of which cron, watchers, etc are built.
//
// channel="" or "web" → write a Notification (Inbox + browser toast).
// channel="telegram"/"slack"/... → look up the user's address via
// the resolver and send through the corresponding IM bot.
func RegisterNotifyTool(
	r *Registry,
	agentID string,
	resolve RecipientResolver,
	send OutboundSender,
	writeNotif NotificationWriter,
) {
	r.Register("notify",
		"Push an unsolicited message to the user. Use this when something interesting happens "+
			"that the user should see right now without you waiting for them to ask. "+
			"Default channel ('' or 'web') goes to the in-app Inbox + browser toast. "+
			"Specify channel='telegram'/'slack'/'discord' to deliver via that IM instead — "+
			"the recipient address is looked up from the user's channel configuration, you "+
			"do NOT need to know their chat ID. Returns 'delivered: <where>' on success.",
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"text": map[string]interface{}{
					"type":        "string",
					"description": "The message body to send.",
				},
				"channel": map[string]interface{}{
					"type":        "string",
					"description": "Optional. Delivery channel: '' (default, web Inbox), 'telegram', 'slack', 'discord'. Use a real channel for time-sensitive things; default for routine updates.",
				},
				"title": map[string]interface{}{
					"type":        "string",
					"description": "Optional. Short title for Inbox display (ignored by IM channels).",
				},
			},
			"required": []string{"text"},
		},
		makeNotify(agentID, resolve, send, writeNotif),
	)
}

func makeNotify(
	agentID string,
	resolve RecipientResolver,
	send OutboundSender,
	writeNotif NotificationWriter,
) ToolFunc {
	return func(ctx context.Context, raw json.RawMessage) (string, error) {
		var args notifyArgs
		if err := json.Unmarshal(raw, &args); err != nil {
			return "", fmt.Errorf("parse args: %w", err)
		}
		if strings.TrimSpace(args.Text) == "" {
			return "", fmt.Errorf("text is required")
		}

		channel := strings.ToLower(strings.TrimSpace(args.Channel))

		// Inbox path — empty / web go to the notifications store so
		// they show up in the dashboard sidebar badge + browser toast.
		if channel == "" || channel == "web" {
			if writeNotif == nil {
				return "", fmt.Errorf("notification store unavailable")
			}
			title := strings.TrimSpace(args.Title)
			if title == "" {
				title = firstLine(args.Text)
			}
			if _, err := writeNotif(ctx, agentID, "agent", title, args.Text); err != nil {
				return "", fmt.Errorf("write notification: %w", err)
			}
			return "delivered: Inbox", nil
		}

		// IM path — resolve the user's address for this channel.
		if resolve == nil {
			return "", fmt.Errorf("no recipient resolver configured for channel %q", channel)
		}
		chatID, ok := resolve(channel, "")
		if !ok || chatID == "" {
			return "", fmt.Errorf("no 'My chat ID' configured for channel %q — ask the user to set it under Channels in the dashboard", channel)
		}
		if send == nil {
			return "", fmt.Errorf("outbound sender unavailable")
		}
		send(bus.OutboundMessage{
			Channel: channel,
			ChatID:  chatID,
			Text:    args.Text,
			// AccountID intentionally left empty — chanMgr.routeOutbound
			// has a single-account fallback that picks the only
			// registered bot for this channel type. For multi-bot
			// setups the resolver should grow to return AccountID too.
		})
		return "delivered: " + channel, nil
	}
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexAny(s, "\r\n"); i > 0 {
		s = s[:i]
	}
	const max = 80
	if len([]rune(s)) > max {
		runes := []rune(s)
		s = string(runes[:max]) + "…"
	}
	return s
}
