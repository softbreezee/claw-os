package agent

import (
	"context"
	"sync"
	"testing"
)

// TestSteerBuffer_DrainAtomicallyClears pins the contract the agent
// loop relies on: Drain returns everything queued so far AND empties
// the buffer in one critical section, so a concurrent Push during
// the drain ends up in the next iteration's batch (where it belongs)
// rather than half-getting consumed in the current iteration's
// "<steer>...</steer>" injection.
func TestSteerBuffer_DrainAtomicallyClears(t *testing.T) {
	b := NewSteerBuffer()
	b.Push("focus on energy stocks")
	b.Push("ignore the previous tool error")

	got := b.Drain()
	if len(got) != 2 {
		t.Fatalf("Drain returned %d messages, want 2", len(got))
	}
	if got[0].Text != "focus on energy stocks" {
		t.Errorf("FIFO order broken: got[0] = %q", got[0].Text)
	}

	if b.Pending() != 0 {
		t.Errorf("buffer not cleared after Drain, Pending = %d", b.Pending())
	}
	if again := b.Drain(); again != nil {
		t.Errorf("second Drain on empty buffer should return nil, got %v", again)
	}
}

// TestSteerBuffer_NilSafe locks down the "steering disabled for this
// run" path: agent loop must be able to Push/Drain a nil buffer
// without panic, since CLI / cron / subagent turns don't have one.
func TestSteerBuffer_NilSafe(t *testing.T) {
	var b *SteerBuffer
	b.Push("anything") // must not panic
	if got := b.Drain(); got != nil {
		t.Errorf("nil buffer Drain should return nil, got %v", got)
	}
	if got := b.Pending(); got != 0 {
		t.Errorf("nil buffer Pending = %d, want 0", got)
	}
}

// TestSteerBuffer_ConcurrentPushDrain catches the obvious data race
// risk — race detector will scream if the locking is wrong.
func TestSteerBuffer_ConcurrentPushDrain(t *testing.T) {
	b := NewSteerBuffer()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			b.Push("steer")
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			b.Drain()
		}
	}()
	wg.Wait()
	// Drain whatever's left so we don't leak buffer contents into
	// the next test run.
	b.Drain()
}

// TestContextWithSteerBuffer pins the round-trip: attaching a buffer
// is observable; nil attach is a no-op; absent buffer returns nil
// without panic.
func TestContextWithSteerBuffer(t *testing.T) {
	ctx := context.Background()
	if got := SteerBufferFromContext(ctx); got != nil {
		t.Errorf("bare ctx should have no buffer, got %v", got)
	}

	if got := ContextWithSteerBuffer(ctx, nil); got != ctx {
		t.Errorf("nil buf attach should return original ctx unchanged")
	}

	buf := NewSteerBuffer()
	ctx2 := ContextWithSteerBuffer(ctx, buf)
	if got := SteerBufferFromContext(ctx2); got != buf {
		t.Errorf("round-trip failed: got %v, want %v", got, buf)
	}
}
