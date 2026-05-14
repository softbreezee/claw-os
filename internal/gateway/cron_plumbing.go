package gateway

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/softbreezee/claw-os/internal/config"
	"github.com/softbreezee/claw-os/internal/cron"
	"github.com/softbreezee/claw-os/internal/store"
)

// schedulerStoreAdapter bridges the unified store.Store interface to
// the narrower cron.StoreInterface that the scheduler talks to.
//
// Why have two interfaces at all? The cron package was originally
// written to be self-contained (only depends on bus) so it could be
// swapped or tested without dragging in store types. We preserve that
// boundary by keeping the conversion at the gateway layer instead of
// letting cron import store directly.
type schedulerStoreAdapter struct {
	st store.Store
}

func (a *schedulerStoreAdapter) GetDueCronJobs(ctx context.Context, now time.Time) ([]cron.StoreJob, error) {
	recs, err := a.st.GetDueCronJobs(ctx, now)
	if err != nil {
		return nil, err
	}
	out := make([]cron.StoreJob, 0, len(recs))
	for _, r := range recs {
		// Skip disabled jobs at the adapter layer; the file-backed
		// store doesn't filter on enabled (only the SQL backend does).
		if !r.Enabled {
			continue
		}
		out = append(out, cron.StoreJob{
			ID:       r.ID,
			AgentID:  r.AgentID,
			Name:     r.Name,
			Type:     r.Type,
			Schedule: r.Schedule,
			Message:  r.Message,
			Channel:  r.Channel,
			ChatID:   r.ChatID,
		})
	}
	return out, nil
}

func (a *schedulerStoreAdapter) LockCronJob(ctx context.Context, jobID, instanceID string) (bool, error) {
	return a.st.LockCronJob(ctx, jobID, instanceID)
}

func (a *schedulerStoreAdapter) UpdateCronJobRun(ctx context.Context, jobID string, lastRun, nextRun time.Time) error {
	return a.st.UpdateCronJobRun(ctx, jobID, lastRun, nextRun)
}

// migrateLegacyCronJobs imports cron jobs that were previously stored
// inline in pawnix.json (cfg.CronJobs) into the unified store.
//
// Returns the number of jobs newly inserted. Existing entries are left
// alone so re-running the migration is idempotent — the dedup key is
// (Name, AgentID, Schedule) which is the natural human-meaningful
// identity of a cron job.
func migrateLegacyCronJobs(st store.Store, legacy []config.CronJob) int {
	if len(legacy) == 0 {
		return 0
	}
	ctx := context.Background()
	existing, err := st.ListCronJobs(ctx, store.DefaultTenantID)
	if err != nil {
		slog.Warn("cron migration: failed to list existing jobs, proceeding without dedup",
			"error", err)
		existing = nil
	}

	seen := make(map[string]struct{}, len(existing))
	for _, j := range existing {
		seen[legacyKey(j.Name, j.AgentID, j.Schedule)] = struct{}{}
	}

	migrated := 0
	for _, lj := range legacy {
		key := legacyKey(lj.Name, lj.AgentID, lj.Schedule)
		if _, ok := seen[key]; ok {
			continue
		}
		now := time.Now()
		nextRun := now // fire on next poll cycle; scheduler will compute the proper next tick after first run
		rec := &store.CronJobRecord{
			ID:        newCronJobID(),
			TenantID:  store.DefaultTenantID,
			AgentID:   lj.AgentID,
			Name:      lj.Name,
			Type:      lj.Type,
			Schedule:  lj.Schedule,
			Message:   lj.Message,
			Channel:   lj.Channel,
			ChatID:    lj.ChatID,
			Timezone:  "Local",
			Enabled:   true,
			NextRun:   &nextRun,
			CreatedAt: now,
		}
		if err := st.SaveCronJob(ctx, store.DefaultTenantID, rec); err != nil {
			slog.Warn("cron migration: failed to import legacy job",
				"name", lj.Name, "error", err)
			continue
		}
		seen[key] = struct{}{}
		migrated++
	}
	return migrated
}

func legacyKey(name, agentID, schedule string) string {
	return name + "\x00" + agentID + "\x00" + schedule
}

// newCronJobID returns a v4-style UUID. Duplicated rather than imported
// from internal/agent/tools/cron.go to keep the dependency direction
// tools → gateway only (not the other way round).
func newCronJobID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Vanishingly unlikely on Linux/macOS; degrade to time-based ID
		// so we never panic just because /dev/urandom hiccupped.
		return fmt.Sprintf("ts-%d", time.Now().UnixNano())
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// persistConfigSnapshot rewrites ~/.pawnix/pawnix.json with the given
// config. We do this from the gateway after migration to drop the
// legacy CronJobs field — keeping it in the file would let the next
// startup re-import the same jobs and cause duplication.
//
// Mirrors setup.saveConfigFile (which lives in a different package
// and isn't exported); duplication is small and avoids creating a
// circular dependency.
func persistConfigSnapshot(cfg *config.Config) error {
	homeDir, err := config.HomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(homeDir, "pawnix.json")
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Store returns the unified store opened by the gateway. Other
// subsystems (setup HTTP API, chat task runner) should call this
// instead of opening their own store handle so we don't end up with
// multiple connection pools / file locks against the same data.
//
// May return nil if the store failed to initialise; callers must
// guard accordingly.
func (g *Gateway) Store() store.Store {
	return g.store
}
