package filewatcher

import (
	"sync"

	"github.com/fsnotify/fsnotify"
)

// Compile-time assertion that fakeBackend satisfies watchBackend. If the
// interface or the fake drifts, this fails at build time. The matching
// fsnotifyBackend assertion lives in backend.go.
var _ watchBackend = (*fakeBackend)(nil)

// fakeBackend is a test double for watchBackend that allows injecting scripted
// events, errors, and Add failures. It enables deterministic testing of
// self-heal, error handling, and the event pipeline without a real filesystem
// notification backend.
type fakeBackend struct {
	mu           sync.Mutex
	events       chan fsnotify.Event
	errCh        chan error
	addFn        func(path string, attempt int) error // nil = always succeed
	addAttempts  map[string]int
	addedPaths   []string
	removedPaths []string
	closeOnce    sync.Once
}

// newFakeBackend creates a fakeBackend with buffered event and error channels.
// By default Add always succeeds; override addFn to script failures.
func newFakeBackend() *fakeBackend {
	return &fakeBackend{
		events:       make(chan fsnotify.Event, 100),
		errCh:        make(chan error, 100),
		addAttempts:  make(map[string]int),
		addedPaths:   make([]string, 0),
		removedPaths: make([]string, 0),
	}
}

func (f *fakeBackend) Add(name string) error {
	f.mu.Lock()
	f.addAttempts[name]++

	attempt := f.addAttempts[name]
	fn := f.addFn

	f.addedPaths = append(f.addedPaths, name)
	f.mu.Unlock()

	if fn != nil {
		return fn(name, attempt)
	}

	return nil
}

func (f *fakeBackend) Remove(name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.removedPaths = append(f.removedPaths, name)

	return nil
}

func (f *fakeBackend) Close() error {
	f.closeOnce.Do(func() {
		close(f.events)
		close(f.errCh)
	})

	return nil
}

func (f *fakeBackend) Events() <-chan fsnotify.Event { return f.events }
func (f *fakeBackend) Errors() <-chan error          { return f.errCh }

// sendEvent injects a filesystem event into the events channel.
func (f *fakeBackend) sendEvent(e fsnotify.Event) {
	f.events <- e
}

// sendError injects an error into the errors channel.
func (f *fakeBackend) sendError(err error) {
	f.errCh <- err
}

// addAttemptCount returns the number of times Add was called for a path.
func (f *fakeBackend) addAttemptCount(path string) int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.addAttempts[path]
}
