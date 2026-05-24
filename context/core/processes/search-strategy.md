# Search Strategy

aveAI supports a progression of search capabilities, each building on the previous.

## Phase 1: Sort-Key + Tag Filtering (Always Available)

The most basic search. Filter by sort-key prefix and/or tags.

```bash
ave search code/conventions
ave search --tag go
ave search code/errors --tag critical
```

**Implementation:** Simple map lookups on `BySortKey` (prefix match) and `ByTag`. No scoring, returns in insertion order. Fast, trivial, works on day one.

## Phase 2: Full-Text Search

### Inverted Index

A custom in-memory inverted index built per-session when loading the `.avdb`:

```
Term "error" → [{EntryID: 5, Freq: 3, Positions: [12,45,78]},
                {EntryID: 12, Freq: 1, Positions: [23]}]
Term "wrap"   → [{EntryID: 5, Freq: 2, Positions: [13,46]}]
```

**Tokenization:** Whitespace + punctuation splitting. Lowercased. Configurable stop-word list.

### Scoring (TF-IDF)

- **TF** — Term frequency in the entry (raw count, normalized by max freq in doc)
- **IDF** — Inverse document frequency across all entries
- **Score** = TF × IDF

### Query Syntax

```bash
ave search "golang error wrapping"
ave search --phrase "always wrap errors with %w"
ave search "conventions" --limit 5
```

Tokens AND by default. `--phrase` for exact phrase search using position data. Results sorted by descending score.

### Storage

The inverted index is **rebuilt in memory on each load**, not persisted to disk. Why:
- Context stores are small (<10k entries typically)
- Building the index is O(N × avg_terms_per_entry) — sub-millisecond for reasonable sizes
- Avoids index consistency issues
- Simpler code

If profiling shows this is a bottleneck, we can persist the index in the `.avdb` file.

## Phase 3: Vector Search

### Embedding Interface

```go
type Embedder interface {
    Embed(text string) ([]float64, error)
    Dims() int
}
```

Built-in embedders:
- **None** (default) — returns error, falls back to text search
- **Local** — calls a local embedding model (e.g., llama.cpp, ONNX)
- **API** — calls OpenAI / Cohere / etc. API

The embedder is selected via `ave search --embed` or configured in `map.yaml`:

```yaml
embedder:
  type: local          # none | local | api
  model: all-MiniLM-L6-v2
  endpoint: "http://localhost:8080/v1/embeddings"
```

### Similarity Search

Flat brute-force cosine similarity:

```go
func (vi *VectorIndex) Search(query []float64, k int) []Result {
    type scored struct {
        id    uint64
        score float64
    }
    results := make([]scored, len(vi.Embeddings))
    for i, emb := range vi.Embeddings {
        results[i] = scored{vi.EntryIDs[i], cosineSimilarity(query, emb)}
    }
    sort.Slice(results, func(i, j int) bool {
        return results[i].score > results[j].score
    })
    // Return top-k
}
```

This is O(N × D) where N = entry count, D = embedding dimensions. For <10k entries with 384-dim embeddings, this is ~3M multiply-adds — under 10ms on modern hardware. **No need for ANN indexes at this scale.**

### Hybrid Search (Future)

Combine text and vector scores:

```go
score = α * textScore + (1-α) * vectorScore
```

Where α is configurable. This is deferred until both text and vector search are stable.

## Fallback Priority

When `--embed` is specified but no embedder is available:

1. Print a warning: "No embedder configured, falling back to text search"
2. Perform text search instead
3. The user never gets empty results — they always get the best available

## Operator Combinations

| Command | Search Method | Fallback |
|---------|---------------|----------|
| `ave search "query"` | Text (inverted index) | Sort-key prefix filter |
| `ave search --embed "query"` | Vector (cosine sim) | Text search if no embedder |
| `ave search -s code/foo` | Sort-key prefix filter | None (exact match) |
| `ave search --tag go` | Tag filter | None (exact match) |
| `ave search "q" --tag go -s code` | Combined (intersection) | Text + filter |
