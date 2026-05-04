// Package store provides a pluggable storage backend for FastClaw.
// Default: file-based (single-user). For cloud multi-tenant: database-backed.
package store

import (
	"context"
	"errors"
	"time"
)

// Store is the unified interface for all persistent data.
// File-based impl reads/writes to ~/.fastclaw; DB impl uses SQL tables with tenant isolation.
type Store interface {
	// Config
	GetConfig(ctx context.Context, tenantID string) (*TenantConfig, error)
	SaveConfig(ctx context.Context, tenantID string, cfg *TenantConfig) error
	DeleteConfig(ctx context.Context, tenantID string) error

	// Agents
	ListAgents(ctx context.Context, tenantID string) ([]AgentRecord, error)
	GetAgent(ctx context.Context, tenantID, agentID string) (*AgentRecord, error)
	SaveAgent(ctx context.Context, tenantID string, agent *AgentRecord) error
	DeleteAgent(ctx context.Context, tenantID, agentID string) error

	// Sessions
	GetSession(ctx context.Context, tenantID, agentID, sessionKey string) (*SessionRecord, error)
	SaveSession(ctx context.Context, tenantID, agentID, sessionKey string, session *SessionRecord) error
	ListSessions(ctx context.Context, tenantID, agentID string) ([]SessionMeta, error)
	DeleteSession(ctx context.Context, tenantID, agentID, sessionKey string) error

	// Memory
	GetMemory(ctx context.Context, tenantID, agentID string) (string, error) // MEMORY.md content
	SaveMemory(ctx context.Context, tenantID, agentID, content string) error
	SearchMemory(ctx context.Context, tenantID, agentID, query string, limit int) ([]MemoryEntry, error)
	AppendMemoryLog(ctx context.Context, tenantID, agentID string, entry MemoryEntry) error

	// Workspace files (SOUL.md, AGENTS.md, etc.)
	GetWorkspaceFile(ctx context.Context, tenantID, agentID, filename string) ([]byte, error)
	SaveWorkspaceFile(ctx context.Context, tenantID, agentID, filename string, data []byte) error
	ListWorkspaceFiles(ctx context.Context, tenantID, agentID string) ([]string, error)

	// Cron Jobs
	ListCronJobs(ctx context.Context, tenantID string) ([]CronJobRecord, error)
	GetCronJob(ctx context.Context, tenantID, jobID string) (*CronJobRecord, error)
	SaveCronJob(ctx context.Context, tenantID string, job *CronJobRecord) error
	DeleteCronJob(ctx context.Context, tenantID, jobID string) error
	GetDueCronJobs(ctx context.Context, now time.Time) ([]CronJobRecord, error) // cross-tenant
	LockCronJob(ctx context.Context, jobID, instanceID string) (bool, error)
	UpdateCronJobRun(ctx context.Context, jobID string, lastRun, nextRun time.Time) error

	// Tasks (web chat async tasks).
	// FileStore implements only the per-task methods; ListChatTasks returns
	// ErrNotSupported on file backend so callers should feature-detect.
	CreateChatTask(ctx context.Context, tenantID string, task *ChatTaskRecord) error
	UpdateChatTask(ctx context.Context, tenantID string, task *ChatTaskRecord) error
	GetChatTask(ctx context.Context, tenantID, taskID string) (*ChatTaskRecord, error)
	ListChatTasks(ctx context.Context, tenantID string, filters ChatTaskFilters) ([]ChatTaskRecord, error)
	DeleteChatTask(ctx context.Context, tenantID, taskID string) error

	// Close releases resources.
	Close() error
}

// ErrNotSupported is returned by Store backends that don't implement an
// optional method (e.g. FileStore.ListChatTasks).
var ErrNotSupported = errors.New("operation not supported by this storage backend")

