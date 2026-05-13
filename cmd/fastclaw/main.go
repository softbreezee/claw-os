package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"

	// Register the pgx stdlib driver so store.NewDBStore can open
	// postgres connections via database/sql. The gateway's own pg pool
	// uses pgxpool directly and doesn't need this, but the chat task
	// store goes through database/sql and would otherwise fail with
	// `sql: unknown driver "pgx"`.
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/fastclaw-ai/fastclaw/internal/agent"
	"github.com/fastclaw-ai/fastclaw/internal/api"
	"github.com/fastclaw-ai/fastclaw/internal/config"
	"github.com/fastclaw-ai/fastclaw/internal/daemon"
	"github.com/fastclaw-ai/fastclaw/internal/eventbus"
	"github.com/fastclaw-ai/fastclaw/internal/gateway"
	"github.com/fastclaw-ai/fastclaw/internal/setup"
	"github.com/fastclaw-ai/fastclaw/internal/store"
	"github.com/fastclaw-ai/fastclaw/internal/taskrunner"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "fastclaw",
		Short: "Pawnix - Lightweight AI Agent Framework",
		// No args = default to gateway (so double-click on Windows works)
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGateway(18953)
		},
	}

	rootCmd.AddCommand(gatewayCmd())
	rootCmd.AddCommand(agentCmd())
	rootCmd.AddCommand(skillCmd())
	rootCmd.AddCommand(sessionCmd())
	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(upgradeCmd())
	rootCmd.AddCommand(doctorCmd())
	rootCmd.AddCommand(backupCmd())
	rootCmd.AddCommand(resetCmd())
	rootCmd.AddCommand(pluginCmd())
	rootCmd.AddCommand(providerCmd())
	rootCmd.AddCommand(sandboxCmd())
	rootCmd.AddCommand(policyCmd())
	rootCmd.AddCommand(daemonCmd())
	rootCmd.AddCommand(migrateCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func gatewayCmd() *cobra.Command {
	var port int
	cmd := &cobra.Command{
		Use:   "gateway",
		Short: "Start the Pawnix gateway (loads all agents)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGateway(port)
		},
	}
	cmd.Flags().IntVar(&port, "port", 18953, "port for setup wizard / web UI")
	return cmd
}

func runGateway(port int) error {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	// Check if config exists
	cfg, err := config.Load()
	if err != nil {
		// Config doesn't exist — run setup wizard
		slog.Info("no config found, starting setup wizard", "url", fmt.Sprintf("http://localhost:%d", port))
		return runSetupWizard(port)
	}

	slog.Info("starting gateway")

	// Write PID file for daemon management
	if err := daemon.WritePIDFile(); err != nil {
		slog.Warn("failed to write PID file", "error", err)
	}
	defer daemon.RemovePIDFile()

	gw, err := gateway.New(cfg)
	if err != nil {
		return fmt.Errorf("create gateway: %w", err)
	}

	// Start web UI server alongside gateway
	gwCfg := &cfg.Gateway
	if gwCfg.Port > 0 {
		port = gwCfg.Port
	}

	webSrv := setup.NewServer(port, nil)
	webSrv.SetAgentProvider(&agentProviderAdapter{mgr: gw.AgentManager()})
	webSrv.SetTaskQueue(gw.TaskQueue())
	webSrv.SetGatewayConfig(gwCfg)

	// Wire up the async web chat task subsystem (PR2):
	//   web UI POST /api/chat/submit  →  taskrunner.Submit
	//                                 →  agent.HandleWebChatStream
	//                                 →  events streamed via eventbus to
	//                                    GET /api/chat/tasks/:id/events
	// We construct a separate store.Store here (the gateway uses its own
	// pg backend for sessions, but doesn't expose a generic Store). For
	// the file backend this is essentially zero-cost; for SQLite/Postgres
	// it opens a second connection pool, which is fine.
	homeDir, _ := config.HomeDir()
	// Open the store WITHOUT full AutoMigrate – the legacy migration
	// set has dialect-specific issues (e.g. memory_logs AUTOINCREMENT
	// only works on SQLite). For the chat_tasks table we need, we run
	// a targeted migration below.
	chatTaskStore, storeErr := store.New(&store.StorageConfig{
		Type: store.StorageType(cfg.Storage.Type),
		DSN:  cfg.Storage.DSN,
	}, homeDir)
	if storeErr != nil {
		slog.Warn("chat task store init failed; async chat tasks disabled", "err", storeErr)
	} else {
		// Run only chat_tasks DDL if we're on a database backend.
		// The file backend doesn't need any schema setup.
		if dbs, ok := chatTaskStore.(*store.DBStore); ok {
			if mErr := dbs.MigrateChatTasks(context.Background()); mErr != nil {
				slog.Warn("chat_tasks migration failed; async chat tasks disabled", "err", mErr)
				chatTaskStore.Close()
				chatTaskStore = nil
			}
		}
		if chatTaskStore != nil {
			evtBus := eventbus.NewMemoryBus()
			runner := taskrunner.New(chatTaskStore, evtBus,
				&chatTaskAgentResolver{mgr: gw.AgentManager()},
				taskrunner.Options{TenantID: store.DefaultTenantID})
			webSrv.SetChatTaskRunner(runner, evtBus, chatTaskStore, store.DefaultTenantID)
			slog.Info("chat task subsystem ready")
		}
	}

	// Set up OpenAI-compatible API and WebSocket gateway
	gatewayToken := cfg.Gateway.Auth.Token
	apiSrv := api.NewServer(gw.AgentManager(), gatewayToken, gwCfg)
	webSrv.SetAPIServer(apiSrv)

	bindMode := gwCfg.Bind
	if bindMode == "" {
		bindMode = "loopback"
	}
	authMode := gwCfg.Auth.Mode
	if authMode == "" {
		authMode = "token"
	}
	slog.Info("gateway API enabled",
		"port", port,
		"bind", bindMode,
		"auth", authMode,
		"chatCompletions", gwCfg.HTTP.Endpoints.ChatCompletions.Enabled,
	)

	// Write fastclaw.gateway.json for ChatClaw auto-detect
	writePawnixGatewayConfig(port, gatewayToken)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := webSrv.Run(ctx); err != nil {
			slog.Error("web server error", "error", err)
		}
	}()

	slog.Info("web UI available", "url", fmt.Sprintf("http://localhost:%d", port))

	return gw.Run()
}

