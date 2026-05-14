package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// DBStore implements Store using a SQL database (PostgreSQL or SQLite).
// All tables are tenant-partitioned for multi-tenant cloud deployments.
type DBStore struct {
	db      *sql.DB
	dialect string // "postgres" or "sqlite"
}

// NewDBStore creates a database-backed store.
// dsn examples:
//   postgres: "postgres://user:pass@host:5432/pawnix?sslmode=disable"
//   sqlite:   "file:pawnix.db?_journal=WAL"
func NewDBStore(dialect, dsn string) (*DBStore, error) {
	// Import drivers via blank import in the caller (main.go) or use pgx/stdlib.
	// Here we use database/sql which requires the driver to be registered.
	db, err := sql.Open(driverName(dialect), dsn)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", dialect, err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping %s: %w", dialect, err)
	}

	return &DBStore{db: db, dialect: dialect}, nil
}

func driverName(dialect string) string {
	switch dialect {
	case "postgres":
		return "pgx"
	case "sqlite":
		return "sqlite3"
	default:
		return dialect
	}
}

// Migrate creates tables if they don't exist.
func (d *DBStore) Migrate(ctx context.Context) error {
	stmts := d.migrationSQL()
	for _, stmt := range stmts {
		if _, err := d.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("migrate: %w\nSQL: %s", err, stmt)
		}
	}
	return nil
}

// MigrateChatTasks creates only the chat_tasks table + its indexes.
// Used by callers that need the chat task subsystem (PR2) but cannot
// run the full Migrate() because legacy table definitions still have
// dialect-specific quirks (e.g. memory_logs uses AUTOINCREMENT which
// is SQLite-only). Idempotent; safe to call on every startup.
func (d *DBStore) MigrateChatTasks(ctx context.Context) error {
	for _, stmt := range d.chatTasksMigrationSQL() {
		if _, err := d.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("migrate chat_tasks: %w\nSQL: %s", err, stmt)
		}
	}
	return nil
}

// MigrateCronJobs creates only the cron_jobs table + its index. Same
// rationale as MigrateChatTasks: callers that want a single subsystem
// shouldn't have to run the full legacy migration set.
//
// As of v0.2.x the gateway calls this on every startup so the unified
// cron-jobs ledger is always available regardless of how Storage.Type
// or Storage.AutoMigrate is configured.
func (d *DBStore) MigrateCronJobs(ctx context.Context) error {
	for _, stmt := range d.cronJobsMigrationSQL() {
		if _, err := d.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("migrate cron_jobs: %w\nSQL: %s", err, stmt)
		}
	}
	return nil
}

func (d *DBStore) cronJobsMigrationSQL() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS cron_jobs (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			agent_id TEXT NOT NULL,
			name TEXT NOT NULL DEFAULT '',
			type TEXT NOT NULL DEFAULT 'cron',
			schedule TEXT NOT NULL,
			message TEXT NOT NULL,
			channel TEXT NOT NULL,
			chat_id TEXT NOT NULL,
			account_id TEXT NOT NULL DEFAULT '',
			timezone TEXT NOT NULL DEFAULT 'UTC',
			enabled BOOLEAN NOT NULL DEFAULT true,
			last_run TIMESTAMP,
			next_run TIMESTAMP,
			locked_by TEXT,
			locked_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_cron_jobs_schedule ON cron_jobs (tenant_id, enabled, next_run)`,
	}
}

func (d *DBStore) chatTasksMigrationSQL() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS chat_tasks (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			agent_id TEXT NOT NULL,
			session_key TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			message TEXT NOT NULL,
			result TEXT NOT NULL DEFAULT '',
			error TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			started_at TIMESTAMP,
			done_at TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_chat_tasks_agent ON chat_tasks (tenant_id, agent_id, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_chat_tasks_status ON chat_tasks (status, updated_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_chat_tasks_session ON chat_tasks (tenant_id, agent_id, session_key, created_at DESC)`,
	}
}

