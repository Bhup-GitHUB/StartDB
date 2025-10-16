package cli

import (
	"os"

	"startdb/internal/storage"

	"github.com/spf13/cobra"
)

var (
	db *storage.Storage
)

var rootCmd = &cobra.Command{
	Use:   "startdb",
	Short: "StartDB - AI-Powered Adaptive Database Management System",
	Long: `StartDB is a next-generation experimental database engine that learns 
from usage patterns and optimizes itself automatically.

Unlike traditional databases that require manual tuning, StartDB uses AI to 
predict query patterns, manage indexes, and adapt to workload changes in real-time.`,
	Version: "1.0.0",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	engine := storage.NewMemoryEngine()
	db = storage.New(engine)
	
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(existsCmd)
}

func Cleanup() {
	if db != nil {
		db.Close()
	}
}
