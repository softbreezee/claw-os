package cron

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/softbreezee/claw-os/internal/bus"
)

// JobType defines the type of cron schedule.
type JobType string

const (
	JobTypeExact    JobType = "exact"    // specific time: "14:30"
	JobTypeInterval JobType = "interval" // duration: "5m", "1h"
	JobTypeCron     JobType = "cron"     // cron expression: "*/5 * * * *"
)

// Job defines a scheduled job.
type Job struct {
	Name      string  `json:"name"`
	Type      JobType `json:"type"`
	Schedule  string  `json:"schedule"`           // depends on type
	AgentID   string  `json:"agentId"`
	Channel   string  `json:"channel"`             // channel to send results back through
	AccountID string  `json:"accountId,omitempty"` // optional bot account within the channel
	ChatID    string  `json:"chatId"`              // chat to send results to
	Message   string  `json:"message"`             // message to send to the agent
}

// CronConfig holds cron job configuration.
type CronConfig struct {
	Jobs []Job `json:"jobs"`
}

// StoreInterface is the subset of store.Store needed by the scheduler.
type StoreInterface interface {
	GetDueCronJobs(ctx context.Context, now time.Time) ([]StoreJob, error)
	LockCronJob(ctx context.Context, jobID, instanceID string) (bool, error)
	UpdateCronJobRun(ctx context.Context, jobID string, lastRun, nextRun time.Time) error
}

// StoreJob mirrors store.CronJobRecord to avoid import cycle.
//
// AccountID is required for IM channels that register multiple bots
// per channel type (e.g. two telegram bots, one personal one team).
// The chanMgr keys outbound messages on "channel:accountID" — drop
// AccountID and the cron-fired reply ends up as "telegram:" which
// matches no registered handler and the message is silently dropped.
// This was the root cause of "Telegram cron didn't deliver" until v0.3.x.
type StoreJob struct {
	ID        string
	AgentID   string
	Name      string
	Type      string
	Schedule  string
	Message   string
	Channel   string
	AccountID string
	ChatID    string
}

// Scheduler manages cron job execution.
type Scheduler struct {
	mu         sync.Mutex
	jobs       []Job
	bus        *bus.MessageBus
	store      StoreInterface
	instanceID string
	// hot-reload support
	parentCtx context.Context
	jobCancel context.CancelFunc
}

// NewScheduler creates a scheduler from config.
func NewScheduler(jobs []Job, mb *bus.MessageBus) *Scheduler {
	return &Scheduler{
		jobs:       jobs,
		bus:        mb,
		instanceID: "default",
	}
}

// SetStore enables DB-backed cron job polling.
func (s *Scheduler) SetStore(st StoreInterface) {
	s.store = st
}

// LoadJobs reads cron jobs from a JSON file.
func LoadJobs(path string) ([]Job, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read cron config: %w", err)
	}

	var cfg CronConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse cron config: %w", err)
	}

	return cfg.Jobs, nil
}

// Start begins the scheduler. It blocks until ctx is cancelled.
func (s *Scheduler) Start(ctx context.Context) {
	s.mu.Lock()
	s.parentCtx = ctx
	slog.Info("cron scheduler started", "jobs", len(s.jobs), "store_backed", s.store != nil)

	// Start goroutines for initial in-memory jobs
	s.startJobGoroutines()
	s.mu.Unlock()

	// If store is set, poll for DB-backed jobs
	if s.store != nil {
		go s.pollStore(ctx)
	}

	<-ctx.Done()
	slog.Info("cron scheduler stopped")
}

func (s *Scheduler) pollStore(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	// Run immediately on start, then on tick
	s.processDueJobs(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.processDueJobs(ctx)
		}
	}
}

func (s *Scheduler) processDueJobs(ctx context.Context) {
	now := time.Now()
	dueJobs, err := s.store.GetDueCronJobs(ctx, now)
	if err != nil {
		slog.Error("failed to get due cron jobs", "error", err)
		return
	}

	for _, j := range dueJobs {
		locked, err := s.store.LockCronJob(ctx, j.ID, s.instanceID)
		if err != nil {
			slog.Error("failed to lock cron job", "id", j.ID, "error", err)
			continue
		}
		if !locked {
			continue
		}

		slog.Info("firing store-backed cron job", "id", j.ID, "name", j.Name)

		text := j.Message
		if text == "" {
			text = fmt.Sprintf("[Cron Job: %s] This is a scheduled task trigger.", j.Name)
		}

		// A blank channel means "deliver in-process via the
		// notifications subsystem" — we tag it as "web" so logs are
		// readable, but routing relies on AgentID + Origin="cron"
		// rather than channel binding.
		ch := j.Channel
		if ch == "" {
			ch = "web"
		}
		s.bus.Inbound <- bus.InboundMessage{
			Channel:   ch,
			AccountID: j.AccountID, // required so chanMgr can route to the right bot instance
			ChatID:    j.ChatID,
			UserID:    "cron",
			Text:      text,
			PeerKind:  "dm",
			AgentID:   j.AgentID, // bypass binding match — cron knows its agent
			Origin:    "cron",    // signals "deliver as Notification, not chat reply"
		}

		// Compute the actual next fire time from the job's schedule. The
		// previous "now + 60s" placeholder caused jobs to refire every
		// minute regardless of cron expression — fine for a smoke test,
		// catastrophic in real use.
		nextRun := ComputeNextRun(j.Type, j.Schedule, now)
		if err := s.store.UpdateCronJobRun(ctx, j.ID, now, nextRun); err != nil {
			slog.Error("failed to update cron job run", "id", j.ID, "error", err)
		}
	}
}

