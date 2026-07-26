// Package demo provides helper functions for examples.
package demo

import (
	"context"
	"log"
	"time"

	filewatcher "github.com/larsartmann/go-filewatcher/v2"
)

const timeFormat = "15:04:05.000"

const defaultTimeout = 10 * time.Second

// PrintEvent logs an event with millisecond precision.
func PrintEvent(event filewatcher.Event) {
	ts := event.Timestamp.Format(timeFormat)
	log.Printf("[%s] %s: %s\n", ts, event.Op.String(), event.Path)
}

// DefaultTimeout returns the default timeout duration for examples.
func DefaultTimeout() time.Duration {
	return defaultTimeout
}

// WithDefaultTimeout creates a context with the default timeout and a cancel function.
func WithDefaultTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), defaultTimeout)
}

// Run creates a context with the default demo timeout, invokes fn with it,
// then cancels the context. Wraps the standard context-setup boilerplate so
// each example program's main body stays focused on filewatcher usage.
func Run(fn func(ctx context.Context)) {
	ctx, cancel := WithDefaultTimeout()
	defer cancel()

	fn(ctx)
}

// MustWatch creates a watcher over paths with the given options, starts
// watching on ctx, and returns the events channel plus a cleanup function.
// It calls log.Fatal on construction or watch errors, making it suitable for
// example programs where startup failure is unrecoverable.
// The caller MUST defer the returned cleanup function to release resources.
func MustWatch(
	ctx context.Context,
	paths []string,
	opts ...filewatcher.Option,
) (<-chan filewatcher.Event, func()) {
	watcher, err := filewatcher.New(paths, opts...)
	if err != nil {
		log.Fatal(err)
	}

	events, err := watcher.Watch(ctx)
	if err != nil {
		_ = watcher.Close()

		log.Fatal(err)
	}

	return events, func() { _ = watcher.Close() }
}
