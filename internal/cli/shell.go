package cli

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"startdb/internal/sql"

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

		PrintHeader("StartDB Interactive Shell (Storage: %s", storageType)
		if walEnabled {
			PrintInfo(", WAL: enabled")
		}
		fmt.Println(")")
		
		if storageType == "disk" {
			PrintMuted("Data file: %s\n", dataFile)
		}
		if walEnabled && walStorage != nil {
			PrintMuted("WAL file: %s\n", walStorage.GetWALPath())
		}
		PrintMuted("Type 'help' for commands, 'quit' to exit\n")
		fmt.Println()

		scanner := bufio.NewScanner(os.Stdin)
		for {
			PrintPrompt("startdb> ")
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
				PrintSuccess("Goodbye!\n")
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
						PrintError("Error: %v\n", err)
					} else {
						PrintSuccess("OK ")
						PrintTransaction("(Transaction: %s)\n", currentTransaction.ID)
					}
				} else {
					err := db.Put(key, []byte(value))
					if err != nil {
						PrintError("Error: %v\n", err)
					} else {
						PrintSuccess("OK\n")
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
						PrintError("Error: %v\n", err)
					} else {
						PrintData("Value: %s ", string(value))
						PrintTransaction("(Transaction: %s)\n", currentTransaction.ID)
					}
				} else {
					value, err = db.Get(parts[1])
					if err != nil {
						PrintError("Error: %v\n", err)
					} else {
						PrintData("Value: %s\n", string(value))
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
						PrintError("Error: %v\n", err)
					} else {
						PrintSuccess("OK ")
						PrintTransaction("(Transaction: %s)\n", currentTransaction.ID)
					}
				} else {
					err := db.Delete(parts[1])
					if err != nil {
						PrintError("Error: %v\n", err)
					} else {
						PrintSuccess("OK\n")
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
						PrintError("Error: %v\n", err)
					} else {
						if exists {
							PrintSuccess("Exists: true ")
						} else {
							PrintWarning("Exists: false ")
						}
						PrintTransaction("(Transaction: %s)\n", currentTransaction.ID)
					}
				} else {
					exists, err = db.Exists(parts[1])
					if err != nil {
						PrintError("Error: %v\n", err)
					} else {
						if exists {
							PrintSuccess("Exists: true\n")
						} else {
							PrintWarning("Exists: false\n")
						}
					}
				}

			case "list":
				keys, err := db.Keys()
				if err != nil {
					PrintError("Error: %v\n", err)
					continue
				}

				if len(keys) == 0 {
					PrintWarning("No keys found in database\n")
					continue
				}

				sort.Strings(keys)
				PrintInfo("Found %d key(s):\n", len(keys))
				for i, key := range keys {
					PrintData("%d. %s\n", i+1, key)
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
					PrintError("Error: Transaction %s already in progress. Use 'commit' or 'rollback' first.\n", currentTransaction.ID)
					continue
				}
				currentTransaction = db.BeginTransaction()
				PrintTransaction("Transaction %s started\n", currentTransaction.ID)

			case "commit":
				if currentTransaction == nil {
					PrintError("Error: No transaction in progress. Use 'begin' first.\n")
					continue
				}
				err := db.CommitTransaction(currentTransaction)
				if err != nil {
					PrintError("Error committing transaction: %v\n", err)
				} else {
					PrintSuccess("Transaction %s committed successfully\n", currentTransaction.ID)
					currentTransaction = nil
				}

			case "rollback":
				if currentTransaction == nil {
					PrintError("Error: No transaction in progress. Use 'begin' first.\n")
					continue
				}
				err := db.AbortTransaction(currentTransaction)
				if err != nil {
					PrintError("Error rolling back transaction: %v\n", err)
				} else {
					PrintWarning("Transaction %s rolled back successfully\n", currentTransaction.ID)
					currentTransaction = nil
				}

			case "status":
				if currentTransaction == nil {
					PrintMuted("No transaction in progress\n")
					continue
				}
				PrintTransaction("Transaction ID: %s\n", currentTransaction.ID)
				PrintInfo("Start Time: %s\n", currentTransaction.StartTime.Format("2006-01-02 15:04:05"))
				PrintSuccess("Status: Active\n")
				
				writeSet := currentTransaction.GetWriteSet()
				deletedSet := currentTransaction.GetDeletedSet()
				
				PrintInfo("Write Set: %d operations\n", len(writeSet))
				for key := range writeSet {
					PrintData("  PUT %s\n", key)
				}
				
				PrintInfo("Delete Set: %d operations\n", len(deletedSet))
				for key := range deletedSet {
					PrintData("  DELETE %s\n", key)
				}

			case "sql":
				if len(parts) < 2 {
					PrintError("Usage: sql <query>\n")
					continue
				}
				
				// Join all parts after "sql" to form the complete query
				query := strings.Join(parts[1:], " ")
				
				// Parse the SQL query
				parser := sql.NewParser(query)
				stmt, err := parser.Parse()
				if err != nil {
					PrintError("SQL Parse Error: %v\n", err)
					continue
				}

				// Create SQL executor
				executor := sql.NewExecutor(db)

				// Execute the statement
				result, err := executor.Execute(stmt)
				if err != nil {
					PrintError("SQL Execution Error: %v\n", err)
					continue
				}

				// Display results
				if result.Count > 0 {
					// Print column headers
					for i, col := range result.Columns {
						if i > 0 {
							PrintMuted(" | ")
						}
						PrintHeader(col)
					}
					fmt.Println()

					// Print separator
					for i, col := range result.Columns {
						if i > 0 {
							PrintMuted("-+-")
						}
						for j := 0; j < len(col); j++ {
							PrintMuted("-")
						}
					}
					fmt.Println()

					// Print rows
					for _, row := range result.Rows {
						for i, value := range row {
							if i > 0 {
								PrintMuted(" | ")
							}
							PrintData("%v", value)
						}
						fmt.Println()
					}
				}

				PrintSuccess("\nQuery executed successfully. %d row(s) returned.\n", result.Count)

			default:
				PrintError("Unknown command: %s (type 'help' for available commands)\n", command)
			}
		}
	},
}

func printHelp() {
	PrintHeader("Available commands:\n")
	PrintData("  set <key> <value>    - Store a key-value pair\n")
	PrintData("  get <key>            - Retrieve a value by key\n")
	PrintData("  delete <key>         - Remove a key-value pair\n")
	PrintData("  exists <key>         - Check if a key exists\n")
	PrintData("  list                 - List all keys\n")
	PrintData("  clear                - Clear screen\n")
	PrintTransaction("  begin                - Begin a new transaction\n")
	PrintSuccess("  commit               - Commit the current transaction\n")
	PrintWarning("  rollback             - Rollback the current transaction\n")
	PrintInfo("  status               - Show transaction status\n")
	PrintSQL("  sql <query>          - Execute a SQL query\n")
	if walEnabled {
		PrintInfo("  checkpoint           - Create a checkpoint (truncate WAL)\n")
		PrintInfo("  recover              - Recover from crash (replay WAL)\n")
		PrintInfo("  wal-status           - Show WAL status\n")
	}
	PrintMuted("  help                 - Show this help\n")
	PrintMuted("  quit/exit            - Exit shell\n")
}
