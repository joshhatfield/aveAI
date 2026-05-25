package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"aveAI/format"
	"aveAI/schema"
	"aveAI/store"
)

var contextCmd = &cobra.Command{
	Use:   "context <subcommand>",
	Short: "Manage project context (pseudocontext)",
}

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Output pseudocontext for LLM seeding",
	RunE:  runContextPull,
}

var addContextCmd = &cobra.Command{
	Use:   "add <sort-key> <value>",
	Short: "Add entry to database",
	Args:  cobra.ExactArgs(2),
	RunE:  runContextAdd,
}

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "DEPRECATED: Open map.yaml in $EDITOR (use 'ave search' instead)",
	RunE:  runContextEdit,
}

var (
	outputFormat string
	pullDepth    int
	pullCounts   bool
	pullLimit    int
	pullSummary  bool
)

func init() {
	contextCmd.AddCommand(pullCmd)
	contextCmd.AddCommand(addContextCmd)
	contextCmd.AddCommand(editCmd)

	for _, cmd := range []*cobra.Command{pullCmd, addContextCmd, editCmd} {
		cmd.Flags().StringVar(&outputFormat, "output", "text", "output format (text|json)")
	}

	// Context pull flags
	pullCmd.Flags().IntVar(&pullDepth, "depth", 0, "max hierarchy depth (0=unlimited)")
	pullCmd.Flags().BoolVar(&pullCounts, "counts", false, "show item counts")
	pullCmd.Flags().IntVar(&pullLimit, "limit", 0, "max keys per branch")
	pullCmd.Flags().BoolVar(&pullSummary, "summary", false, "summary mode (categories with counts)")
}

func runContextPull(cmd *cobra.Command, args []string) error {
	dbPath := GetDBPath()

	// Load entries from database
	s, err := format.Load(dbPath)
	if err != nil {
		return fmt.Errorf("load .avdb: %w", err)
	}

	entries := s.All()
	if len(entries) == 0 {
		if outputFormat == "json" {
			return printJSON(map[string]any{
				"error": "no entries found in database",
				"hint":  "run 'ave add' to add entries first",
			})
		}
		fmt.Println("No entries found in database. Run 'ave add' to add entries first.")
		return nil
	}

	// Build hierarchy from DB entries
	hierarchy := buildKeyHierarchy(entries)
	notes := schema.NotesHierarchy(hierarchy)

	// Apply context optimization options in correct order:
	// 1. TruncateDepth (before WithCounts corrupts structure)
	// 2. TruncateKeys
	// 3. WithCounts (for counts/summary)
	if pullDepth > 0 {
		notes = notes.TruncateDepth(pullDepth)
	}
	if pullLimit > 0 {
		notes = notes.TruncateKeys(pullLimit)
	}
	if pullCounts || pullSummary {
		notes = notes.WithCounts()
	}

	// Apply context optimization options
	if pullSummary {
		// Summary mode: just show top-level categories with counts
		type summaryEntry struct {
			Category string `json:"category"`
			Count    int    `json:"count"`
		}
		var summaryList []summaryEntry
		for key, value := range notes {
			// value is WithCounts result: {count: N, children: {...}} or {leaf: true, count: N}
			if m, ok := value.(map[string]any); ok {
				if count, ok := m["count"].(int); ok {
					summaryList = append(summaryList, summaryEntry{Category: key, Count: count})
					continue
				}
			}
			summaryList = append(summaryList, summaryEntry{Category: key, Count: 1})
		}

		if outputFormat == "json" {
			return printJSON(map[string]any{
				"summary":    summaryList,
				"total_keys": len(entries),
			})
		}
	}

	if outputFormat == "json" {
		return printJSON(map[string]any{
			"notes": notes,
			"total_keys": len(entries),
		})
	}

	// Markdown output
	var sb strings.Builder
	sb.WriteString("# Database Index\n\n")
	sb.WriteString(fmt.Sprintf("Total entries: %d\n\n", len(entries)))

	if pullSummary {
		sb.WriteString("## Categories (summary)\n")
		for key, value := range notes {
			// value is WithCounts result: {count: N, children: {...}} or {leaf: true, count: N}
			if m, ok := value.(map[string]any); ok {
				if count, ok := m["count"].(int); ok {
					sb.WriteString(fmt.Sprintf("- %s: {count: %d}\n", key, count))
					continue
				}
			}
			sb.WriteString(fmt.Sprintf("- %s\n", key))
		}
	} else {
		sb.WriteString("## Notes\n")
		for key, value := range notes {
			printHierarchy(&sb, key, value, 0)
		}
	}

	fmt.Print(sb.String())
	return nil
}

