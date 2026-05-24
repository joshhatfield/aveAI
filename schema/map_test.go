package schema

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValidSchema(t *testing.T) {
	content := []byte(`
version: 1
keys:
  code:
    description: "Code conventions and patterns"
    children:
      conventions:
        description: "Stylistic and structural conventions"
      errors:
        description: "Error handling patterns"
  notes:
    description: "General project notes"
`)
	dir := t.TempDir()
	path := filepath.Join(dir, "map.yaml")
	os.WriteFile(path, content, 0644)

	s, err := LoadSchema(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Version != 1 {
		t.Errorf("expected version 1, got %d", s.Version)
	}
	if len(s.Keys) != 2 {
		t.Errorf("expected 2 top-level keys, got %d", len(s.Keys))
	}
}

func TestLoadSchemaWithMissingDescription(t *testing.T) {
	content := []byte(`
version: 1
keys:
  code:
    description: ""
`)
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	os.WriteFile(path, content, 0644)

	_, err := LoadSchema(path)
	if err == nil {
		t.Fatal("expected error for empty description")
	}
}

func TestLoadSchemaWithNoKeys(t *testing.T) {
	content := []byte(`version: 1`)
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.yaml")
	os.WriteFile(path, content, 0644)

	_, err := LoadSchema(path)
	if err == nil {
		t.Fatal("expected error for no keys")
	}
}

func TestLoadNonexistentFile(t *testing.T) {
	_, err := LoadSchema("/nonexistent/path.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestValidateNodesHappyPath(t *testing.T) {
	s := &Schema{
		Version: 1,
		Keys: map[string]*Node{
			"code": {
				Description: "Code stuff",
				Children: map[string]*Node{
					"go": {Description: "Go-specific"},
				},
			},
		},
	}
	if err := s.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
