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

var infoOut string

func init() {
	infoCmd.Flags().StringVarP(&infoOut, "output", "o", "text", "output format (text|json)")
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

	if infoOut == "json" {
		return printJSON(stats)
	}

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