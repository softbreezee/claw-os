package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/softbreezee/claw-os/internal/config"
	"github.com/softbreezee/claw-os/internal/provider"
	pgstore "github.com/softbreezee/claw-os/internal/store/pg"
)

func migrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate local JSONL sessions to PostgreSQL",
		Long: `Reads all existing JSONL session files from ~/.pawnix/agents/*/sessions/
and upserts them into the PostgreSQL sessions table configured in storage.dsn.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigrate()
		},
	}
	return cmd
}

func runMigrate() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if cfg.Storage.Type != "postgres" || cfg.Storage.DSN == "" {
		return fmt.Errorf("storage.type must be 'postgres' with a valid dsn")
	}

	ctx := context.Background()
	db, err := pgstore.Open(ctx, cfg.Storage.DSN)
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	defer db.Close()

	if err := db.Migrate(ctx); err != nil {
		return fmt.Errorf("migrate schema: %w", err)
	}

	store := pgstore.NewSessionStore(db)
	homeDir, _ := config.HomeDir()
	agentsDir := filepath.Join(homeDir, "agents")

	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return fmt.Errorf("read agents dir: %w", err)
	}

	total, skipped, migrated := 0, 0, 0

	for _, agentEntry := range entries {
		if !agentEntry.IsDir() {
			continue
		}
		agentID := agentEntry.Name()
		sessionsDir := filepath.Join(agentsDir, agentID, "agent", "sessions")

		files, err := filepath.Glob(filepath.Join(sessionsDir, "*.jsonl"))
		if err != nil || len(files) == 0 {
			continue
		}

		for _, file := range files {
			total++
			base := filepath.Base(file)
			// filename: "web_s-xxx.jsonl"  →  channel="web"  sessionID="s-xxx"
			noExt := strings.TrimSuffix(base, ".jsonl")
			parts := strings.SplitN(noExt, "_", 2)
			if len(parts) != 2 {
				fmt.Printf("  skip (unexpected filename): %s\n", base)
				skipped++
				continue
			}
			channel, sessionID := parts[0], parts[1]

			msgs, err := loadJSONL(file)
			if err != nil || len(msgs) == 0 {
				skipped++
				continue
			}

			if err := store.Save(ctx, agentID, channel, sessionID, msgs); err != nil {
				fmt.Printf("  ERROR saving %s/%s/%s: %v\n", agentID, channel, sessionID, err)
				skipped++
				continue
			}

			fmt.Printf("  migrated: agent=%-20s channel=%-10s session=%-30s msgs=%d\n",
				agentID, channel, sessionID, len(msgs))
			migrated++
		}
	}

	fmt.Printf("\nDone: %d total, %d migrated, %d skipped\n", total, migrated, skipped)
	return nil
}

func loadJSONL(path string) ([]provider.Message, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var msgs []provider.Message
	for scanner.Scan() {
		var m provider.Message
		if json.Unmarshal(scanner.Bytes(), &m) == nil {
			msgs = append(msgs, m)
		}
	}
	return msgs, scanner.Err()
}
