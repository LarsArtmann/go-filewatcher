//nolint:testpackage,varnamelen // Tests need internal access; idiomatic short names in tests
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

	debounceMulti(d, []DebounceKey{
		NewDebounceKey("key1"),
		NewDebounceKey("key1"),
		NewDebounceKey("key1"),
	}, &count)

	time.Sleep(100 * time.Millisecond)

	assertCount(t, &count, 1)
}

func TestDebouncer_DifferentKeys(t *testing.T) {
	t.Parallel()

	var count atomic.Int32

	d := NewDebouncer(50 * time.Millisecond)

	debounceMulti(d, []DebounceKey{NewDebounceKey("key1"), NewDebounceKey("key2")}, &count)

	time.Sleep(100 * time.Millisecond)

	assertCount(t, &count, 2)
}

func TestDebouncer_Flush(t *testing.T) {
	t.Parallel()

	var count atomic.Int32

	d := NewDebouncer(200 * time.Millisecond)

	debounceSingle(d, NewDebounceKey("key1"), &count)
	d.Flush()

	time.Sleep(50 * time.Millisecond)

	assertCount(t, &count, 1)
	assertPending(t, d, 0)
}

func TestDebouncer_Stop(t *testing.T) {
	t.Parallel()

	var count atomic.Int32

	d := NewDebouncer(50 * time.Millisecond)

	debounceSingle(d, NewDebounceKey("key1"), &count)
	d.Stop()

	time.Sleep(100 * time.Millisecond)

	assertCount(t, &count, 0)
}

func TestDebouncer_Pending(t *testing.T) {
	t.Parallel()

	d := NewDebouncer(200 * time.Millisecond)

	debounceMultiNoCount(d, []DebounceKey{
		NewDebounceKey("key1"),
		NewDebounceKey("key2"),
		NewDebounceKey("key3"),
	})

	assertPending(t, d, 3)

	d.Stop()

	assertPending(t, d, 0)
}

func TestDebouncer_DefaultDelay(t *testing.T) {
	t.Parallel()

	d := NewDebouncer(0)

	var count atomic.Int32
	debounceSingle(d, NewDebounceKey("key"), &count)

	time.Sleep(600 * time.Millisecond)

	assertCount(t, &count, 1)
}

func TestDebouncer_NegativeDelay(t *testing.T) {
	t.Parallel()

	d := NewDebouncer(-1 * time.Second)

	var count atomic.Int32
	debounceSingle(d, NewDebounceKey("key"), &count)

	time.Sleep(600 * time.Millisecond)

	assertCount(t, &count, 1)
}

func TestDebouncer_RapidCalls(t *testing.T) {
	t.Parallel()

	var count atomic.Int32

	d := NewDebouncer(30 * time.Millisecond)

	debounceSingle(d, NewDebounceKey("key1"), &count)

	time.Sleep(100 * time.Millisecond)

	assertCount(t, &count, 1)
}

func TestGlobalDebouncer_Flush(t *testing.T) {
	t.Parallel()

	var count atomic.Int32

	d := NewGlobalDebouncer(200 * time.Millisecond)

	debounceGlobalMulti(d, &count, 1)

	assertGlobalPending(t, d, 1)

	d.Flush()

	time.Sleep(50 * time.Millisecond)

	assertCount(t, &count, 1)
	assertGlobalPending(t, d, 0)
}

func TestGlobalDebouncer_Stop(t *testing.T) {
	t.Parallel()

	var count atomic.Int32

	d := NewGlobalDebouncer(50 * time.Millisecond)

	debounceGlobalMulti(d, &count, 1)
	d.Stop()

	time.Sleep(100 * time.Millisecond)

	assertCount(t, &count, 0)
}

func TestGlobalDebouncer_Debounce(t *testing.T) {
	t.Parallel()

	var count atomic.Int32

	d := NewGlobalDebouncer(50 * time.Millisecond)

	debounceGlobalMulti(d, &count, 3)

	time.Sleep(100 * time.Millisecond)

	assertCount(t, &count, 1)
}

func TestGlobalDebouncer_DefaultDelay(t *testing.T) {
	t.Parallel()

	d := NewGlobalDebouncer(0)

	var count atomic.Int32
	debounceGlobalMulti(d, &count, 1)

	time.Sleep(600 * time.Millisecond)

	assertCount(t, &count, 1)
}

func TestGlobalDebouncer_Pending(t *testing.T) {
	t.Parallel()

	d := NewGlobalDebouncer(200 * time.Millisecond)

	assertGlobalPending(t, d, 0)

	debounceGlobalNoCount(d)

	assertGlobalPending(t, d, 1)

	d.Stop()

	assertGlobalPending(t, d, 0)
}

func BenchmarkDebouncer_Debounce(b *testing.B) {
	d := NewDebouncer(1 * time.Second)
	defer d.Stop()

	runDebouncerBenchmark(b, d, NewDebounceKey("key"))
}

func BenchmarkDebouncer_DifferentKeys(b *testing.B) {
	d := NewDebouncer(1 * time.Second)
	defer d.Stop()

	b.ResetTimer()

	for i := range b.N {
		d.Debounce(NewDebounceKey(fmt.Sprintf("key-%d", i%100)), func() {})
	}
}

func BenchmarkGlobalDebouncer_Debounce(b *testing.B) {
	d := NewGlobalDebouncer(1 * time.Second)
	defer d.Stop()

	runGlobalDebouncerBenchmark(b, d)
}

func runDebouncerBenchmark(b *testing.B, d *Debouncer, key DebounceKey) {
	b.Helper()
	b.ResetTimer()

	for i := range b.N {
		d.Debounce(key, func() {})

		_ = i
	}
}

func runGlobalDebouncerBenchmark(b *testing.B, d *GlobalDebouncer) {
	b.Helper()
	b.ResetTimer()

	for i := range b.N {
		d.Debounce(NewDebounceKey(""), func() {})

		_ = i
	}
}
