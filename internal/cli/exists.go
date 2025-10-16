package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var existsCmd = &cobra.Command{
	Use:   "exists <key>",
	Short: "Check if a key exists",
	Long: `Check if a key exists in the database.
Returns true if the key exists, false otherwise.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		exists, err := db.Exists(key)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("Exists: %t\n", exists)
	},
}
