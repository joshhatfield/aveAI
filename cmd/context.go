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

var updateCmd = &cobra.Command{
	Use:   "update <key> <value>",
	Short: "Update a context value (e.g., project.conventions)",
	Args:  cobra.ExactArgs(2),
	RunE:  runContextUpdate,
}

var addContextCmd = &cobra.Command{
	Use:   "add <sort-key> <value>",
	Short: "Add entry AND append to pseudocontext",
	Args:  cobra.ExactArgs(2),
	RunE:  runContextAdd,
}

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open map.yaml in $EDITOR",
	RunE:  runContextEdit,
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Regenerate context from all .avdb entries",
	RunE:  runContextSync,
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
	contextCmd.AddCommand(updateCmd)
	contextCmd.AddCommand(addContextCmd)
	contextCmd.AddCommand(editCmd)
	contextCmd.AddCommand(syncCmd)

	for _, cmd := range []*cobra.Command{pullCmd, updateCmd, addContextCmd, editCmd, syncCmd} {
		cmd.Flags().StringVar(&outputFormat, "output", "text", "output format (text|json)")
	}

	// Context pull flags
	pullCmd.Flags().IntVar(&pullDepth, "depth", 0, "max hierarchy depth (0=unlimited)")
	pullCmd.Flags().BoolVar(&pullCounts, "counts", false, "show item counts")
	pullCmd.Flags().IntVar(&pullLimit, "limit", 0, "max keys per branch")
	pullCmd.Flags().BoolVar(&pullSummary, "summary", false, "summary mode (categories with counts)")
}

func runContextPull(cmd *cobra.Command, args []string) error {
	mapPath := getContextMapPath()

	s, err := schema.LoadSchema(mapPath)
	if err != nil {
		return fmt.Errorf("load map.yaml: %w", err)
	}

	if s.Context == nil || !s.HasContext() {
		if outputFormat == "json" {
			return printJSON(map[string]any{
				"error": "no context found in map.yaml",
				"hint":  "run 'ave context add' to add context or edit map.yaml directly",
			})
		}
		fmt.Println("No context found in map.yaml. Run 'ave context add' or 'ave context edit' to add context.")
		return nil
	}

	if outputFormat == "json" {
		return printJSON(s.Context)
	}

	// Markdown output
	var sb strings.Builder
	proj := s.Context.Project

	sb.WriteString("# Project Context")
	if proj.Name != "" {
		sb.WriteString(" — " + proj.Name)
	}
	sb.WriteString("\n\n")

	if proj.Description != "" {
		sb.WriteString(proj.Description + "\n\n")
	}

	if len(proj.Conventions) > 0 {
		sb.WriteString("## Conventions\n")
		for _, c := range proj.Conventions {
			sb.WriteString("- " + c + "\n")
		}
		sb.WriteString("\n")
	}

	if len(proj.Patterns) > 0 {
		sb.WriteString("## Patterns\n")
		for _, p := range proj.Patterns {
			sb.WriteString("- " + p + "\n")
		}
		sb.WriteString("\n")
	}

	if proj.Notes != nil && len(proj.Notes) > 0 {
		// Apply context optimization options
		notes := schema.NotesHierarchy(proj.Notes)

		if pullSummary {
			// Summary mode: just show top-level categories with counts
			sb.WriteString("## Notes (summary)\n")
			for key, value := range notes {
				if nested, ok := value.(map[string]any); ok {
					count := schema.NotesHierarchy(nested).CountLeaves()
					sb.WriteString(fmt.Sprintf("- %s: {count: %d}\n", key, count))
				} else {
					sb.WriteString(fmt.Sprintf("- %s\n", key))
				}
			}
		} else {
			// Normal mode with options
			if pullDepth > 0 {
				notes = notes.TruncateDepth(pullDepth)
			}
			if pullLimit > 0 {
				notes = notes.TruncateKeys(pullLimit)
			}
			if pullCounts {
				notes = notes.WithCounts()
			}

			sb.WriteString("## Notes\n")
			for key, value := range notes {
				printHierarchy(&sb, key, value, 0)
			}
		}
		sb.WriteString("\n")
	}

	if s.Context.Commands.Add.Usage != "" {
		sb.WriteString("## Command Usage\n")
		if s.Context.Commands.Add.Usage != "" {
			sb.WriteString("### add\n")
			sb.WriteString("`" + s.Context.Commands.Add.Usage + "`\n")
			for _, ex := range s.Context.Commands.Add.Examples {
				sb.WriteString("- " + ex + "\n")
			}
			sb.WriteString("\n")
		}
		if s.Context.Commands.Search.Usage != "" {
			sb.WriteString("### search\n")
			sb.WriteString("`" + s.Context.Commands.Search.Usage + "`\n")
			for _, ex := range s.Context.Commands.Search.Examples {
				sb.WriteString("- " + ex + "\n")
			}
			sb.WriteString("\n")
		}
	}

	fmt.Print(sb.String())
	return nil
}

