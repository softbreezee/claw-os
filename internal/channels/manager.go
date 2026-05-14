package channels

import (
	"context"
	"log/slog"
	"strings"
	"sync"

	"github.com/softbreezee/claw-os/internal/bus"
)

// Manager manages all channel instances and routes outbound messages.
type Manager struct {
	channels map[string]Channel // key: "channel:accountID"
	bus      *bus.MessageBus
}

// NewManager creates a new channel manager.
func NewManager(mb *bus.MessageBus) *Manager {
	return &Manager{
		channels: make(map[string]Channel),
		bus:      mb,
	}
}

// Register adds a channel to the manager keyed by channel:accountID.
func (m *Manager) Register(ch Channel) {
	key := channelKey(ch.Name(), ch.AccountID())
	m.channels[key] = ch
}

// Start launches all channels and the outbound message router.
func (m *Manager) Start(ctx context.Context) {
	var wg sync.WaitGroup

	// Start outbound router
	wg.Add(1)
	go func() {
		defer wg.Done()
		m.routeOutbound(ctx)
	}()

	// Start each channel
	for key, ch := range m.channels {
		wg.Add(1)
		go func(k string, c Channel) {
			defer wg.Done()
			slog.Info("starting channel", "key", k)
			if err := c.Start(ctx); err != nil {
				slog.Error("channel stopped with error", "key", k, "error", err)
			}
		}(key, ch)
	}

	wg.Wait()
}

func (m *Manager) routeOutbound(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-m.bus.Outbound:
			key := channelKey(msg.Channel, msg.AccountID)
			ch, ok := m.channels[key]
			if !ok {
				// Fallback: when caller specified the channel type
				// but left AccountID empty AND only one account of
				// that type is registered, use it automatically.
				// This handles the common single-bot case where a
				// cron job created from the Web UI says
				// "push to telegram, chat 12345" without knowing
				// the bot's account name. Without this the message
				// is silently dropped.
				if msg.AccountID == "" {
					if fallback := m.onlyChannelOfType(msg.Channel); fallback != nil {
						slog.Info("outbound: empty accountID matched single registered account, auto-resolving",
							"channel", msg.Channel,
							"account", fallback.AccountID(),
						)
						ch = fallback
						ok = true
					}
				}
			}
			if !ok {
				slog.Warn("unknown outbound channel; check that the channel is enabled and the account ID matches",
					"channel", msg.Channel,
					"account", msg.AccountID,
					"chat_id", msg.ChatID,
					"key", key,
					"available_keys", m.channelKeys(),
				)
				continue
			}
			if err := ch.SendMessage(msg); err != nil {
				slog.Error("send message failed", "key", key, "error", err)
			}
		}
	}
}

// onlyChannelOfType returns the sole registered channel matching
// the given type, or nil if there are zero or multiple. Used by the
// outbound fallback so callers that don't know the AccountID
// (typically cross-channel cron jobs) still hit the right bot when
// the user only has one configured.
func (m *Manager) onlyChannelOfType(channelType string) Channel {
	prefix := channelType + ":"
	var found Channel
	count := 0
	for k, c := range m.channels {
		if strings.HasPrefix(k, prefix) {
			found = c
			count++
		}
	}
	if count == 1 {
		return found
	}
	return nil
}

// channelKeys returns the registered channel keys for diagnostic
// logging. Slow path — only called on lookup failure.
func (m *Manager) channelKeys() []string {
	keys := make([]string, 0, len(m.channels))
	for k := range m.channels {
		keys = append(keys, k)
	}
	return keys
}

// BotUsername returns the bot username for a given channel:accountID pair.
func (m *Manager) BotUsername(channel, accountID string) string {
	key := channelKey(channel, accountID)
	ch, ok := m.channels[key]
	if !ok {
		return ""
	}
	return ch.BotUsername()
}

// SendTyping sends a typing indicator for the given channel and chat.
func (m *Manager) SendTyping(channel, accountID, chatID string) {
	key := channelKey(channel, accountID)
	ch, ok := m.channels[key]
	if !ok {
		return
	}
	if err := ch.SendTyping(chatID); err != nil {
		slog.Debug("send typing failed", "key", key, "error", err)
	}
}

func channelKey(channel, accountID string) string {
	if accountID == "" {
		return channel + ":"
	}
	return channel + ":" + accountID
}
