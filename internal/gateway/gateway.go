package gateway

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/softbreezee/claw-os/internal/agent"
	"github.com/softbreezee/claw-os/internal/bus"
	"github.com/softbreezee/claw-os/internal/channels"
	"github.com/softbreezee/claw-os/internal/config"
	"github.com/softbreezee/claw-os/internal/cron"
	"github.com/softbreezee/claw-os/internal/plugin"
	"github.com/softbreezee/claw-os/internal/provider"
	"github.com/softbreezee/claw-os/internal/store"
	pgstore "github.com/softbreezee/claw-os/internal/store/pg"
	"github.com/softbreezee/claw-os/internal/taskqueue"
	"github.com/softbreezee/claw-os/internal/webhook"
)

// Gateway is the main orchestrator that starts all services.
type Gateway struct {
	config       *config.Config
	bus          *bus.MessageBus
	agents       *agent.Manager
	chanMgr      *channels.Manager
	bindings     []config.Binding
	botUsernames map[string]string           // agentID -> bot username
	teams        map[string]config.TeamEntry // team name -> team config
	mu           sync.RWMutex
	dedup        sync.Map                    // dedup key -> dedupEntry
	heartbeats   []*agent.Heartbeat
	scheduler    *cron.Scheduler
	webhookSrv   *webhook.Server
	pluginMgr    *plugin.Manager
	taskQueue    *taskqueue.Queue
	pgDB         *pgstore.DB // nil when storage.type != "postgres"

	// store is the unified persistence layer. As of v0.2.x it is the
	// single source of truth for cron jobs (UI handlers, agent tools
	// and the scheduler all read/write through here). Other subsystems
	// — chat tasks, agent sessions, memory — are gradually moving to
	// the same store; until that's complete several legacy files in
	// ~/.pawnix still exist alongside it.
	store store.Store
}

