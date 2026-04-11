# Examples

Runnable examples demonstrating go-filewatcher usage.

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
