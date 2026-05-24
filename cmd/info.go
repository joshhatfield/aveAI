package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"aveAI/format"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show database statistics",
	RunE:  runInfo,
}

func runInfo(cmd *cobra.Command, args []string) error {
	dbPath := GetDBPath()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Println("No database found. Run 'ave init' to create one.")
		return nil
	}

	s, err := format.Load(dbPath)
	if err != nil {
		return fmt.Errorf("load .avdb: %w", err)
	}

	stats := s.Stats()

	fmt.Println("aveAI Database Info")
	fmt.Println("====================")
	fmt.Printf("Entries:    %d\n", stats["entries"])
	fmt.Printf("Sort keys:  %d\n", stats["sort_keys"])
	fmt.Printf("Tags:       %d\n", stats["tags"])

	if summary, ok := stats["key_summary"].(map[string]int); ok {
		fmt.Println("\nSort-key breakdown:")
		for key, count := range summary {
			fmt.Printf("  %s: %d\n", key, count)
		}
	}

	return nil
}