// ComputeNextRun returns the next time a job of the given type+schedule
// should fire after `from`. On any parse error it falls back to one hour
// out, which is a safe-enough rate-limit to avoid hot-spinning if a user
// somehow stored a malformed schedule.
//
// Supported types (matches cron.JobType + a few synonyms used by the
// agent tool's create_cron_job for ergonomic LLM input):
//   - "interval" / "duration"  — go duration like "5m", "1h"
//   - "exact" / "daily"        — HH:MM, fires once per day at that wall-clock time
//   - "cron"                   — five-field cron expression (* and */N supported)
//   - "once"                   — RFC3339 datetime; after firing nextRun is far future (year 9999)
func ComputeNextRun(jobType, schedule string, from time.Time) time.Time {
	const fallback = time.Hour
	jobType = strings.ToLower(strings.TrimSpace(jobType))
	schedule = strings.TrimSpace(schedule)

	switch jobType {
	case "interval", "duration":
		if d, err := time.ParseDuration(schedule); err == nil && d > 0 {
			return from.Add(d)
		}
	case "exact", "daily":
		parts := strings.Split(schedule, ":")
		if len(parts) == 2 {
			hour, err1 := strconv.Atoi(parts[0])
			minute, err2 := strconv.Atoi(parts[1])
			if err1 == nil && err2 == nil {
				next := time.Date(from.Year(), from.Month(), from.Day(), hour, minute, 0, 0, from.Location())
				if !next.After(from) {
					next = next.Add(24 * time.Hour)
				}
				return next
			}
		}
	case "once":
		if t, err := time.Parse(time.RFC3339, schedule); err == nil {
			if t.After(from) {
				return t
			}
			// Already fired-or-past — push way out so it never re-fires.
			return time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC)
		}
	case "cron", "":
		if next := nextCronTick(schedule, from); !next.IsZero() {
			return next
		}
	}
	// keep the unexported alias so in-package callers compile unchanged

	slog.Warn("cron: cannot parse schedule, falling back to +1h",
		"type", jobType, "schedule", schedule)
	return from.Add(fallback)
}

// nextCronTick walks forward minute-by-minute from `from+1m` looking for
// the next time that matches the 5-field cron expression. Capped at one
// year so a hopelessly impossible expression like "0 0 30 2 *" returns
// the zero time and the caller can fall back. Inefficient but trivially
// correct, which matches the rest of this scheduler's "small and obvious"
// design (we already do per-minute polling for the in-memory cron path).
func nextCronTick(expr string, from time.Time) time.Time {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return time.Time{}
	}
	// Round up to the next whole minute so we don't return `from` itself.
	t := from.Truncate(time.Minute).Add(time.Minute)
	limit := from.Add(365 * 24 * time.Hour)
	for !t.After(limit) {
		if cronMatch(fields, t) {
			return t
		}
		t = t.Add(time.Minute)
	}
	return time.Time{}
}

func (s *Scheduler) runJob(ctx context.Context, job Job) {
	slog.Info("cron job registered", "name", job.Name, "type", job.Type, "schedule", job.Schedule)

	switch job.Type {
	case JobTypeInterval:
		s.runInterval(ctx, job)
	case JobTypeExact:
		s.runExact(ctx, job)
	case JobTypeCron:
		s.runCronExpr(ctx, job)
	default:
		slog.Warn("unknown cron job type", "name", job.Name, "type", job.Type)
	}
}

func (s *Scheduler) runInterval(ctx context.Context, job Job) {
	dur, err := time.ParseDuration(job.Schedule)
	if err != nil {
		slog.Error("invalid interval duration", "name", job.Name, "schedule", job.Schedule, "error", err)
		return
	}

	ticker := time.NewTicker(dur)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.fireJob(job)
		}
	}
}

