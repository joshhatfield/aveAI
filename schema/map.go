package schema

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Schema represents a parsed map.yaml file.
type Schema struct {
	Version int              `yaml:"version"`
	Keys    map[string]*Node `yaml:"keys"`
}

// Node represents a single key or namespace in the key hierarchy.
type Node struct {
	Description string            `yaml:"description"`
	Aliases     []string          `yaml:"aliases,omitempty"`
	Children    map[string]*Node  `yaml:"children,omitempty"`
}

// LoadSchema reads and parses a map.yaml file.
func LoadSchema(path string) (*Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var s Schema
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("validate %s: %w", path, err)
	}

	return &s, nil
}

// Validate checks that the schema is well-formed.
func (s *Schema) Validate() error {
	if s.Version < 1 {
		return fmt.Errorf("version must be >= 1, got %d", s.Version)
	}
	if len(s.Keys) == 0 {
		return fmt.Errorf("at least one key is required")
	}
	return validateNodes(s.Keys, "")
}

func validateNodes(nodes map[string]*Node, path string) error {
	for name, node := range nodes {
		currentPath := path + "/" + name
		if node == nil {
			return fmt.Errorf("key '%s' is empty", currentPath)
		}
		if node.Description == "" {
			return fmt.Errorf("key '%s' requires a description", currentPath)
		}
		if len(node.Children) > 0 {
			if err := validateNodes(node.Children, currentPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// Resolve validates that a sort-key exists in the schema.
func (s *Schema) Resolve(sortKey string) error {
	if s == nil {
		// No schema loaded — accept any sort-key
		return nil
	}
	// TODO: traverse key hierarchy to validate
	return nil
}
