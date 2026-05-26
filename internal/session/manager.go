package session

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/softbreezee/claw-os/internal/provider"
)

// RemoteStore is an optional backend that persists sessions in addition to
// (or instead of) JSONL files.  Implement this interface to plug in PostgreSQL
// or any other store.
type RemoteStore interface {
	Load(ctx context.Context, agentID, channel, sessionID string) ([]provider.Message, error)
	Save(ctx context.Context, agentID, channel, sessionID string, msgs []provider.Message) error
	Delete(ctx context.Context, agentID, channel, sessionID string) error
	ListWebSessions(ctx context.Context, agentID string) ([]map[string]string, error)
	ListExternalSessions(ctx context.Context, agentID string) ([]map[string]string, error)
}

// Session holds the message history for a channel:chat_id pair.
type Session struct {
	mu               sync.Mutex
	Messages         []provider.Message
	LastConsolidated int // index of last consolidated message
	filePath         string
	snapshot         []provider.Message // undo snapshot
	// remote backend fields
	remote    RemoteStore
	agentID   string
	channel   string
	sessionID string
	// ephemeral sessions live entirely in-memory and never touch
	// disk or the remote store. Used for stateless "internal-origin"
	// agent invocations (cron fires, webhook deliveries) where the
	// trigger should NOT pollute the user's chat history. See
	// NewEphemeralSession().
	ephemeral bool
}

// Manager manages sessions, keyed by "channel:chat_id".
type Manager struct {
	mu       sync.Mutex
	sessions map[string]*Session
	dataDir  string
	agentID  string
	remote   RemoteStore // optional PG/remote backend
}

// NewManager creates a new session manager backed by JSONL files.
func NewManager(dataDir string) *Manager {
	return &Manager{
		sessions: make(map[string]*Session),
		dataDir:  dataDir,
	}
}

// NewManagerWithRemote creates a session manager that uses a remote store as
// the primary backend (JSONL files are kept as a local cache/fallback).
func NewManagerWithRemote(dataDir, agentID string, remote RemoteStore) *Manager {
	return &Manager{
		sessions: make(map[string]*Session),
		dataDir:  dataDir,
		agentID:  agentID,
		remote:   remote,
	}
}

func sessionKey(channel, chatID string) string {
	return channel + ":" + chatID
}

// NewEphemeralSession returns an in-memory-only session that never
// writes to the file system or remote store. Use this for stateless
// agent invocations like cron-fire / webhook-delivery where you want
// the agent to process a message WITHOUT it being recorded in the
// user's web chat history.
//
// The returned Session is not registered with any Manager — it's
// caller-owned and garbage-collected when references drop. Append /
// GetMessages / ReplaceMessages all work but are no-ops on disk.
func NewEphemeralSession() *Session {
	return &Session{ephemeral: true}
}

// Get returns or creates a session for the given channel and chat ID.
// When a RemoteStore is configured, messages are loaded from PG on first access.
func (m *Manager) Get(channel, chatID string) *Session {
	key := sessionKey(channel, chatID)

	m.mu.Lock()
	defer m.mu.Unlock()

	if s, ok := m.sessions[key]; ok {
		return s
	}

	safeKey := strings.ReplaceAll(key, ":", "_")
	filePath := filepath.Join(m.dataDir, safeKey+".jsonl")

	s := &Session{
		filePath: filePath,
		remote:   m.remote,
		agentID:  m.agentID,
		channel:  channel,
		sessionID: chatID,
	}

	if m.remote != nil {
		// Load from remote store; fall back to local file if remote fails.
		msgs, err := m.remote.Load(context.Background(), m.agentID, channel, chatID)
		if err == nil && len(msgs) > 0 {
			s.Messages = msgs
		} else {
			s.load()
		}
	} else {
		s.load()
	}

	m.sessions[key] = s
	return s
}

// Append adds a message to the session and persists it to all configured backends.
func (s *Session) Append(msg provider.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Messages = append(s.Messages, msg)
	if s.ephemeral {
		return // in-memory only — skip file + remote
	}
	s.appendToFile(msg)

	if s.remote != nil {
		go func(msgs []provider.Message) {
			_ = s.remote.Save(context.Background(), s.agentID, s.channel, s.sessionID, msgs)
		}(append([]provider.Message(nil), s.Messages...))
	}
}

// GetMessages returns a copy of all messages.
func (s *Session) GetMessages() []provider.Message {
	s.mu.Lock()
	defer s.mu.Unlock()

	msgs := make([]provider.Message, len(s.Messages))
	copy(msgs, s.Messages)
	return msgs
}

// UnconsolidatedCount returns the number of messages since last consolidation.
func (s *Session) UnconsolidatedCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.Messages) - s.LastConsolidated
}

// MarkConsolidated updates the consolidation pointer.
func (s *Session) MarkConsolidated(index int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastConsolidated = index
}

// ReplaceMessages replaces all session messages with the given list.
// This is used after context compaction to trim the session.
func (s *Session) ReplaceMessages(msgs []provider.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Messages = make([]provider.Message, len(msgs))
	copy(s.Messages, msgs)
	s.LastConsolidated = 0

	if s.ephemeral {
		return // in-memory only — skip file + remote
	}
	s.rewriteFile()

	if s.remote != nil {
		snapshot := make([]provider.Message, len(msgs))
		copy(snapshot, msgs)
		go func() {
			_ = s.remote.Save(context.Background(), s.agentID, s.channel, s.sessionID, snapshot)
		}()
	}
}

