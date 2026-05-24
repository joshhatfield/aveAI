# Phase 2: Full-Text Search

**Status:** Not started
**Dependencies:** Phase 1 (Store + .avdb + CRUD working)
**Estimated effort:** ~4-6 hours

## Goals

- Custom inverted index built in memory on load
- `ave search <query>` command with TF-IDF scoring
- Phrase search support
- Combine with sort-key + tag filtering

## Implementation Steps

### Step 1: Tokenizer

Create `search/tokenizer.go`:

- Split on whitespace and punctuation
- Lowercase all tokens
- Configurable stop-word list (the, a, an, in, on, at, for, to, of, is, it)
- Return `[]Token` where `Token` has `Term string` and `Position int`

### Step 2: Inverted Index

Create `search/text.go`:

```go
type Posting struct {
    EntryID  uint64
    Freq     int
    Positions []int
}

type InvertedIndex struct {
    mu       sync.RWMutex
    Postings map[string][]Posting  // term → postings
    DocCount int                   // total documents indexed
}
```

- `Build(entries []Entry)` — tokenize all entries, build postings list
- `Search(query string) []ScoredResult` — tokenize query, intersect postings, score

### Step 3: TF-IDF Scoring

```go
type ScoredResult struct {
    EntryID uint64
    Score   float64
}
```

- **TF(t,d)** = raw term frequency in doc d (log-scaled)
- **IDF(t)** = log(N / df_t) where N = total docs, df_t = docs containing t
- **Score(q,d)** = Σ TF(t,d) × IDF(t) for each query term t

### Step 4: Query Parser

Support:

| Syntax | Example | Behavior |
|--------|---------|----------|
| Bare tokens | `error wrapping` | AND — all terms must match |
| `--phrase` | `--phrase "always wrap"` | Exact phrase using position data |
| `--limit N` | `--limit 10` | Top-K results |
| `--sort-key` | `-s code/` | Filter to sort-key prefix |
| `--tag` | `--tag go` | Filter by tag |

### Step 5: CLI Integration

```bash
ave search "golang error handling"
ave search "always wrap" --phrase
ave search "panic" -s code/conventions --tag go --limit 5
```

Results formatted as:

```
 5: code/conventions    "Always wrap errors with %w, never panic"
    Score: 0.874 | Tags: [go, conventions]

12: code/conventions    "Use recover() only in goroutines"
    Score: 0.231 | Tags: [go]
```

## Acceptance Criteria

- [ ] `ave search "query"` returns relevant results sorted by score
- [ ] `ave search "phrase words" --phrase` returns only exact phrase matches
- [ ] Sort-key + tag filters narrow results correctly
- [ ] `--limit` truncates results
- [ ] Empty query returns error message
- [ ] No results returns empty with "no results" message
- [ ] Index build is <1ms for 1000 entries
- [ ] Query latency is <1ms for typical queries
- [ ] `go test ./...` passes

## Files to Create / Modify

```
search/tokenizer.go
search/tokenizer_test.go
search/text.go          (inverted index)
search/text_test.go
cmd/search.go           (new command)
store/store.go          (add BuildIndex method, integrate on load)
```
