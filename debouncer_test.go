package filewatcher

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestDebouncer_Debounce(t *testing.T) {
	t.Parallel()

	var count atomic.Int32
	d := NewDebouncer(50 * time.Millisecond)

	d.Debounce("key1", func() { count.Add(1) })
	d.Debounce("key1", func() { count.Add(1) })
	d.Debounce("key1", func() { count.Add(1) })

	time.Sleep(100 * time.Millisecond)

	if got := count.Load(); got != 1 {
		t.Errorf("expected 1 execution after debouncing 3 calls, got %d", got)
	}
}

func TestDebouncer_DifferentKeys(t *testing.T) {
	t.Parallel()

	var count atomic.Int32
	d := NewDebouncer(50 * time.Millisecond)

	d.Debounce("key1", func() { count.Add(1) })
	d.Debounce("key2", func() { count.Add(1) })

	time.Sleep(100 * time.Millisecond)

	if got := count.Load(); got != 2 {
		t.Errorf("expected 2 executions for different keys, got %d", got)
	}
}

func TestDebouncer_Flush(t *testing.T) {
	t.Parallel()

	var count atomic.Int32
	d := NewDebouncer(200 * time.Millisecond)

	d.Debounce("key1", func() { count.Add(1) })
	d.Flush()

	time.Sleep(50 * time.Millisecond)

	// Flush executes pending functions immediately, so we expect 1 execution
	if got := count.Load(); got != 1 {
		t.Errorf("expected 1 execution after flush (executes pending), got %d", got)
	}

	if got := d.Pending(); got != 0 {
		t.Errorf("expected 0 pending after flush, got %d", got)
	}
}

func TestDebouncer_Stop(t *testing.T) {
	t.Parallel()

	var count atomic.Int32
	d := NewDebouncer(50 * time.Millisecond)

	d.Debounce("key1", func() { count.Add(1) })
	d.Stop()

	time.Sleep(100 * time.Millisecond)

	if got := count.Load(); got != 0 {
		t.Errorf("expected 0 executions after stop, got %d", got)
	}
}

func TestDebouncer_Pending(t *testing.T) {
	t.Parallel()

	d := NewDebouncer(200 * time.Millisecond)

	d.Debounce("key1", func() {})
	d.Debounce("key2", func() {})
	d.Debounce("key3", func() {})

	if got := d.Pending(); got != 3 {
		t.Errorf("expected 3 pending, got %d", got)
	}

	d.Stop()

	if got := d.Pending(); got != 0 {
		t.Errorf("expected 0 pending after stop, got %d", got)
	}
}

func TestDebouncer_DefaultDelay(t *testing.T) {
	t.Parallel()

	d := NewDebouncer(0)

	var count atomic.Int32
	d.Debounce("key", func() { count.Add(1) })

	time.Sleep(600 * time.Millisecond)

	if got := count.Load(); got != 1 {
		t.Errorf("expected 1 execution with default delay, got %d", got)
	}
}

func TestDebouncer_NegativeDelay(t *testing.T) {
	t.Parallel()

	d := NewDebouncer(-1 * time.Second)

	var count atomic.Int32
	d.Debounce("key", func() { count.Add(1) })

	time.Sleep(600 * time.Millisecond)

	if got := count.Load(); got != 1 {
		t.Errorf("expected 1 execution with negative delay (should default to 500ms), got %d", got)
	}
}

func TestDebouncer_RapidCalls(t *testing.T) {
	t.Parallel()

	var count atomic.Int32
	d := NewDebouncer(30 * time.Millisecond)

	for range 100 {
		d.Debounce("key1", func() { count.Add(1) })
	}

	time.Sleep(100 * time.Millisecond)

	if got := count.Load(); got != 1 {
		t.Errorf("expected 1 execution after 100 rapid calls, got %d", got)
	}
}

func TestGlobalDebouncer_Flush(t *testing.T) {
	t.Parallel()

	var count atomic.Int32
	d := NewGlobalDebouncer(200 * time.Millisecond)

	d.Debounce("", func() { count.Add(1) })

	if got := d.Pending(); got != 1 {
		t.Errorf("expected 1 pending before flush, got %d", got)
	}

	d.Flush()

	time.Sleep(50 * time.Millisecond)

	if got := count.Load(); got != 1 {
		t.Errorf("expected 1 execution after flush, got %d", got)
	}

	if got := d.Pending(); got != 0 {
		t.Errorf("expected 0 pending after flush, got %d", got)
	}
}

func TestGlobalDebouncer_Stop(t *testing.T) {
	t.Parallel()

	var count atomic.Int32
	d := NewGlobalDebouncer(50 * time.Millisecond)

	d.Debounce("", func() { count.Add(1) })
	d.Stop()

	time.Sleep(100 * time.Millisecond)

	if got := count.Load(); got != 0 {
		t.Errorf("expected 0 executions after stop, got %d", got)
	}
}

func TestGlobalDebouncer_Debounce(t *testing.T) {
	t.Parallel()

	var count atomic.Int32
	d := NewGlobalDebouncer(50 * time.Millisecond)

	d.Debounce("", func() { count.Add(1) })
	d.Debounce("", func() { count.Add(1) })
	d.Debounce("", func() { count.Add(1) })

	time.Sleep(100 * time.Millisecond)

	if got := count.Load(); got != 1 {
		t.Errorf("expected 1 execution after debouncing 3 global calls, got %d", got)
	}
}

func TestGlobalDebouncer_DefaultDelay(t *testing.T) {
	t.Parallel()

	d := NewGlobalDebouncer(0)

	var count atomic.Int32
	d.Debounce("", func() { count.Add(1) })

	time.Sleep(600 * time.Millisecond)

	if got := count.Load(); got != 1 {
		t.Errorf("expected 1 execution with default delay, got %d", got)
	}
}

func TestGlobalDebouncer_Pending(t *testing.T) {
	t.Parallel()

	d := NewGlobalDebouncer(200 * time.Millisecond)

	if got := d.Pending(); got != 0 {
		t.Errorf("expected 0 pending initially, got %d", got)
	}

	d.Debounce("", func() {})

	if got := d.Pending(); got != 1 {
		t.Errorf("expected 1 pending after debounce, got %d", got)
	}

	d.Stop()

	if got := d.Pending(); got != 0 {
		t.Errorf("expected 0 pending after stop, got %d", got)
	}
}

func BenchmarkDebouncer_Debounce(b *testing.B) {
	d := NewDebouncer(1 * time.Second)
	defer d.Stop()

	b.ResetTimer()
	for i := range b.N {
		d.Debounce("key", func() {})
		// Use index to avoid compiler optimization
		_ = i
	}
}

func BenchmarkDebouncer_DifferentKeys(b *testing.B) {
	d := NewDebouncer(1 * time.Second)
	defer d.Stop()

	b.ResetTimer()
	for i := range b.N {
		d.Debounce(fmt.Sprintf("key-%d", i%100), func() {})
	}
}

func BenchmarkGlobalDebouncer_Debounce(b *testing.B) {
	d := NewGlobalDebouncer(1 * time.Second)
	defer d.Stop()

	b.ResetTimer()
	for i := range b.N {
		d.Debounce("", func() {})
		_ = i
	}
}
