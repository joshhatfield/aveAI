package schema

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Schema represents a parsed map.yaml file.
type Schema struct {
	Version int              `yaml:"version"`
	Context *ContextBlock    `yaml:"context,omitempty"`
	Keys    map[string]*Node `yaml:"keys"`
}

// ContextBlock holds the pseudocontext section of the map.yaml.
type ContextBlock struct {
	Project  ProjectContext `yaml:"project,omitempty"`
	Commands CommandHelp    `yaml:"commands,omitempty"`
}

// ProjectContext holds project-level metadata and conventions.
type ProjectContext struct {
	Name          string          `yaml:"name,omitempty"`
	Description  string          `yaml:"description,omitempty"`
	Conventions  []string        `yaml:"conventions,omitempty"`
	Patterns     []string        `yaml:"patterns,omitempty"`
	Notes        NotesHierarchy  `yaml:"notes,omitempty"`
}

// NotesHierarchy represents arbitrary nested YAML for context index.
type NotesHierarchy map[string]any

// CommandHelp holds usage examples for each command.
type CommandHelp struct {
	Add    CommandSpec `yaml:"add,omitempty"`
	Search CommandSpec `yaml:"search,omitempty"`
	List   CommandSpec `yaml:"list,omitempty"`
	Get    CommandSpec `yaml:"get,omitempty"`
	Init   CommandSpec `yaml:"init,omitempty"`
	Info   CommandSpec `yaml:"info,omitempty"`
	Delete CommandSpec `yaml:"delete,omitempty"`
	Context CommandSpec `yaml:"context,omitempty"`
}

// CommandSpec holds usage and examples for a command.
type CommandSpec struct {
	Usage    string   `yaml:"usage,omitempty"`
	Examples []string `yaml:"examples,omitempty"`
}

// Node represents a single key or namespace in the key hierarchy.
type Node struct {
	Description string           `yaml:"description"`
	Aliases     []string         `yaml:"aliases,omitempty"`
	Children    map[string]*Node `yaml:"children,omitempty"`
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

// Save writes the schema back to a map.yaml file.
func (s *Schema) Save(path string) error {
	data, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshal schema: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
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
		return nil
	}
	parts := strings.Split(sortKey, "/")
	current := s.Keys
	for i, part := range parts {
		node, ok := current[part]
		if !ok {
			return fmt.Errorf("sort-key '%s' not found in schema", sortKey)
		}
		if i < len(parts)-1 {
			if node.Children == nil {
				return fmt.Errorf("sort-key '%s' has no children at '%s'", sortKey, part)
			}
			current = node.Children
		}
	}
	return nil
}

// HasContext returns true if the schema has a context block.
func (s *Schema) HasContext() bool {
	if s.Context == nil {
		return false
	}
	// Check if there's any meaningful context content
	if s.Context.Project.Name != "" {
		return true
	}
	if len(s.Context.Project.Conventions) > 0 {
		return true
	}
	if len(s.Context.Project.Patterns) > 0 {
		return true
	}
	// Check notes hierarchy (NotesHierarchy is map[string]any)
	if s.Context.Project.Notes != nil && len(s.Context.Project.Notes) > 0 {
		return true
	}
	return false
}

// GetContextDescription returns the project's description if set.
func (s *Schema) GetContextDescription() string {
	if s.Context == nil {
		return ""
	}
	return s.Context.Project.Description
}

// AppendConvention appends a convention to the project's conventions list.
func (s *Schema) AppendConvention(convention string) {
	if s.Context == nil {
		s.Context = &ContextBlock{Project: ProjectContext{}}
	}
	if s.Context.Project.Conventions == nil {
		s.Context.Project.Conventions = []string{}
	}
	s.Context.Project.Conventions = append(s.Context.Project.Conventions, convention)
}

// AppendPattern appends a pattern to the project's patterns list.
func (s *Schema) AppendPattern(pattern string) {
	if s.Context == nil {
		s.Context = &ContextBlock{Project: ProjectContext{}}
	}
	if s.Context.Project.Patterns == nil {
		s.Context.Project.Patterns = []string{}
	}
	s.Context.Project.Patterns = append(s.Context.Project.Patterns, pattern)
}

// AppendNote appends a note to the project's notes list.
// DEPRECATED: Notes is now a hierarchy (map[string]any), not []string.
// Use context sync to rebuild the notes hierarchy from sort-keys.
func (s *Schema) AppendNote(note string) {
	// Deprecated: Notes is now a nested hierarchy, not a flat list.
	// This function is kept for compatibility but does nothing.
}

func (s *Schema) SetProjectName(name string) {
	if s.Context == nil {
		s.Context = &ContextBlock{Project: ProjectContext{}}
	}
	s.Context.Project.Name = name
}

// SetProjectDescription sets the project's description.
func (s *Schema) SetProjectDescription(desc string) {
	if s.Context == nil {
		s.Context = &ContextBlock{Project: ProjectContext{}}
	}
	s.Context.Project.Description = desc
}

// TruncateDepth returns a new NotesHierarchy truncated to maxDepth levels.
// Depth 0 = no truncation, depth 1 = top-level only, depth 2 = top + one level, etc.
func (h NotesHierarchy) TruncateDepth(maxDepth int) NotesHierarchy {
	if maxDepth <= 0 {
		return h
	}
	return truncateDepthRecursive(h, 0, maxDepth)
}

func truncateDepthRecursive(h map[string]any, currentDepth, maxDepth int) map[string]any {
	result := make(map[string]any)
	for k, v := range h {
		if nested, ok := v.(map[string]any); ok && currentDepth < maxDepth {
			result[k] = truncateDepthRecursive(nested, currentDepth+1, maxDepth)
		} else {
			// At max depth or leaf node - truncate to empty marker
			result[k] = map[string]any{}
		}
	}
	return result
}

// CountLeaves returns total number of leaf entries in the hierarchy.
func (h NotesHierarchy) CountLeaves() int {
	return countLeavesRecursive(h)
}

func countLeavesRecursive(h map[string]any) int {
	count := 0
	for _, v := range h {
		if nested, ok := v.(map[string]any); ok {
			count += countLeavesRecursive(nested)
		} else {
			count++
		}
	}
	return count
}

// WithCounts returns a new hierarchy where each branch has a {count: N} annotation.
func (h NotesHierarchy) WithCounts() NotesHierarchy {
	return withCountsRecursive(h)
}

func withCountsRecursive(h map[string]any) map[string]any {
	result := make(map[string]any)
	for k, v := range h {
		if nested, ok := v.(map[string]any); ok {
			count := countLeavesRecursive(nested)
			result[k] = map[string]any{"count": count}
		} else {
			result[k] = v
		}
	}
	return result
}

// TruncateKeys returns a new hierarchy with at most limit keys per branch.
// If limit > 0 and there are more keys, the remainder is indicated with {remaining: N}.
func (h NotesHierarchy) TruncateKeys(limit int) NotesHierarchy {
	if limit <= 0 {
		return h
	}
	return truncateKeysRecursive(h, limit)
}

func truncateKeysRecursive(h map[string]any, limit int) map[string]any {
	result := make(map[string]any)
	keys := make([]string, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	for i, k := range keys {
		if i >= limit {
			continue
		}
		v := h[k]
		if nested, ok := v.(map[string]any); ok {
			result[k] = truncateKeysRecursive(nested, limit)
		} else {
			result[k] = v
		}
	}
	if len(keys) > limit {
		result["..."] = map[string]any{"remaining": len(keys) - limit}
	}
	return result
}