// CronJobRecord holds a scheduled job.
type CronJobRecord struct {
	ID        string     `json:"id"`
	TenantID  string     `json:"tenantId"`
	AgentID   string     `json:"agentId"`
	Name      string     `json:"name"`
	Type      string     `json:"type"`      // cron, interval, once
	Schedule  string     `json:"schedule"`
	Message   string     `json:"message"`
	Channel   string     `json:"channel"`
	ChatID    string     `json:"chatId"`
	AccountID string     `json:"accountId"`
	Timezone  string     `json:"timezone"`
	Enabled   bool       `json:"enabled"`
	LastRun   *time.Time `json:"lastRun,omitempty"`
	NextRun   *time.Time `json:"nextRun,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
}

// TenantConfig holds the full config for a tenant (maps to fastclaw.json for file store).
type TenantConfig struct {
	TenantID  string                 `json:"tenantId"`
	Data      map[string]interface{} `json:"data"` // raw config JSON
	CreatedAt time.Time              `json:"createdAt"`
	UpdatedAt time.Time              `json:"updatedAt"`
}

// AgentRecord is the persisted state for one agent.
type AgentRecord struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Model       string            `json:"model"`
	Config      map[string]interface{} `json:"config"` // agent.json content
	Workspace   map[string]string `json:"workspace"` // filename -> content (SOUL.md, etc.)
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

// SessionRecord holds a conversation session.
type SessionRecord struct {
	Messages  []SessionMessage `json:"messages"`
	UpdatedAt time.Time        `json:"updatedAt"`
}

// SessionMessage is a single message in a session.
type SessionMessage struct {
	Role       string      `json:"role"`
	Content    string      `json:"content"`
	ToolCalls  interface{} `json:"toolCalls,omitempty"`
	ToolCallID string      `json:"toolCallId,omitempty"`
	Timestamp  time.Time   `json:"timestamp"`
}

// SessionMeta is summary info for a session (for listing).
type SessionMeta struct {
	Key          string    `json:"key"`
	MessageCount int       `json:"messageCount"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// MemoryEntry is one searchable memory log entry.
type MemoryEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	SessionID string    `json:"sessionId,omitempty"`
}

// ChatTaskStatus is the lifecycle state of a web chat task.
// Distinct from internal/taskqueue.TaskStatus (which models IM task queue);
// kept separate so the two systems can evolve independently.
type ChatTaskStatus string

const (
	ChatTaskPending   ChatTaskStatus = "pending"
	ChatTaskRunning   ChatTaskStatus = "running"
	ChatTaskDone      ChatTaskStatus = "done"
	ChatTaskFailed    ChatTaskStatus = "failed"
	ChatTaskCancelled ChatTaskStatus = "cancelled"
)

// ChatTaskRecord is one persisted web chat task. Web chat tasks are the
// async unit of work created by POST /api/chat/submit; see internal/taskrunner.
type ChatTaskRecord struct {
	ID         string         `json:"id"`
	TenantID   string         `json:"tenantId"`
	AgentID    string         `json:"agentId"`
	SessionKey string         `json:"sessionKey"`
	Status     ChatTaskStatus `json:"status"`
	Message    string         `json:"message"`           // user input
	Result     string         `json:"result,omitempty"`  // final assistant reply
	Error      string         `json:"error,omitempty"`   // error message if failed
	CreatedAt  time.Time      `json:"createdAt"`
	StartedAt  *time.Time     `json:"startedAt,omitempty"`
	DoneAt     *time.Time     `json:"doneAt,omitempty"`
	UpdatedAt  time.Time      `json:"updatedAt"`
}

// ChatTaskFilters narrows ListChatTasks results.
type ChatTaskFilters struct {
	AgentID    string         // optional, exact match
	SessionKey string         // optional, exact match
	Status     ChatTaskStatus // optional, exact match
	Limit      int            // 0 = backend default (typically 50)
	Offset     int
}

// StorageType identifies the storage backend.
type StorageType string

const (
	StorageFile     StorageType = "file"
	StoragePostgres StorageType = "postgres"
	StorageSQLite   StorageType = "sqlite"
)

// StorageConfig is the config block for choosing and configuring the store.
type StorageConfig struct {
	Type     StorageType `json:"type"`               // "file" (default), "postgres", "sqlite"
	DSN      string      `json:"dsn,omitempty"`       // database connection string
	AutoMigrate bool    `json:"autoMigrate,omitempty"` // auto-create tables on startup
}

// DefaultTenantID is used for single-user file-based mode.
const DefaultTenantID = "default"
