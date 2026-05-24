# Data Model

## The Sort-Key Model

The organizing principle of aveAI is the **sort-key**: a hierarchical, human-readable string that classifies every entry into a namespace.

```
code/conventions/go
     ↑         ↑
  domain    subdomain
```

Sort-keys are defined in `map.yaml`. Every entry has exactly one sort-key. They form a tree, but each entry sits at a leaf (or any node).

### Entry

```go
type Entry struct {
    ID        uint64            `json:"id"`
    SortKey   string            `json:"sort_key"`     // e.g. "code/conventions/go"
    Value     string            `json:"value"`         // The actual context content
    Tags      []string          `json:"tags,omitempty"`
    Metadata  map[string]string `json:"metadata,omitempty"`
    Created   int64             `json:"created"`
    Updated   int64             `json:"updated"`
    Embedding []float64         `json:"embedding,omitempty"`
}
```

### Indexes

| Index | Type | Purpose |
|-------|------|---------|
| `BySortKey` | `map[string][]uint64` | Fast lookup by sort-key prefix |
| `ByTag` | `map[string][]uint64` | Lookup by tag |
| `TextIndex` | Custom inverted index | Full-text search (Phase 2) |
| `VecIndex` | Flat list | Cosine similarity (Phase 3) |

## .avdb File Format

A single binary file with a header, then serialized entries:

```
┌──────────────────────────────┐
│ Magic: "AVE0" (4 bytes)      │
│ Version: uint16              │
│ EntryCount: uint32           │
│ IndexOffset: uint64          │    ← points to serialized indexes
│ ┌──────────────────────────┐ │
│ │ Entry 1 (gob-encoded)    │ │
│ │ Entry 2 (gob-encoded)    │ │
│ │ ...                      │ │
│ └──────────────────────────┘ │
│ ┌──────────────────────────┐ │
│ │ Index data (gob-encoded) │ │
│ │ - BySortKey map          │ │
│ │ - ByTag map              │ │
│ │ - TextIndex (Phase 2)    │ │
│ │ - VecIndex (Phase 3)     │ │
│ └──────────────────────────┘ │
└──────────────────────────────┘
```

**Format choices considered:**

| Format | Verdict | Rationale |
|--------|---------|-----------|
| **gob** | ✅ Primary | Built-in, fast, handles Go types natively |
| JSON | ❌ | Slow to parse, no binary support for embeddings |
| protobuf | ❌ | Adds codegen step, overkill |
| custom binary | ❌ | Premature optimization |

**Compression:** Not applied at the file level. If entries are large, individual values may be gzip-compressed with a flag byte.

## map.yaml Schema

```yaml
# map.yaml
version: 1
keys:
  <namespace>:
    description: "<human-readable description>"
    aliases: ["<alternative key paths>"]   # optional
    children:
      <sub-namespace>:
        description: "..."
```

The `map.yaml` lives at the root of where the `.avdb` file is found. Default lookup path:

1. `./ave.yaml` / `./ave.yml`
2. `./.ave/map.yaml`
3. `~/.config/ave/map.yaml`

## Single-Table Principle

All entries live in one flat list. There are no tables, no relationships, no joins. The sort-key + tags provide all the grouping needed. This keeps the mental model simple and avoids the complexity the user explicitly wants to sidestep.