// New creates a new Gateway with multi-agent support.
func New(cfg *config.Config) (*Gateway, error) {
	mb := bus.New()

	// Build the provider registry: one Provider instance per entry
	// under cfg.Providers. Agent->provider routing is driven by the
	// "provider/" prefix on each agent's model field; see
	// provider.Registry for the resolution rules.
	registry := buildProviderRegistry(cfg)

	// Resolve agent configs
	resolved := config.ResolveAgents(cfg)

	// Create agent manager
	agentMgr, err := agent.NewManager(resolved, registry, mb)
	if err != nil {
		return nil, err
	}

	slog.Info("agents loaded", "count", len(resolved), "names", agentMgr.Names())

	// Connect to PostgreSQL if configured and inject PG backend into each agent.
	var pgDB *pgstore.DB
	if cfg.Storage.Type == "postgres" && cfg.Storage.DSN != "" {
		db, err := pgstore.Open(context.Background(), cfg.Storage.DSN)
		if err != nil {
			slog.Warn("pg: failed to connect, continuing with file storage", "error", err)
		} else {
			if err := db.Migrate(context.Background()); err != nil {
				slog.Warn("pg: migration failed", "error", err)
				db.Close()
			} else {
				pgDB = db
				slog.Info("pg: storage backend active")
				sessionStore := pgstore.NewSessionStore(db)
				memStore := pgstore.NewMemoryStore(db)
				pgBackend := &agent.PGBackend{
					SessionStore:    sessionStore,
					MemoryStore:     memStore,
					DBQuerier:       db,
					SchemaRegistrar: db,
				}
				for _, ag := range agentMgr.All() {
					ag.SetPGBackend(pgBackend)
				}
			}
		}
	}

	// Create channel manager and register channel instances
	chanMgr := channels.NewManager(mb)

	if err := registerChannels(cfg, mb, chanMgr); err != nil {
		return nil, err
	}

	// Build agentID -> botUsername mapping from bindings + channel manager
	botUsernames := buildBotUsernames(cfg.Bindings, chanMgr)
	if len(botUsernames) > 0 {
		slog.Info("bot username mappings", "map", botUsernames)
	}

	teams := cfg.Teams
	if teams == nil {
		teams = make(map[string]config.TeamEntry)
	}

	// Set up group context for agents in teams
	for _, team := range teams {
		for _, agentID := range team.Agents {
			ag := agentMgr.AgentByID(agentID)
			if ag == nil {
				continue
			}
			var teammates []string
			for _, otherID := range team.Agents {
				if otherID != agentID {
					if uname, ok := botUsernames[otherID]; ok {
						teammates = append(teammates, "@"+uname)
					} else {
						teammates = append(teammates, otherID)
					}
				}
			}
			if botUname, ok := botUsernames[agentID]; ok {
				ag.SetGroupContext(&agent.GroupContext{
					BotUsername: botUname,
					Teammates:  teammates,
				})
			}
		}
	}

	// Set up heartbeats for each agent
	heartbeatInterval := time.Duration(cfg.Heartbeat.IntervalMinutes) * time.Minute
	if heartbeatInterval <= 0 {
		heartbeatInterval = agent.DefaultHeartbeatInterval
	}
	var heartbeats []*agent.Heartbeat
	for _, ag := range agentMgr.All() {
		hb := agent.NewHeartbeat(ag, mb, heartbeatInterval)
		heartbeats = append(heartbeats, hb)
	}

	// Open the unified store. We open this BEFORE the cron scheduler so
	// the scheduler can be wired to the store-backed polling path that
	// the UI and agent tools also write through. On any failure we fall
	// back to a no-op (in-memory only) configuration: the in-memory job
	// list from cfg.CronJobs still works, just without UI <→ agent
	// visibility.
	homeDir, _ := config.HomeDir()
	unifiedStore, storeErr := store.New(&store.StorageConfig{
		Type: store.StorageType(cfg.Storage.Type),
		DSN:  cfg.Storage.DSN,
	}, homeDir)
	if storeErr != nil {
		slog.Warn("unified store init failed; cron will be in-memory only",
			"error", storeErr)
		unifiedStore = nil
	}

	// Ensure the cron_jobs + notifications tables exist on database
	// backends. FileStore creates the corresponding JSON files lazily
	// so it doesn't need explicit migration calls. We deliberately
	// migrate just these tables rather than running the full legacy
	// Migrate() set — the historical migrationSQL includes
	// memory_logs (AUTOINCREMENT, SQLite-only) and other dialect-
	// specific quirks that fail under Postgres.
	if dbs, ok := unifiedStore.(*store.DBStore); ok {
		if err := dbs.MigrateCronJobs(context.Background()); err != nil {
			slog.Warn("cron_jobs migration failed; UI/agent cron will return errors",
				"error", err)
		} else {
			slog.Info("cron_jobs table ready")
		}
		if err := dbs.MigrateNotifications(context.Background()); err != nil {
			slog.Warn("notifications migration failed; inbox will return errors",
				"error", err)
		} else {
			slog.Info("notifications table ready")
		}
	}

	// One-time migration: anything declared under cfg.CronJobs in the
	// JSON config gets imported into the store, then the JSON list is
	// emptied and rewritten. This collapses the historical "two ledgers"
	// design — UI used to read JSON, agent tool wrote DB, scheduler
	// polled DB — into a single source of truth in `unifiedStore`.
	//
	// Idempotent: if the store already contains a job with the same
	// (Name, AgentID, Schedule) signature, we skip rather than duplicate.
	if unifiedStore != nil && len(cfg.CronJobs) > 0 {
		migrated := migrateLegacyCronJobs(unifiedStore, cfg.CronJobs)
		if migrated > 0 {
			slog.Info("cron: migrated legacy JSON jobs into store",
				"count", migrated, "total_in_config", len(cfg.CronJobs))
		}
		// Drop the legacy field so subsequent saves don't re-introduce
		// the duplicate ledger. We persist this immediately so a crash
		// before clean shutdown still leaves the config in a sane state.
		cfg.CronJobs = nil
		if err := persistConfigSnapshot(cfg); err != nil {
			slog.Warn("cron: failed to clear legacy CronJobs from config",
				"error", err)
		}
	}

	// In-memory legacy path is now empty by design; scheduler will
	// pull everything from the store via SetStore below.
	scheduler := cron.NewScheduler(nil, mb)
	if unifiedStore != nil {
		scheduler.SetStore(&schedulerStoreAdapter{st: unifiedStore})
	}

	// Register web search tool for all agents if configured
	if cfg.WebSearch.APIKey != "" {
		for _, ag := range agentMgr.All() {
			ag.RegisterWebSearchTool(cfg.WebSearch.APIKey)
		}
		slog.Info("web search registered", "provider", cfg.WebSearch.Provider)
	}

	// Register sub-agent spawner for all agents
	spawner := &gatewaySubAgentSpawner{agents: agentMgr}
	for _, ag := range agentMgr.All() {
		ag.SetSubAgentSpawner(spawner)
	}

	// Set up webhook server if enabled
	var webhookSrv *webhook.Server
	if cfg.Hooks.Enabled {
		webhookSrv = webhook.NewServer(cfg.Hooks.Token, cfg.Hooks.Path, &webhookAgentHandler{agents: agentMgr})
	}

	// Set up plugin manager
	var pluginMgr *plugin.Manager
	if cfg.Plugins.Enabled {
		pluginMgr = plugin.NewManager(mb)

		homeDir, _ := config.HomeDir()
		pluginPaths := []string{filepath.Join(homeDir, "plugins")}
		for _, p := range cfg.Plugins.Paths {
			pluginPaths = append(pluginPaths, p)
		}

		if err := pluginMgr.Discover(pluginPaths); err != nil {
			slog.Warn("plugin discovery error", "error", err)
		}

		// Apply per-plugin config
		if len(cfg.Plugins.Entries) > 0 {
			entries := make(map[string]plugin.PluginEntryCfg, len(cfg.Plugins.Entries))
			for k, v := range cfg.Plugins.Entries {
				entries[k] = plugin.PluginEntryCfg{
					Enabled: v.Enabled,
					Config:  v.Config,
				}
			}
			pluginMgr.ApplyConfig(entries)
		}
	}

	// Create task queue with config values
	maxConcurrent := cfg.TaskQueue.MaxConcurrent
	if maxConcurrent <= 0 {
		maxConcurrent = 10
	}
	taskTimeoutSec := cfg.TaskQueue.TaskTimeoutSec
	if taskTimeoutSec <= 0 {
		taskTimeoutSec = 300
	}
	taskTimeout := time.Duration(taskTimeoutSec) * time.Second

	g := &Gateway{
		config:       cfg,
		bus:          mb,
		agents:       agentMgr,
		chanMgr:      chanMgr,
		bindings:     cfg.Bindings,
		botUsernames: botUsernames,
		teams:        teams,
		heartbeats:   heartbeats,
		scheduler:    scheduler,
		webhookSrv:   webhookSrv,
		pluginMgr:    pluginMgr,
		pgDB:         pgDB,
		store:        unifiedStore,
	}

	// Recipient resolver: looks up the user's "send to me" address
	// for any (channel, accountID) by walking cfg.Channels. Shared
	// by the cron tool (chatID auto-fill for cross-channel jobs)
	// and the notify tool (any-agent push to the user).
	resolver := makeRecipientResolver(cfg)
	sender := func(out bus.OutboundMessage) {
		mb.Outbound <- out
	}
	notifWriter := makeNotificationWriter(unifiedStore)

	// Hand the unified store + resolver to every agent so their tool
	// registries can register cron management AND the notify tool.
	// Without this the agent has no in-band way to schedule a task
	// (falls back to writing shell scripts + crontab via exec —
	// invisible to the platform and a security footgun) or to push
	// to the user.
	if unifiedStore != nil {
		for _, ag := range agentMgr.All() {
			ag.SetCronStore(unifiedStore, resolver)
		}
	}
	for _, ag := range agentMgr.All() {
		ag.SetNotifyDeps(resolver, sender, notifWriter)
	}

	tq := taskqueue.NewQueue(maxConcurrent, taskTimeout, func(ctx context.Context, task *taskqueue.Task) (string, error) {
		ag := agentMgr.AgentByID(task.AgentID)
		if ag == nil {
			return "", fmt.Errorf("agent %q not found", task.AgentID)
		}

		// Two axes drive delivery:
		//   (1) Origin: was this a real human input or an internal
		//       trigger (cron/webhook/etc)?
		//   (2) Channel: '' / 'web' = Inbox-only; real channel name
		//       = also push through that IM.
		//
		// Internal-origin replies ALWAYS land in the Inbox as a
		// notification — the Inbox is the OS-level "system log" for
		// agent activity, every cron fire should be auditable there
		// regardless of whether it also pings Telegram. Real-channel
		// internal triggers ALSO push through the IM bot. Plain
		// human-channel input (real chat) keeps the existing
		// outbound-only path so we don't double-notify.
		isInternalOrigin := task.Message.Origin == "cron" ||
			task.Message.Origin == "webhook" ||
			task.Message.Origin == "internal"
		channelIsReal := task.Message.Channel != "" && task.Message.Channel != "web"

		var typingDone chan struct{}
		// Typing indicator: only meaningful for real channels with a
		// real chat thread. A cron-on-telegram fire still wants the
		// "typing…" affordance so the user knows a reply is coming.
		if isInternalOrigin && channelIsReal {
			chanMgr.SendTyping(task.Message.Channel, task.AccountID, task.Message.ChatID)
			typingDone = make(chan struct{})
			go func() {
				ticker := time.NewTicker(5 * time.Second)
				defer ticker.Stop()
				for {
					select {
					case <-typingDone:
						return
					case <-ctx.Done():
						return
					case <-ticker.C:
						chanMgr.SendTyping(task.Message.Channel, task.AccountID, task.Message.ChatID)
					}
				}
			}()
		}

		reply := ag.HandleMessage(ctx, task.Message)
		if typingDone != nil {
			close(typingDone)
		}

		if isInternalOrigin {
			// Always write Inbox — it's the audit log for every
			// cron/webhook/internal fire, even when an IM also
			// receives a copy. Lets the user catch up on what their
			// agents have been doing without scrolling Telegram
			// history.
			if unifiedStore != nil {
				if err := writeOriginNotification(unifiedStore, ag.Name(), task.Message, reply); err != nil {
					slog.Warn("failed to persist notification for internal origin",
						"origin", task.Message.Origin,
						"agent", ag.Name(),
						"error", err,
					)
				}
			} else {
				slog.Warn("internal origin inbox copy dropped: no store",
					"origin", task.Message.Origin, "agent", ag.Name())
			}

			// Then, if the cron job named a real IM channel, push
			// the reply there too. Outbound goes through chanMgr,
			// which has its own single-account fallback so empty
			// AccountID still resolves to the right bot.
			if channelIsReal {
				mb.Outbound <- bus.OutboundMessage{
					Channel:      task.Message.Channel,
					AccountID:    task.AccountID,
					ChatID:       task.Message.ChatID,
					Text:         reply,
					ReplyToMsgID: task.Message.MessageID,
				}
			}
			return reply, nil
		}

		// Real human input — original behaviour, single outbound.
		mb.Outbound <- bus.OutboundMessage{
			Channel:      task.Message.Channel,
			AccountID:    task.AccountID,
			ChatID:       task.Message.ChatID,
			Text:         reply,
			ReplyToMsgID: task.Message.MessageID,
		}
		return reply, nil
	})
	g.taskQueue = tq

	return g, nil
}

