package filewatcher

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

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

func countHandler(count *atomic.Int32) Handler {
	return func(_ context.Context, _ Event) error {
		count.Add(1)
		return nil
	}
}

func noopHandler() Handler {
	return func(_ context.Context, _ Event) error {
		return nil
	}
}

func assertLogContains(t *testing.T, content, substr string) {
	t.Helper()
	if !strings.Contains(content, substr) {
		t.Errorf("expected log to contain %q, got %q", substr, content)
	}
}

func testWriteEventGo(path string) Event {
	return Event{Path: path, Op: Write, Timestamp: time.Now(), IsDir: false}
}

func testContextTimeout(t *testing.T, d time.Duration) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(t.Context(), d)
}

func setupTestContext(t *testing.T, timeout time.Duration) context.Context {
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

func waitForEventOrFailMsg(t *testing.T, events <-chan Event, timeout time.Duration, msg string) Event {
	t.Helper()
	event := waitForEvent(t, events, timeout)
	if event == nil {
		t.Fatal(msg)
	}
	return *event
}

func waitForEvents(t *testing.T, events <-chan Event, count int, timeout time.Duration) []Event {
	t.Helper()
	var result []Event
	deadline := time.After(timeout)
	for {
		select {
		case event := <-events:
			result = append(result, event)
			if len(result) >= count {
				return result
			}
		case <-deadline:
			return result
		}
	}
}

func newWatcherWithTimeout(t *testing.T, tmpDir string, timeout time.Duration) (*Watcher, <-chan Event) {
	t.Helper()
	w, err := New([]string{tmpDir})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)

	events, err := w.Watch(ctx)
	if err != nil {
		_ = w.Close()
		t.Fatalf("Watch failed: %v", err)
	}
	t.Cleanup(func() { _ = w.Close() })
	return w, events
}

func newWatcherWithTimeoutDefault(t *testing.T, tmpDir string) (*Watcher, <-chan Event) {
	t.Helper()
	return newWatcherWithTimeout(t, tmpDir, 5*time.Second)
}

func setupDebouncer(count *atomic.Int32, delay time.Duration) *Debouncer {
	d := NewDebouncer(delay)
	return d
}

func debounceAndCount(d *Debouncer, key string, count *atomic.Int32) {
	d.Debounce(key, func() { count.Add(1) })
}

func debounceMulti(d *Debouncer, keys []string, count *atomic.Int32) {
	for _, key := range keys {
		debounceAndCount(d, key, count)
	}
}

func debounceGlobalMulti(d *GlobalDebouncer, count *atomic.Int32, times int) {
	for range times {
		d.Debounce("", func() { count.Add(1) })
	}
}

func debounceGlobalSingle(d *GlobalDebouncer, count *atomic.Int32) {
	d.Debounce("", func() { count.Add(1) })
}

func debounceSingle(d *Debouncer, key string, count *atomic.Int32) {
	d.Debounce(key, func() { count.Add(1) })
}

func debounceGlobalNoCount(d *GlobalDebouncer) {
	d.Debounce("", func() {})
}

func debounceNoCount(d *Debouncer, key string) {
	d.Debounce(key, func() {})
}

func debounceMultiNoCount(d *Debouncer, keys []string) {
	for _, key := range keys {
		debounceNoCount(d, key)
	}
}

func benchmarkMiddlewareEvent() (context.Context, Event) {
	return context.Background(), Event{Op: Write, Path: "/tmp/test.go", Timestamp: time.Now(), IsDir: false}
}

func createTestFile(t *testing.T, tmpDir, filename, content string) string {
	t.Helper()
	path := filepath.Join(tmpDir, filename)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func createTestFileDefault(t *testing.T, tmpDir string) string {
	t.Helper()
	return createTestFile(t, tmpDir, "test.go", "package test")
}

func assertTimeout(t *testing.T, events <-chan Event, timeout time.Duration) {
	t.Helper()
	select {
	case <-events:
	case <-time.After(timeout):
		t.Fatal("timed out waiting for event")
	}
}

func assertTimeoutMsg(t *testing.T, events <-chan Event, timeout time.Duration, msg string) {
	t.Helper()
	select {
	case <-events:
	case <-time.After(timeout):
		t.Fatal(msg)
	}
}