func (d *DBStore) migrationSQL() []string {
	// Use TEXT for JSON columns (works in both postgres and sqlite).
	// Postgres users can alter to JSONB later for indexing.
	return []string{
		`CREATE TABLE IF NOT EXISTS configs (
			tenant_id TEXT NOT NULL,
			data TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (tenant_id)
		)`,
		`CREATE TABLE IF NOT EXISTS agents (
			tenant_id TEXT NOT NULL,
			agent_id TEXT NOT NULL,
			name TEXT NOT NULL DEFAULT '',
			model TEXT NOT NULL DEFAULT '',
			config TEXT NOT NULL DEFAULT '{}',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (tenant_id, agent_id)
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_files (
			tenant_id TEXT NOT NULL,
			agent_id TEXT NOT NULL,
			filename TEXT NOT NULL,
			content TEXT NOT NULL DEFAULT '',
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (tenant_id, agent_id, filename)
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			tenant_id TEXT NOT NULL,
			agent_id TEXT NOT NULL,
			session_key TEXT NOT NULL,
			messages TEXT NOT NULL DEFAULT '[]',
			message_count INTEGER NOT NULL DEFAULT 0,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (tenant_id, agent_id, session_key)
		)`,
		`CREATE TABLE IF NOT EXISTS memory_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			agent_id TEXT NOT NULL,
			session_id TEXT NOT NULL DEFAULT '',
			role TEXT NOT NULL DEFAULT '',
			content TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		d.memoryLogsIndex(),
		`CREATE TABLE IF NOT EXISTS cron_jobs (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			agent_id TEXT NOT NULL,
			name TEXT NOT NULL DEFAULT '',
			type TEXT NOT NULL DEFAULT 'cron',
			schedule TEXT NOT NULL,
			message TEXT NOT NULL,
			channel TEXT NOT NULL,
			chat_id TEXT NOT NULL,
			account_id TEXT NOT NULL DEFAULT '',
			timezone TEXT NOT NULL DEFAULT 'UTC',
			enabled BOOLEAN NOT NULL DEFAULT true,
			last_run TIMESTAMP,
			next_run TIMESTAMP,
			locked_by TEXT,
			locked_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_cron_jobs_schedule ON cron_jobs (tenant_id, enabled, next_run)`,
		`CREATE TABLE IF NOT EXISTS chat_tasks (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			agent_id TEXT NOT NULL,
			session_key TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			message TEXT NOT NULL,
			result TEXT NOT NULL DEFAULT '',
			error TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			started_at TIMESTAMP,
			done_at TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_chat_tasks_agent ON chat_tasks (tenant_id, agent_id, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_chat_tasks_status ON chat_tasks (status, updated_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_chat_tasks_session ON chat_tasks (tenant_id, agent_id, session_key, created_at DESC)`,
	}
}

func (d *DBStore) memoryLogsIndex() string {
	return `CREATE INDEX IF NOT EXISTS idx_memory_logs_search 
		ON memory_logs (tenant_id, agent_id, created_at DESC)`
}

func (d *DBStore) Close() error {
	return d.db.Close()
}

// placeholder returns the correct placeholder for the dialect.
func (d *DBStore) ph(n int) string {
	if d.dialect == "postgres" {
		return fmt.Sprintf("$%d", n)
	}
	return "?"
}

// --- Config ---

func (d *DBStore) GetConfig(ctx context.Context, tenantID string) (*TenantConfig, error) {
	row := d.db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT data, created_at, updated_at FROM configs WHERE tenant_id = %s", d.ph(1)),
		tenantID)

	var dataStr string
	var cfg TenantConfig
	cfg.TenantID = tenantID
	if err := row.Scan(&dataStr, &cfg.CreatedAt, &cfg.UpdatedAt); err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(dataStr), &cfg.Data)
	return &cfg, nil
}

func (d *DBStore) SaveConfig(ctx context.Context, tenantID string, cfg *TenantConfig) error {
	data, _ := json.Marshal(cfg.Data)
	now := time.Now()

	if d.dialect == "postgres" {
		_, err := d.db.ExecContext(ctx,
			`INSERT INTO configs (tenant_id, data, created_at, updated_at)
			 VALUES ($1, $2, $3, $4)
			 ON CONFLICT (tenant_id) DO UPDATE SET data = $2, updated_at = $4`,
			tenantID, string(data), now, now)
		return err
	}
	// SQLite
	_, err := d.db.ExecContext(ctx,
		`INSERT INTO configs (tenant_id, data, created_at, updated_at)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT (tenant_id) DO UPDATE SET data = excluded.data, updated_at = excluded.updated_at`,
		tenantID, string(data), now, now)
	return err
}

func (d *DBStore) DeleteConfig(ctx context.Context, tenantID string) error {
	_, err := d.db.ExecContext(ctx,
		fmt.Sprintf("DELETE FROM configs WHERE tenant_id = %s", d.ph(1)), tenantID)
	return err
}

// --- Agents ---

func (d *DBStore) ListAgents(ctx context.Context, tenantID string) ([]AgentRecord, error) {
	rows, err := d.db.QueryContext(ctx,
		fmt.Sprintf("SELECT agent_id, name, model, config, created_at, updated_at FROM agents WHERE tenant_id = %s ORDER BY created_at", d.ph(1)),
		tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []AgentRecord
	for rows.Next() {
		var ag AgentRecord
		var cfgStr string
		if err := rows.Scan(&ag.ID, &ag.Name, &ag.Model, &cfgStr, &ag.CreatedAt, &ag.UpdatedAt); err != nil {
			continue
		}
		json.Unmarshal([]byte(cfgStr), &ag.Config)
		agents = append(agents, ag)
	}
	return agents, nil
}

func (d *DBStore) GetAgent(ctx context.Context, tenantID, agentID string) (*AgentRecord, error) {
	row := d.db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT agent_id, name, model, config, created_at, updated_at FROM agents WHERE tenant_id = %s AND agent_id = %s", d.ph(1), d.ph(2)),
		tenantID, agentID)

	var ag AgentRecord
	var cfgStr string
	if err := row.Scan(&ag.ID, &ag.Name, &ag.Model, &cfgStr, &ag.CreatedAt, &ag.UpdatedAt); err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(cfgStr), &ag.Config)

	// Load workspace files
	ag.Workspace = make(map[string]string)
	files, _ := d.ListWorkspaceFiles(ctx, tenantID, agentID)
	for _, fname := range files {
		data, err := d.GetWorkspaceFile(ctx, tenantID, agentID, fname)
		if err == nil {
			ag.Workspace[fname] = string(data)
		}
	}

	return &ag, nil
}

func (d *DBStore) SaveAgent(ctx context.Context, tenantID string, agent *AgentRecord) error {
	cfgData, _ := json.Marshal(agent.Config)
	now := time.Now()

	if d.dialect == "postgres" {
		_, err := d.db.ExecContext(ctx,
			`INSERT INTO agents (tenant_id, agent_id, name, model, config, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)
			 ON CONFLICT (tenant_id, agent_id) DO UPDATE SET name=$3, model=$4, config=$5, updated_at=$7`,
			tenantID, agent.ID, agent.Name, agent.Model, string(cfgData), now, now)
		if err != nil {
			return err
		}
	} else {
		_, err := d.db.ExecContext(ctx,
			`INSERT INTO agents (tenant_id, agent_id, name, model, config, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?)
			 ON CONFLICT (tenant_id, agent_id) DO UPDATE SET
			   name=excluded.name, model=excluded.model, config=excluded.config, updated_at=excluded.updated_at`,
			tenantID, agent.ID, agent.Name, agent.Model, string(cfgData), now, now)
		if err != nil {
			return err
		}
	}

	// Save workspace files
	for fname, content := range agent.Workspace {
		if err := d.SaveWorkspaceFile(ctx, tenantID, agent.ID, fname, []byte(content)); err != nil {
			return err
		}
	}

	return nil
}

func (d *DBStore) DeleteAgent(ctx context.Context, tenantID, agentID string) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, table := range []string{"workspace_files", "sessions", "memory_logs", "agents"} {
		if d.dialect == "postgres" {
			tx.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE tenant_id = $1 AND agent_id = $2", table), tenantID, agentID)
		} else {
			tx.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE tenant_id = ? AND agent_id = ?", table), tenantID, agentID)
		}
	}

	return tx.Commit()
}

// --- Sessions ---

func (d *DBStore) GetSession(ctx context.Context, tenantID, agentID, sessionKey string) (*SessionRecord, error) {
	row := d.db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT messages, updated_at FROM sessions WHERE tenant_id = %s AND agent_id = %s AND session_key = %s", d.ph(1), d.ph(2), d.ph(3)),
		tenantID, agentID, sessionKey)

	var msgsStr string
	var rec SessionRecord
	if err := row.Scan(&msgsStr, &rec.UpdatedAt); err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(msgsStr), &rec.Messages)
	return &rec, nil
}

func (d *DBStore) SaveSession(ctx context.Context, tenantID, agentID, sessionKey string, session *SessionRecord) error {
	msgsData, _ := json.Marshal(session.Messages)
	now := time.Now()
	count := len(session.Messages)

	if d.dialect == "postgres" {
		_, err := d.db.ExecContext(ctx,
			`INSERT INTO sessions (tenant_id, agent_id, session_key, messages, message_count, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6)
			 ON CONFLICT (tenant_id, agent_id, session_key) DO UPDATE SET messages=$4, message_count=$5, updated_at=$6`,
			tenantID, agentID, sessionKey, string(msgsData), count, now)
		return err
	}
	_, err := d.db.ExecContext(ctx,
		`INSERT INTO sessions (tenant_id, agent_id, session_key, messages, message_count, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT (tenant_id, agent_id, session_key) DO UPDATE SET
		   messages=excluded.messages, message_count=excluded.message_count, updated_at=excluded.updated_at`,
		tenantID, agentID, sessionKey, string(msgsData), count, now)
	return err
}

func (d *DBStore) ListSessions(ctx context.Context, tenantID, agentID string) ([]SessionMeta, error) {
	rows, err := d.db.QueryContext(ctx,
		fmt.Sprintf("SELECT session_key, message_count, updated_at FROM sessions WHERE tenant_id = %s AND agent_id = %s ORDER BY updated_at DESC", d.ph(1), d.ph(2)),
		tenantID, agentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metas []SessionMeta
	for rows.Next() {
		var m SessionMeta
		rows.Scan(&m.Key, &m.MessageCount, &m.UpdatedAt)
		metas = append(metas, m)
	}
	return metas, nil
}

func (d *DBStore) DeleteSession(ctx context.Context, tenantID, agentID, sessionKey string) error {
	_, err := d.db.ExecContext(ctx,
		fmt.Sprintf("DELETE FROM sessions WHERE tenant_id = %s AND agent_id = %s AND session_key = %s", d.ph(1), d.ph(2), d.ph(3)),
		tenantID, agentID, sessionKey)
	return err
}

// --- Memory ---

func (d *DBStore) GetMemory(ctx context.Context, tenantID, agentID string) (string, error) {
	data, err := d.GetWorkspaceFile(ctx, tenantID, agentID, "MEMORY.md")
	if err != nil {
		return "", nil
	}
	return string(data), nil
}

func (d *DBStore) SaveMemory(ctx context.Context, tenantID, agentID, content string) error {
	return d.SaveWorkspaceFile(ctx, tenantID, agentID, "MEMORY.md", []byte(content))
}

func (d *DBStore) SearchMemory(ctx context.Context, tenantID, agentID, query string, limit int) ([]MemoryEntry, error) {
	if limit <= 0 {
		limit = 20
	}

	var rows *sql.Rows
	var err error

	if d.dialect == "postgres" {
		// Postgres: use ILIKE for case-insensitive search
		rows, err = d.db.QueryContext(ctx,
			`SELECT content, role, session_id, created_at FROM memory_logs
			 WHERE tenant_id = $1 AND agent_id = $2 AND content ILIKE '%' || $3 || '%'
			 ORDER BY created_at DESC LIMIT $4`,
			tenantID, agentID, query, limit)
	} else {
		// SQLite: LIKE is case-insensitive by default for ASCII
		rows, err = d.db.QueryContext(ctx,
			`SELECT content, role, session_id, created_at FROM memory_logs
			 WHERE tenant_id = ? AND agent_id = ? AND content LIKE '%' || ? || '%'
			 ORDER BY created_at DESC LIMIT ?`,
			tenantID, agentID, query, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []MemoryEntry
	for rows.Next() {
		var e MemoryEntry
		rows.Scan(&e.Content, &e.Role, &e.SessionID, &e.Timestamp)
		entries = append(entries, e)
	}
	return entries, nil
}

func (d *DBStore) AppendMemoryLog(ctx context.Context, tenantID, agentID string, entry MemoryEntry) error {
	if d.dialect == "postgres" {
		_, err := d.db.ExecContext(ctx,
			`INSERT INTO memory_logs (tenant_id, agent_id, session_id, role, content, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			tenantID, agentID, entry.SessionID, entry.Role, entry.Content, entry.Timestamp)
		return err
	}
	_, err := d.db.ExecContext(ctx,
		`INSERT INTO memory_logs (tenant_id, agent_id, session_id, role, content, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		tenantID, agentID, entry.SessionID, entry.Role, entry.Content, entry.Timestamp)
	return err
}

// --- Workspace Files ---

func (d *DBStore) GetWorkspaceFile(ctx context.Context, tenantID, agentID, filename string) ([]byte, error) {
	row := d.db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT content FROM workspace_files WHERE tenant_id = %s AND agent_id = %s AND filename = %s", d.ph(1), d.ph(2), d.ph(3)),
		tenantID, agentID, filename)

	var content string
	if err := row.Scan(&content); err != nil {
		return nil, err
	}
	return []byte(content), nil
}

func (d *DBStore) SaveWorkspaceFile(ctx context.Context, tenantID, agentID, filename string, data []byte) error {
	now := time.Now()
	if d.dialect == "postgres" {
		_, err := d.db.ExecContext(ctx,
			`INSERT INTO workspace_files (tenant_id, agent_id, filename, content, updated_at)
			 VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (tenant_id, agent_id, filename) DO UPDATE SET content=$4, updated_at=$5`,
			tenantID, agentID, filename, string(data), now)
		return err
	}
	_, err := d.db.ExecContext(ctx,
		`INSERT INTO workspace_files (tenant_id, agent_id, filename, content, updated_at)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT (tenant_id, agent_id, filename) DO UPDATE SET content=excluded.content, updated_at=excluded.updated_at`,
		tenantID, agentID, filename, string(data), now)
	return err
}

func (d *DBStore) ListWorkspaceFiles(ctx context.Context, tenantID, agentID string) ([]string, error) {
	rows, err := d.db.QueryContext(ctx,
		fmt.Sprintf("SELECT filename FROM workspace_files WHERE tenant_id = %s AND agent_id = %s ORDER BY filename", d.ph(1), d.ph(2)),
		tenantID, agentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []string
	for rows.Next() {
		var f string
		rows.Scan(&f)
		files = append(files, f)
	}
	return files, nil
}

// --- Cron Jobs ---

func (d *DBStore) ListCronJobs(ctx context.Context, tenantID string) ([]CronJobRecord, error) {
	rows, err := d.db.QueryContext(ctx,
		fmt.Sprintf("SELECT id, tenant_id, agent_id, name, type, schedule, message, channel, chat_id, account_id, timezone, enabled, last_run, next_run, created_at FROM cron_jobs WHERE tenant_id = %s ORDER BY created_at", d.ph(1)),
		tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return d.scanCronJobs(rows)
}

func (d *DBStore) GetCronJob(ctx context.Context, tenantID, jobID string) (*CronJobRecord, error) {
	row := d.db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT id, tenant_id, agent_id, name, type, schedule, message, channel, chat_id, account_id, timezone, enabled, last_run, next_run, created_at FROM cron_jobs WHERE tenant_id = %s AND id = %s", d.ph(1), d.ph(2)),
		tenantID, jobID)
	var j CronJobRecord
	var lastRun, nextRun sql.NullTime
	if err := row.Scan(&j.ID, &j.TenantID, &j.AgentID, &j.Name, &j.Type, &j.Schedule, &j.Message, &j.Channel, &j.ChatID, &j.AccountID, &j.Timezone, &j.Enabled, &lastRun, &nextRun, &j.CreatedAt); err != nil {
		return nil, err
	}
	if lastRun.Valid {
		j.LastRun = &lastRun.Time
	}
	if nextRun.Valid {
		j.NextRun = &nextRun.Time
	}
	return &j, nil
}

func (d *DBStore) SaveCronJob(ctx context.Context, tenantID string, job *CronJobRecord) error {
	now := time.Now()
	if d.dialect == "postgres" {
		_, err := d.db.ExecContext(ctx,
			`INSERT INTO cron_jobs (id, tenant_id, agent_id, name, type, schedule, message, channel, chat_id, account_id, timezone, enabled, last_run, next_run, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
			 ON CONFLICT (id) DO UPDATE SET name=$4, type=$5, schedule=$6, message=$7, channel=$8, chat_id=$9, account_id=$10, timezone=$11, enabled=$12, last_run=$13, next_run=$14`,
			job.ID, tenantID, job.AgentID, job.Name, job.Type, job.Schedule, job.Message, job.Channel, job.ChatID, job.AccountID, job.Timezone, job.Enabled, job.LastRun, job.NextRun, now)
		return err
	}
	_, err := d.db.ExecContext(ctx,
		`INSERT INTO cron_jobs (id, tenant_id, agent_id, name, type, schedule, message, channel, chat_id, account_id, timezone, enabled, last_run, next_run, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT (id) DO UPDATE SET
		   name=excluded.name, type=excluded.type, schedule=excluded.schedule, message=excluded.message,
		   channel=excluded.channel, chat_id=excluded.chat_id, account_id=excluded.account_id,
		   timezone=excluded.timezone, enabled=excluded.enabled, last_run=excluded.last_run, next_run=excluded.next_run`,
		job.ID, tenantID, job.AgentID, job.Name, job.Type, job.Schedule, job.Message, job.Channel, job.ChatID, job.AccountID, job.Timezone, job.Enabled, job.LastRun, job.NextRun, now)
	return err
}

func (d *DBStore) DeleteCronJob(ctx context.Context, tenantID, jobID string) error {
	_, err := d.db.ExecContext(ctx,
		fmt.Sprintf("DELETE FROM cron_jobs WHERE tenant_id = %s AND id = %s", d.ph(1), d.ph(2)),
		tenantID, jobID)
	return err
}

func (d *DBStore) GetDueCronJobs(ctx context.Context, now time.Time) ([]CronJobRecord, error) {
	var rows *sql.Rows
	var err error
	if d.dialect == "postgres" {
		rows, err = d.db.QueryContext(ctx,
			`SELECT id, tenant_id, agent_id, name, type, schedule, message, channel, chat_id, account_id, timezone, enabled, last_run, next_run, created_at
			 FROM cron_jobs WHERE enabled = true AND next_run <= $1 ORDER BY next_run`, now)
	} else {
		rows, err = d.db.QueryContext(ctx,
			`SELECT id, tenant_id, agent_id, name, type, schedule, message, channel, chat_id, account_id, timezone, enabled, last_run, next_run, created_at
			 FROM cron_jobs WHERE enabled = 1 AND next_run <= ? ORDER BY next_run`, now)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return d.scanCronJobs(rows)
}

func (d *DBStore) LockCronJob(ctx context.Context, jobID, instanceID string) (bool, error) {
	now := time.Now()
	fiveMinAgo := now.Add(-5 * time.Minute)
	var res sql.Result
	var err error
	if d.dialect == "postgres" {
		res, err = d.db.ExecContext(ctx,
			`UPDATE cron_jobs SET locked_by=$1, locked_at=$2 WHERE id=$3 AND (locked_by IS NULL OR locked_at < $4)`,
			instanceID, now, jobID, fiveMinAgo)
	} else {
		res, err = d.db.ExecContext(ctx,
			`UPDATE cron_jobs SET locked_by=?, locked_at=? WHERE id=? AND (locked_by IS NULL OR locked_at < ?)`,
			instanceID, now, jobID, fiveMinAgo)
	}
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func (d *DBStore) UpdateCronJobRun(ctx context.Context, jobID string, lastRun, nextRun time.Time) error {
	if d.dialect == "postgres" {
		_, err := d.db.ExecContext(ctx,
			`UPDATE cron_jobs SET last_run=$1, next_run=$2, locked_by=NULL, locked_at=NULL WHERE id=$3`,
			lastRun, nextRun, jobID)
		return err
	}
	_, err := d.db.ExecContext(ctx,
		`UPDATE cron_jobs SET last_run=?, next_run=?, locked_by=NULL, locked_at=NULL WHERE id=?`,
		lastRun, nextRun, jobID)
	return err
}

func (d *DBStore) scanCronJobs(rows *sql.Rows) ([]CronJobRecord, error) {
	var jobs []CronJobRecord
	for rows.Next() {
		var j CronJobRecord
		var lastRun, nextRun sql.NullTime
		if err := rows.Scan(&j.ID, &j.TenantID, &j.AgentID, &j.Name, &j.Type, &j.Schedule, &j.Message, &j.Channel, &j.ChatID, &j.AccountID, &j.Timezone, &j.Enabled, &lastRun, &nextRun, &j.CreatedAt); err != nil {
			continue
		}
		if lastRun.Valid {
			j.LastRun = &lastRun.Time
		}
		if nextRun.Valid {
			j.NextRun = &nextRun.Time
		}
		jobs = append(jobs, j)
	}
	return jobs, nil
}

// --- Chat Tasks ---

func (d *DBStore) CreateChatTask(ctx context.Context, tenantID string, t *ChatTaskRecord) error {
	if t.Status == "" {
		t.Status = ChatTaskPending
	}
	now := time.Now().UTC()
	if t.CreatedAt.IsZero() {
		t.CreatedAt = now
	}
	t.UpdatedAt = now
	t.TenantID = tenantID

	q := fmt.Sprintf(
		`INSERT INTO chat_tasks (id, tenant_id, agent_id, session_key, status, message, result, error, created_at, started_at, done_at, updated_at)
		 VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)`,
		d.ph(1), d.ph(2), d.ph(3), d.ph(4), d.ph(5), d.ph(6), d.ph(7), d.ph(8), d.ph(9), d.ph(10), d.ph(11), d.ph(12))
	_, err := d.db.ExecContext(ctx, q,
		t.ID, tenantID, t.AgentID, t.SessionKey, string(t.Status),
		t.Message, t.Result, t.Error,
		t.CreatedAt, nullableTime(t.StartedAt), nullableTime(t.DoneAt), t.UpdatedAt,
	)
	return err
}

func (d *DBStore) UpdateChatTask(ctx context.Context, tenantID string, t *ChatTaskRecord) error {
	t.UpdatedAt = time.Now().UTC()
	q := fmt.Sprintf(
		`UPDATE chat_tasks
		 SET status=%s, result=%s, error=%s, started_at=%s, done_at=%s, updated_at=%s
		 WHERE id=%s AND tenant_id=%s`,
		d.ph(1), d.ph(2), d.ph(3), d.ph(4), d.ph(5), d.ph(6), d.ph(7), d.ph(8))
	res, err := d.db.ExecContext(ctx, q,
		string(t.Status), t.Result, t.Error,
		nullableTime(t.StartedAt), nullableTime(t.DoneAt), t.UpdatedAt,
		t.ID, tenantID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (d *DBStore) GetChatTask(ctx context.Context, tenantID, taskID string) (*ChatTaskRecord, error) {
	q := fmt.Sprintf(
		`SELECT id, tenant_id, agent_id, session_key, status, message, result, error, created_at, started_at, done_at, updated_at
		 FROM chat_tasks WHERE id=%s AND tenant_id=%s`, d.ph(1), d.ph(2))
	row := d.db.QueryRowContext(ctx, q, taskID, tenantID)
	return scanChatTask(row.Scan)
}

func (d *DBStore) ListChatTasks(ctx context.Context, tenantID string, f ChatTaskFilters) ([]ChatTaskRecord, error) {
	limit := f.Limit
	if limit <= 0 || limit > 500 {
		limit = 50
	}

	// Build WHERE dynamically; placeholder index increments per condition
	// to stay compatible with both postgres ($N) and sqlite (?).
	var (
		where = []string{fmt.Sprintf("tenant_id=%s", d.ph(1))}
		args  = []any{tenantID}
		idx   = 2
	)
	if f.AgentID != "" {
		where = append(where, fmt.Sprintf("agent_id=%s", d.ph(idx)))
		args = append(args, f.AgentID)
		idx++
	}
	if f.SessionKey != "" {
		where = append(where, fmt.Sprintf("session_key=%s", d.ph(idx)))
		args = append(args, f.SessionKey)
		idx++
	}
	if f.Status != "" {
		where = append(where, fmt.Sprintf("status=%s", d.ph(idx)))
		args = append(args, string(f.Status))
		idx++
	}
	limPlaceholder := d.ph(idx)
	offPlaceholder := d.ph(idx + 1)
	args = append(args, limit, f.Offset)

	q := fmt.Sprintf(
		`SELECT id, tenant_id, agent_id, session_key, status, message, result, error, created_at, started_at, done_at, updated_at
		 FROM chat_tasks WHERE %s
		 ORDER BY created_at DESC LIMIT %s OFFSET %s`,
		strings.Join(where, " AND "), limPlaceholder, offPlaceholder)

	rows, err := d.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ChatTaskRecord
	for rows.Next() {
		t, err := scanChatTask(rows.Scan)
		if err != nil {
			continue
		}
		out = append(out, *t)
	}
	return out, rows.Err()
}

func (d *DBStore) DeleteChatTask(ctx context.Context, tenantID, taskID string) error {
	q := fmt.Sprintf(`DELETE FROM chat_tasks WHERE id=%s AND tenant_id=%s`, d.ph(1), d.ph(2))
	_, err := d.db.ExecContext(ctx, q, taskID, tenantID)
	return err
}

// scanChatTask works with both *sql.Row.Scan and *sql.Rows.Scan via the
// passed function. Centralises the column order in one place.
func scanChatTask(scan func(...any) error) (*ChatTaskRecord, error) {
	var t ChatTaskRecord
	var status string
	var startedAt, doneAt sql.NullTime
	if err := scan(&t.ID, &t.TenantID, &t.AgentID, &t.SessionKey, &status,
		&t.Message, &t.Result, &t.Error,
		&t.CreatedAt, &startedAt, &doneAt, &t.UpdatedAt); err != nil {
		return nil, err
	}
	t.Status = ChatTaskStatus(status)
	if startedAt.Valid {
		t.StartedAt = &startedAt.Time
	}
	if doneAt.Valid {
		t.DoneAt = &doneAt.Time
	}
	return &t, nil
}

// nullableTime converts a *time.Time into a sql.NullTime for DB writes.
func nullableTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

// Ensure DBStore implements Store.
var _ Store = (*DBStore)(nil)

// Suppress unused import warning for strings.
var _ = strings.TrimSpace
