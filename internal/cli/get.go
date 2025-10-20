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
		if err := initStorage(); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		defer Cleanup()

		key := args[0]

		var value []byte
		var err error

		if currentTransaction != nil {
			// If we're in a transaction, use the transaction's Get method
			value, err = currentTransaction.Get(key)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Printf("Value: %s (Transaction: %s)\n", string(value), currentTransaction.ID)
		} else {
			// Direct operation
			value, err = db.Get(key)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Printf("Value: %s\n", string(value))
		}
	},
}
