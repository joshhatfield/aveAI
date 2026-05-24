package store

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// Store is the in-memory context database.
type Store struct {
	mu       sync.RWMutex
	Entries  []Entry
	BySortKey map[string][]uint64 // sort-key prefix → entry IDs
	ByTag     map[string][]uint64 // tag → entry IDs
	nextID   uint64
}

// New creates an empty Store.
func New() *Store {
	return &Store{
		BySortKey: make(map[string][]uint64),
		ByTag:     make(map[string][]uint64),
		nextID:    1,
	}
}

// Add inserts an entry and returns its ID.
func (s *Store) Add(e Entry) (uint64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if e.SortKey == "" {
		return 0, fmt.Errorf("sort_key is required")
	}
	if e.Value == "" {
		return 0, fmt.Errorf("value is required")
	}

	now := time.Now().Unix()
	e.ID = s.nextID
	e.Created = now
	e.Updated = now
	s.nextID++

	s.Entries = append(s.Entries, e)
	s.indexEntry(e)
	return e.ID, nil
}

// Get retrieves an entry by ID.
func (s *Store) Get(id uint64) (*Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for i := range s.Entries {
		if s.Entries[i].ID == id {
			return &s.Entries[i], nil
		}
	}
	return nil, fmt.Errorf("entry %d not found", id)
}

// List returns entries filtered by optional sort-key prefix.
func (s *Store) List(sortKeyPrefix string) []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if sortKeyPrefix == "" {
		result := make([]Entry, len(s.Entries))
		copy(result, s.Entries)
		return result
	}

	ids, ok := s.BySortKey[sortKeyPrefix]
	if !ok {
		return nil
	}

	result := make([]Entry, 0, len(ids))
	for _, id := range ids {
		if e, err := s.getByID(id); err == nil {
			result = append(result, *e)
		}
	}
	return result
}

// Delete removes an entry by ID.
func (s *Store) Delete(id uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx := -1
	for i := range s.Entries {
		if s.Entries[i].ID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("entry %d not found", id)
	}

	s.Entries = append(s.Entries[:idx], s.Entries[idx+1:]...)
	s.rebuildIndexes()
	return nil
}

// Len returns the number of entries.
func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.Entries)
}

// All returns a copy of all entries.
func (s *Store) All() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Entry, len(s.Entries))
	copy(result, s.Entries)
	return result
}

// --- internal ---

func (s *Store) getByID(id uint64) (*Entry, error) {
	for i := range s.Entries {
		if s.Entries[i].ID == id {
			return &s.Entries[i], nil
		}
	}
	return nil, fmt.Errorf("entry %d not found", id)
}

func (s *Store) indexEntry(e Entry) {
	// Index by sort-key prefixes (e.g., "code/conventions" → match "code/", "code/conventions/")
	parts := strings.Split(e.SortKey, "/")
	for i := range parts {
		prefix := strings.Join(parts[:i+1], "/")
		s.BySortKey[prefix] = append(s.BySortKey[prefix], e.ID)
	}

	// Index by tags
	for _, tag := range e.Tags {
		tag = strings.ToLower(strings.TrimSpace(tag))
		s.ByTag[tag] = append(s.ByTag[tag], e.ID)
	}
}

func (s *Store) rebuildIndexes() {
	s.BySortKey = make(map[string][]uint64)
	s.ByTag = make(map[string][]uint64)
	for _, e := range s.Entries {
		s.indexEntry(e)
	}
}

// Stats returns summary information about the store.
func (s *Store) Stats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keyCount := make(map[string]int)
	for _, e := range s.Entries {
		top := strings.SplitN(e.SortKey, "/", 2)[0]
		keyCount[top]++
	}

	return map[string]interface{}{
		"entries":     len(s.Entries),
		"sort_keys":   len(s.BySortKey),
		"tags":        len(s.ByTag),
		"key_summary": keyCount,
	}
}

// Ensure Entry sorting by insertion order.
type ByID []Entry

func (a ByID) Len() int           { return len(a) }
func (a ByID) Less(i, j int) bool { return a[i].ID < a[j].ID }
func (a ByID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func init() {
	// Verify ByID implements sort.Interface
	var _ sort.Interface = ByID{}
}
