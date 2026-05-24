package format

import (
	"os"
	"path/filepath"
	"testing"

	"aveAI/store"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.avdb")

	// Create a store and add some entries
	s := store.New()
	s.Add(store.Entry{SortKey: "code/conventions", Value: "use gofmt", Tags: []string{"go"}})
	s.Add(store.Entry{SortKey: "notes/decisions", Value: "use sqlite for storage"})
	s.Add(store.Entry{SortKey: "code/errors", Value: "always wrap errors with %w", Tags: []string{"go", "best-practice"}})

	// Save
	if err := Save(dbPath, s); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	// Verify file exists and has content
	info, err := os.Stat(dbPath)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("saved file is empty")
	}

	// Load
	loaded, err := Load(dbPath)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	// Verify contents
	if loaded.Len() != s.Len() {
		t.Errorf("entry count mismatch: got %d, want %d", loaded.Len(), s.Len())
	}

	entries := loaded.All()
	if entries[0].SortKey != "code/conventions" {
		t.Errorf("expected sort_key 'code/conventions', got '%s'", entries[0].SortKey)
	}
	if entries[0].Value != "use gofmt" {
		t.Errorf("expected value 'use gofmt', got '%s'", entries[0].Value)
	}
	if len(entries[0].Tags) != 1 || entries[0].Tags[0] != "go" {
		t.Errorf("expected tag 'go', got %v", entries[0].Tags)
	}
}

func TestLoadInvalidFile(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "bad.avdb")
	os.WriteFile(dbPath, []byte("not an avdb file"), 0644)

	_, err := Load(dbPath)
	if err == nil {
		t.Fatal("expected error loading invalid file")
	}
}

func TestSaveLoadEmptyStore(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "empty.avdb")

	s := store.New()
	if err := Save(dbPath, s); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := Load(dbPath)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if loaded.Len() != 0 {
		t.Errorf("expected empty store, got %d entries", loaded.Len())
	}
}
