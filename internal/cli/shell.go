package cli

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Start interactive StartDB shell",
	Long: `Start an interactive shell where you can run multiple commands.
Type 'help' for available commands, 'quit' to exit.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if err := initStorage(); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		defer Cleanup()

		fmt.Printf("StartDB Interactive Shell (Storage: %s", storageType)
		if walEnabled {
			fmt.Printf(", WAL: enabled")
		}
		fmt.Println(")")
		
		if storageType == "disk" {
			fmt.Printf("Data file: %s\n", dataFile)
		}
		if walEnabled && walStorage != nil {
			fmt.Printf("WAL file: %s\n", walStorage.GetWALPath())
		}
		fmt.Println("Type 'help' for commands, 'quit' to exit")
		fmt.Println()

		scanner := bufio.NewScanner(os.Stdin)
		for {
			fmt.Print("startdb> ")
			if !scanner.Scan() {
				break
			}

			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			parts := strings.Fields(line)
			if len(parts) == 0 {
				continue
			}

			command := strings.ToLower(parts[0])

			switch command {
			case "quit", "exit":
				fmt.Println("Goodbye!")
				return

			case "help":
				printHelp()

			case "set":
				if len(parts) < 3 {
					fmt.Println("Usage: set <key> <value>")
					continue
				}
				key := parts[1]
				value := strings.Join(parts[2:], " ")
				
				if currentTransaction != nil {
					err := currentTransaction.Put(key, []byte(value))
					if err != nil {
						fmt.Printf("Error: %v\n", err)
					} else {
						fmt.Printf("OK (Transaction: %s)\n", currentTransaction.ID)
					}
				} else {
					err := db.Put(key, []byte(value))
					if err != nil {
						fmt.Printf("Error: %v\n", err)
					} else {
						fmt.Println("OK")
					}
				}

			case "get":
				if len(parts) != 2 {
					fmt.Println("Usage: get <key>")
					continue
				}
				
				var value []byte
				var err error
				
				if currentTransaction != nil {
					value, err = currentTransaction.Get(parts[1])
					if err != nil {
						fmt.Printf("Error: %v\n", err)
					} else {
						fmt.Printf("Value: %s (Transaction: %s)\n", string(value), currentTransaction.ID)
					}
				} else {
					value, err = db.Get(parts[1])
					if err != nil {
						fmt.Printf("Error: %v\n", err)
					} else {
						fmt.Printf("Value: %s\n", string(value))
					}
				}

			case "delete":
				if len(parts) != 2 {
					fmt.Println("Usage: delete <key>")
					continue
				}
				
				if currentTransaction != nil {
					err := currentTransaction.Delete(parts[1])
					if err != nil {
						fmt.Printf("Error: %v\n", err)
					} else {
						fmt.Printf("OK (Transaction: %s)\n", currentTransaction.ID)
					}
				} else {
					err := db.Delete(parts[1])
					if err != nil {
						fmt.Printf("Error: %v\n", err)
					} else {
						fmt.Println("OK")
					}
				}

			case "exists":
				if len(parts) != 2 {
					fmt.Println("Usage: exists <key>")
					continue
				}
				
				var exists bool
				var err error
				
				if currentTransaction != nil {
					exists, err = currentTransaction.Exists(parts[1])
					if err != nil {
						fmt.Printf("Error: %v\n", err)
					} else {
						fmt.Printf("Exists: %t (Transaction: %s)\n", exists, currentTransaction.ID)
					}
				} else {
					exists, err = db.Exists(parts[1])
					if err != nil {
						fmt.Printf("Error: %v\n", err)
					} else {
						fmt.Printf("Exists: %t\n", exists)
					}
				}

			case "list":
				keys, err := db.Keys()
				if err != nil {
					fmt.Printf("Error: %v\n", err)
					continue
				}

				if len(keys) == 0 {
					fmt.Println("No keys found in database")
					continue
				}

				sort.Strings(keys)
				fmt.Printf("Found %d key(s):\n", len(keys))
				for i, key := range keys {
					fmt.Printf("%d. %s\n", i+1, key)
				}

			case "clear":
				fmt.Print("\033[H\033[2J")

			case "checkpoint":
				if !walEnabled {
					fmt.Println("Error: WAL is not enabled")
					continue
				}
				if walStorage == nil {
					fmt.Println("Error: WAL storage not initialized")
					continue
				}
				err := walStorage.Checkpoint()
				if err != nil {
					fmt.Printf("Error creating checkpoint: %v\n", err)
				} else {
					fmt.Println("Checkpoint created successfully")
				}

			case "recover":
				if !walEnabled {
					fmt.Println("Error: WAL is not enabled")
					continue
				}
				if walStorage == nil {
					fmt.Println("Error: WAL storage not initialized")
					continue
				}
				err := walStorage.Recover()
				if err != nil {
					fmt.Printf("Error during recovery: %v\n", err)
				} else {
					fmt.Println("Recovery completed successfully")
				}

			case "wal-status":
				if walEnabled {
					if walStorage != nil {
						fmt.Printf("WAL enabled - File: %s\n", walStorage.GetWALPath())
					} else {
						fmt.Println("WAL enabled but not initialized")
					}
				} else {
					fmt.Println("WAL disabled")
				}

			case "begin":
				if currentTransaction != nil {
					fmt.Printf("Error: Transaction %s already in progress. Use 'commit' or 'rollback' first.\n", currentTransaction.ID)
					continue
				}
				currentTransaction = db.BeginTransaction()
				fmt.Printf("Transaction %s started\n", currentTransaction.ID)

			case "commit":
				if currentTransaction == nil {
					fmt.Println("Error: No transaction in progress. Use 'begin' first.")
					continue
				}
				err := db.CommitTransaction(currentTransaction)
				if err != nil {
					fmt.Printf("Error committing transaction: %v\n", err)
				} else {
					fmt.Printf("Transaction %s committed successfully\n", currentTransaction.ID)
					currentTransaction = nil
				}

			case "rollback":
				if currentTransaction == nil {
					fmt.Println("Error: No transaction in progress. Use 'begin' first.")
					continue
				}
				err := db.AbortTransaction(currentTransaction)
				if err != nil {
					fmt.Printf("Error rolling back transaction: %v\n", err)
				} else {
					fmt.Printf("Transaction %s rolled back successfully\n", currentTransaction.ID)
					currentTransaction = nil
				}

			case "status":
				if currentTransaction == nil {
					fmt.Println("No transaction in progress")
					continue
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

			default:
				fmt.Printf("Unknown command: %s (type 'help' for available commands)\n", command)
			}
		}
	},
}

func printHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  set <key> <value>    - Store a key-value pair")
	fmt.Println("  get <key>            - Retrieve a value by key")
	fmt.Println("  delete <key>         - Remove a key-value pair")
	fmt.Println("  exists <key>         - Check if a key exists")
	fmt.Println("  list                 - List all keys")
	fmt.Println("  clear                - Clear screen")
	fmt.Println("  begin                - Begin a new transaction")
	fmt.Println("  commit               - Commit the current transaction")
	fmt.Println("  rollback             - Rollback the current transaction")
	fmt.Println("  status               - Show transaction status")
	if walEnabled {
		fmt.Println("  checkpoint           - Create a checkpoint (truncate WAL)")
		fmt.Println("  recover              - Recover from crash (replay WAL)")
		fmt.Println("  wal-status           - Show WAL status")
	}
	fmt.Println("  help                 - Show this help")
	fmt.Println("  quit/exit            - Exit shell")
}