// Clear resets the session messages.
func (s *Session) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Messages = nil
	s.LastConsolidated = 0
	// Truncate the file
	os.Remove(s.filePath)
}

func (s *Session) load() {
	f, err := os.Open(s.filePath)
	if err != nil {
		return // file doesn't exist yet
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		var msg provider.Message
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			continue
		}
		s.Messages = append(s.Messages, msg)
	}
}

func (s *Session) rewriteFile() {
	dir := filepath.Dir(s.filePath)
	os.MkdirAll(dir, 0o755)

	f, err := os.Create(s.filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "session rewrite error: %v\n", err)
		return
	}
	defer f.Close()

	for _, msg := range s.Messages {
		data, err := json.Marshal(msg)
		if err != nil {
			continue
		}
		f.Write(data)
		f.Write([]byte("\n"))
	}
}

func (s *Session) appendToFile(msg provider.Message) {
	dir := filepath.Dir(s.filePath)
	os.MkdirAll(dir, 0o755)

	f, err := os.OpenFile(s.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "session persist error: %v\n", err)
		return
	}
	defer f.Close()

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	f.Write(data)
	f.Write([]byte("\n"))
}

// ListWebSessions returns web chat sessions. Uses remote store when available.
func (m *Manager) ListExternalSessions() []map[string]string {
	if m.remote != nil {
		sessions, err := m.remote.ListExternalSessions(context.Background(), m.agentID)
		if err == nil {
			return sessions
		}
	}
	return m.listExternalSessionsFromFiles()
}

func (m *Manager) listExternalSessionsFromFiles() []map[string]string {
	var sessions []map[string]string
	// Read from the sessions directory for non-web files
	entries, err := os.ReadDir(m.dataDir)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".jsonl") || strings.HasPrefix(name, "web_") {
			continue
		}
		// Only user-facing channels
		if !strings.HasPrefix(name, "discord_") && !strings.HasPrefix(name, "telegram_") {
			continue
		}
		// Parse "discord_1504039153787076611.jsonl" → channel=discord, chatId=...
		base := strings.TrimSuffix(name, ".jsonl")
		parts := strings.SplitN(base, "_", 2)
		if len(parts) < 2 {
			continue
		}
		ch := parts[0]
		chatID := parts[1]
		// Quick preview: first user message
		preview := ""
		f, err := os.Open(filepath.Join(m.dataDir, name))
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			var msg struct { Role, Content string }
			if err := json.Unmarshal(scanner.Bytes(), &msg); err == nil && msg.Role == "user" && msg.Content != "" {
				preview = msg.Content
				break
			}
		}
		f.Close()
		if preview == "" {
			continue
		}
		if len(preview) > 80 {
			preview = preview[:80] + "..."
		}
		sessions = append(sessions, map[string]string{
			"id":      ch + ":" + chatID,
			"channel": ch,
			"chatId":  chatID,
			"preview": preview,
		})
	}
	return sessions
}

func (m *Manager) ListWebSessions() []map[string]string {
	if m.remote != nil {
		sessions, err := m.remote.ListWebSessions(context.Background(), m.agentID)
		if err == nil {
			return sessions
		}
	}
	return m.listWebSessionsFromFiles()
}

func (m *Manager) listWebSessionsFromFiles() []map[string]string {
	pattern := filepath.Join(m.dataDir, "web_*.jsonl")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil
	}

	var sessions []map[string]string
	for _, f := range files {
		base := filepath.Base(f)
		// "web_<sessionId>.jsonl" -> "<sessionId>"
		sessionId := strings.TrimPrefix(base, "web_")
		sessionId = strings.TrimSuffix(sessionId, ".jsonl")

		// Read first user message as preview
		preview := ""
		fh, err := os.Open(f)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(fh)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			var msg struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			}
			if json.Unmarshal(scanner.Bytes(), &msg) == nil && msg.Role == "user" && msg.Content != "" {
				preview = msg.Content
				if len(preview) > 50 {
					preview = preview[:50] + "..."
				}
				break
			}
		}
		fh.Close()

		if preview == "" {
			continue // skip empty sessions
		}

		sessions = append(sessions, map[string]string{
			"id":      sessionId,
			"preview": preview,
		})
	}
	return sessions
}

// DeleteWebSession deletes a web chat session by session ID from all backends.
func (m *Manager) DeleteWebSession(sessionId string) error {
	key := "web:" + sessionId
	m.mu.Lock()
	delete(m.sessions, key)
	m.mu.Unlock()

	// Remove from disk
	safeKey := strings.ReplaceAll(key, ":", "_")
	filePath := filepath.Join(m.dataDir, safeKey+".jsonl")
	_ = os.Remove(filePath)

	// Remove from remote store
	if m.remote != nil {
		return m.remote.Delete(context.Background(), m.agentID, "web", sessionId)
	}
	return nil
}

// Snapshot saves the current message list as a restore point (for undo).
func (s *Session) Snapshot() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshot = make([]provider.Message, len(s.Messages))
	copy(s.snapshot, s.Messages)
}

// Undo restores the last snapshot. Returns false if no snapshot exists.
func (s *Session) Undo() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.snapshot == nil {
		return false
	}
	s.Messages = make([]provider.Message, len(s.snapshot))
	copy(s.Messages, s.snapshot)
	s.snapshot = nil
	s.LastConsolidated = 0
	s.rewriteFile()
	return true
}

// HasSnapshot returns true if an undo snapshot exists.
func (s *Session) HasSnapshot() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.snapshot != nil
}
