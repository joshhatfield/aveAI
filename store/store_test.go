package store

import (
	"testing"
)

func TestNewStoreIsEmpty(t *testing.T) {
	s := New()
	if s.Len() != 0 {
		t.Errorf("expected empty store, got %d entries", s.Len())
	}
}

func TestAddEntry(t *testing.T) {
	s := New()
	id, err := s.Add(Entry{
		SortKey: "code/conventions",
		Value:   "always wrap errors with %w",
		Tags:    []string{"go", "errors"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 1 {
		t.Errorf("expected id 1, got %d", id)
	}
	if s.Len() != 1 {
		t.Errorf("expected 1 entry, got %d", s.Len())
	}
}

func TestAddEntryRequiresSortKey(t *testing.T) {
	s := New()
	_, err := s.Add(Entry{Value: "no sort key"})
	if err == nil {
		t.Fatal("expected error for missing sort_key")
	}
}

func TestAddEntryRequiresValue(t *testing.T) {
	s := New()
	_, err := s.Add(Entry{SortKey: "test"})
	if err == nil {
		t.Fatal("expected error for missing value")
	}
}

func TestGetEntry(t *testing.T) {
	s := New()
	id, _ := s.Add(Entry{SortKey: "test", Value: "hello"})
	e, err := s.Get(id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Value != "hello" {
		t.Errorf("expected 'hello', got '%s'", e.Value)
	}
}

func TestGetMissingEntry(t *testing.T) {
	s := New()
	_, err := s.Get(999)
	if err == nil {
		t.Fatal("expected error for missing entry")
	}
}

func TestListBySortKeyPrefix(t *testing.T) {
	s := New()
	s.Add(Entry{SortKey: "code/conventions/go", Value: "use gofmt"})
	s.Add(Entry{SortKey: "code/conventions/python", Value: "use black"})
	s.Add(Entry{SortKey: "notes/decisions", Value: "use postgres"})

	results := s.List("code/conventions")
	if len(results) != 2 {
		t.Errorf("expected 2 entries, got %d", len(results))
	}

	results = s.List("code")
	if len(results) != 2 {
		t.Errorf("expected 2 entries, got %d", len(results))
	}
}

func TestListAll(t *testing.T) {
	s := New()
	s.Add(Entry{SortKey: "a", Value: "1"})
	s.Add(Entry{SortKey: "b", Value: "2"})

	results := s.List("")
	if len(results) != 2 {
		t.Errorf("expected 2 entries, got %d", len(results))
	}
}

func TestDeleteEntry(t *testing.T) {
	s := New()
	id, _ := s.Add(Entry{SortKey: "test", Value: "delete me"})
	err := s.Delete(id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Len() != 0 {
		t.Errorf("expected 0 entries after delete, got %d", s.Len())
	}
	// Index should be cleaned up
	if len(s.BySortKey["test"]) != 0 {
		t.Errorf("expected 0 indexed entries, got %d", len(s.BySortKey["test"]))
	}
}

func TestStats(t *testing.T) {
	s := New()
	s.Add(Entry{SortKey: "code/conventions", Value: "a", Tags: []string{"go"}})
	s.Add(Entry{SortKey: "code/patterns", Value: "b", Tags: []string{"go"}})
	s.Add(Entry{SortKey: "notes/decisions", Value: "c", Tags: []string{"general"}})

	stats := s.Stats()
	if stats["entries"] != 3 {
		t.Errorf("expected 3 entries in stats, got %v", stats["entries"])
	}
}

func TestAutoIncrementID(t *testing.T) {
	s := New()
	id1, _ := s.Add(Entry{SortKey: "a", Value: "first"})
	id2, _ := s.Add(Entry{SortKey: "a", Value: "second"})
	if id2 != id1+1 {
		t.Errorf("expected id %d, got %d", id1+1, id2)
	}
}

func TestEntryTimestamps(t *testing.T) {
	s := New()
	id, _ := s.Add(Entry{SortKey: "test", Value: "timestamps"})
	e, _ := s.Get(id)
	if e.Created == 0 {
		t.Error("expected non-zero Created timestamp")
	}
	if e.Updated == 0 {
		t.Error("expected non-zero Updated timestamp")
	}
	if e.Created != e.Updated {
		t.Errorf("expected Created == Updated for new entry, got %d vs %d", e.Created, e.Updated)
	}
}
