package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	dbPath  string
	mapPath string
)

var rootCmd = &cobra.Command{
	Use:   "ave",
	Short: "aveAI — local context store for AI agents",
	Long: `aveAI is a lightweight CLI tool for storing and retrieving
structured context data that AI agents can query.

It stores entries with hierarchical sort-keys in a portable .avdb file,
supporting text search, vector search, and tag-based filtering.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Resolve .avdb and map.yaml paths
		// This will be implemented in Phase 1
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "", "path to .avdb file (default: ./.ave/data.avdb)")
	rootCmd.PersistentFlags().StringVar(&mapPath, "map", "", "path to map.yaml (default: ./.ave/map.yaml)")
}