// agentProviderAdapter adapts agent.Manager to setup.AgentProvider.
type agentProviderAdapter struct {
	mgr *agent.Manager
}

func (a *agentProviderAdapter) AllAgents() []setup.AgentHandle {
	agents := a.mgr.All()
	result := make([]setup.AgentHandle, len(agents))
	for i, ag := range agents {
		result[i] = ag
	}
	return result
}

func (a *agentProviderAdapter) AgentByID(id string) setup.AgentHandle {
	ag := a.mgr.AgentByID(id)
	if ag == nil {
		return nil
	}
	return ag
}

// chatTaskAgentResolver adapts agent.Manager to the (much narrower)
// taskrunner.AgentResolver contract. We can't pass the manager directly
// because its AgentByID returns *agent.Agent and Go interface satisfaction
// requires the exact return type.
type chatTaskAgentResolver struct {
	mgr *agent.Manager
}

func (r *chatTaskAgentResolver) AgentByID(id string) taskrunner.AgentHandle {
	ag := r.mgr.AgentByID(id)
	if ag == nil {
		return nil
	}
	return ag
}

// writePawnixGatewayConfig writes ~/.fastclaw/fastclaw.gateway.json for ChatClaw auto-detect.
func writePawnixGatewayConfig(port int, token string) {
	if token == "" {
		return
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	dir := filepath.Join(home, ".fastclaw")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}

	cfg := map[string]any{
		"gateway": map[string]any{
			"port": port,
			"auth": map[string]string{
				"token": token,
			},
		},
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	if err := os.WriteFile(filepath.Join(dir, "fastclaw.gateway.json"), data, 0o644); err != nil {
		slog.Warn("failed to write fastclaw.gateway.json", "error", err)
	} else {
		slog.Info("wrote fastclaw.gateway.json for ChatClaw auto-detect")
	}
}

func runSetupWizard(port int) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := setup.NewServer(port, func(cfg *config.Config) {
		slog.Info("setup complete, config saved")
		// Stop the setup wizard and restart as gateway
		go func() {
			cancel()
		}()
	})

	// Open browser
	url := fmt.Sprintf("http://localhost:%d", port)
	go openBrowser(url)

	if err := srv.Run(ctx); err != nil {
		return err
	}

	// Config was saved, now start the gateway
	slog.Info("restarting as gateway")
	return runGateway(port)
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return
	}
	cmd.Run()
}

