package setup

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/fastclaw-ai/fastclaw/internal/config"
)

// --- Channels ---
//
// Channel CRUD endpoints. Because Telegram/Discord/Slack channels are
// long-poll / WebSocket clients booted at gateway start, any change to
// the account roster (new bot, deleted bot, swapped token) requires a
// daemon restart. The handlers therefore *only* persist to disk and
// flag `needsRestart: true` in the response — the UI is expected to
// follow up with POST /api/daemon/restart and then poll /api/status.

// channelAccountResponse describes a single account inside a channel,
// enriched with the bound agent ID derived from cfg.Bindings.
//
// BotToken is always masked on the wire. The UI re-submits the masked
// placeholder when the user didn't change it; mergeSecret() then keeps
// the on-disk value intact.
type channelAccountResponse struct {
	ID       string `json:"id"`
	BotToken string `json:"botToken"`
	AgentID  string `json:"agentId,omitempty"`
}

type channelDetailResponse struct {
	Type     string                   `json:"type"`
	Enabled  bool                     `json:"enabled"`
	BotToken string                   `json:"botToken,omitempty"` // masked; only meaningful for legacy/single-bot setups
	AppToken string                   `json:"appToken,omitempty"` // masked; slack-only
	Accounts []channelAccountResponse `json:"accounts"`
	Status   string                   `json:"status,omitempty"`
}

func (s *Server) handleListChannels(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.Load()
	if err != nil {
		jsonResponse(w, http.StatusOK, []any{})
		return
	}

	// Stable order for deterministic UI rendering.
	types := make([]string, 0, len(cfg.Channels))
	for t := range cfg.Channels {
		types = append(types, t)
	}
	sort.Strings(types)

	channels := make([]channelDetailResponse, 0, len(types))
	for _, t := range types {
		channels = append(channels, buildChannelDetail(t, cfg.Channels[t], cfg.Bindings))
	}
	jsonResponse(w, http.StatusOK, channels)
}

// channelUpdateRequest is the wire format for PUT /api/channels/{type}.
// Tokens may arrive masked ("****") to indicate "keep existing".
type channelUpdateRequest struct {
	Enabled  bool                  `json:"enabled"`
	BotToken string                `json:"botToken"`
	AppToken string                `json:"appToken"`
	Accounts []channelAccountInput `json:"accounts"`
}

type channelAccountInput struct {
	ID       string `json:"id"`
	BotToken string `json:"botToken"`
	AgentID  string `json:"agentId"`
}

func (s *Server) handleUpsertChannel(w http.ResponseWriter, r *http.Request) {
	chType := r.PathValue("type")
	if chType == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "channel type required"})
		return
	}

	var req channelUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid request"})
		return
	}

	cfg, err := config.Load()
	if err != nil {
		// First-time creation: nothing on disk yet.
		cfg = &config.Config{Channels: map[string]config.ChannelConfig{}}
	}
	if cfg.Channels == nil {
		cfg.Channels = map[string]config.ChannelConfig{}
	}

	old := cfg.Channels[chType]

	// Validate accounts: must have non-empty ID and (after merge) a real token.
	mergedAccounts := map[string]config.AccountConfig{}
	seenIDs := map[string]bool{}
	for _, a := range req.Accounts {
		id := strings.TrimSpace(a.ID)
		if id == "" {
			jsonResponse(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "account id is required"})
			return
		}
		if seenIDs[id] {
			jsonResponse(w, http.StatusBadRequest, map[string]any{"ok": false, "error": fmt.Sprintf("duplicate account id %q", id)})
			return
		}
		seenIDs[id] = true

		oldAcct := old.Accounts[id]
		token := mergeSecret(a.BotToken, oldAcct.BotToken)
		if token == "" {
			jsonResponse(w, http.StatusBadRequest, map[string]any{"ok": false, "error": fmt.Sprintf("bot token is required for account %q", id)})
			return
		}
		mergedAccounts[id] = config.AccountConfig{BotToken: token}
	}

	// Channel-level botToken is only meaningful for backwards-compat /
	// legacy single-bot setups. New configs always use the accounts map.
	mergedBotToken := mergeSecret(req.BotToken, old.BotToken)
	mergedAppToken := mergeSecret(req.AppToken, old.AppToken)

	// If there's at least one account, the channel-level token is redundant —
	// drop it so we don't carry two sources of truth.
	if len(mergedAccounts) > 0 {
		mergedBotToken = ""
	}

	cfg.Channels[chType] = config.ChannelConfig{
		Enabled:  req.Enabled,
		BotToken: mergedBotToken,
		AppToken: mergedAppToken,
		Accounts: mergedAccounts,
	}

	// Sync bindings: drop any account-level binding for this channel
	// (peer-specific bindings are preserved — those are managed elsewhere),
	// then add new ones from the account.agentId fields.
	pruned := pruneChannelBindings(cfg.Bindings, chType)
	for _, a := range req.Accounts {
		if a.AgentID == "" || a.ID == "" {
			continue
		}
		pruned = append(pruned, config.Binding{
			AgentID: a.AgentID,
			Match: config.Match{
				Channel:   chType,
				AccountID: a.ID,
			},
		})
	}
	cfg.Bindings = pruned

	if err := saveConfigFile(cfg); err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}

	slog.Info("channel saved", "type", chType, "accounts", len(mergedAccounts), "enabled", req.Enabled)
	jsonResponse(w, http.StatusOK, map[string]any{"ok": true, "needsRestart": true})
}

