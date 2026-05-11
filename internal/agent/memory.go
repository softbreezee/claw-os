package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fastclaw-ai/fastclaw/internal/privacy"
	"github.com/fastclaw-ai/fastclaw/internal/provider"
)

// PGMemoryStore is the interface for PostgreSQL-backed memory persistence.
// Implemented by *pg.MemoryStore; defined here to avoid import cycles.
type PGMemoryStore interface {
	Insert(ctx context.Context, agentID, kind, content string, embedding []float32, tags []string) (string, error)
	LoadAll(ctx context.Context, agentID string) ([]interface{ GetContent() string }, error)
}

// Memory manages the dual-layer memory system (MEMORY.md + HISTORY.md).
// When pgStore is set, facts are also persisted to PostgreSQL.
type Memory struct {
	mu        sync.Mutex // serialises all file writes to prevent concurrent-write races
	workspace string
	agentID   string
	pgStore   interface {
		Insert(ctx context.Context, agentID, kind, content string, embedding []float32, tags []string) (string, error)
	}
}

// SetPGStore attaches a PostgreSQL memory store for dual-write.
func (m *Memory) SetPGStore(store interface {
	Insert(ctx context.Context, agentID, kind, content string, embedding []float32, tags []string) (string, error)
}, agentID string) {
	m.agentID = agentID
	m.pgStore = store
}

// NewMemory creates a new memory manager.
//
// Defensive against empty / relative workspace paths: a blank workspace
// would make memoryPath() return the bare string "MEMORY.md", which then
// gets written relative to whatever the daemon's startup cwd happens to
// be — historically that meant MEMORY.md polluting the developer's repo
// root when fastclaw was launched from a source checkout. Always anchor
// the workspace to an absolute path; if the caller provided nothing,
// fall back to a clearly-marked orphan dir under ~/.fastclaw so the
// data is still recoverable but never lands somewhere surprising.
func NewMemory(workspace string) *Memory {
	if strings.TrimSpace(workspace) == "" {
		fallback := orphanWorkspaceDir()
		slog.Warn("Memory created with empty workspace; using fallback to avoid cwd pollution",
			"fallback", fallback)
		workspace = fallback
	}
	if abs, err := filepath.Abs(workspace); err == nil {
		workspace = abs
	}
	return &Memory{workspace: workspace}
}

// orphanWorkspaceDir returns ~/.fastclaw/agents/_orphan/agent — a
// quarantine location for Memory instances built without a real
// workspace. We don't propagate the error from UserHomeDir because
// SaveMemory will surface any real I/O failure later; the only goal
// here is "absolute path that is NOT the process cwd".
func orphanWorkspaceDir() string {
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".fastclaw", "agents", "_orphan", "agent")
	}
	// Last resort: temp dir is still better than cwd.
	return filepath.Join(os.TempDir(), "fastclaw-orphan-agent")
}

// memoryPath returns the path to MEMORY.md.
func (m *Memory) memoryPath() string {
	return filepath.Join(m.workspace, "MEMORY.md")
}

// historyPath returns the path to HISTORY.md.
func (m *Memory) historyPath() string {
	return filepath.Join(m.workspace, "HISTORY.md")
}

// LoadMemory reads the long-term memory file.
func (m *Memory) LoadMemory() string {
	data, err := os.ReadFile(m.memoryPath())
	if err != nil {
		return ""
	}
	return string(data)
}

// SaveMemory overwrites the long-term memory file.
func (m *Memory) SaveMemory(content string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	os.MkdirAll(m.workspace, 0o755)
	return os.WriteFile(m.memoryPath(), []byte(content), 0o644)
}

