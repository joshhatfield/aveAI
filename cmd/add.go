package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"aveAI/config"
	"aveAI/format"
	"aveAI/store"
)

var addCmd = &cobra.Command{
	Use:   "add <sort-key> <value> [flags]",
	Short: "Add a new entry to the database",
	Args:  cobra.ExactArgs(2),
	RunE:  runAdd,
}

var (
	addTags []string
)

func init() {
	addCmd.Flags().StringArrayVarP(&addTags, "tag", "t", nil, "tag to associate with entry")
}

func runAdd(cmd *cobra.Command, args []string) error {
	sortKey := args[0]
	value := args[1]

	storePath := GetDBPath()
	config.Info("adding entry", "path", storePath)

	s, err := loadStore(storePath)
	if err != nil {
		return err
	}

	var tags []string
	if len(addTags) > 0 {
		tags = addTags
	}

	id, err := s.Add(store.Entry{
		SortKey: sortKey,
		Value:   value,
		Tags:    tags,
	})
	if err != nil {
		return fmt.Errorf("add entry: %w", err)
	}

	if err := format.Save(storePath, s); err != nil {
		return fmt.Errorf("save .avdb: %w", err)
	}

	fmt.Printf("Added entry %d: %s → %s\n", id, sortKey, truncate(value, 50))
	return nil
}

func loadStore(path string) (*store.Store, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return store.New(), nil
	}
	return format.Load(path)
}