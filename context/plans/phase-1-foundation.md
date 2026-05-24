# Phase 1: Foundation — CLI, .avdb, CRUD, map.yaml

**Status:** Not started
**Dependencies:** None
**Estimated effort:** ~3-5 hours

## Goals

- Working `ave` CLI with `init`, `add`, `list`, `get` commands
- `.avdb` file format read/write
- `map.yaml` parsing
- Sort-key based indexing (map lookup, no search yet)
- `ave info` command to inspect the database

## Implementation Steps

### Step 1: Project Scaffold

```bash
go mod init aveAI
go get github.com/spf13/cobra
```

Create `main.go` with Cobra root command.

### Step 2: map.yaml Schema Parser



Create `schema/map.go`:

- Parse YAML into `Schema` struct
- Validate no duplicate keys
- Support aliases
- Default lookup paths (`./ave.yaml` → `./.ave/map.yaml` → `~/.config/ave/map.yaml`)

### Step 3: Entry + Store Types

Create `store/entry.go` and `store/store.go`:

- `Entry` struct (ID, SortKey, Value, Tags, Metadata, Created, Updated)
- `Store` struct with `Entries` slice + `BySortKey` map
- `New()` — create empty store
- `Add(entry)` — append, update index
- `Get(id)` — direct lookup
- `List(sortKeyPrefix)` — filter by sort-key
- `Delete(id)` — remove from slice + index

### Step 4: .avdb Format

Create `format/avdb.go`:

- `Save(path string, store *Store) error`
- `Load(path string) (*Store, error)`
- Binary format: magic bytes `"AVE0"` + version + entry count + gob-encoded entries
- Write tests for round-trip save/load

### Step 5: CLI Commands

Create `cmd/root.go`:

| Command | Args | Description |
|---------|------|-------------|
| `ave init` | [path] | Create empty .avdb + map.yaml |
| `ave add <sort-key> <value>` | sort-key, value | Add a new entry |
| `ave get <id>` | id | Show entry by ID |
| `ave list [sort-key]` | [prefix] | List entries |
| `ave info` | — | Show db stats |
| `ave delete <id>` | id | Remove entry |

### Step 6: Directory Resolution

- Determine which `.avdb` file to use:
  1. `--db` flag explicit path
  2. `./.ave/data.avdb`
  3. `~/.config/ave/data.avdb`
- Determine which `map.yaml` to use (same directories)

## Acceptance Criteria

- [ ] `ave init` creates a valid, loadable `.avdb` file
- [ ] `ave add code/conventions "always wrap errors"` persists the entry
- [ ] `ave list` shows all entries sorted by insertion order
- [ ] `ave list code/` lists only entries with that prefix
- [ ] `ave get 1` returns the exact entry
- [ ] `ave info` shows entry count, file size, schema keys
- [ ] `go test ./...` passes
- [ ] Round-trip save/load produces identical data

## Files to Create

```
main.go
cmd/root.go
store/entry.go
store/store.go
store/store_test.go
format/avdb.go
format/avdb_test.go
schema/map.go
schema/map_test.go
go.mod
```
