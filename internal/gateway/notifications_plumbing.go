package gateway

import (
	"context"
	"strings"
	"time"

	"github.com/softbreezee/claw-os/internal/bus"
	"github.com/softbreezee/claw-os/internal/store"
)

// writeOriginNotification persists a NotificationRecord summarising
// the result of an internal-origin agent run (cron / webhook / etc).
// This is the entry point that turns "scheduler fires → agent runs"
// into something the user actually sees in the inbox.
//
// Title heuristic: the first line of the trigger prompt (truncated)
// gives the user enough context to know "ah, that's my drink-water
// reminder". Body is the agent's reply.
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
		Title:     summariseTitle(source, in.Text),
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
// row. We prefer the first line of the trigger text (the cron job
// message), capped to ~60 chars. Falls back to a per-source label
// when the trigger is empty.
func summariseTitle(source, trigger string) string {
	t := strings.TrimSpace(trigger)
	if t == "" {
		switch source {
		case "cron":
			return "Cron job triggered"
		case "webhook":
			return "Webhook delivered"
		default:
			return "Agent message"
		}
	}
	// First line only — multi-line cron messages happen but the row
	// preview only has space for one line.
	if i := strings.IndexAny(t, "\r\n"); i > 0 {
		t = t[:i]
	}
	const maxLen = 80
	if len([]rune(t)) > maxLen {
		runes := []rune(t)
		t = string(runes[:maxLen]) + "…"
	}
	switch source {
	case "cron":
		return "⏰ " + t
	case "webhook":
		return "🌐 " + t
	default:
		return t
	}
}
