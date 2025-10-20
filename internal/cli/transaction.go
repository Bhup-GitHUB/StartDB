package cli

import (
	"fmt"
	"os"

	"startdb/internal/storage"

	"github.com/spf13/cobra"
)

var (
	currentTransaction *storage.Transaction
)

var beginCmd = &cobra.Command{
	Use:   "begin",
	Short: "Begin a new transaction",
	Long:  `Begin a new database transaction. All subsequent operations will be part of this transaction until commit or rollback.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := initStorage(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing storage: %v\n", err)
			os.Exit(1)
		}

		if currentTransaction != nil {
			fmt.Fprintf(os.Stderr, "Error: Transaction already in progress. Use 'commit' or 'rollback' first.\n")
			os.Exit(1)
		}

		currentTransaction = db.BeginTransaction()
		fmt.Printf("Transaction %s started\n", currentTransaction.ID)
	},
}

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Commit the current transaction",
	Long:  `Commit the current transaction, making all changes permanent.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := initStorage(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing storage: %v\n", err)
			os.Exit(1)
		}

		if currentTransaction == nil {
			fmt.Fprintf(os.Stderr, "Error: No transaction in progress. Use 'begin' first.\n")
			os.Exit(1)
		}

		if err := db.CommitTransaction(currentTransaction); err != nil {
			fmt.Fprintf(os.Stderr, "Error committing transaction: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Transaction %s committed successfully\n", currentTransaction.ID)
		currentTransaction = nil
	},
}

var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback the current transaction",
	Long:  `Rollback the current transaction, discarding all changes made in this transaction.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := initStorage(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing storage: %v\n", err)
			os.Exit(1)
		}

		if currentTransaction == nil {
			fmt.Fprintf(os.Stderr, "Error: No transaction in progress. Use 'begin' first.\n")
			os.Exit(1)
		}

		if err := db.AbortTransaction(currentTransaction); err != nil {
			fmt.Fprintf(os.Stderr, "Error rolling back transaction: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Transaction %s rolled back successfully\n", currentTransaction.ID)
		currentTransaction = nil
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show transaction status",
	Long:  `Show the current transaction status and information.`,
	Run: func(cmd *cobra.Command, args []string) {
		if currentTransaction == nil {
			fmt.Println("No transaction in progress")
			return
		}

		fmt.Printf("Transaction ID: %s\n", currentTransaction.ID)
		fmt.Printf("Start Time: %s\n", currentTransaction.StartTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("Status: Active\n")
		
		writeSet := currentTransaction.GetWriteSet()
		deletedSet := currentTransaction.GetDeletedSet()
		
		fmt.Printf("Write Set: %d operations\n", len(writeSet))
		for key := range writeSet {
			fmt.Printf("  PUT %s\n", key)
		}
		
		fmt.Printf("Delete Set: %d operations\n", len(deletedSet))
		for key := range deletedSet {
			fmt.Printf("  DELETE %s\n", key)
		}
	},
}
