package filewatcher

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

// --- Add() routes through the fake backend ---

func TestFakeBackend_AddRoutesToBackend(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	fake := newFakeBackend()

	watcher := newTestWatcherWithFake(t, tmpDir, fake)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	events, err := watcher.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Create a real subdirectory so Add's directory walk has something to find.
	subDir := filepath.Join(tmpDir, "subpkg")
	if mkErr := os.MkdirAll(subDir, 0o750); mkErr != nil { //nolint:gosec // standard temp directory permissions
		t.Fatalf("mkdir: %v", mkErr)
	}

	addErr := watcher.Add(subDir)
	if addErr != nil {
		t.Fatalf("Add failed: %v", addErr)
	}

	// The fake backend should have received an Add call for the subdirectory.
	waitForCondition(t, time.Second, "fake backend should record Add for subDir", func() bool {
		fake.mu.Lock()
		defer fake.mu.Unlock()

		for _, p := range fake.addedPaths {
			if p == subDir {
				return true
			}
		}

		return false
	})

	// The path should appear in the watcher's watch list.
	if !pathInWatchList(t, watcher, subDir) {
		t.Errorf("subDir %q not in WatchList after Add", subDir)
	}

	cancelAndDrain(cancel, events)
}

// --- Remove() routes through the fake backend ---

func TestFakeBackend_RemoveRoutesToBackend(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	fake := newFakeBackend()

	watcher := newTestWatcherWithFake(t, tmpDir, fake)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	events, err := watcher.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	removePath := watcher.WatchList()
	if len(removePath) == 0 {
		t.Fatal("expected at least one path in WatchList")
	}

	target := removePath[0]

	removeErr := watcher.Remove(target)
	if removeErr != nil {
		t.Fatalf("Remove failed: %v", removeErr)
	}

	// The fake backend should have received a Remove call.
	waitForCondition(t, time.Second, "fake backend should record Remove", func() bool {
		fake.mu.Lock()
		defer fake.mu.Unlock()

		for _, p := range fake.removedPaths {
			if p == target {
				return true
			}
		}

		return false
	})

	// The path should no longer be in the watcher's watch list.
	if pathInWatchList(t, watcher, target) {
		t.Errorf("path %q still in WatchList after Remove", target)
	}

	cancelAndDrain(cancel, events)
}

// --- Reset() clears runtime state and allows restart ---

func TestFakeBackend_ResetClearsStateAndRestarts(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	fake := newFakeBackend()

	watcher := newTestWatcherWithFake(t, tmpDir, fake)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	events, err := watcher.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Inject some events to populate stats counters.
	fake.sendEvent(fsnotify.Event{
		Name: filepath.Join(tmpDir, "a.go"),
		Op:   fsnotify.Create,
	})

	waitForCondition(t, time.Second, "event processed", func() bool {
		return watcher.Stats().EventsProcessed > 0
	})

	cancelAndDrain(cancel, events)

	// Close the watcher — Reset requires a closed state.
	if closeErr := watcher.Close(); closeErr != nil {
		t.Fatalf("Close failed: %v", closeErr)
	}

	// Reset clears runtime state while preserving configuration.
	if resetErr := watcher.Reset(); resetErr != nil {
		t.Fatalf("Reset failed: %v", resetErr)
	}

	stats := watcher.Stats()
	if stats.EventsProcessed != 0 {
		t.Errorf("EventsProcessed = %d after Reset, want 0", stats.EventsProcessed)
	}

	if stats.IsClosed {
		t.Error("IsClosed = true after Reset, want false")
	}

	if stats.IsWatching {
		t.Error("IsWatching = true after Reset, want false")
	}

	if len(watcher.WatchList()) != 0 {
		t.Errorf("WatchList len = %d after Reset, want 0", len(watcher.WatchList()))
	}

	// After Reset, the watcher should be restartable. Reset replaces the
	// fakeBackend with a real fsnotify watcher, so we use the real backend.
	ctx2, cancel2 := context.WithCancel(t.Context())
	defer cancel2()

	events2, watchErr := watcher.Watch(ctx2)
	if watchErr != nil {
		t.Fatalf("Watch after Reset failed: %v", watchErr)
	}

	if !watcher.Stats().IsWatching {
		t.Error("IsWatching = false after Watch post-Reset")
	}

	cancelAndDrain(cancel2, events2)
}