// AppendHistory adds an entry to the history log.
func (m *Memory) AppendHistory(entry string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	os.MkdirAll(m.workspace, 0o755)
	f, err := os.OpenFile(m.historyPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	_, err = fmt.Fprintf(f, "- [%s] %s\n", timestamp, entry)
	return err
}

// LoadHistory reads the history log.
func (m *Memory) LoadHistory() string {
	data, err := os.ReadFile(m.historyPath())
	if err != nil {
		return ""
	}
	return string(data)
}

// ReviewAndUpdateMemory scans recent history entries and appends new key facts
// to MEMORY.md. This is called by the heartbeat to keep long-term memory fresh.
func (m *Memory) ReviewAndUpdateMemory(workspace string) {
	history := m.LoadHistory()
	if history == "" {
		return
	}

	// Get the last N lines of history to review
	lines := strings.Split(strings.TrimSpace(history), "\n")
	reviewCount := 50
	if len(lines) < reviewCount {
		reviewCount = len(lines)
	}
	recentLines := lines[len(lines)-reviewCount:]

	// Extract key facts from recent history (simple keyword-based extraction)
	currentMemory := m.LoadMemory()
	var newFacts []string

	for _, line := range recentLines {
		lower := strings.ToLower(line)
		// Look for lines that contain important keywords
		if containsAny(lower, []string{
			"learned", "discovered", "user prefers", "important",
			"remember", "note:", "key fact", "decision",
			"preference", "configured", "set up",
		}) {
			// Extract the content after the timestamp
			if idx := strings.Index(line, "] "); idx >= 0 {
				fact := strings.TrimSpace(line[idx+2:])
				if fact != "" && !strings.Contains(currentMemory, fact) {
					newFacts = append(newFacts, fact)
				}
			}
		}
	}

	if len(newFacts) == 0 {
		slog.Debug("memory review: no new facts to add")
		return
	}

	// Append new facts to MEMORY.md
	var sb strings.Builder
	sb.WriteString(currentMemory)
	if currentMemory != "" && !strings.HasSuffix(currentMemory, "\n") {
		sb.WriteString("\n")
	}
	sb.WriteString(fmt.Sprintf("\n## Auto-updated: %s\n", time.Now().Format("2006-01-02 15:04")))
	for _, fact := range newFacts {
		sb.WriteString(fmt.Sprintf("- %s\n", fact))
	}

	if err := m.SaveMemory(sb.String()); err != nil {
		slog.Warn("failed to update memory", "error", err)
		return
	}

	slog.Info("memory updated", "new_facts", len(newFacts))
}

func containsAny(s string, keywords []string) bool {
	for _, kw := range keywords {
		if strings.Contains(s, kw) {
			return true
		}
	}
	return false
}

// SaveMemoryWithScan scans content for threats before writing to MEMORY.md.
// Logs warnings for any detected threats but still writes (to avoid data loss).
func (m *Memory) SaveMemoryWithScan(content string) error {
	if threats := privacy.Scan(content); len(threats) > 0 {
		for _, t := range threats {
			slog.Warn("memory safety threat detected in MEMORY.md write",
				"type", t.Type,
				"pattern", t.Pattern,
				"context", t.Context,
			)
		}
	}
	return m.SaveMemory(content)
}

// SaveUserFile writes USER.md with threat scanning.
func (m *Memory) SaveUserFile(content string) error {
	if threats := privacy.Scan(content); len(threats) > 0 {
		for _, t := range threats {
			slog.Warn("memory safety threat detected in USER.md write",
				"type", t.Type,
				"pattern", t.Pattern,
				"context", t.Context,
			)
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	os.MkdirAll(m.workspace, 0o755)
	return os.WriteFile(filepath.Join(m.workspace, "USER.md"), []byte(content), 0o644)
}

// LoadUserFile reads the USER.md file.
func (m *Memory) LoadUserFile() string {
	data, err := os.ReadFile(filepath.Join(m.workspace, "USER.md"))
	if err != nil {
		return ""
	}
	return string(data)
}

// AutoPersistMemory uses an LLM to extract facts from recent messages and
// append them to MEMORY.md and USER.md. Called every N turns.
func AutoPersistMemory(ctx context.Context, mem *Memory, prov provider.Provider, model string, messages []provider.Message) {
	// Build a summary of recent messages for the LLM
	var sb strings.Builder
	// Only look at last 20 messages to keep prompt small
	start := 0
	if len(messages) > 20 {
		start = len(messages) - 20
	}
	for _, m := range messages[start:] {
		if m.Role == "system" {
			continue
		}
		content := m.Content
		if len(content) > 300 {
			content = content[:300] + "..."
		}
		sb.WriteString(fmt.Sprintf("[%s]: %s\n", m.Role, content))
	}

	currentMemory := mem.LoadMemory()
	currentUser := mem.LoadUserFile()

	extractPrompt := fmt.Sprintf(`Analyze this conversation and extract:
1. Key facts, decisions, or learnings worth remembering (for MEMORY.md)
2. User preferences, profile details, or work style notes (for USER.md)

Current MEMORY.md:
%s

Current USER.md:
%s

Recent conversation:
%s

Output JSON only (no markdown fences):
{"memory_facts": ["fact1", "fact2"], "user_notes": ["note1"]}
If nothing worth saving, output: {"memory_facts": [], "user_notes": []}`,
		truncateStr(currentMemory, 500),
		truncateStr(currentUser, 500),
		sb.String(),
	)

	resp, err := prov.Chat(ctx, []provider.Message{
		{Role: "user", Content: extractPrompt},
	}, nil, model, 200, 0.3)
	if err != nil {
		slog.Debug("auto-persist: LLM call failed", "error", err)
		return
	}

	var result struct {
		MemoryFacts []string `json:"memory_facts"`
		UserNotes   []string `json:"user_notes"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(resp.Content)), &result); err != nil {
		slog.Debug("auto-persist: failed to parse LLM response", "error", err)
		return
	}

	// Atomically append new memory facts (read-modify-write under lock)
	if len(result.MemoryFacts) > 0 {
		mem.mu.Lock()
		latestMemory := mem.loadMemoryLocked()
		var memSB strings.Builder
		memSB.WriteString(latestMemory)
		if latestMemory != "" && !strings.HasSuffix(latestMemory, "\n") {
			memSB.WriteString("\n")
		}
		memSB.WriteString(fmt.Sprintf("\n## Auto-persisted: %s\n", time.Now().Format("2006-01-02 15:04")))
		for _, fact := range result.MemoryFacts {
			memSB.WriteString(fmt.Sprintf("- %s\n", fact))
		}
		newContent := memSB.String()
		mem.mu.Unlock()

		if err := mem.SaveMemoryWithScan(newContent); err != nil {
			slog.Warn("auto-persist: failed to save MEMORY.md", "error", err)
		} else {
			slog.Info("auto-persist: updated MEMORY.md", "facts", len(result.MemoryFacts))
		}

		// Dual-write to PostgreSQL if configured
		if mem.pgStore != nil {
			for _, fact := range result.MemoryFacts {
				if _, err := mem.pgStore.Insert(ctx, mem.agentID, "fact", fact, nil, nil); err != nil {
					slog.Warn("auto-persist: pg memory insert failed", "error", err)
				}
			}
		}
	}

	// Atomically append user notes
	if len(result.UserNotes) > 0 {
		mem.mu.Lock()
		latestUser := mem.loadUserFileLocked()
		var userSB strings.Builder
		userSB.WriteString(latestUser)
		if latestUser != "" && !strings.HasSuffix(latestUser, "\n") {
			userSB.WriteString("\n")
		}
		userSB.WriteString(fmt.Sprintf("\n## Auto-persisted: %s\n", time.Now().Format("2006-01-02 15:04")))
		for _, note := range result.UserNotes {
			userSB.WriteString(fmt.Sprintf("- %s\n", note))
		}
		newContent := userSB.String()
		mem.mu.Unlock()

		if err := mem.SaveUserFile(newContent); err != nil {
			slog.Warn("auto-persist: failed to save USER.md", "error", err)
		} else {
			slog.Info("auto-persist: updated USER.md", "notes", len(result.UserNotes))
		}

		// Dual-write user notes to PostgreSQL
		if mem.pgStore != nil {
			for _, note := range result.UserNotes {
				if _, err := mem.pgStore.Insert(ctx, mem.agentID, "user_note", note, nil, nil); err != nil {
					slog.Warn("auto-persist: pg user_note insert failed", "error", err)
				}
			}
		}
	}
}

// loadMemoryLocked reads MEMORY.md without acquiring the lock (caller must hold it).
func (m *Memory) loadMemoryLocked() string {
	data, err := os.ReadFile(m.memoryPath())
	if err != nil {
		return ""
	}
	return string(data)
}

// loadUserFileLocked reads USER.md without acquiring the lock (caller must hold it).
func (m *Memory) loadUserFileLocked() string {
	data, err := os.ReadFile(filepath.Join(m.workspace, "USER.md"))
	if err != nil {
		return ""
	}
	return string(data)
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
