# Architecture

## CLI Lifecycle (Stateless Mode)

aveAI is primarily a CLI tool. Each invocation follows this flow:

```
┌──────────┐     ┌──────────┐     ┌────────────┐     ┌───────────┐
│  Parse   │ ──▶ │  Load    │ ──▶ │  Execute   │ ──▶ │  Write    │
│  CLI &   │     │  .avdb   │     │  Command   │     │  .avdb    │
│  map.yaml│     │  → mem   │     │  in mem    │     │  to disk  │
└──────────┘     └──────────┘     └────────────┘     └───────────┘
```

1. **Parse** — Read CLI args and locate the `.avdb` file + `map.yaml`
2. **Load** — Memory-map or deserialize the `.avdb` file into Go structs
3. **Execute** — Run the command (add, search, list, etc.) against in-memory data
4. **Write** — If the command mutated data, serialize and write back to `.avdb`

## In-Memory Store

At the heart of the system is a single `Store` struct:

```go
type Store struct {
    Entries    []Entry          // All entries in insertion order
    BySortKey  map[string][]int // sort-key → entry indices (hierarchical)
    TextIndex  *InvertedIndex   // term → posting list (Phase 2)
    VecIndex   *VectorIndex     // embedding → entries (Phase 3)
    Schema     Schema           // Parsed map.yaml
    Mut        sync.RWMutex
}
```

## Entry Structure

Each entry in the `.avdb` file:

| Field | Type | Description |
|-------|------|-------------|
| `ID` | uint64 | Auto-incrementing ID |
| `SortKey` | string | Hierarchical key, e.g. `code/conventions/go` |
| `Value` | string | The context content |
| `Tags` | []string | Optional tags for filtering |
| `Metadata` | map[string]string | Key-value metadata |
| `Created` | int64 | Unix timestamp |
| `Updated` | int64 | Unix timestamp |
| `Embedding` | []float64 | Optional vector embedding |

## map.yaml

The `map.yaml` file defines the schema — what sort keys are valid, their descriptions, and aliases:

```yaml
# map.yaml — defines the sort-key namespace
keys:
  code:
    description: "Code conventions and patterns"
    children:
      conventions:
        description: "Stylistic and structural conventions"
      patterns:
        description: "Design patterns used in the project"
      errors:
        description: "Error handling patterns"
  notes:
    description: "General project notes"
    children:
      decisions:
        description: "Architectural decision records"
      setup:
        description: "Setup and configuration notes"
  tools:
    description: "Tool usage patterns"
```

The sort-key is the central organizing concept. It acts as both a namespace and a primary filter for searches.

## Daemon Mode (Future)

For faster repeated queries, a `ave daemon` mode can keep the `.avdb` hot in memory and serve requests via a Unix socket. This is a future optimization, not the primary design.
