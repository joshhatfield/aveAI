package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"aveAI/config"
	"aveAI/schema"
)

// truncate shortens a string to max length, adding "..." if truncated.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// joinTags concatenates tags into a comma-separated string.
func joinTags(tags []string) string {
	result := ""
	for i, t := range tags {
		if i > 0 {
			result += ", "
		}
		result += t
	}
	return result
}

// autoGrowSchema adds missing keys to map.yaml when new sort-keys are used.
// This keeps the schema in sync with actual data without manual context sync.
func autoGrowSchema(mapPath, sortKey string) error {
	s, err := schema.LoadSchema(mapPath)
	if err != nil {
		return fmt.Errorf("load schema: %w", err)
	}

	parts := strings.Split(sortKey, "/")
	if len(parts) < 1 {
		return nil
	}

	top := parts[0]
	if s.Keys == nil {
		s.Keys = make(map[string]*schema.Node)
	}

	// Ensure top-level key exists
	if _, ok := s.Keys[top]; !ok {
		s.Keys[top] = &schema.Node{
			Description: defaultDescription(top),
		}
	}

	// Ensure nested keys exist
	current := s.Keys[top]
	for i := 1; i < len(parts); i++ {
		part := parts[i]
		if current.Children == nil {
			current.Children = make(map[string]*schema.Node)
		}
		if _, ok := current.Children[part]; !ok {
			current.Children[part] = &schema.Node{
				Description: defaultDescription(part),
			}
		}
		current = current.Children[part]
	}

	return s.Save(mapPath)
}

// defaultDescription generates a description from a key name.
func defaultDescription(key string) string {
	desc := strings.ToLower(key)
	if len(desc) > 0 {
		desc = strings.ToUpper(string(desc[0])) + desc[1:]
	}
	return desc + " patterns"
}

// getContextMapPath derives map.yaml path from db path when possible.
// Falls back to config resolution if derived path doesn't exist.
func getContextMapPath() string {
	dbPath := GetDBPath()
	dir := filepath.Dir(dbPath)
	base := filepath.Base(dbPath)
	mapBase := strings.Replace(base, "data.avdb", "map.yaml", 1)
	mapPath := filepath.Join(dir, mapBase)

	// If derived map exists alongside db, use it
	if _, err := os.Stat(mapPath); err == nil {
		return mapPath
	}

	// Fallback to config resolution
	return config.Global().ResolveMapPath("")
}