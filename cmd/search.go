package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"aveAI/format"
	"aveAI/search"
	"aveAI/store"
)

var searchCmd = &cobra.Command{
	Use:   "search <query> [flags]",
	Short: "Search entries by text query",
	Args:  cobra.ExactArgs(1),
	RunE:  runSearch,
}

var (
	limitCount    int
	filterSortKey string
	filterTag     string
)

func init() {
	searchCmd.Flags().IntVarP(&limitCount, "limit", "l", 0, "limit number of results (0 = all)")
	searchCmd.Flags().StringVarP(&filterSortKey, "sort-key", "s", "", "filter by sort-key prefix")
	searchCmd.Flags().StringVarP(&filterTag, "tag", "t", "", "filter by tag")
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := args[0]

	if query == "" {
		return fmt.Errorf("query cannot be empty")
	}

	dbPath := GetDBPath()
	s, err := format.Load(dbPath)
	if err != nil {
		return fmt.Errorf("load .avdb: %w", err)
	}

	entries := s.All()

	// Apply pre-search filters if specified
	if filterSortKey != "" {
		filtered := make([]store.Entry, 0)
		for _, e := range entries {
			if len(e.SortKey) >= len(filterSortKey) && e.SortKey[:len(filterSortKey)] == filterSortKey {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}

	if filterTag != "" {
		filtered := make([]store.Entry, 0)
		for _, e := range entries {
			for _, tag := range e.Tags {
				if tag == filterTag {
					filtered = append(filtered, e)
					break
				}
			}
		}
		entries = filtered
	}

	// Build inverted index and search
	idx := search.NewInvertedIndex()
	idx.Build(entries)
	results := idx.Search(query)

	if len(results) == 0 {
		fmt.Println("No results found.")
		return nil
	}

	// Apply limit
	if limitCount > 0 && len(results) > limitCount {
		results = results[:limitCount]
	}

	// Print results
	for _, r := range results {
		entry, err := s.Get(r.EntryID)
		if err != nil {
			continue
		}
		fmt.Printf("%d: %s → %s", entry.ID, entry.SortKey, truncate(entry.Value, 60))
		if len(entry.Tags) > 0 {
			fmt.Printf(" [%s]", joinTags(entry.Tags))
		}
		fmt.Printf(" (%.2f)\n", r.Score)
	}

	fmt.Printf("\n%d results\n", len(results))
	return nil
}