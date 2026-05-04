// Package eventbus provides a topic-based publish/subscribe primitive for
// decoupling task producers (TaskRunner) from event consumers (HTTP SSE
// handlers, future webhook fan-out, etc).
//
// Why a separate package instead of reusing internal/bus?
//   internal/bus is a pair of global channels modelling IM message routing
//   (inbound/outbound). It has no concept of topic and no fan-out. Forcing
//   task events through it would corrupt that semantics. A clean split keeps
//   each abstraction aligned with its actual use case.
//
// Design choices:
//   * Topic is a free-form string. Convention: "task:{taskID}".
//   * One Publish fans out to every Subscribe call active at that moment.
//   * Each subscription has a buffered channel (default 32). If a slow
//     subscriber's buffer fills up, that subscriber drops the event – the
//     publisher is never blocked. Slow consumers fall behind silently rather
//     than stalling the entire bus.
//   * Subscribe returns an unsubscribe func. Callers (HTTP handlers) MUST
//     defer it; otherwise goroutines leak.
//   * Closing a topic is implicit: when the last subscriber unsubscribes,
//     the topic entry is removed.
package eventbus

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Event is a single message broadcast on a topic.
//
// Seq is a monotonically-increasing per-topic sequence number assigned
// by the producer (see TaskRunner). It enables resumable subscribers –
// a reconnecting client passes its last-seen Seq via ?after=N and the
// handler replays buffered events with greater Seq before subscribing
// to live updates.
type Event struct {
	Seq       int64          `json:"seq,omitempty"`
	Type      string         `json:"type"`
	Data      map[string]any `json:"data,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

// Bus is the abstract interface; we currently only ship an in-memory
// implementation. Splitting via interface lets PR3 swap in a Redis-backed
// version (or add event replay) without touching call sites.
type Bus interface {
	// Publish broadcasts evt to every current subscriber of topic.
	// Never blocks. Returns the number of slow subscribers that dropped
	// the event (for observability; safe to ignore).
	Publish(ctx context.Context, topic string, evt Event) int

	// Subscribe returns a channel that will receive events published to
	// topic, plus a cancel func that removes the subscription. Callers
	// must always call the cancel func to avoid goroutine leaks.
	Subscribe(topic string) (<-chan Event, func())

	// Close releases all resources. Concurrent Publish/Subscribe calls
	// after Close return zero-value channels.
	Close() error
}

// MemoryBus is the default in-process implementation backed by a
// map[topic][]*subscriber. Safe for concurrent use.
type MemoryBus struct {
	mu     sync.RWMutex
	topics map[string]map[*subscriber]struct{}
	closed bool
}

type subscriber struct {
	ch chan Event
}

// NewMemoryBus constructs an empty MemoryBus.
func NewMemoryBus() *MemoryBus {
	return &MemoryBus{
		topics: make(map[string]map[*subscriber]struct{}),
	}
}

// defaultBufferSize controls how many events a subscriber may fall behind
// before events get dropped. 32 is large enough for typical SSE jitter
// (network buffering, client GC pause) but small enough to apply backpressure
// quickly so memory doesn't balloon when a client disappears.
const defaultBufferSize = 32

// Publish implements Bus.Publish. It snapshots subscribers under the read
// lock and then sends without holding the lock to avoid latency spikes.
func (b *MemoryBus) Publish(_ context.Context, topic string, evt Event) int {
	b.mu.RLock()
	if b.closed {
		b.mu.RUnlock()
		return 0
	}
	subs := b.topics[topic]
	// Snapshot to allow Subscribe/Unsubscribe to proceed during fan-out.
	snapshot := make([]*subscriber, 0, len(subs))
	for s := range subs {
		snapshot = append(snapshot, s)
	}
	b.mu.RUnlock()

	dropped := 0
	for _, s := range snapshot {
		select {
		case s.ch <- evt:
		default:
			// Subscriber's buffer is full – count it as a drop and move on.
			// We deliberately do NOT block; one slow consumer must not
			// stall the producer or other consumers.
			dropped++
		}
	}
	if dropped > 0 {
		slog.Warn("eventbus: dropped events for slow subscribers",
			"topic", topic, "dropped", dropped, "subscribers", len(snapshot))
	}
	return dropped
}

// Subscribe implements Bus.Subscribe. The unsubscribe func is idempotent:
// calling it multiple times is a no-op after the first call.
func (b *MemoryBus) Subscribe(topic string) (<-chan Event, func()) {
	sub := &subscriber{ch: make(chan Event, defaultBufferSize)}

	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		closed := make(chan Event)
		close(closed)
		return closed, func() {}
	}
	subs, ok := b.topics[topic]
	if !ok {
		subs = make(map[*subscriber]struct{})
		b.topics[topic] = subs
	}
	subs[sub] = struct{}{}
	b.mu.Unlock()

	var once sync.Once
	cancel := func() {
		once.Do(func() {
			b.mu.Lock()
			// Only close if WE still own the channel. If Close() ran first,
			// it already removed the topic and closed the channel; closing
			// again would panic.
			stillOwned := false
			if subs, ok := b.topics[topic]; ok {
				if _, present := subs[sub]; present {
					stillOwned = true
					delete(subs, sub)
					if len(subs) == 0 {
						// Reclaim the topic entry; otherwise long-lived
						// buses accumulate empty maps per completed task.
						delete(b.topics, topic)
					}
				}
			}
			b.mu.Unlock()
			if stillOwned {
				close(sub.ch)
			}
		})
	}
	return sub.ch, cancel
}

// Close releases all resources and terminates pending subscribers.
func (b *MemoryBus) Close() error {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil
	}
	b.closed = true
	for topic, subs := range b.topics {
		for s := range subs {
			close(s.ch)
		}
		delete(b.topics, topic)
	}
	b.mu.Unlock()
	return nil
}

// SubscriberCount returns the number of active subscribers on a topic.
// Useful for tests and admin endpoints.
func (b *MemoryBus) SubscriberCount(topic string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.topics[topic])
}

// Compile-time interface check.
var _ Bus = (*MemoryBus)(nil)
