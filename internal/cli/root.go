package cli

import (
	"fmt"
	"os"

	"startdb/internal/storage"

	"github.com/spf13/cobra"
)

var (
	db        *storage.Storage
	walStorage storage.WALEngine
	storageType string
	dataFile  string
	walEnabled bool
	walFile   string
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
	rootCmd.PersistentFlags().BoolVarP(&walEnabled, "wal", "w", false, "Enable Write-Ahead Logging for crash recovery")
	rootCmd.PersistentFlags().StringVarP(&walFile, "wal-file", "", "", "WAL file path (auto-generated if not specified)")
	
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(shellCmd)
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(existsCmd)
	rootCmd.AddCommand(checkpointCmd)
	rootCmd.AddCommand(recoverCmd)
	rootCmd.AddCommand(beginCmd)
	rootCmd.AddCommand(commitCmd)
	rootCmd.AddCommand(rollbackCmd)
	rootCmd.AddCommand(statusCmd)
}

func initStorage() error {
	var engine storage.Engine
	var err error

	var walPath string
	if walEnabled {
		if walFile != "" {
			walPath = walFile
		} else {
			if storageType == "disk" {
				walPath = dataFile + ".wal"
			} else {
				walPath = "startdb.wal"
			}
		}
	}

	switch storageType {
	case "memory":
		if walEnabled {
			walStorage, err = storage.NewWALMemoryEngine(walPath)
			if err != nil {
				return fmt.Errorf("failed to initialize WAL memory storage: %w", err)
			}
			db = storage.New(walStorage)
		} else {
			engine = storage.NewMemoryEngine()
			db = storage.New(engine)
		}
	case "disk":
		if walEnabled {
			walStorage, err = storage.NewWALDiskEngine(dataFile, walPath)
			if err != nil {
				return fmt.Errorf("failed to initialize WAL disk storage: %w", err)
			}
			db = storage.New(walStorage)
		} else {
			engine, err = storage.NewDiskEngine(dataFile)
			if err != nil {
				return fmt.Errorf("failed to initialize disk storage: %w", err)
			}
			db = storage.New(engine)
		}
	default:
		return fmt.Errorf("invalid storage type: %s (use 'memory' or 'disk')", storageType)
	}

	return nil
}

func Cleanup() {
	if db != nil {
		db.Close()
	}
}
