package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <key>",
	Short: "Remove a key-value pair",
	Long: `Remove a key-value pair from the database.
Returns an error if the key does not exist.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		err := db.Delete(key)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Println("OK")
	},
}
