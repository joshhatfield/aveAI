package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"aveAI/config"
)

var (
	dbPath string
)

var rootCmd = &cobra.Command{
	Use:   "ave",
	Short: "aveAI — local context store for AI agents",
	Long: `aveAI is a lightweight CLI tool for storing and retrieving
structured context data that AI agents can query.

It stores entries with hierarchical sort-keys in a portable .avdb file,
supporting text search, vector search, and tag-based filtering.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load config and set log level
		cfg := config.Global()
		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("config validation: %w", err)
		}
		if err := config.SetLevel(cfg.Logging.Level); err != nil {
			return fmt.Errorf("set log level: %w", err)
		}
		config.Debug("config loaded", "db_path", cfg.DB.Path, "log_level", cfg.Logging.Level)
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
	rootCmd.PersistentFlags().StringVarP(&dbPath, "db", "d", "", "path to .avdb file (overrides config)")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(contextCmd)
}

// GetDBPath returns the effective .avdb path using config precedence.
func GetDBPath() string {
	return config.Global().ResolveDBPath(dbPath)
}

// GetMapPath returns the effective map.yaml path using config precedence.
func GetMapPath() string {
	return config.Global().ResolveMapPath("")
}