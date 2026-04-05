# go-filewatcher

A high-level, composable file system watcher for Go, built on [fsnotify](https://github.com/fsnotify/fsnotify).

Eliminates the boilerplate of raw fsnotify usage by providing sensible defaults for common patterns: automatic recursive directory watching, configurable debounce, composable filters, middleware chains, and graceful context-based shutdown.

## Installation

```bash
go get github.com/larsartmann/go-filewatcher
```

## Quick Start

```go
watcher, err := filewatcher.New(
    []string{"./src"},
    filewatcher.WithExtensions(".go"),
    filewatcher.WithDebounce(500*time.Millisecond),
    filewatcher.WithIgnoreDirs("vendor", "node_modules"),
)
if err != nil {
    log.Fatal(err)
}
defer watcher.Close()

events, err := watcher.Watch(ctx)
for event := range events {
    fmt.Printf("%s: %s\n", event.Op, event.Path)
}
```

## Options

| Option                    | Description                                                 |
| ------------------------- | ----------------------------------------------------------- |
| `WithDebounce(d)`         | Global debounce — all events coalesced into one after delay |
| `WithPerPathDebounce(d)`  | Per-path debounce — each file debounced independently       |
| `WithFilter(f)`           | Add a custom filter function                                |
| `WithExtensions(exts...)` | Only emit events for given file extensions                  |
| `WithIgnoreDirs(dirs...)` | Discard events from given directory names                   |
| `WithIgnoreHidden()`      | Discard events for hidden files/dirs (dot prefix)           |
| `WithRecursive(b)`        | Enable/disable recursive directory watching (default: true) |
| `WithMiddleware(m...)`    | Add middleware to the event processing pipeline             |
| `WithErrorHandler(fn)`    | Set custom error handler for watcher errors                 |

## Filters

Built-in filters combine with AND/OR/NOT logic:

```go
filter := filewatcher.FilterAnd(
    filewatcher.FilterExtensions(".go"),
    filewatcher.FilterNot(filewatcher.FilterIgnoreDirs("vendor")),
    filewatcher.FilterOperations(filewatcher.Write, filewatcher.Create),
)
```

Available: `FilterExtensions`, `FilterIgnoreExtensions`, `FilterIgnoreDirs`, `FilterIgnoreHidden`, `FilterOperations`, `FilterNotOperations`, `FilterGlob`, `FilterRegex`, `FilterMinSize`, `FilterAnd`, `FilterOr`, `FilterNot`.

## Middleware

```go
watcher, _ := filewatcher.New(paths,
    filewatcher.WithMiddleware(
        filewatcher.MiddlewareRecovery(),
        filewatcher.MiddlewareLogging(nil),
    ),
)
```

Available: `MiddlewareLogging`, `MiddlewareRecovery`, `MiddlewareRateLimit`, `MiddlewareFilter`, `MiddlewareOnError`, `MiddlewareMetrics`, `MiddlewareWriteFileLog`.

## Event Types

| Op       | Description               |
| -------- | ------------------------- |
| `Create` | File or directory created |
| `Write`  | File modified             |
| `Remove` | File or directory removed |
| `Rename` | File or directory renamed |

## Design

Follows functional options, sentinel errors (`errors`/`fmt.Errorf`), middleware chains, channel-based streaming, minimal dependencies (only [fsnotify](https://github.com/fsnotify/fsnotify)).

## License

Proprietary — See [LICENSE](LICENSE) file.
