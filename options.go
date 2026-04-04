package filewatcher

import "time"

// Option configures a Watcher during creation.
type Option func(*Watcher)

// WithDebounce sets a global debounce delay. All events are coalesced
// into a single emission after the delay since the last event.
// Default is no debouncing.
func WithDebounce(d time.Duration) Option {
	return func(w *Watcher) {
		w.globalDebounce = d
	}
}

// WithPerPathDebounce sets a per-path debounce delay. Events for different
// file paths are debounced independently. This is useful when watching
// many files and changes to different files should trigger separate actions.
func WithPerPathDebounce(d time.Duration) Option {
	return func(w *Watcher) {
		w.perPathDebounce = d
	}
}

// WithFilter adds an event filter. Only events that pass all registered
// filters are emitted. Multiple filters are ANDed together.
func WithFilter(f Filter) Option {
	return func(w *Watcher) {
		w.filters = append(w.filters, f)
	}
}

// WithExtensions filters events to only those matching the given file
// extensions. Extensions should include the dot prefix (e.g., ".go", ".md").
func WithExtensions(exts ...string) Option {
	return func(w *Watcher) {
		w.filters = append(w.filters, FilterExtensions(exts...))
	}
}

// WithIgnoreDirs discards events for files within the given directory names.
// Common values: "vendor", "node_modules", ".git", "dist", "build", "bin".
func WithIgnoreDirs(dirs ...string) Option {
	return func(w *Watcher) {
		w.filters = append(w.filters, FilterIgnoreDirs(dirs...))
	}
}

// WithIgnoreHidden discards events for hidden files and directories
// (those starting with a dot).
func WithIgnoreHidden() Option {
	return func(w *Watcher) {
		w.filters = append(w.filters, FilterIgnoreHidden())
	}
}

// WithRecursive enables recursive directory watching. When enabled,
// subdirectories are added to the watcher automatically, and newly
// created directories are added dynamically. Default is true.
func WithRecursive(b bool) Option {
	return func(w *Watcher) {
		w.recursive = b
	}
}

// WithMiddleware adds middleware to the event processing pipeline.
// Middleware is applied in reverse order (last added runs first),
// matching the go-cqrs-lite convention.
func WithMiddleware(m ...Middleware) Option {
	return func(w *Watcher) {
		w.middleware = append(w.middleware, m...)
	}
}

// WithErrorHandler sets a callback for watcher errors that occur during
// the event loop. Errors are passed to this handler instead of being
// silently dropped. If not set, errors are logged to stderr.
func WithErrorHandler(handler func(error)) Option {
	return func(w *Watcher) {
		w.errorHandler = handler
	}
}

// WithSkipDotDirs controls whether directories starting with a dot (.
// are skipped during recursive directory walking. Default is true.
// Set to false to watch dot-directories like .config, .vscode, etc.
func WithSkipDotDirs(skip bool) Option {
	return func(w *Watcher) {
		w.skipDotDirs = skip
	}
}
