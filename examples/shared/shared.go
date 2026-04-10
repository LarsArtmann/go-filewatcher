// Package shared provides helper functions for examples.
package shared

import (
	"fmt"
	"time"

	filewatcher "github.com/larsartmann/go-filewatcher"
)

const timeFormat = "15:04:05.000"

const defaultTimeout = 10 * time.Second

// PrintEvent prints an event with millisecond precision.
func PrintEvent(event filewatcher.Event) {
	ts := event.Timestamp.Format(timeFormat)
	fmt.Printf("[%s] %s: %s\n", ts, event.Op.String(), event.Path)
}

// DefaultTimeout returns the default timeout duration for examples.
func DefaultTimeout() time.Duration {
	return defaultTimeout
}
