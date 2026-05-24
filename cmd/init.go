package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"aveAI/config"
	"aveAI/format"
	"aveAI/store"
)

const defaultMapYAML = `version: 1
keys:
  code:
    description: "Code conventions and patterns"
    children:
      conventions:
        description: "Stylistic and structural conventions"
      patterns:
        description: "Design patterns used in the project"
      errors:
        description: "Error handling patterns"
  notes:
    description: "General project notes"
    children:
      decisions:
        description: "Architectural decision records"
      setup:
        description: "Setup and configuration notes"
  tools:
    description: "Tool usage patterns"
`

const defaultConfigYAML = `version: 1
db:
  path: ".ave/data.avdb"
map:
  path: ".ave/map.yaml"
logging:
  level: "info"
  format: "text"
search:
  default_limit: 10
`

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a new .avdb database",
	Args:  cobra.RangeArgs(0, 1),
	RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	basePath := ".ave"
	if len(args) > 0 {
		basePath = args[0]
	}

	// Ensure directory exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	dbPath := filepath.Join(basePath, "data.avdb")
	mapPath := filepath.Join(basePath, "map.yaml")
	configPath := filepath.Join(basePath, "config.yaml")

	// Write default config.yaml if not present
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := os.WriteFile(configPath, []byte(defaultConfigYAML), 0644); err != nil {
			return fmt.Errorf("write config.yaml: %w", err)
		}
		fmt.Printf("Created %s\n", configPath)
	}

	// Write default map.yaml if not present
	if _, err := os.Stat(mapPath); os.IsNotExist(err) {
		if err := os.WriteFile(mapPath, []byte(defaultMapYAML), 0644); err != nil {
			return fmt.Errorf("write map.yaml: %w", err)
		}
		fmt.Printf("Created %s\n", mapPath)
	}

	// Create empty .avdb
	s := store.New()
	if err := format.Save(dbPath, s); err != nil {
		return fmt.Errorf("save .avdb: %w", err)
	}
	fmt.Printf("Created %s\n", dbPath)

	// Reload config to pick up the new config file
	config.ResetGlobal()

	return nil
}