func runContextAdd(cmd *cobra.Command, args []string) error {
	sortKey := args[0]
	value := args[1]
	dbPath := GetDBPath()

	// Load or create store
	s, err := loadStore(dbPath)
	if err != nil {
		return err
	}

	// Add entry to .avdb
	id, err := s.Add(store.Entry{
		SortKey: sortKey,
		Value:   value,
	})
	if err != nil {
		return fmt.Errorf("add entry: %w", err)
	}

	if err := format.Save(dbPath, s); err != nil {
		return fmt.Errorf("save .avdb: %w", err)
	}

	if outputFormat == "json" {
		return printJSON(map[string]any{
			"id":      id,
			"sortKey": sortKey,
			"value":   value,
		})
	}

	fmt.Printf("Added entry %d: %s → %s\n", id, sortKey, truncate(value, 50))
	return nil
}

func runContextEdit(cmd *cobra.Command, args []string) error {
	fmt.Println("WARNING: context edit is deprecated. Use 'ave search' to find entries or 'ave add' to add new ones.")
	fmt.Println()

	mapPath := getContextMapPath()
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim" // fallback
	}

	editCmd := exec.Command(editor, mapPath)
	editCmd.Stdin = os.Stdin
	editCmd.Stdout = os.Stdout
	editCmd.Stderr = os.Stderr

	if err := editCmd.Run(); err != nil {
		return fmt.Errorf("editor: %w", err)
	}

	// Validate after edit
	s, err := schema.LoadSchema(mapPath)
	if err != nil {
		return fmt.Errorf("map.yaml after edit is invalid: %w", err)
	}

	if outputFormat == "json" {
		return printJSON(map[string]any{"status": "saved", "path": mapPath})
	}

	fmt.Printf("Saved %s\n", mapPath)
	_ = s // suppress unused warning
	return nil
}

// buildKeyHierarchy builds a nested map from sort-key paths.
// e.g., "items/cards/card1" → {"items": {"cards": {"card1": {}}}}
func buildKeyHierarchy(entries []store.Entry) map[string]any {
	hierarchy := make(map[string]any)

	for _, e := range entries {
		parts := strings.Split(e.SortKey, "/")
		current := hierarchy
		for i, part := range parts {
			if i == len(parts)-1 {
				// Leaf: use empty struct or true
				current[part] = map[string]any{}
			} else {
				// Branch: create nested map if not exists
				if _, ok := current[part]; !ok {
					current[part] = make(map[string]any)
				}
			}
			current = current[part].(map[string]any)
		}
	}

	return hierarchy
}

// mapToNotes converts a nested map hierarchy to dot-notation strings for context.
func mapToNotes(hierarchy map[string]any, prefix string) []string {
	var notes []string

	for key, value := range hierarchy {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}
		notes = append(notes, fullKey)
		if nested, ok := value.(map[string]any); ok {
			notes = append(notes, mapToNotes(nested, fullKey)...)
		}
	}

	return notes
}

// printHierarchy recursively prints a hierarchy map as indented markdown list.
func printHierarchy(sb *strings.Builder, key string, value any, indent int) {
	prefix := strings.Repeat("  ", indent)

	// Check if this is a WithCounts node (has count and children keys)
	if m, ok := value.(map[string]any); ok {
		if count, hasCount := m["count"].(int); hasCount {
			// WithCounts structure
			if _, isLeaf := m["leaf"]; isLeaf {
				// Leaf node - just show the count
				sb.WriteString(fmt.Sprintf("%s- %s: {count: %d}\n", prefix, key, count))
				return
			}
			// Branch node - show count, then children
			sb.WriteString(fmt.Sprintf("%s- %s: {count: %d}\n", prefix, key, count))
			if children, ok := m["children"].(map[string]any); ok {
				for k, v := range children {
					printHierarchy(sb, k, v, indent+1)
				}
			}
			return
		}
	}

	sb.WriteString(prefix + "- " + key + "\n")
	// Handle nested hierarchies (schema.NotesHierarchy is map[string]any)
	switch v := value.(type) {
	case schema.NotesHierarchy:
		if len(v) > 0 {
			for k, val := range v {
				printHierarchy(sb, k, val, indent+1)
			}
		}
	case map[string]any:
		if len(v) > 0 {
			for k, val := range v {
				printHierarchy(sb, k, val, indent+1)
			}
		}
	}
}

func printJSON(v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	fmt.Println(string(data))
	return nil
}