package cli

import (
	"fmt"
	"os"
	"strings"

	"startdb/internal/sql"

	"github.com/spf13/cobra"
)

var sqlCmd = &cobra.Command{
	Use:   "sql <query>",
	Short: "Execute a SQL query",
	Long: `Execute a SQL query against the database.
Supports SELECT, INSERT, UPDATE, DELETE, CREATE TABLE, and DROP TABLE statements.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initStorage(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing storage: %v\n", err)
			os.Exit(1)
		}
		defer Cleanup()

		// Join all arguments to form the complete SQL query
		query := strings.Join(args, " ")

		// Parse the SQL query
		parser := sql.NewParser(query)
		stmt, err := parser.Parse()
		if err != nil {
			fmt.Fprintf(os.Stderr, "SQL Parse Error: %v\n", err)
			os.Exit(1)
		}

		// Create SQL executor
		executor := sql.NewExecutor(db)

		// Execute the statement
		result, err := executor.Execute(stmt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "SQL Execution Error: %v\n", err)
			os.Exit(1)
		}

		// Display results
		if result.Count > 0 {
			// Print column headers
			for i, col := range result.Columns {
				if i > 0 {
					fmt.Print(" | ")
				}
				fmt.Print(col)
			}
			fmt.Println()

			// Print separator
			for i, col := range result.Columns {
				if i > 0 {
					fmt.Print("-+-")
				}
				for j := 0; j < len(col); j++ {
					fmt.Print("-")
				}
			}
			fmt.Println()

			// Print rows
			for _, row := range result.Rows {
				for i, value := range row {
					if i > 0 {
						fmt.Print(" | ")
					}
					fmt.Print(value)
				}
				fmt.Println()
			}
		}

		fmt.Printf("\nQuery executed successfully. %d row(s) returned.\n", result.Count)
	},
}
