package gateway

import (
	"context"
	"log/slog"
	"path/filepath"
	"sync"
	"time"

	"github.com/softbreezee/claw-os/internal/agent"
	"github.com/softbreezee/claw-os/internal/bus"
	"github.com/softbreezee/claw-os/internal/channels"
	"github.com/softbreezee/claw-os/internal/config"
	"github.com/softbreezee/claw-os/internal/cron"
	"github.com/fsnotify/fsnotify"
)

// startConfigWatcher watches the config file and workspace files for changes,
// triggering a hot-reload when modifications are detected.
func (g *Gateway) startConfigWatcher(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Error("failed to create file watcher", "error", err)
		return
	}
	defer watcher.Close()

	// Watch config directory
	homeDir, err := config.HomeDir()
	if err != nil {
		slog.Error("cannot determine config dir for watcher", "error", err)
		return
	}
	configPath := filepath.Join(homeDir, "pawnix.json")
	configDir := filepath.Dir(configPath)

	if err := watcher.Add(configDir); err != nil {
		slog.Error("failed to watch config dir", "dir", configDir, "error", err)
		return
	}
	slog.Info("config watcher started", "path", configDir)

	// Also watch agent workspace directories for SOUL.md, AGENTS.md, etc.
	for _, ag := range g.agents.All() {
		wsPath := ag.WorkspacePath()
		if wsPath != "" {
			if err := watcher.Add(wsPath); err != nil {
				slog.Warn("failed to watch workspace", "path", wsPath, "error", err)
			}
		}
	}

	// Debounce: wait for writes to settle before reloading
	var debounceTimer *time.Timer
	var debounceMu sync.Mutex

	for {
		select {
		case <-ctx.Done():
			return

		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// Only react to writes and creates
			if !event.Has(fsnotify.Write) && !event.Has(fsnotify.Create) {
				continue
			}

			filename := filepath.Base(event.Name)

			// Determine what changed
			isConfig := filename == "pawnix.json"
			isWorkspaceFile := isWatchedWorkspaceFile(filename)

			if !isConfig && !isWorkspaceFile {
				continue
			}

			slog.Info("file change detected", "file", event.Name, "op", event.Op.String())

			// Debounce: many editors write multiple events per save
			debounceMu.Lock()
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			debounceTimer = time.AfterFunc(500*time.Millisecond, func() {
				if isConfig {
					g.reloadConfig()
				} else if isWorkspaceFile {
					g.reloadWorkspaceFile(event.Name, filename)
				}
			})
			debounceMu.Unlock()

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			slog.Error("file watcher error", "error", err)
		}
	}
}

// isWatchedWorkspaceFile returns true if this is a file we hot-reload.
func isWatchedWorkspaceFile(filename string) bool {
	switch filename {
	case "SOUL.md", "AGENTS.md", "IDENTITY.md", "TOOLS.md",
		"BOOTSTRAP.md", "HEARTBEAT.md", "MEMORY.md", "USER.md",
		"agent.json":
		return true
	}
	return false
}

// reloadConfig reloads the main config file and applies changes.
func (g *Gateway) reloadConfig() {
	slog.Info("hot-reloading config...")

	newCfg, err := config.Load()
	if err != nil {
		slog.Error("hot-reload: failed to load config", "error", err)
		return
	}

	// 1. Update LLM provider if changed
	g.reloadProvider(newCfg)

	// 2. Update agent configs (model, temperature, etc.)
	g.reloadAgents(newCfg)

	// 3. Update bindings
	g.mu.Lock()
	g.bindings = newCfg.Bindings
	g.config = newCfg
	g.mu.Unlock()

	// 4. Update cron jobs
	g.reloadCron(newCfg)

	// 5. Update teams and group context
	g.reloadTeams(newCfg)

	slog.Info("hot-reload complete ✅")
}

// resolveProviderCfg picks the active provider config from a Config.
// reloadProvider rebuilds the provider registry whenever any provider
// entry changes, then asks the agent manager to re-route every agent
// to the appropriate provider based on its (possibly also reloaded)
// model field.
//
// We don't try to be clever about diffing individual providers — the
// previous "is the default provider's apiKey/apiBase the same?" check
// missed the case where a *non-default* provider was added/removed,
// which silently broke routing for any agent that targeted it. A full
// rebuild is cheap (it's just constructing a few HTTP clients) and
// always correct.
func (g *Gateway) reloadProvider(newCfg *config.Config) {
	newReg := buildProviderRegistry(newCfg)
	if reg := g.agents.Registry(); reg != nil {
		reg.Replace(newReg)
		slog.Info("hot-reload: provider registry rebuilt", "providers", reg.Names())
	}
}

