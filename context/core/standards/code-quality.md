# Code Quality Standards — aveAI

## Project Structure

```
.
├── main.go                 # CLI entrypoint
├── cmd/
│   └── root.go             # Cobra command definitions
├── store/
│   ├── store.go            # Store struct, load/save, CRUD
│   ├── entry.go            # Entry type
│   └── index.go            # Sort-key and tag indexes
├── search/
│   ├── text.go             # Inverted index (Phase 2)
│   ├── vector.go           # Vector index (Phase 3)
│   └── search.go           # Search orchestration
├── schema/
│   └── map.go              # map.yaml parsing
├── format/
│   ├── avdb.go             # .avdb read/write
│   └── avdb_test.go        # Format tests
├── go.mod
└── context/                # Project context (you are here)
```

## Go Conventions

- **Go 1.22+** — use range-over-func, clear(), improved loops
- **No global state** — everything should be injectable. The store is created in `main()` and passed through.
- **No panics** — always return errors. Panics are for truly unrecoverable states.
- **Functional options** for config — [functional options pattern](https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis)
- **`internal/`** for packages not meant for external consumption

## Testing

- **Tests live beside code** — `store_test.go` next to `store.go`
- **Table-driven tests** for search/query logic
- **`go test ./...` must pass** before any commit
- Aim for >80% coverage on the `store/` and `search/` packages
- Use `testing/slogtest` or a test logger, not `log.Println`

## Error Handling

```go
// Always wrap errors with context
if err != nil {
    return fmt.Errorf("load .avdb: %w", err)
}
// Use %w for wrapping, not %v or %s
```

## CLI (Cobra)

- Use [spf13/cobra](https://github.com/spf13/cobra) for CLI
- Commands in `cmd/root.go`, subcommands in `cmd/` package
- Short descriptions only — no novel-length help text
- Flags use `--` long form by default, `-` short form for common ones

## Imports

Standard library first, third-party second, internal third. Group with blank lines.

```go
import (
    "fmt"
    "os"

    "github.com/spf13/cobra"

    "aveAI/store"
)
```
