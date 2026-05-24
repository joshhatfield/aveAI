package cmd

import (
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