package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"aveAI/format"
	"aveAI/store"
)

var getCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get an entry by ID",
	Args:  cobra.ExactArgs(1),
	RunE:  runGet,
}

var getOut string

func init() {
	getCmd.Flags().StringVarP(&getOut, "output", "o", "text", "output format (text|json)")
}

func runGet(cmd *cobra.Command, args []string) error {
	id, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid id '%s': %w", args[0], err)
	}

	s, err := format.Load(GetDBPath())
	if err != nil {
		return fmt.Errorf("load .avdb: %w", err)
	}

	entry, err := s.Get(id)
	if err != nil {
		return err
	}

	if getOut == "json" {
		return printJSON(entry)
	}

	printEntry(entry)
	return nil
}

func printEntry(e *store.Entry) {
	fmt.Printf("ID:      %d\n", e.ID)
	fmt.Printf("SortKey: %s\n", e.SortKey)
	fmt.Printf("Value:   %s\n", e.Value)
	if len(e.Tags) > 0 {
		fmt.Printf("Tags:    %v\n", e.Tags)
	}
	fmt.Printf("Created: %d\n", e.Created)
	fmt.Printf("Updated: %d\n", e.Updated)
}

func printEntryJSON(e *store.Entry) error {
	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal entry: %w", err)
	}
	fmt.Println(string(data))
	return nil
}