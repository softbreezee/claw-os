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

	"github.com/softbreezee/claw-os/internal/agent"
	"github.com/softbreezee/claw-os/internal/api"
	"github.com/softbreezee/claw-os/internal/config"
	"github.com/softbreezee/claw-os/internal/daemon"
	"github.com/softbreezee/claw-os/internal/eventbus"
	"github.com/softbreezee/claw-os/internal/gateway"
	"github.com/softbreezee/claw-os/internal/setup"
	"github.com/softbreezee/claw-os/internal/store"
	"github.com/softbreezee/claw-os/internal/taskrunner"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "pawnix",
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
	rootCmd.AddCommand(mcpReasonixCmd())

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
	//
	// We deliberately reuse the gateway's unified store rather than
	// opening a second handle. Pre-v0.2.x there were two parallel
	// store handles for cron + chat_tasks, which caused subtle write
	// races on the file backend (two writers to ~/.pawnix/cron_jobs.json)
	// and doubled connection pools on PG/SQLite. One store, one source
	// of truth.
	chatTaskStore := gw.Store()
	if chatTaskStore == nil {
		slog.Warn("chat task store unavailable (gateway store init failed); async chat tasks disabled")
	} else {
		// Targeted DDL for chat_tasks on database backends. The file
		// backend doesn't need any schema setup.
		if dbs, ok := chatTaskStore.(*store.DBStore); ok {
			if mErr := dbs.MigrateChatTasks(context.Background()); mErr != nil {
				slog.Warn("chat_tasks migration failed; async chat tasks disabled", "err", mErr)
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

	// Hand the same unified store to the setup HTTP server so the
	// Cron Jobs page reads/writes the same ledger as the agent tools
	// and scheduler.
	if cs := gw.Store(); cs != nil {
		webSrv.SetCronStore(cs, store.DefaultTenantID)
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

	// Write pawnix.gateway.json for ChatClaw auto-detect
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

// writePawnixGatewayConfig writes ~/.pawnix/pawnix.gateway.json for ChatClaw auto-detect.
func writePawnixGatewayConfig(port int, token string) {
	if token == "" {
		return
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	dir := filepath.Join(home, ".pawnix")
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
	if err := os.WriteFile(filepath.Join(dir, "pawnix.gateway.json"), data, 0o644); err != nil {
		slog.Warn("failed to write pawnix.gateway.json", "error", err)
	} else {
		slog.Info("wrote pawnix.gateway.json for ChatClaw auto-detect")
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

