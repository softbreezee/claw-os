package eventbus

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestPublishToSingleSubscriber(t *testing.T) {
	b := NewMemoryBus()
	defer b.Close()

	ch, cancel := b.Subscribe("task:1")
	defer cancel()

	go b.Publish(context.Background(), "task:1", Event{Type: "running"})

	select {
	case evt := <-ch:
		if evt.Type != "running" {
			t.Fatalf("got %q, want %q", evt.Type, "running")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for event")
	}
}

func TestPublishFanOutToMultipleSubscribers(t *testing.T) {
	b := NewMemoryBus()
	defer b.Close()

	const n = 5
	chans := make([]<-chan Event, n)
	cancels := make([]func(), n)
	for i := 0; i < n; i++ {
		chans[i], cancels[i] = b.Subscribe("task:1")
		defer cancels[i]()
	}

	dropped := b.Publish(context.Background(), "task:1", Event{Type: "x"})
	if dropped != 0 {
		t.Fatalf("unexpected drops: %d", dropped)
	}

	for i, ch := range chans {
		select {
		case evt := <-ch:
			if evt.Type != "x" {
				t.Fatalf("subscriber %d: got %q", i, evt.Type)
			}
		case <-time.After(time.Second):
			t.Fatalf("subscriber %d: timed out", i)
		}
	}
}

func TestPublishToOtherTopicIsIgnored(t *testing.T) {
	b := NewMemoryBus()
	defer b.Close()

	ch, cancel := b.Subscribe("task:1")
	defer cancel()

	b.Publish(context.Background(), "task:2", Event{Type: "wrong"})

	select {
	case evt := <-ch:
		t.Fatalf("unexpected event: %+v", evt)
	case <-time.After(50 * time.Millisecond):
		// expected: no event
	}
}

func TestUnsubscribeStopsDelivery(t *testing.T) {
	b := NewMemoryBus()
	defer b.Close()

	ch, cancel := b.Subscribe("task:1")
	cancel()

	// After cancel, the channel is closed; Publish should not panic.
	b.Publish(context.Background(), "task:1", Event{Type: "after-cancel"})

	// Closed channel reads return zero value immediately.
	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("got event after unsubscribe")
		}
	case <-time.After(time.Second):
		t.Fatal("channel should be closed")
	}

	if got := b.SubscriberCount("task:1"); got != 0 {
		t.Fatalf("SubscriberCount=%d, want 0", got)
	}
}

func TestUnsubscribeIsIdempotent(t *testing.T) {
	b := NewMemoryBus()
	defer b.Close()

	_, cancel := b.Subscribe("task:1")
	cancel()
	cancel() // must not panic or double-close
}

func TestSlowSubscriberDropsInsteadOfBlocking(t *testing.T) {
	b := NewMemoryBus()
	defer b.Close()

	_, cancel := b.Subscribe("task:1")
	defer cancel()

	// Publish far more than the buffer (32) so drops are guaranteed.
	const overrun = 100
	totalDropped := 0
	for i := 0; i < overrun; i++ {
		totalDropped += b.Publish(context.Background(), "task:1", Event{Type: "x"})
	}

	if totalDropped == 0 {
		t.Fatal("expected at least one drop, got none")
	}
	// The bus must still be usable after drops.
	if got := b.SubscriberCount("task:1"); got != 1 {
		t.Fatalf("SubscriberCount=%d, want 1", got)
	}
}

func TestCloseTerminatesPendingSubscribers(t *testing.T) {
	b := NewMemoryBus()

	ch, cancel := b.Subscribe("task:1")
	defer cancel()

	b.Close()

	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("expected closed channel, got value")
		}
	case <-time.After(time.Second):
		t.Fatal("subscriber channel was not closed by Close()")
	}

	// Publish after Close must be a no-op.
	if got := b.Publish(context.Background(), "task:1", Event{}); got != 0 {
		t.Fatalf("Publish after Close should report 0 drops, got %d", got)
	}
}

func TestConcurrentPublishSubscribe(t *testing.T) {
	b := NewMemoryBus()
	defer b.Close()

	var wg sync.WaitGroup
	const subs = 10
	const events = 50

	received := make([]int, subs)
	for i := 0; i < subs; i++ {
		ch, cancel := b.Subscribe("task:hot")
		defer cancel()
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for range ch {
				received[idx]++
			}
		}(i)
		_ = ch
	}

	for i := 0; i < events; i++ {
		b.Publish(context.Background(), "task:hot", Event{Type: "tick"})
	}

	// Allow the bus to drain.
	time.Sleep(100 * time.Millisecond)
	b.Close()
	wg.Wait()

	// Every subscriber should have received some events (exact count varies
	// because of buffer sizing; the only invariant is "at least one").
	for i, n := range received {
		if n == 0 {
			t.Errorf("subscriber %d received 0 events", i)
		}
	}
}
