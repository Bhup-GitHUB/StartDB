package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"startdb/internal/storage"
)

func main() {
	
	fmt.Println("Starting StartDB...")
	engine := storage.NewMemoryEngine()
	db := storage.New(engine)
	defer db.Close()

	fmt.Println("StartDB - In-Memory Key-Value Store")
	fmt.Println("Commands: get <key>, put <key> <value>, delete <key>, exists <key>, quit")
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

		case "put":
			if len(parts) < 3 {
				fmt.Println("Usage: put <key> <value>")
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

		default:
			fmt.Printf("Unknown command: %s\n", command)
		}
	}
}
