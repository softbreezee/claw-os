package gateway

import (
	"context"
	"strings"
	"time"

	"github.com/softbreezee/claw-os/internal/bus"
	"github.com/softbreezee/claw-os/internal/config"
	"github.com/softbreezee/claw-os/internal/store"
)

// makeRecipientResolver builds the lookup function passed to the
// notify tool. It walks AccountConfig.MyChatID → ChannelConfig.MyChatID
// for the requested (channel, accountID) pair so any agent can ask
// "where do I send a message to the user on telegram?" without
// knowing the bot config layout.
//
// Returns ("", false) when nothing is configured — the notify tool
// surfaces this back to the LLM as a "please ask the user to set it"
// error so the failure mode is visible instead of silently dropping.
func makeRecipientResolver(cfg *config.Config) func(channel, accountID string) (string, bool) {
	return func(channel, accountID string) (string, bool) {
		if cfg == nil {
			return "", false
		}
		ch, ok := cfg.Channels[channel]
		if !ok || !ch.Enabled {
			return "", false
		}
		if accountID != "" {
			if a, ok := ch.Accounts[accountID]; ok && a.MyChatID != "" {
				return a.MyChatID, true
			}
		}
		if ch.MyChatID != "" {
			return ch.MyChatID, true
		}
		// Last-ditch: if there's exactly one account with a MyChatID,
		// use it. Mirrors chanMgr's outbound single-account fallback.
		if accountID == "" && len(ch.Accounts) == 1 {
			for _, a := range ch.Accounts {
				if a.MyChatID != "" {
					return a.MyChatID, true
				}
			}
		}
		return "", false
	}
}

// makeNotificationWriter wraps store.SaveNotification with the source
// defaulted to "agent" so the inbox can distinguish self-initiated
// agent pings from cron-fired ones.
func makeNotificationWriter(st store.Store) func(ctx context.Context, agentID, source, title, body string) (string, error) {
	return func(ctx context.Context, agentID, source, title, body string) (string, error) {
		if st == nil {
			return "", nil
		}
		if source == "" {
			source = "agent"
		}
		rec := &store.NotificationRecord{
			ID:        newCronJobID(),
			TenantID:  store.DefaultTenantID,
			AgentID:   agentID,
			Source:    source,
			Title:     title,
			Body:      body,
			CreatedAt: time.Now().UTC(),
		}
		if err := st.SaveNotification(ctx, store.DefaultTenantID, rec); err != nil {
			return "", err
		}
		return rec.ID, nil
	}
}

// writeOriginNotification persists a NotificationRecord summarising
// the result of an internal-origin agent run (cron / webhook / etc).
// This is the entry point that turns "scheduler fires → agent runs"
// into something the user actually sees in the inbox.
//
// Title heuristic: the first line of the agent's REPLY (not the
// cron trigger prompt). The trigger is system plumbing — what the
// user wants to see in the inbox row preview is the agent's actual
// message ("☀️ 早安~"), not the prompt that produced it ("发一句
// 简短温暖的鼓励语. 根据当前时间..."). Body is the full reply.
//
// Source/Link policy is per-origin so the inbox can group/filter and
// "open" a notification can deep-link the user back to the source
// (cron job page, webhook config, etc).
func writeOriginNotification(st store.Store, agentID string, in bus.InboundMessage, reply string) error {
	source := in.Origin
	if source == "" {
		source = "internal"
	}
	rec := &store.NotificationRecord{
		ID:        newCronJobID(), // reuse the UUID helper
		TenantID:  store.DefaultTenantID,
		AgentID:   agentID,
		Source:    source,
		Title:     summariseTitle(source, reply, in.Text),
		Body:      strings.TrimSpace(reply),
		CreatedAt: time.Now().UTC(),
	}
	switch source {
	case "cron":
		// Future: thread cron job ID through InboundMessage so we can
		// set rec.SourceID and link directly to the row in the Cron
		// Jobs page. For now a generic deep link is enough.
		rec.Link = "/cron"
	case "webhook":
		rec.Link = "/settings"
	}
	return st.SaveNotification(context.Background(), store.DefaultTenantID, rec)
}

// summariseTitle picks a short, human-meaningful title for the inbox
// row. Resolution order:
//   1. First line of the agent reply (the actual content the user
//      came to see).
//   2. First line of the trigger prompt (fallback when the agent
//      somehow returned an empty reply — debugging breadcrumb).
//   3. A generic per-source label.
//
// Capped at ~80 visible chars + an ellipsis. The source icon
// rendered by the UI already conveys the origin, but we keep a
// leading emoji so the title still reads sensibly when copied
// out of the notification (toast, browser title, etc).
func summariseTitle(source, primary, fallback string) string {
	pick := strings.TrimSpace(primary)
	if pick == "" {
		pick = strings.TrimSpace(fallback)
	}
	if pick == "" {
		switch source {
		case "cron":
			return "Cron job triggered"
		case "webhook":
			return "Webhook delivered"
		default:
			return "Agent message"
		}
	}
	// First line only — multi-line replies are common and the row
	// preview only has space for one line.
	if i := strings.IndexAny(pick, "\r\n"); i > 0 {
		pick = pick[:i]
	}
	const maxLen = 80
	if len([]rune(pick)) > maxLen {
		runes := []rune(pick)
		pick = string(runes[:maxLen]) + "…"
	}
	switch source {
	case "cron":
		return "⏰ " + pick
	case "webhook":
		return "🌐 " + pick
	default:
		return pick
	}
}
