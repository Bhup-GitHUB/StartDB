package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var checkpointCmd = &cobra.Command{
	Use:   "checkpoint",
	Short: "Create a checkpoint by truncating the WAL",
	Long: `Create a checkpoint by truncating the Write-Ahead Log.
This operation consolidates all pending changes and clears the WAL file.
Use this command periodically to manage WAL file size.`,
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

		if err := walStorage.Checkpoint(); err != nil {
			fmt.Printf("Error creating checkpoint: %v\n", err)
			return
		}

		fmt.Println("Checkpoint created successfully")
		fmt.Printf("WAL file: %s\n", walStorage.GetWALPath())
	},
}