func (s *Server) handleDeleteChannel(w http.ResponseWriter, r *http.Request) {
	chType := r.PathValue("type")
	if chType == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "channel type required"})
		return
	}

	cfg, err := config.Load()
	if err != nil {
		jsonResponse(w, http.StatusOK, map[string]any{"ok": true})
		return
	}
	if cfg.Channels != nil {
		delete(cfg.Channels, chType)
	}
	cfg.Bindings = pruneChannelBindings(cfg.Bindings, chType)

	if err := saveConfigFile(cfg); err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	slog.Info("channel deleted", "type", chType)
	jsonResponse(w, http.StatusOK, map[string]any{"ok": true, "needsRestart": true})
}

// channelTestRequest probes a bot token's validity without persisting it.
// Currently telegram-only (calls https://api.telegram.org/bot<token>/getMe).
type channelTestRequest struct {
	Type     string `json:"type"`
	BotToken string `json:"botToken"`
}

func (s *Server) handleTestChannel(w http.ResponseWriter, r *http.Request) {
	var req channelTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid request"})
		return
	}
	switch req.Type {
	case "telegram":
		// fall through
	default:
		jsonResponse(w, http.StatusOK, map[string]any{
			"ok":    false,
			"error": fmt.Sprintf("connection test not yet implemented for %q", req.Type),
		})
		return
	}
	if req.BotToken == "" || strings.Contains(req.BotToken, "****") {
		jsonResponse(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "bot token required (paste the real token, not the masked value)"})
		return
	}

	url := "https://api.telegram.org/bot" + req.BotToken + "/getMe"
	httpReq, err := http.NewRequestWithContext(r.Context(), http.MethodGet, url, nil)
	if err != nil {
		jsonResponse(w, http.StatusOK, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		jsonResponse(w, http.StatusOK, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode != http.StatusOK {
		jsonResponse(w, http.StatusOK, map[string]any{"ok": false, "error": fmt.Sprintf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))})
		return
	}
	var probe struct {
		Ok     bool `json:"ok"`
		Result struct {
			ID        int64  `json:"id"`
			IsBot     bool   `json:"is_bot"`
			FirstName string `json:"first_name"`
			Username  string `json:"username"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &probe); err != nil || !probe.Ok {
		jsonResponse(w, http.StatusOK, map[string]any{"ok": false, "error": "unexpected Telegram response"})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]any{
		"ok":          true,
		"botUsername": probe.Result.Username,
		"firstName":   probe.Result.FirstName,
	})
}

// --- helpers ---

func buildChannelDetail(chType string, ch config.ChannelConfig, bindings []config.Binding) channelDetailResponse {
	resp := channelDetailResponse{
		Type:     chType,
		Enabled:  ch.Enabled,
		AppToken: maskToken(ch.AppToken),
	}
	if ch.Enabled {
		resp.Status = "connected"
	} else {
		resp.Status = "disconnected"
	}

	// Convert legacy "channel-level botToken without accounts" into a
	// synthetic "default" account so the UI doesn't have to handle two
	// different shapes. The next save will persist it as a real account.
	if len(ch.Accounts) == 0 && ch.BotToken != "" {
		resp.Accounts = []channelAccountResponse{{
			ID:       "default",
			BotToken: maskToken(ch.BotToken),
			AgentID:  agentForAccount(bindings, chType, ""),
		}}
		return resp
	}

	// Sort accounts by ID for stable rendering.
	ids := make([]string, 0, len(ch.Accounts))
	for id := range ch.Accounts {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	resp.Accounts = make([]channelAccountResponse, 0, len(ids))
	for _, id := range ids {
		acct := ch.Accounts[id]
		token := acct.BotToken
		if token == "" {
			token = ch.BotToken
		}
		resp.Accounts = append(resp.Accounts, channelAccountResponse{
			ID:       id,
			BotToken: maskToken(token),
			AgentID:  agentForAccount(bindings, chType, id),
		})
	}
	return resp
}

// agentForAccount returns the agentId of the binding that targets this
// (channel, accountID) pair without a peer filter — i.e. the "default
// agent for this bot" binding the channels UI manages.
func agentForAccount(bindings []config.Binding, channel, accountID string) string {
	for _, b := range bindings {
		if b.Match.Channel != channel {
			continue
		}
		if b.Match.AccountID != accountID {
			continue
		}
		if b.Match.Peer != nil && b.Match.Peer.ID != "" {
			continue
		}
		return b.AgentID
	}
	return ""
}

// pruneChannelBindings drops account-level bindings (no peer.id) for the
// given channel. Peer-specific bindings (per-chat overrides, used by
// cron tasks etc.) are preserved.
func pruneChannelBindings(bindings []config.Binding, channel string) []config.Binding {
	out := make([]config.Binding, 0, len(bindings))
	for _, b := range bindings {
		if b.Match.Channel != channel {
			out = append(out, b)
			continue
		}
		if b.Match.Peer != nil && b.Match.Peer.ID != "" {
			out = append(out, b)
			continue
		}
		// drop
	}
	return out
}

// maskToken keeps the first/last few characters and obscures the middle.
// Returns the empty string for empty input so the field can be omitted
// from the JSON response.
func maskToken(token string) string {
	if token == "" {
		return ""
	}
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}

// mergeSecret returns the new value when it's a real secret (non-empty
// and not the masked placeholder), otherwise it falls back to the old
// value. This lets the UI round-trip the masked token without leaking
// it to the client and still preserve the underlying secret.
func mergeSecret(incoming, existing string) string {
	if strings.Contains(incoming, "****") {
		return existing
	}
	if incoming == "" {
		return existing
	}
	return incoming
}
