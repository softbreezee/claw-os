package pg

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/fastclaw-ai/fastclaw/internal/provider"
)

// SessionStore provides PostgreSQL-backed session persistence.
type SessionStore struct {
	db *DB
}

// NewSessionStore creates a SessionStore backed by the given DB.
func NewSessionStore(db *DB) *SessionStore {
	return &SessionStore{db: db}
}

// Load reads messages for a session. Returns nil slice if not found.
func (s *SessionStore) Load(ctx context.Context, agentID, channel, sessionID string) ([]provider.Message, error) {
	var raw []byte
	err := s.db.Pool.QueryRow(ctx,
		`SELECT messages FROM sessions
		 WHERE agent_id = $1 AND channel = $2 AND session_id = $3`,
		agentID, channel, sessionID,
	).Scan(&raw)
	if err != nil {
		// Row not found — return empty, not an error.
		return nil, nil
	}

	var msgs []provider.Message
	if err := json.Unmarshal(raw, &msgs); err != nil {
		return nil, fmt.Errorf("pg: session unmarshal: %w", err)
	}
	return msgs, nil
}

// Save upserts the full message list for a session.
func (s *SessionStore) Save(ctx context.Context, agentID, channel, sessionID string, msgs []provider.Message) error {
	data, err := json.Marshal(msgs)
	if err != nil {
		return fmt.Errorf("pg: session marshal: %w", err)
	}

	_, err = s.db.Pool.Exec(ctx,
		`INSERT INTO sessions (agent_id, channel, session_id, messages, updated_at)
		 VALUES ($1, $2, $3, $4, now())
		 ON CONFLICT (agent_id, channel, session_id)
		 DO UPDATE SET messages = EXCLUDED.messages, updated_at = now()`,
		agentID, channel, sessionID, data,
	)
	if err != nil {
		return fmt.Errorf("pg: session save: %w", err)
	}
	return nil
}

// Delete removes a session.
func (s *SessionStore) Delete(ctx context.Context, agentID, channel, sessionID string) error {
	_, err := s.db.Pool.Exec(ctx,
		`DELETE FROM sessions WHERE agent_id=$1 AND channel=$2 AND session_id=$3`,
		agentID, channel, sessionID,
	)
	return err
}

// ListWebSessions returns all web sessions for an agent with a preview of the
// first user message.
func (s *SessionStore) ListWebSessions(ctx context.Context, agentID string) ([]map[string]string, error) {
	rows, err := s.db.Pool.Query(ctx,
		`SELECT session_id, messages FROM sessions
		 WHERE agent_id = $1 AND channel = 'web'
		 ORDER BY updated_at DESC`,
		agentID,
	)
	if err != nil {
		return nil, fmt.Errorf("pg: list web sessions: %w", err)
	}
	defer rows.Close()

	var result []map[string]string
	for rows.Next() {
		var sid string
		var raw []byte
		if err := rows.Scan(&sid, &raw); err != nil {
			continue
		}

		var msgs []provider.Message
		if err := json.Unmarshal(raw, &msgs); err != nil {
			continue
		}

		preview := ""
		for _, m := range msgs {
			if m.Role == "user" && m.Content != "" {
				preview = m.Content
				if len(preview) > 50 {
					preview = preview[:50] + "..."
				}
				break
			}
		}
		if preview == "" {
			continue
		}
		result = append(result, map[string]string{"id": sid, "preview": preview})
	}
	return result, nil
}

// LogUnused suppresses the unused import warning during development.
var _ = slog.Info
