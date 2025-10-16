package cli

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all keys in the database",
	Long: `List all keys currently stored in the database.
Keys are displayed in alphabetical order.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if err := initStorage(); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		defer Cleanup()

		keys, err := db.Keys()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if len(keys) == 0 {
			fmt.Println("No keys found in database")
			return
		}

		sort.Strings(keys)

		fmt.Printf("Found %d key(s):\n", len(keys))
		for i, key := range keys {
			fmt.Printf("%d. %s\n", i+1, key)
		}
	},
}