// AgentManager returns the gateway's agent manager.
func (g *Gateway) AgentManager() *agent.Manager {
	return g.agents
}

// TaskQueue returns the gateway's task queue.
func (g *Gateway) TaskQueue() *taskqueue.Queue {
	return g.taskQueue
}

// Run starts the gateway and blocks until shutdown signal.
// buildProviderRegistry walks cfg.Providers and builds one Provider
// per entry, then picks the default Provider for unprefixed model
// names. Resolution order for the default:
//
//  1. The provider whose name matches the prefix of cfg.Agents.Defaults.Model
//     (so a default model of "openai/foo" makes "openai" the default).
//  2. The first of "default" / "openai" / "openrouter" that is registered.
//  3. Any registered provider — deterministic-ish via Go's map iteration
//     would be bad, so we sort lexicographically before picking.
//
// We log every registered provider on start so misconfigurations show
// up in the gateway log instead of silently 4xx-ing later.
func buildProviderRegistry(cfg *config.Config) *provider.Registry {
	reg := provider.NewRegistry()

	// Collect names so the default-resolution log line is reproducible.
	names := make([]string, 0, len(cfg.Providers))
	for name, pCfg := range cfg.Providers {
		if pCfg.APIBase == "" && pCfg.APIKey == "" {
			slog.Warn("provider entry empty, skipping", "name", name)
			continue
		}
		p := provider.NewProvider(pCfg.APIKey, pCfg.APIBase, pCfg.APIType)
		reg.Set(name, p)
		names = append(names, name)
		slog.Info("provider registered", "name", name, "apiBase", pCfg.APIBase, "apiType", pCfg.APIType)
	}

	// Pick the default. We mirror the historical fallback chain so
	// existing single-provider configs (no provider/ prefix on the
	// model) keep working.
	defaultModel := cfg.Agents.Defaults.Model
	chosen := ""
	if idx := strings.Index(defaultModel, "/"); idx > 0 {
		candidate := defaultModel[:idx]
		if reg.Has(candidate) {
			chosen = candidate
		}
	}
	if chosen == "" {
		for _, key := range []string{"default", "openai", "openrouter"} {
			if reg.Has(key) {
				chosen = key
				break
			}
		}
	}
	if chosen == "" && len(names) > 0 {
		// Pick deterministically rather than relying on map order.
		chosen = names[0]
		for _, n := range names {
			if n < chosen {
				chosen = n
			}
		}
	}
	if chosen != "" {
		reg.SetDefault(chosen)
	}
	slog.Info("provider registry built",
		"providers", reg.Names(),
		"default", chosen,
		"defaultModel", defaultModel,
	)
	return reg
}

