package agent

import (
	"context"
	"log/slog"
	"time"

	"github.com/softbreezee/claw-os/internal/bus"
)

const (
	// DefaultHeartbeatInterval is the default interval between heartbeat checks.
	DefaultHeartbeatInterval = 30 * time.Minute
)

// HeartbeatConfig holds heartbeat configuration.
type HeartbeatConfig struct {
	Interval time.Duration
}

// Heartbeat runs periodic checks and triggers agent actions.
type Heartbeat struct {
	agent    *Agent
	bus      *bus.MessageBus
	interval time.Duration
}

// NewHeartbeat creates a new heartbeat for the given agent.
func NewHeartbeat(ag *Agent, mb *bus.MessageBus, interval time.Duration) *Heartbeat {
	if interval <= 0 {
		interval = DefaultHeartbeatInterval
	}
	return &Heartbeat{
		agent:    ag,
		bus:      mb,
		interval: interval,
	}
}

// Start begins the heartbeat goroutine. It blocks until ctx is cancelled.
func (hb *Heartbeat) Start(ctx context.Context) {
	slog.Info("heartbeat started",
		"agent", hb.agent.Name(),
		"interval", hb.interval,
	)

	ticker := time.NewTicker(hb.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("heartbeat stopped", "agent", hb.agent.Name())
			return
		case <-ticker.C:
			hb.tick(ctx)
		}
	}
}

func (hb *Heartbeat) tick(ctx context.Context) {
	slog.Info("heartbeat tick", "agent", hb.agent.Name())

	// Trigger memory consolidation from recent HISTORY.md entries.
	hb.updateMemory()

	// Sniff embedding-pipeline health. Cheap (one indexed COUNT) and
	// catches the silent-failure modes described in
	// docs/memory-verification.md — ones that would otherwise leave
	// the agent appearing to "not remember anything" with no visible
	// error.
	hb.checkMemoryHealth(ctx)
}

func (hb *Heartbeat) updateMemory() {
	slog.Info("heartbeat: triggering memory update", "agent", hb.agent.Name())
	hb.agent.memory.ReviewAndUpdateMemory(hb.agent.workspace())
}

// memoryHealthRecentNullThresholdPct is the line below which we
// emit a warn-level log. Two consecutive ticks (~1h) under this
// threshold means the embedding pipeline is broken. The threshold
// is intentionally permissive — we'd rather miss a flaky single tick
// than spam logs when the LLM extractor genuinely produced zero new
// facts in 30m. Tune via experience, not theory.
const memoryHealthRecentNullThresholdPct = 50

func (hb *Heartbeat) checkMemoryHealth(ctx context.Context) {
	store := hb.agent.memory.PGStore()
	if store == nil {
		return // file-only mode; nothing to verify
	}
	agentID := hb.agent.memory.AgentID()
	if agentID == "" {
		return
	}
	h, err := store.HealthStats(ctx, agentID)
	if err != nil {
		slog.Warn("heartbeat: memory health probe failed",
			"agent", hb.agent.Name(),
			"error", err)
		return
	}
	recent := h.RecentCoveragePct()
	slog.Info("heartbeat: memory health",
		"agent", hb.agent.Name(),
		"total", h.Total,
		"with_embed", h.WithEmbedding,
		"recent_total", h.RecentTotal,
		"recent_embed", h.RecentEmbedded,
		"recent_coverage_pct", recent)
	// recent == -1 means zero rows in the last 24h — that's normal
	// after a quiet period and not a signal worth alarming on.
	if recent >= 0 && recent < memoryHealthRecentNullThresholdPct && h.RecentTotal >= 3 {
		slog.Warn("heartbeat: memory embedding coverage dropping — check embed model / API key",
			"agent", hb.agent.Name(),
			"recent_coverage_pct", recent,
			"recent_total", h.RecentTotal,
			"hint", "verify cfg.Memory.EmbedModel is set, embed provider registered, API key valid")
	}
}
