package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Store a key-value pair",
	Long: `Store a key-value pair in the database.
The value can contain spaces and will be stored as provided.`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initStorage(); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		defer Cleanup()

		key := args[0]
		value := args[1]
		
		if len(args) > 2 {
			value = ""
			for i, arg := range args[1:] {
				if i > 0 {
					value += " "
				}
				value += arg
			}
		}

		if currentTransaction != nil {
			// If we're in a transaction, use the transaction's Put method
			err := currentTransaction.Put(key, []byte(value))
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Printf("OK (Transaction: %s)\n", currentTransaction.ID)
		} else {
			// Direct operation
			err := db.Put(key, []byte(value))
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Println("OK")
		}
	},
}