func (s *Scheduler) runExact(ctx context.Context, job Job) {
	// Parse time in HH:MM format
	parts := strings.Split(job.Schedule, ":")
	if len(parts) != 2 {
		slog.Error("invalid exact time format (expected HH:MM)", "name", job.Name, "schedule", job.Schedule)
		return
	}

	hour, err1 := strconv.Atoi(parts[0])
	minute, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		slog.Error("invalid exact time values", "name", job.Name, "schedule", job.Schedule)
		return
	}

	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
		if !next.After(now) {
			next = next.Add(24 * time.Hour)
		}

		waitDur := time.Until(next)
		slog.Info("cron exact job scheduled", "name", job.Name, "next_fire", next.Format("2006-01-02 15:04:05"))

		select {
		case <-ctx.Done():
			return
		case <-time.After(waitDur):
			s.fireJob(job)
		}
	}
}

func (s *Scheduler) runCronExpr(ctx context.Context, job Job) {
	// Simple cron expression parser: "minute hour day month weekday"
	// Supports * and */N for each field
	fields := strings.Fields(job.Schedule)
	if len(fields) != 5 {
		slog.Error("invalid cron expression (expected 5 fields)", "name", job.Name, "schedule", job.Schedule)
		return
	}

	// Check every minute
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			if cronMatch(fields, t) {
				s.fireJob(job)
			}
		}
	}
}

// cronMatch checks if the current time matches a 5-field cron expression.
func cronMatch(fields []string, t time.Time) bool {
	values := []int{t.Minute(), t.Hour(), t.Day(), int(t.Month()), int(t.Weekday())}
	for i, field := range fields {
		if !fieldMatch(field, values[i]) {
			return false
		}
	}
	return true
}

// fieldMatch checks if a value matches a single cron field.
// Supported syntax (standard cron subset):
//
//	*        any value
//	*/N      every N units  (e.g. */5 = every 5 minutes)
//	A-B      inclusive range (e.g. 1-5 = Mon–Fri, 9-17 = office hours)
//	A,B,C    explicit list   (e.g. 0,15,30,45)
//	N        exact value
func fieldMatch(field string, value int) bool {
	if field == "*" {
		return true
	}

	// */N  – step
	if strings.HasPrefix(field, "*/") {
		n, err := strconv.Atoi(field[2:])
		if err != nil || n <= 0 {
			return false
		}
		return value%n == 0
	}

	// A-B  – inclusive range
	if strings.Contains(field, "-") {
		parts := strings.SplitN(field, "-", 2)
		if len(parts) == 2 {
			lo, err1 := strconv.Atoi(parts[0])
			hi, err2 := strconv.Atoi(parts[1])
			if err1 == nil && err2 == nil {
				return value >= lo && value <= hi
			}
		}
		return false
	}

	// A,B,C  – list
	if strings.Contains(field, ",") {
		for _, part := range strings.Split(field, ",") {
			n, err := strconv.Atoi(strings.TrimSpace(part))
			if err == nil && n == value {
				return true
			}
		}
		return false
	}

	// exact number
	n, err := strconv.Atoi(field)
	if err != nil {
		return false
	}
	return n == value
}

func (s *Scheduler) fireJob(job Job) {
	slog.Info("cron job firing", "name", job.Name, "agent", job.AgentID)

	text := job.Message
	if text == "" {
		text = fmt.Sprintf("[Cron Job: %s] This is a scheduled task trigger.", job.Name)
	}

	ch := job.Channel
	if ch == "" {
		ch = "web"
	}
	s.bus.Inbound <- bus.InboundMessage{
		Channel:   ch,
		AccountID: job.AccountID,
		ChatID:    job.ChatID,
		UserID:    "cron",
		Text:      text,
		PeerKind:  "dm",
		AgentID:   job.AgentID, // bypass binding match
		Origin:    "cron",      // signals "deliver as Notification, not chat reply"
	}
}

// UpdateJobs replaces the scheduler's job list (hot-reload).
// It cancels goroutines for old jobs and starts new ones.
func (s *Scheduler) UpdateJobs(jobs []Job) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.jobs = jobs

	// If the scheduler is running, restart job goroutines
	if s.parentCtx != nil {
		s.startJobGoroutines()
	}

	slog.Info("cron jobs updated (hot-reload)", "jobs", len(jobs))
}

// startJobGoroutines cancels any existing job goroutines and starts new ones.
// Must be called with s.mu held.
func (s *Scheduler) startJobGoroutines() {
	// Cancel previous batch of job goroutines
	if s.jobCancel != nil {
		s.jobCancel()
	}

	// Create a new child context for this batch
	jobCtx, cancel := context.WithCancel(s.parentCtx)
	s.jobCancel = cancel

	for _, job := range s.jobs {
		go s.runJob(jobCtx, job)
	}
}