// reloadAgents updates agent model settings from new config and then
// re-points each agent at its (possibly newly-relevant) provider.
//
// Order matters: UpdateConfig must run before Rewire so Rewire sees
// the new model field when computing which provider to use.
func (g *Gateway) reloadAgents(newCfg *config.Config) {
	resolved := config.ResolveAgents(newCfg)
	for _, rc := range resolved {
		ag := g.agents.AgentByID(rc.ID)
		if ag == nil {
			slog.Info("hot-reload: new agent detected (restart required to add)", "id", rc.ID)
			continue
		}
		ag.UpdateConfig(rc)
		slog.Info("hot-reload: agent config updated", "id", rc.ID, "model", rc.Model)
	}
	// Re-resolve provider for every agent — model field may have moved
	// between providers (e.g. "deepseek/x" -> "openai/y").
	g.agents.Rewire()
}

// reloadCron updates the cron scheduler with new jobs.
func (g *Gateway) reloadCron(newCfg *config.Config) {
	var cronJobs []cron.Job
	for _, cj := range newCfg.CronJobs {
		cronJobs = append(cronJobs, cron.Job{
			Name:     cj.Name,
			Type:     cron.JobType(cj.Type),
			Schedule: cj.Schedule,
			AgentID:  cj.AgentID,
			Channel:  cj.Channel,
			ChatID:   cj.ChatID,
			Message:  cj.Message,
		})
	}
	g.scheduler.UpdateJobs(cronJobs)
	slog.Info("hot-reload: cron jobs updated", "count", len(cronJobs))
}

// reloadTeams updates team config and group context.
func (g *Gateway) reloadTeams(newCfg *config.Config) {
	teams := newCfg.Teams
	if teams == nil {
		teams = make(map[string]config.TeamEntry)
	}
	g.mu.Lock()
	g.teams = teams
	g.mu.Unlock()

	// Refresh group context for agents in teams
	for _, team := range teams {
		for _, agentID := range team.Agents {
			ag := g.agents.AgentByID(agentID)
			if ag == nil {
				continue
			}
			var teammates []string
			for _, otherID := range team.Agents {
				if otherID != agentID {
					if uname, ok := g.botUsernames[otherID]; ok {
						teammates = append(teammates, "@"+uname)
					} else {
						teammates = append(teammates, otherID)
					}
				}
			}
			if botUname, ok := g.botUsernames[agentID]; ok {
				ag.SetGroupContext(&agent.GroupContext{
					BotUsername: botUname,
					Teammates:  teammates,
				})
			}
		}
	}
}

// reloadWorkspaceFile handles changes to agent workspace files (SOUL.md, etc.)
func (g *Gateway) reloadWorkspaceFile(fullPath, filename string) {
	wsDir := filepath.Dir(fullPath)

	// Find which agent owns this workspace
	for _, ag := range g.agents.All() {
		if ag.WorkspacePath() == wsDir {
			ag.ReloadWorkspaceFiles()
			slog.Info("hot-reload: workspace file updated", "agent", ag.Name(), "file", filename)
			return
		}
	}
	slog.Warn("hot-reload: changed file doesn't match any agent workspace", "path", fullPath)
}

// Minimal channel hot-reload: new channels require restart,
// but we can update existing channel configs.
func (g *Gateway) reloadChannels(newCfg *config.Config) {
	// For now, log which channels changed. Full channel hot-reload
	// (adding/removing Telegram bots) requires restart.
	for name, chCfg := range newCfg.Channels {
		if !chCfg.Enabled {
			continue
		}
		slog.Info("hot-reload: channel config noted (restart needed for new channels)", "channel", name)
	}
}

// registerChannelsHot would add new channels at runtime.
// Currently not implemented — new channels require restart.
func registerChannelsHot(cfg *config.Config, mb *bus.MessageBus, chanMgr *channels.Manager) error {
	_ = cfg
	_ = mb
	_ = chanMgr
	return nil
}
