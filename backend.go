package filewatcher

import "github.com/fsnotify/fsnotify"

// watchBackend abstracts the filesystem notification backend so that tests can
// inject a fake that scripts events, errors, and Add failures deterministically.
type watchBackend interface {
	Add(name string) error
	Remove(name string) error
	Close() error
	Events() <-chan fsnotify.Event
	Errors() <-chan error
}

// fsnotifyBackend adapts *fsnotify.Watcher to satisfy watchBackend.
// fsnotify exposes Events and Errors as channel fields rather than methods,
// so this wrapper exposes them as methods for interface conformance.
type fsnotifyBackend struct {
	*fsnotify.Watcher
}

func (b fsnotifyBackend) Events() <-chan fsnotify.Event { return b.Watcher.Events }
func (b fsnotifyBackend) Errors() <-chan error          { return b.Watcher.Errors }

// withBackend overrides the filesystem notification backend.
// This is an unexported option for test injection only.
func withBackend(b watchBackend) Option {
	return func(w *Watcher) {
		w.fswatcher = b
	}
}
