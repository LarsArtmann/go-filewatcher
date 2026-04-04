package filewatcher

import "github.com/cockroachdb/errors"

// ErrWatcherClosed is returned when operations are attempted on a closed watcher.
var ErrWatcherClosed = errors.New("watcher is closed")

// ErrNoPaths is returned when no paths are provided to the watcher.
var ErrNoPaths = errors.New("at least one path is required")

// ErrPathNotFound is returned when a specified path does not exist.
var ErrPathNotFound = errors.New("path not found")

// ErrPathNotDir is returned when a path is not a directory.
var ErrPathNotDir = errors.New("path is not a directory")
