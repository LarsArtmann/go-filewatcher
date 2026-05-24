package filewatcher

import (
	"fmt"
	"log/slog"
	"time"
)

// Option configures a Watcher during creation.
type Option func(*Watcher)

// WithDebounce sets a global debounce delay. All events are coalesced
// into a single emission after the delay since the last event.
// Default is no debouncing. Panics if delay is negative.
func WithDebounce(delay time.Duration) Option {
	if delay < 0 {
		panic(fmt.Sprintf("filewatcher: WithDebounce: negative duration %v", delay))
	}

	return func(w *Watcher) {
		w.globalDebounce = delay
	}
}

// WithPerPathDebounce sets a per-path debounce delay. Events for different
// file paths are debounced independently. This is useful when watching
// many files and changes to different files should trigger separate actions.
// Panics if delay is negative.
func WithPerPathDebounce(delay time.Duration) Option {
	if delay < 0 {
		panic(fmt.Sprintf("filewatcher: WithPerPathDebounce: negative duration %v", delay))
	}

	return func(w *Watcher) {
		w.perPathDebounce = delay
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
// Also skips these directories during recursive walking.
func WithIgnoreDirs(dirs ...string) Option {
	return func(w *Watcher) {
		w.filters = append(w.filters, FilterIgnoreDirs(dirs...))
		w.ignoreDirNames = append(w.ignoreDirNames, dirs...)
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
// the event loop. Errors are passed to this handler with context about
// what operation was being performed. If not set, errors are logged to stderr.
func WithErrorHandler(handler ErrorHandler) Option {
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

// WithBuffer sets the buffer size for the event channel.
// A larger buffer helps handle event bursts without dropping events.
// Default is 64. A value of 0 creates an unbuffered channel which may
// cause deadlocks if the consumer is slow; use with caution.
func WithBuffer(size int) Option {
	return func(w *Watcher) {
		if size >= 0 {
			w.bufferSize = size
		}
	}
}

// WithOnAdd sets a callback that is invoked whenever a new path is added
// to the watcher. This is useful for logging or tracking which directories
// are being watched.
func WithOnAdd(fn func(path string)) Option {
	return func(w *Watcher) {
		w.onAdd = fn
	}
}

// WithOnError sets a simple callback for errors that occur during watching.
// This is a convenience wrapper around WithErrorHandler for simple use cases.
func WithOnError(fn func(error)) Option {
	return func(w *Watcher) {
		w.errorHandler = func(_ ErrorContext, err error) {
			fn(err)
		}
	}
}

// WithLazyIsDir skips the os.Stat call in convertEvent for better performance.
// When enabled, Event.IsDir will always be false. This is useful when you
// don't need directory information and want to minimize filesystem calls.
// Default is false (IsDir is populated accurately).
func WithLazyIsDir() Option {
	return func(w *Watcher) {
		w.lazyIsDir = true
	}
}

// WithIgnorePatterns adds glob patterns to ignore at the filter level.
// Files matching any of the patterns are filtered out.
// Patterns are applied to the filename only (not the full path).
// Example patterns: "*.log", "*.tmp", ".*".
func WithIgnorePatterns(patterns ...string) Option {
	return func(w *Watcher) {
		w.filters = append(w.filters, FilterIgnoreGlobs(patterns...))
	}
}

// WithPollInterval sets a polling interval for filesystem change detection.
// When set to a non-zero duration, the watcher will periodically poll watched
// paths instead of relying solely on OS-native file events. This is useful
// for NFS mounts, FUSE filesystems, and other environments where inotify/FSEvents
// are unreliable or unavailable.
//
// A typical value is 1–5 seconds. Shorter intervals detect changes faster
// but consume more CPU. Set to 0 (default) to disable polling.
func WithPollInterval(interval time.Duration) Option {
	return func(w *Watcher) {
		w.pollInterval = interval
	}
}

const defaultPollInterval = 2 * time.Second // Default polling interval for NFS/FUSE

// WithPolling enables or disables polling-based filesystem watching as a
// complement to OS-native events. When enabled (fallback=true), the watcher
// supplements fsnotify with periodic polling at the configured poll interval
// (default 2s, customizable via WithPollInterval). This detects changes on
// NFS mounts, FUSE filesystems, Docker volumes, and other environments where
// inotify/FSEvents may not fire.
//
// When fallback=false, polling is fully disabled (identical to not calling
// this option).
func WithPolling(fallback bool) Option {
	return func(w *Watcher) {
		w.polling = fallback
		if fallback && w.pollInterval == 0 {
			w.pollInterval = defaultPollInterval
		}
	}
}

// WithDebug enables verbose debug logging to the provided logger.
// If logger is nil, log/slog.Default() is used. When enabled, the watcher
// logs detailed information about event processing, filter decisions,
// debounce actions, and error handling.
func WithDebug(logger *slog.Logger) Option {
	return func(w *Watcher) {
		w.debug = true
		w.debugLogger = logger
	}
}

// WithWatchedIgnoreDirs adds directory names that are ignored at the event
// filtering level only, without affecting directory walking. This differs
// from WithIgnoreDirs which both filters events AND skips directories during
// recursive walking. Use this when you want to walk into a directory (to watch
// its subdirectories) but ignore file events within it.
//
// Deprecated: Use WithFilter(FilterIgnoreDirs(dirs...)) instead for better composability.
func WithWatchedIgnoreDirs(dirs ...string) Option {
	return func(w *Watcher) {
		w.filters = append(w.filters, FilterIgnoreDirs(dirs...))
	}
}

// WithFollowSymlinks enables following symbolic links during directory walking.
// When enabled, symlinked directories are resolved to their targets and
// added to the watcher. This allows monitoring directories accessed through
// symlinks. Default is false (symlinks are not followed).
func WithFollowSymlinks(follow bool) Option {
	return func(w *Watcher) {
		w.followSymlinks = follow
	}
}
