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
				err := db.Put(key, []byte(value))
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				} else {
					fmt.Println("OK")
				}

			case "get":
				if len(parts) != 2 {
					fmt.Println("Usage: get <key>")
					continue
				}
				value, err := db.Get(parts[1])
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				} else {
					fmt.Printf("Value: %s\n", string(value))
				}

			case "delete":
				if len(parts) != 2 {
					fmt.Println("Usage: delete <key>")
					continue
				}
				err := db.Delete(parts[1])
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				} else {
					fmt.Println("OK")
				}

			case "exists":
				if len(parts) != 2 {
					fmt.Println("Usage: exists <key>")
					continue
				}
				exists, err := db.Exists(parts[1])
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				} else {
					fmt.Printf("Exists: %t\n", exists)
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
	if walEnabled {
		fmt.Println("  checkpoint           - Create a checkpoint (truncate WAL)")
		fmt.Println("  recover              - Recover from crash (replay WAL)")
		fmt.Println("  wal-status           - Show WAL status")
	}
	fmt.Println("  help                 - Show this help")
	fmt.Println("  quit/exit            - Exit shell")
}
