package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"aveAI/format"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete an entry by ID",
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

func runDelete(cmd *cobra.Command, args []string) error {
	id, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid id '%s': %w", args[0], err)
	}

	dbPath := GetDBPath()
	s, err := format.Load(dbPath)
	if err != nil {
		return fmt.Errorf("load .avdb: %w", err)
	}

	// Verify entry exists before deleting
	if _, err := s.Get(id); err != nil {
		return err
	}

	if err := s.Delete(id); err != nil {
		return fmt.Errorf("delete entry: %w", err)
	}

	if err := format.Save(dbPath, s); err != nil {
		return fmt.Errorf("save .avdb: %w", err)
	}

	fmt.Printf("Deleted entry %d\n", id)
	return nil
}