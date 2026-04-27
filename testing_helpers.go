package filewatcher

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const testFilePermission = 0o600 // rw------- (owner read/write only)

func testEvent(path string, op Op) Event {
	return Event{Path: path, Op: op, Timestamp: time.Now(), IsDir: false}
}

func testWriteEvent(path string) Event {
	return testEvent(path, Write)
}

func fixedTimeEvent(path string, op Op, hour int) Event {
	return Event{
		Path:      path,
		Op:        op,
		Timestamp: time.Date(2025, 1, 1, hour, 0, 0, 0, time.UTC),
		IsDir:     false,
	}
}

const fixedEventTimestamp = 12

// FixedTime creates a time.Time with a fixed UTC timestamp.
// Use this for deterministic test timestamps.
func FixedTime(year int, month time.Month, day, hour, minute, second int) time.Time {
	return time.Date(year, month, day, hour, minute, second, 0, time.UTC)
}

func fixedWriteEvent(path string) Event {
	return fixedTimeEvent(path, Write, fixedEventTimestamp)
}

func assertCount(t *testing.T, count *atomic.Int32, want int32) {
	t.Helper()

	if got := count.Load(); got != want {
		t.Errorf("expected count %d, got %d", want, got)
	}
}

type pendingChecker interface {
	Pending() int
}

func assertPendingFunc(t *testing.T, p pendingChecker, want int) {
	t.Helper()

	if got := p.Pending(); got != want {
		t.Errorf("expected pending %d, got %d", want, got)
	}
}

func assertPending(t *testing.T, d *Debouncer, want int) {
	t.Helper()
	assertPendingFunc(t, d, want)
}

func assertGlobalPending(t *testing.T, d *GlobalDebouncer, want int) {
	t.Helper()
	assertPendingFunc(t, d, want)
}

func noopHandler() Handler {
	return func(_ context.Context, _ Event) error {
		return nil
	}
}

func testHandlerFunc(counter *int) Handler {
	return func(_ context.Context, _ Event) error {
		*counter++

		return nil
	}
}

func assertOpCount(t *testing.T, opCounts [4]int, op Op, want int) {
	t.Helper()

	if got := opCounts[op]; got != want {
		t.Errorf("expected %s count to be %d, got %d", op, want, got)
	}
}

var errTest = errors.New("test")

func testWatcherError(category ErrorCategory) *WatcherError {
	return &WatcherError{
		Op:       OpString("test"),
		Path:     "test-path",
		Err:      errTest,
		Category: category,
	}
}

func testError(err error, category ErrorCategory) *WatcherError {
	return &WatcherError{
		Op:       OpString("test"),
		Path:     "test-path",
		Err:      err,
		Category: category,
	}
}

func assertLogContains(t *testing.T, content string, substr LogSubstring) {
	t.Helper()

	if !strings.Contains(content, string(substr)) {
		t.Errorf("expected log to contain %q, got %q", substr, content)
	}
}

func setupTestContext(t *testing.T, timeout time.Duration) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), timeout)
	t.Cleanup(cancel)

	return ctx
}

func waitForEvent(t *testing.T, events <-chan Event, timeout time.Duration) *Event {
	t.Helper()

	select {
	case event := <-events:
		return &event
	case <-time.After(timeout):
		return nil
	}
}

func waitForEventOrFail(t *testing.T, events <-chan Event, timeout time.Duration) Event {
	t.Helper()

	event := waitForEvent(t, events, timeout)
	if event == nil {
		t.Fatal("timed out waiting for event")
	}

	return *event
}

// waitForEventOrTimeout waits for a single event from the channel.
// Returns true if an event was received, false if timeout occurred.
func waitForEventOrTimeout(t *testing.T, events <-chan Event, timeout time.Duration) bool {
	t.Helper()

	select {
	case <-events:
		return true
	case <-time.After(timeout):
		return false
	}
}

func receiveEventOrTimeout(t *testing.T, events <-chan Event, timeout time.Duration) {
	t.Helper()

	if !waitForEventOrTimeout(t, events, timeout) {
		t.Fatal("timed out waiting for event")
	}
}

func receiveEventMatchingOrTimeout(
	t *testing.T,
	events <-chan Event,
	timeout time.Duration,
	check func(Event),
	msg string,
) {
	t.Helper()

	select {
	case event := <-events:
		check(event)
	case <-time.After(timeout):
		t.Fatal(msg)
	}
}

// assertChannelClosed verifies that a receive-only channel is closed within the timeout.
func assertChannelClosed[T any](t *testing.T, ch <-chan T, timeout time.Duration, name string) {
	t.Helper()

	select {
	case _, ok := <-ch:
		if ok {
			t.Errorf("expected %s to be closed after Close()", name)
		}
	case <-time.After(timeout):
		t.Errorf("timed out waiting for %s to close", name)
	}
}

// assertEventPath verifies that an event has the expected path.
func assertEventPath(t *testing.T, event Event, expectedPath string) {
	t.Helper()

	if event.Path != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, event.Path)
	}
}

func debounceMulti(d *Debouncer, keys []DebounceKey, count *atomic.Int32) {
	for _, key := range keys {
		d.Debounce(key, func() { count.Add(1) })
	}
}

func debounceSingle(d *Debouncer, key DebounceKey, count *atomic.Int32) {
	debounceMulti(d, []DebounceKey{key}, count)
}

func debounceGlobalMulti(d *GlobalDebouncer, count *atomic.Int32, times int) {
	for range times {
		d.Debounce(DebounceKey(""), func() { count.Add(1) })
	}
}

func debounceNoCount(d *Debouncer, key DebounceKey) {
	d.Debounce(key, func() {})
}

func debounceMultiNoCount(d *Debouncer, keys []DebounceKey) {
	for _, key := range keys {
		debounceNoCount(d, key)
	}
}

func debounceGlobalNoCount(d *GlobalDebouncer) {
	d.Debounce(DebounceKey(""), func() {})
}

func createTestFile(t *testing.T, tmpDir TempDir, filename, content string) string {
	t.Helper()

	path := filepath.Join(string(tmpDir), filename)

	err := os.WriteFile(path, []byte(content), testFilePermission)
	if err != nil {
		t.Fatal(err)
	}

	return path
}

// waitForChannel waits for a value from a channel with timeout.
// Fatal if timeout occurs.
//
//nolint:ireturn // Generic channel helper must return generic type
func waitForChannel[T any](t *testing.T, ch <-chan T, timeout time.Duration, msg string) T {
	t.Helper()

	select {
	case v := <-ch:
		return v
	case <-time.After(timeout):
		t.Fatal(msg)
	}

	var zero T

	return zero
}

// waitForClose waits for a channel to be closed within the timeout.
// Returns true if closed, false if timeout occurred.
func waitForClose[T any](t *testing.T, channel <-chan T, timeout time.Duration) bool {
	t.Helper()

	var closed bool

	select {
	case _, ok := <-channel:
		closed = !ok
	case <-time.After(timeout):
		t.Fatal("timeout waiting for channel to close")
	}

	return closed
}

// newTestWatcher creates a new watcher with the given options.
// Fatal if creation fails.
func newTestWatcher(t *testing.T, tmpDir string, opts ...Option) *Watcher {
	t.Helper()

	w, err := New([]string{tmpDir}, opts...)
	if err != nil {
		t.Fatal(err)
	}

	return w
}
