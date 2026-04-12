package filewatcher

import (
	"sync"
	"sync/atomic"
	"time"
)

const defaultDebounceDelay = 500 * time.Millisecond // Default delay for debouncing when none is specified

// debounceEntry holds a timer and its associated function for per-key debouncing.
type debounceEntry struct {
	fn    func()
	timer *time.Timer
}

// Debouncer prevents rapid successive function executions by coalescing
// calls within a delay window. It supports per-key debouncing so that
// different keys (e.g., file paths) are debounced independently.
type Debouncer struct {
	delay   time.Duration
	mu      sync.Mutex
	entries map[DebounceKey]*debounceEntry
	stopped atomic.Bool
	wg      sync.WaitGroup // tracks in-flight callbacks
}

// NewDebouncer creates a new Debouncer with the specified delay.
func NewDebouncer(delay time.Duration) *Debouncer {
	if delay <= 0 {
		delay = defaultDebounceDelay
	}

	return &Debouncer{
		delay:   delay,
		mu:      sync.Mutex{},
		entries: make(map[DebounceKey]*debounceEntry),
		stopped: atomic.Bool{},
		wg:      sync.WaitGroup{},
	}
}

// Debounce schedules fn to run after the delay, resetting any pending
// execution for the same key. This ensures callback runs only once for a burst
// of events sharing the same key.
func (d *Debouncer) Debounce(key DebounceKey, callback func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.stopped.Load() {
		return
	}

	if entry, exists := d.entries[key]; exists {
		entry.timer.Stop()
	}

	entry := &debounceEntry{
		fn:    callback,
		timer: nil,
	}
	entry.timer = time.AfterFunc(d.delay, func() {
		d.mu.Lock()
		delete(d.entries, key)
		stopped := d.stopped.Load()
		d.mu.Unlock()

		if stopped {
			return
		}

		callback()
		d.wg.Done()
	})
	d.entries[key] = entry
	d.wg.Add(1)
}

// Flush executes all pending functions immediately and clears all timers.
func (d *Debouncer) Flush() {
	d.mu.Lock()

	callbacks := make([]func(), 0, len(d.entries))

	for key, entry := range d.entries {
		entry.timer.Stop()
		callbacks = append(callbacks, entry.fn)

		delete(d.entries, key)
	}

	d.mu.Unlock()

	for _, fn := range callbacks {
		fn()
	}
}

// Stop cancels all pending executions without running them.
// Waits for any in-flight callbacks to complete before returning.
func (d *Debouncer) Stop() {
	d.mu.Lock()
	d.stopped.Store(true)

	for key, entry := range d.entries {
		entry.timer.Stop()
		delete(d.entries, key)
	}
	d.mu.Unlock()

	// Wait for any in-flight callbacks to complete
	d.wg.Wait()
}

// Pending returns the number of keys with pending executions.
func (d *Debouncer) Pending() int {
	d.mu.Lock()
	defer d.mu.Unlock()

	return len(d.entries)
}

// GlobalDebouncer coalesces all events into a single timer, regardless of key.
// Useful when you want to batch all file changes into one action.
type GlobalDebouncer struct {
	fn      func()
	timer   *time.Timer
	delay   time.Duration
	mu      sync.Mutex
	stopped atomic.Bool
	wg      sync.WaitGroup // tracks in-flight callbacks
}

// NewGlobalDebouncer creates a new GlobalDebouncer with the specified delay.
func NewGlobalDebouncer(delay time.Duration) *GlobalDebouncer {
	if delay <= 0 {
		delay = defaultDebounceDelay
	}

	return &GlobalDebouncer{
		delay:   delay,
		mu:      sync.Mutex{},
		stopped: atomic.Bool{},
		wg:      sync.WaitGroup{},
		fn:      nil,
		timer:   nil,
	}
}

// Debounce resets the global timer. callback runs only once after the delay
// since the last call, regardless of how many times Debounce is called.
func (g *GlobalDebouncer) Debounce(_ DebounceKey, callback func()) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.stopped.Load() {
		return
	}

	if g.timer != nil {
		g.timer.Stop()
	}

	g.fn = callback
	g.timer = time.AfterFunc(g.delay, func() {
		g.mu.Lock()
		g.timer = nil
		g.fn = nil
		stopped := g.stopped.Load()
		g.mu.Unlock()

		if stopped {
			return
		}

		callback()
		g.wg.Done()
	})
	g.wg.Add(1)
}

// Flush executes the pending function immediately and clears the timer.
func (g *GlobalDebouncer) Flush() {
	g.mu.Lock()

	var callback func()

	if g.timer != nil {
		g.timer.Stop()
		g.timer = nil
		callback = g.fn
		g.fn = nil
	}

	g.mu.Unlock()

	if callback != nil {
		callback()
	}
}

// Stop cancels the pending execution.
// Waits for any in-flight callback to complete before returning.
func (g *GlobalDebouncer) Stop() {
	g.mu.Lock()
	g.stopped.Store(true)

	if g.timer != nil {
		g.timer.Stop()
		g.timer = nil
	}

	g.fn = nil

	g.mu.Unlock()

	// Wait for any in-flight callback to complete
	g.wg.Wait()
}

// Pending returns whether there is a pending execution.
func (g *GlobalDebouncer) Pending() int {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.timer != nil {
		return 1
	}

	return 0
}