func (g *Gateway) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		slog.Info("received signal, shutting down", "signal", sig)
		cancel()
	}()

	var wg sync.WaitGroup

	// Start config file watcher for hot-reload
	wg.Add(1)
	go g.startConfigWatcher(ctx, &wg)

	// Start dedup cleanup goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		g.cleanupDedup(ctx)
	}()

	// Start inbound message processor
	wg.Add(1)
	go func() {
		defer wg.Done()
		g.processInbound(ctx)
	}()

	// Start channel manager
	wg.Add(1)
	go func() {
		defer wg.Done()
		g.chanMgr.Start(ctx)
	}()

	// Start heartbeats for each agent
	for _, hb := range g.heartbeats {
		wg.Add(1)
		go func(h *agent.Heartbeat) {
			defer wg.Done()
			h.Start(ctx)
		}(hb)
	}

	// Start cron scheduler
	wg.Add(1)
	go func() {
		defer wg.Done()
		g.scheduler.Start(ctx)
	}()

	// Start webhook server if configured
	if g.webhookSrv != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			port := g.config.Hooks.Port
			if port == 0 {
				port = 18954
			}
			addr := fmt.Sprintf(":%d", port)
			if err := g.webhookSrv.ListenAndServe(ctx, addr); err != nil {
				slog.Error("webhook server error", "error", err)
			}
		}()
	}

	// Start plugins if enabled
	if g.pluginMgr != nil {
		if err := g.pluginMgr.StartAll(ctx); err != nil {
			slog.Error("plugin start error", "error", err)
		}

		// Register channel adapters for channel plugins
		for _, inst := range g.pluginMgr.ChannelPlugins() {
			adapter := plugin.NewChannelAdapter(g.pluginMgr, inst.Manifest.ID)
			g.chanMgr.Register(adapter)
			slog.Info("registered plugin channel", "plugin", inst.Manifest.ID)
		}

		// Register tool plugins with all agents
		for _, inst := range g.pluginMgr.ToolPlugins() {
			for _, ag := range g.agents.All() {
				if err := plugin.RegisterPluginTools(ctx, g.pluginMgr, inst.Manifest.ID, ag.ToolRegistry()); err != nil {
					slog.Error("register plugin tools failed", "plugin", inst.Manifest.ID, "agent", ag.Name(), "error", err)
				}
			}
		}

		// Register hook plugins with all agents
		for _, inst := range g.pluginMgr.HookPlugins() {
			for _, ag := range g.agents.All() {
				if err := plugin.RegisterPluginHooks(ctx, g.pluginMgr, inst.Manifest.ID, ag.HookRegistry(), ag.Name()); err != nil {
					slog.Error("register plugin hooks failed", "plugin", inst.Manifest.ID, "agent", ag.Name(), "error", err)
				}
			}
		}
	}

	slog.Info("gateway started")

	wg.Wait()

	// Stop task queue
	if g.taskQueue != nil {
		g.taskQueue.Stop()
	}

	// Stop plugins on shutdown
	if g.pluginMgr != nil {
		g.pluginMgr.StopAll()
	}

	// Close PostgreSQL connection pool
	if g.pgDB != nil {
		g.pgDB.Close()
		slog.Info("pg: connection pool closed")
	}

	slog.Info("gateway stopped")
	return nil
}