func runContextUpdate(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]
	mapPath := getContextMapPath()

	s, err := schema.LoadSchema(mapPath)
	if err != nil {
		return fmt.Errorf("load map.yaml: %w", err)
	}

	if s.Context == nil {
		s.Context = &schema.ContextBlock{Project: schema.ProjectContext{}}
	}

	// Parse dot-notation key: project.conventions, project.name, etc.
	parts := strings.SplitN(key, ".", 2)
	if len(parts) != 2 {
		if outputFormat == "json" {
			return printJSON(map[string]any{"error": "key must be in dot notation (e.g., project.conventions)"})
		}
		return fmt.Errorf("key must be in dot notation (e.g., project.conventions)")
	}

	section := parts[0]
	field := parts[1]

	switch section {
	case "project":
		switch field {
		case "name":
			s.SetProjectName(value)
		case "description":
			s.SetProjectDescription(value)
		case "conventions":
			s.AppendConvention(value)
		case "patterns":
			s.AppendPattern(value)
		case "notes":
			s.AppendNote(value)
		default:
			if outputFormat == "json" {
				return printJSON(map[string]any{"error": "unknown project field: " + field})
			}
			return fmt.Errorf("unknown project field: %s (known: name, description, conventions, patterns, notes)", field)
		}
	default:
		if outputFormat == "json" {
			return printJSON(map[string]any{"error": "unknown section: " + section})
		}
		return fmt.Errorf("unknown section: %s (known: project)", section)
	}

	if err := s.Save(mapPath); err != nil {
		return fmt.Errorf("save map.yaml: %w", err)
	}

	if outputFormat == "json" {
		return printJSON(map[string]any{
			"key":   key,
			"value": value,
			"saved": mapPath,
		})
	}

	fmt.Printf("Updated %s = %q in %s\n", key, value, mapPath)
	return nil
}

func runContextAdd(cmd *cobra.Command, args []string) error {
	sortKey := args[0]
	value := args[1]
	dbPath := GetDBPath()
	mapPath := getContextMapPath()

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

	// Also append to pseudocontext in map.yaml
	sc, err := schema.LoadSchema(mapPath)
	if err == nil && sc != nil {
		// Infer what to append based on sort-key prefix
		if strings.HasPrefix(sortKey, "code/conventions") || strings.HasPrefix(sortKey, "code/patterns") || strings.HasPrefix(sortKey, "code/") {
			sc.AppendConvention(value)
		} else if strings.HasPrefix(sortKey, "notes/") {
			sc.AppendNote(value)
		} else {
			// Generic append to notes
			sc.AppendNote(value)
		}
		sc.Save(mapPath)
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

func runContextSync(cmd *cobra.Command, args []string) error {
	dbPath := GetDBPath()
	mapPath := getContextMapPath()

	s, err := format.Load(dbPath)
	if err != nil {
		return fmt.Errorf("load .avdb: %w", err)
	}

	entries := s.All()
	if len(entries) == 0 {
		if outputFormat == "json" {
			return printJSON(map[string]any{"message": "no entries to sync"})
		}
		fmt.Println("No entries to sync.")
		return nil
	}

	// Build hierarchy from entries: e.g., "items/cards/card1" → nested map
	hierarchy := buildKeyHierarchy(entries)

	// Load existing schema and update context
	sc, err := schema.LoadSchema(mapPath)
	if err != nil {
		return fmt.Errorf("load map.yaml: %w", err)
	}

	if sc.Context == nil {
		sc.Context = &schema.ContextBlock{Project: schema.ProjectContext{}}
	}

	// Clear and rebuild
	sc.Context.Project.Conventions = []string{}
	sc.Context.Project.Patterns = []string{}
	sc.Context.Project.Notes = schema.NotesHierarchy{}

	// Add hierarchy directly (becomes nested YAML)
	for k, v := range hierarchy {
		sc.Context.Project.Notes[k] = v
	}

	// Also update the schema keys section with actual entries
	for _, e := range entries {
		autoGrowSchema(mapPath, e.SortKey)
	}

	if err := sc.Save(mapPath); err != nil {
		return fmt.Errorf("save map.yaml: %w", err)
	}

	if outputFormat == "json" {
		return printJSON(map[string]any{
			"synced_entries": len(entries),
			"hierarchy":      hierarchy,
			"saved":          mapPath,
		})
	}

	fmt.Printf("Synced %d entries into context\n", len(entries))
	return nil
}

func printJSON(v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	fmt.Println(string(data))
	return nil
}