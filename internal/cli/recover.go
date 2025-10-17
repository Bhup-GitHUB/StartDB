package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var recoverCmd = &cobra.Command{
	Use:   "recover",
	Short: "Recover from a crash by replaying the WAL",
	Long: `Recover from a crash by replaying the Write-Ahead Log.
This command replays all operations from the WAL to restore the database
to its last consistent state before the crash.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if !walEnabled {
			fmt.Println("Error: WAL is not enabled. Use --wal flag to enable Write-Ahead Logging.")
			return
		}

		if err := initStorage(); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		defer Cleanup()

		if walStorage == nil {
			fmt.Println("Error: WAL storage not initialized")
			return
		}

		if err := walStorage.Recover(); err != nil {
			fmt.Printf("Error during recovery: %v\n", err)
			return
		}

		fmt.Println("Recovery completed successfully")
		fmt.Printf("WAL file: %s\n", walStorage.GetWALPath())
	},
}
