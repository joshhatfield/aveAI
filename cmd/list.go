package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"aveAI/format"
)

var listCmd = &cobra.Command{
	Use:   "list [sort-key]",
	Short: "List entries, optionally filtered by sort-key prefix",
	Args:  cobra.RangeArgs(0, 1),
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	var prefix string
	if len(args) > 0 {
		prefix = args[0]
	}

	s, err := format.Load(GetDBPath())
	if err != nil {
		return fmt.Errorf("load .avdb: %w", err)
	}

	entries := s.List(prefix)
	if len(entries) == 0 {
		fmt.Println("No entries found.")
		return nil
	}

	for _, e := range entries {
		fmt.Printf("%d: %s → %s", e.ID, e.SortKey, truncate(e.Value, 60))
		if len(e.Tags) > 0 {
			fmt.Printf(" [%s]", joinTags(e.Tags))
		}
		fmt.Println()
	}
	fmt.Printf("\n%d entries\n", len(entries))
	return nil
}