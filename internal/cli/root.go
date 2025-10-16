package cli

import (
	"fmt"
	"os"

	"startdb/internal/storage"

	"github.com/spf13/cobra"
)

var (
	db        *storage.Storage
	storageType string
	dataFile  string
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
	rootCmd.PersistentFlags().StringVarP(&storageType, "storage", "s", "memory", "Storage type: memory or disk")
	rootCmd.PersistentFlags().StringVarP(&dataFile, "data", "d", "startdb.json", "Data file path for disk storage")
	
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(existsCmd)
}

func initStorage() error {
	var engine storage.Engine
	var err error

	switch storageType {
	case "memory":
		engine = storage.NewMemoryEngine()
	case "disk":
		engine, err = storage.NewDiskEngine(dataFile)
		if err != nil {
			return fmt.Errorf("failed to initialize disk storage: %w", err)
		}
	default:
		return fmt.Errorf("invalid storage type: %s (use 'memory' or 'disk')", storageType)
	}

	db = storage.New(engine)
	return nil
}

func Cleanup() {
	if db != nil {
		db.Close()
	}
}
