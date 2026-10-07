# Named Types (phantom types)

## Named Types (phantom types)

Plain `type X string` named types for compile-time type safety on path-like strings:

| Type           | Purpose                             |
| -------------- | ----------------------------------- |
| `EventPath`    | Event file/directory paths          |
| `RootPath`     | Root directory paths during walking |
| `DebounceKey`  | Debouncer keys                      |
| `LogSubstring` | Log substring assertions (tests)    |
| `TempDir`      | Temp directory paths (tests)        |
| `OpString`     | Operation names                     |

**Usage:** Use constructor functions (e.g., `NewEventPath()`, `NewRootPath()`).

**EventPath has domain methods:** `.Base()`, `.Dir()`, `.Ext()`, `.Join()` for path operations.

---

_Moved from `AGENTS.md` on 2026-10-07 (carrying-capacity split); AGENTS.md links here._
