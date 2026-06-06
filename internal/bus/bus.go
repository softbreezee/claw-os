package bus

// Origin identifies what produced an InboundMessage. The default
// (empty string) is "real user input from a channel/web", which keeps
// every existing producer correct without touching it. Runtime-
// originated messages (cron / webhook / agent-to-agent / goal
// continuations / mid-run user steering) carry a non-empty origin so
// downstream policy can distinguish them.
//
// The reply path uses this to decide between sending out via channel
// (real chat) and writing a Notification record (cron / webhook /
// internal). The agent loop and goal/steer hooks branch on it too.
//
// IMPORTANT: OriginUser MUST stay "". Many producers (web POST, IM
// adapters, plugins) leave Origin unset; if OriginUser ever became a
// non-empty string the zero-value InboundMessage would silently
// reclassify as a non-user message and break routing.
const (
	OriginUser        = "" // default — real user input from a channel/web
	OriginChannel     = "" // alias of OriginUser, for callers that want to spell it out
	OriginCron        = "cron"
	OriginWebhook     = "webhook"
	OriginInternal    = "internal" // agent-to-agent, gateway-self injected
	OriginHeartbeat   = "heartbeat"
	OriginSubAgent    = "subagent"
	OriginGoalContext = "goal_context" // goal-runtime continuation prompt
	OriginUserSteer   = "user_steer"   // mid-run user steering message
)

// IsRuntimeInjected reports whether a message was produced by the
// runtime rather than a real channel user. Used by reply routing to
// pick "Notification record" vs. "send out via channel", and by the
// agent loop to gate behaviors like compaction-on-user-turn.
//
// OriginHeartbeat / OriginGoalContext / OriginUserSteer / OriginSubAgent
// are currently treated as "non-internal" because they ARE conceptually
// part of the user's conversation (or driven by it) — only cron /
// webhook / internal are out-of-band relative to the chat. Keep this
// list in sync with isInternalOrigin checks elsewhere; centralizing
// here is the whole point.
func IsRuntimeInjected(origin string) bool {
	switch origin {
	case OriginCron, OriginWebhook, OriginInternal:
		return true
	default:
		return false
	}
}

// Attachment represents a file attached to an inbound message — image,
// (in future) PDF, audio, etc.
//
// Path is a pawnix-local filesystem path produced by internal/upload.
// We deliberately do NOT carry inline bytes through the bus: large
// payloads pollute logs, blow up channel buffers, and make persistence
// in the session store wasteful. The agent loop reads bytes on demand
// when building the LLM request.
type Attachment struct {
	Path     string // absolute local path, e.g. ~/.pawnix/uploads/<sha>.png
	MimeType string // e.g. "image/png", "application/pdf"
	Name     string // original filename for display purposes
	Size     int64  // bytes (informational; the file on disk is the source of truth)
}

// IsImage reports whether the attachment is an image type the LLM
// vision pipeline can consume directly.
func (a Attachment) IsImage() bool {
	return len(a.MimeType) >= 6 && a.MimeType[:6] == "image/"
}

// InboundMessage represents a message received from a channel.
type InboundMessage struct {
	Channel      string       // channel type, e.g. "telegram"
	AccountID    string       // account within the channel (e.g. which bot)
	ChatID       string       // unique chat identifier within the channel
	UserID       string       // user identifier
	MessageID    string       // unique message identifier within the chat
	Text         string       // message text
	PeerKind     string       // "group" or "dm"
	SenderName   string       // display name of the sender
	Mentions     []string     // @usernames mentioned in the message
	IsBotMessage bool         // true if the message was sent by a bot
	PhotoURL     string       // legacy: single attached photo URL (Telegram). Prefer Attachments for new code.
	Attachments  []Attachment // multimodal attachments (images today; pdf/audio later)
	ReplyToMsgID string       // message ID being replied to

	// AgentID, when non-empty, bypasses cfg.Bindings-based agent
	// matching and routes the message directly to the named agent.
	// Used by the cron scheduler (jobs carry their target agent in
	// the store) and the webhook system. For real channel messages
	// this stays empty and routing falls back to binding match.
	AgentID string

	// Origin tags the source of the inbound message so downstream
	// routing can apply per-origin policy. See the Origin* constants
	// above for the full enum. The reply path uses this to decide
	// whether to send via outbound channel (real chat) vs. write a
	// Notification record (cron / webhook / internal); v0.3 Goal +
	// Steering features tag continuation / steer messages here too.
	//
	// Empty string == OriginUser (real user input). Keep it that way
	// so producers that don't set Origin stay correct.
	Origin string
}

// OutboundButton represents a button in an inline keyboard.
type OutboundButton struct {
	Text         string
	CallbackData string
	URL          string
}

// OutboundMessage represents a message to be sent to a channel.
type OutboundMessage struct {
	Channel      string              // target channel type
	AccountID    string              // target account within the channel
	ChatID       string              // target chat identifier
	Text         string              // message text
	ReplyToMsgID string              // reply to specific message
	ParseMode    string              // "MarkdownV2", "HTML", ""
	Buttons      [][]OutboundButton  // inline keyboard rows
	EditMsgID    string              // edit existing message instead of sending new
	MediaPaths   []string            // file paths to attach (from MEDIA: protocol)
}

// MessageBus is an async message queue backed by Go channels.
type MessageBus struct {
	Inbound  chan InboundMessage
	Outbound chan OutboundMessage
}

// New creates a new MessageBus with buffered channels.
func New() *MessageBus {
	return &MessageBus{
		Inbound:  make(chan InboundMessage, 100),
		Outbound: make(chan OutboundMessage, 100),
	}
}
