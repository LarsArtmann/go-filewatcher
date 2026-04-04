package filewatcher

import (
	"sync"
	"time"
)

// debounceEntry holds a timer and its associated function.
type debounceEntry struct {
	timer *time.Timer
	fn    func()
}

// Debouncer prevents rapid successive function executions by coalescing
// calls within a delay window. It supports per-key debouncing so that
// different keys (e.g., file paths) are debounced independently.
type Debouncer struct {
	delay   time.Duration
	mu      sync.Mutex
	entries map[string]*debounceEntry
}

// NewDebouncer creates a new Debouncer with the specified delay.
func NewDebouncer(delay time.Duration) *Debouncer {
	if delay <= 0 {
		delay = 500 * time.Millisecond
	}
	return &Debouncer{
		delay:   delay,
		mu:      sync.Mutex{},
		entries: make(map[string]*debounceEntry),
	}
}

// Debounce schedules fn to run after the delay, resetting any pending
// execution for the same key. This ensures fn runs only once for a burst
// of events sharing the same key.
func (d *Debouncer) Debounce(key string, fn func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if entry, exists := d.entries[key]; exists {
		entry.timer.Stop()
	}

	entry := &debounceEntry{
		timer: nil,
		fn:    fn,
	}
	entry.timer = time.AfterFunc(d.delay, func() {
		fn()
		d.mu.Lock()
		delete(d.entries, key)
		d.mu.Unlock()
	})
	d.entries[key] = entry
}

// Flush executes all pending functions immediately and clears all timers.
func (d *Debouncer) Flush() {
	d.mu.Lock()
	defer d.mu.Unlock()

	for key, entry := range d.entries {
		entry.timer.Stop()
		entry.fn()
		delete(d.entries, key)
	}
}

// Stop cancels all pending executions without running them.
func (d *Debouncer) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()

	for key, entry := range d.entries {
		entry.timer.Stop()
		delete(d.entries, key)
	}
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
	delay time.Duration
	mu    sync.Mutex
	timer *time.Timer
	fn    func()
}

// NewGlobalDebouncer creates a new GlobalDebouncer with the specified delay.
func NewGlobalDebouncer(delay time.Duration) *GlobalDebouncer {
	if delay <= 0 {
		delay = 500 * time.Millisecond
	}
	return &GlobalDebouncer{
		delay: delay,
		mu:    sync.Mutex{},
		timer: nil,
		fn:    nil,
	}
}

// Debounce resets the global timer. fn runs only once after the delay
// since the last call, regardless of how many times Debounce is called.
func (g *GlobalDebouncer) Debounce(_ string, fn func()) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.timer != nil {
		g.timer.Stop()
	}

	g.fn = fn
	g.timer = time.AfterFunc(g.delay, fn)
}

// Flush executes the pending function immediately and clears the timer.
func (g *GlobalDebouncer) Flush() {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.timer != nil {
		g.timer.Stop()
		g.timer = nil
		if g.fn != nil {
			g.fn()
			g.fn = nil
		}
	}
}

// Stop cancels the pending execution.
func (g *GlobalDebouncer) Stop() {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.timer != nil {
		g.timer.Stop()
		g.timer = nil
	}
	g.fn = nil
}
