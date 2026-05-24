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

	if len(proj.Notes) > 0 {
		sb.WriteString("## Notes\n")
		for _, n := range proj.Notes {
			sb.WriteString("- " + n + "\n")
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

	// Group entries by top-level sort-key
	byTop := make(map[string][]string)
	for _, e := range entries {
		parts := strings.SplitN(e.SortKey, "/", 2)
		top := parts[0]
		byTop[top] = append(byTop[top], e.Value)
	}

	// Load existing schema and update context
	sc, err := schema.LoadSchema(mapPath)
	if err != nil {
		return fmt.Errorf("load map.yaml: %w", err)
	}

	if sc.Context == nil {
		sc.Context = &schema.ContextBlock{Project: schema.ProjectContext{}}
	}

	// Merge into context
	for top, values := range byTop {
		switch top {
		case "code":
			if sc.Context.Project.Conventions == nil {
				sc.Context.Project.Conventions = []string{}
			}
			for _, v := range values {
				sc.Context.Project.Conventions = append(sc.Context.Project.Conventions, v)
			}
		case "notes":
			if sc.Context.Project.Notes == nil {
				sc.Context.Project.Notes = []string{}
			}
			for _, v := range values {
				sc.Context.Project.Notes = append(sc.Context.Project.Notes, v)
			}
		default:
			if sc.Context.Project.Notes == nil {
				sc.Context.Project.Notes = []string{}
			}
			for _, v := range values {
				sc.Context.Project.Notes = append(sc.Context.Project.Notes, v)
			}
		}
	}

	if err := sc.Save(mapPath); err != nil {
		return fmt.Errorf("save map.yaml: %w", err)
	}

	if outputFormat == "json" {
		return printJSON(map[string]any{
			"synced":   len(entries),
			"key_count": len(byTop),
			"saved":    mapPath,
		})
	}

	fmt.Printf("Synced %d entries into context (%d top-level keys)\n", len(entries), len(byTop))
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