// --- Circuit breaker integrated into the pipeline (closed state) ---
// The circuit breaker's opening behavior is tested at the middleware level
// (see TestCircuitBreaker_DropsEventsAfterFailures). This test verifies the
// integration point: events flow through the pipeline when the breaker is
// closed and the breaker does not interfere with event delivery.

func TestFakeBackend_CircuitBreakerInPipelinePassesEvents(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	fake := newFakeBackend()

	var delivered atomic.Int32

	watcher := newTestWatcherWithFake(
		t, tmpDir, fake,
		WithMiddleware(
			MiddlewareRecovery(),
			MiddlewareCircuitBreaker(100, time.Second),
		),
	)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	events, err := watcher.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Send 5 events; all should pass through the closed circuit breaker.
	for i := range 5 {
		fake.sendEvent(fsnotify.Event{
			Name: filepath.Join(tmpDir, "file"+string(rune('A'+i))+".go"),
			Op:   fsnotify.Create,
		})
	}

	for range 5 {
		select {
		case <-events:
			delivered.Add(1)
		case <-time.After(time.Second):
			t.Fatalf("timeout: only received %d/5 events", delivered.Load())
		}
	}

	cancelAndDrain(cancel, events)
}

// --- Concurrent event burst: goroutine leak detection ---

func TestFakeBackend_ConcurrentBurstNoGoroutineLeak(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	fake := newFakeBackend()

	watcher := newTestWatcherWithFake(t, tmpDir, fake)

	// Snapshot goroutine count before the burst (after Watch starts its
	// goroutines so they are part of the baseline).
	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	events, err := watcher.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	const numSenders = 8
	const eventsPerSender = 200

	var sent atomic.Int32

	// Launch concurrent senders that flood the fake backend.
	for s := range numSenders {
		go func(senderID int) {
			for i := range eventsPerSender {
				fake.sendEvent(fsnotify.Event{
					Name: filepath.Join(tmpDir,
						"s"+itoa(senderID)+"_f"+itoa(i)+".go"),
					Op: fsnotify.Create,
				})
				sent.Add(1)
			}
		}(s)
	}

	// Drain events until all are received or timeout.
	target := int32(numSenders * eventsPerSender)
	deadline := time.After(5 * time.Second)

	for received := 0; received < int(target); {
		select {
		case <-events:
			received++
		case <-deadline:
			t.Fatalf("timeout: received %d/%d events", received, target)
		}
	}

	cancelAndDrain(cancel, events)

	// Wait for goroutines to settle after shutdown.
	settled := waitForGoroutineSettle(t, before, 2*time.Second)

	after := runtime.NumGoroutine()
	// Allow a small delta for GC/runtime goroutines that may not have settled.
	if !settled && after > before+5 {
		t.Errorf("goroutine leak: before=%d after=%d (delta=%d)",
			before, after, after-before)
	}
}

// --- helpers ---

// pathInWatchList returns true if the given path is in the watcher's WatchList.
func pathInWatchList(t *testing.T, w *Watcher, path string) bool {
	t.Helper()

	for _, p := range w.WatchList() {
		if p == path {
			return true
		}
	}

	return false
}

// waitForGoroutineSettle polls NumGoroutine until it returns to at most
// before+2 or the timeout elapses. Returns true if settled, false on timeout.
func waitForGoroutineSettle(t *testing.T, before int, timeout time.Duration) bool {
	t.Helper()

	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if runtime.NumGoroutine() <= before+2 {
			return true
		}

		runtime.Gosched()
		time.Sleep(20 * time.Millisecond)
	}

	return false
}

// itoa converts a non-negative int to its decimal string representation without
// importing strconv (keeps the test file import list lean).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	var buf [20]byte
	pos := len(buf)

	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}

	return string(buf[pos:])
}
