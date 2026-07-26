# Examples

Runnable examples demonstrating go-filewatcher usage.

## Shared Helpers

All examples use the `examples/demo` package which provides:

- **`demo.MustWatch(ctx, paths, opts...)`** — creates a watcher, starts
  watching, and returns an events channel plus a cleanup function. Calls
  `log.Fatal` on startup failure (suitable for example programs where errors
  are unrecoverable). **Always defer the cleanup function.**
- **`demo.Run(fn)`** — wraps the context-setup boilerplate with a 10s timeout.
- **`demo.PrintEvent(event)`** — logs an event with timestamp and operation.

## Running Examples

```bash
# Basic usage
go run ./examples/basic

# Per-path debounce
go run ./examples/per-path-debounce

# Middleware chain
go run ./examples/middleware

# Filter auto-generated code
go run ./examples/filter-generated
```

## Examples

| Example                                  | Description                                        |
| ---------------------------------------- | -------------------------------------------------- |
| [basic](./basic)                         | Simplest usage with extensions filter and debounce |
| [per-path-debounce](./per-path-debounce) | Each file debounced independently                  |
| [middleware](./middleware)               | Logging, recovery, and metrics middleware          |
| [filter-generated](./filter-generated)   | Exclude auto-generated Go files from events        |
