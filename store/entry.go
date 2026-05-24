package store

// Entry represents a single piece of context in the store.
type Entry struct {
	ID        uint64            `json:"id"`
	SortKey   string            `json:"sort_key"`
	Value     string            `json:"value"`
	Tags      []string          `json:"tags,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Created   int64             `json:"created"`
	Updated   int64             `json:"updated"`
	Embedding []float64         `json:"embedding,omitempty"`
}
