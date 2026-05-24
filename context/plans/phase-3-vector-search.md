# Phase 3: Vector Search

**Status:** Not started
**Dependencies:** Phase 2 (Text search working)
**Estimated effort:** ~4-6 hours

## Goals

- Embedding interface with pluggable backends
- Brute-force cosine similarity search
- Hybrid text + vector search (optional weighting)
- Graceful fallback when no embedder is available

## Implementation Steps

### Step 1: Embedder Interface

Create `search/embed.go`:

```go
package search

type Embedder interface {
    Embed(text string) ([]float64, error)
    Dims() int
    Name() string
}

// NoopEmbedder — returns error, triggers text fallback
type NoopEmbedder struct{}

func (e *NoopEmbedder) Embed(text string) ([]float64, error) {
    return nil, fmt.Errorf("no embedder configured")
}
```

### Step 2: Built-in Embedders

**Local Embedder** — calls a local HTTP endpoint:

```go
type LocalEmbedder struct {
    Endpoint string  // e.g. "http://localhost:8080/v1/embeddings"
    Model    string
    Dim      int
}
```

Expected API contract: `POST /v1/embeddings` with `{"model": "...", "input": "..."}` returning `{"data": [{"embedding": [...]}]}` — compatible with the OpenAI embeddings API format.

**API Embedder** — calls OpenAI / Cohere / etc.:

```go
type APIEmbedder struct {
    Provider string  // "openai" | "cohere" | etc.
    APIKey   string
    Model    string
}
```

### Step 3: Vector Index

Create `search/vector.go`:

```go
type VectorIndex struct {
    EntryIDs   []uint64
    Embeddings [][]float64
}

func (vi *VectorIndex) Build(entries []Entry, embedder Embedder) error
func (vi *VectorIndex) Search(query []float64, k int) []ScoredResult
```

- `Build` — iterate entries with non-empty Embedding field, build flat list
- `Search` — brute-force cosine similarity, return top-K

### Step 4: Embedding on Add

When an embedder is configured, `ave add` can optionally compute and store the embedding:

```bash
ave add --embed code/conventions "always wrap errors"
```

This calls the embedder, stores the resulting vector in the entry's `Embedding` field, and persists it in the `.avdb`.

### Step 5: Vector Search CLI

```bash
ave search --embed "error handling patterns"       # pure vector search
ave search --embed --alpha 0.5 "error handling"     # hybrid (vector + text)
```

- `--embed` flag enables vector search (or hybrid if combined with text)
- `--alpha` controls hybrid weighting: 0 = pure text, 1 = pure vector, 0.5 = equal
- If no embedder configured, prints warning and falls back to text

### Step 6: map.yaml Embedder Config

```yaml
embedder:
  type: none              # none | local | api
  model: ""               # model name
  endpoint: ""            # local endpoint URL
  api_key_env: ""         # env var name for API key
  dims: 384              # embedding dimensions
```

## Acceptance Criteria

- [ ] `--embed` flag present on `ave add` and `ave search`
- [ ] `ave add --embed` stores vector in `.avdb`
- [ ] `ave search --embed` performs cosine similarity search
- [ ] Results are sorted by descending similarity
- [ ] Without embedder, `--embed` falls back gracefully to text search with a warning
- [ ] Hybrid search (`--alpha`) combines text + vector scores
- [ ] Vector search on 1000 entries with 384-dim vectors completes in <50ms
- [ ] `go test ./...` passes

## Files to Create / Modify

```
search/embed.go         (Embedder interface)
search/embed_test.go
search/vector.go        (VectorIndex)
search/vector_test.go
cmd/search.go           (modify — add --embed, --alpha flags)
cmd/add.go              (modify — add --embed flag)
schema/map.go           (modify — add embedder config parsing)
store/entry.go          (modify — add Embedding field if not already)
store/store.go          (modify — add VecIndex, rebuild on load)
format/avdb.go          (modify — handle embeddings in save/load)
```
