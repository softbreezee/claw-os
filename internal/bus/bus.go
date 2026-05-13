package bus

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
