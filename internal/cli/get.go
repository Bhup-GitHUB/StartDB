package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Retrieve a value by key",
	Long: `Retrieve and display the value associated with the given key.
Returns an error if the key does not exist.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		value, err := db.Get(key)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("Value: %s\n", string(value))
	},
}
