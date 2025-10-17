package storage

import (
	"fmt"
	"os"
	"testing"
)

func TestWALBasicOperations(t *testing.T) {
	tempFile := "test_wal.log"
	defer os.Remove(tempFile)

	wal, err := NewWAL(tempFile)
	if err != nil {
		t.Fatalf("Failed to create WAL: %v", err)
	}
	defer wal.Close()

	// Test LogPut
	err = wal.LogPut("key1", []byte("value1"))
	if err != nil {
		t.Fatalf("LogPut failed: %v", err)
	}

	// Test LogDelete
	err = wal.LogDelete("key2")
	if err != nil {
		t.Fatalf("LogDelete failed: %v", err)
	}

	// Test LogCommit
	err = wal.LogCommit()
	if err != nil {
		t.Fatalf("LogCommit failed: %v", err)
	}
}

func TestWALReplay(t *testing.T) {
	tempFile := "test_wal_replay.log"
	defer os.Remove(tempFile)

	// Create WAL and log some operations
	wal, err := NewWAL(tempFile)
	if err != nil {
		t.Fatalf("Failed to create WAL: %v", err)
	}

	err = wal.LogPut("key1", []byte("value1"))
	if err != nil {
		t.Fatalf("LogPut failed: %v", err)
	}

	err = wal.LogPut("key2", []byte("value2"))
	if err != nil {
		t.Fatalf("LogPut failed: %v", err)
	}

	err = wal.LogDelete("key1")
	if err != nil {
		t.Fatalf("LogDelete failed: %v", err)
	}

	wal.Close()

	// Create a new memory engine and replay the WAL
	engine := NewMemoryEngine()
	defer engine.Close()

	wal2, err := NewWAL(tempFile)
	if err != nil {
		t.Fatalf("Failed to create WAL for replay: %v", err)
	}
	defer wal2.Close()

	err = wal2.Replay(engine)
	if err != nil {
		t.Fatalf("WAL replay failed: %v", err)
	}

	// Verify the operations were replayed correctly
	// key1 should not exist (was deleted)
	_, err = engine.Get("key1")
	if err != ErrKeyNotFound {
		t.Fatalf("Expected ErrKeyNotFound for key1, got %v", err)
	}

	// key2 should exist with value "value2"
	value, err := engine.Get("key2")
	if err != nil {
		t.Fatalf("Failed to get key2: %v", err)
	}
	if string(value) != "value2" {
		t.Fatalf("Expected 'value2', got '%s'", string(value))
	}
}

func TestWALStorageIntegration(t *testing.T) {
	tempDataFile := "test_wal_storage_data.json"
	tempWALFile := "test_wal_storage.wal"
	defer os.Remove(tempDataFile)
	defer os.Remove(tempWALFile)

	// Create WAL-enabled storage
	storage, err := NewWALDiskEngine(tempDataFile, tempWALFile)
	if err != nil {
		t.Fatalf("Failed to create WAL storage: %v", err)
	}
	defer storage.Close()

	// Test basic operations
	err = storage.Put("key1", []byte("value1"))
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	err = storage.Put("key2", []byte("value2"))
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Verify data
	value, err := storage.Get("key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if string(value) != "value1" {
		t.Fatalf("Expected 'value1', got '%s'", string(value))
	}

	// Test delete
	err = storage.Delete("key1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	_, err = storage.Get("key1")
	if err != ErrKeyNotFound {
		t.Fatalf("Expected ErrKeyNotFound, got %v", err)
	}
}

func TestWALCrashRecovery(t *testing.T) {
	tempDataFile := "test_crash_recovery_data.json"
	tempWALFile := "test_crash_recovery.wal"
	defer os.Remove(tempDataFile)
	defer os.Remove(tempWALFile)

	// Simulate crash recovery scenario
	// 1. Create storage and perform operations
	storage1, err := NewWALDiskEngine(tempDataFile, tempWALFile)
	if err != nil {
		t.Fatalf("Failed to create WAL storage: %v", err)
	}

	err = storage1.Put("key1", []byte("value1"))
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	err = storage1.Put("key2", []byte("value2"))
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Simulate crash by closing without proper cleanup
	storage1.Close()

	// 2. Recover by creating new storage (WAL replay happens automatically)
	storage2, err := NewWALDiskEngine(tempDataFile, tempWALFile)
	if err != nil {
		t.Fatalf("Failed to create WAL storage for recovery: %v", err)
	}
	defer storage2.Close()

	// 3. Verify data was recovered
	value, err := storage2.Get("key1")
	if err != nil {
		t.Fatalf("Get failed after recovery: %v", err)
	}
	if string(value) != "value1" {
		t.Fatalf("Expected 'value1' after recovery, got '%s'", string(value))
	}

	value, err = storage2.Get("key2")
	if err != nil {
		t.Fatalf("Get failed after recovery: %v", err)
	}
	if string(value) != "value2" {
		t.Fatalf("Expected 'value2' after recovery, got '%s'", string(value))
	}
}

func TestWALCheckpoint(t *testing.T) {
	tempDataFile := "test_checkpoint_data.json"
	tempWALFile := "test_checkpoint.wal"
	defer os.Remove(tempDataFile)
	defer os.Remove(tempWALFile)

	// Create WAL storage
	storage, err := NewWALDiskEngine(tempDataFile, tempWALFile)
	if err != nil {
		t.Fatalf("Failed to create WAL storage: %v", err)
	}
	defer storage.Close()

	// Perform some operations
	err = storage.Put("key1", []byte("value1"))
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Create checkpoint
	err = storage.Checkpoint()
	if err != nil {
		t.Fatalf("Checkpoint failed: %v", err)
	}

	// Verify data still exists
	value, err := storage.Get("key1")
	if err != nil {
		t.Fatalf("Get failed after checkpoint: %v", err)
	}
	if string(value) != "value1" {
		t.Fatalf("Expected 'value1' after checkpoint, got '%s'", string(value))
	}
}

func TestWALMemoryEngine(t *testing.T) {
	tempWALFile := "test_wal_memory.wal"
	defer os.Remove(tempWALFile)

	// Create WAL-enabled memory engine
	storage, err := NewWALMemoryEngine(tempWALFile)
	if err != nil {
		t.Fatalf("Failed to create WAL memory engine: %v", err)
	}
	defer storage.Close()

	// Test operations
	err = storage.Put("key1", []byte("value1"))
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	value, err := storage.Get("key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if string(value) != "value1" {
		t.Fatalf("Expected 'value1', got '%s'", string(value))
	}
}

func TestWALAutoPath(t *testing.T) {
	tempDataFile := "test_auto_path_data.json"
	defer os.Remove(tempDataFile)
	defer os.Remove("test_auto_path_data.wal")

	// Create WAL storage with auto-generated WAL path
	storage, err := NewWALDiskEngineWithAutoPath(tempDataFile)
	if err != nil {
		t.Fatalf("Failed to create WAL storage with auto path: %v", err)
	}
	defer storage.Close()

	// Verify WAL path was generated correctly
	expectedWALPath := "test_auto_path_data.wal"
	if storage.GetWALPath() != expectedWALPath {
		t.Fatalf("Expected WAL path '%s', got '%s'", expectedWALPath, storage.GetWALPath())
	}

	// Test operations
	err = storage.Put("key1", []byte("value1"))
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}
}

func TestWALConcurrency(t *testing.T) {
	tempFile := "test_wal_concurrency.log"
	defer os.Remove(tempFile)

	wal, err := NewWAL(tempFile)
	if err != nil {
		t.Fatalf("Failed to create WAL: %v", err)
	}
	defer wal.Close()

	// Test concurrent writes
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(i int) {
			defer func() { done <- true }()
			key := fmt.Sprintf("key%d", i)
			value := fmt.Sprintf("value%d", i)
			if err := wal.LogPut(key, []byte(value)); err != nil {
				t.Errorf("Concurrent LogPut failed: %v", err)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all entries were written
	engine := NewMemoryEngine()
	defer engine.Close()

	if err := wal.Replay(engine); err != nil {
		t.Fatalf("WAL replay failed: %v", err)
	}

	keys, err := engine.Keys()
	if err != nil {
		t.Fatalf("Failed to get keys: %v", err)
	}

	if len(keys) != 10 {
		t.Fatalf("Expected 10 keys, got %d", len(keys))
	}
}

func TestWALChecksum(t *testing.T) {
	tempFile := "test_wal_checksum.log"
	defer os.Remove(tempFile)

	wal, err := NewWAL(tempFile)
	if err != nil {
		t.Fatalf("Failed to create WAL: %v", err)
	}
	defer wal.Close()

	// Log an entry
	err = wal.LogPut("testkey", []byte("testvalue"))
	if err != nil {
		t.Fatalf("LogPut failed: %v", err)
	}

	// Manually corrupt the file to test checksum verification
	file, err := os.OpenFile(tempFile, os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to open file for corruption: %v", err)
	}
	defer file.Close()

	// Seek to a position and write garbage
	file.Seek(10, 0)
	file.Write([]byte("CORRUPTED"))
	file.Close()

	// Try to replay - should fail due to checksum mismatch
	engine := NewMemoryEngine()
	defer engine.Close()

	err = wal.Replay(engine)
	if err == nil {
		t.Fatal("Expected checksum verification to fail, but it succeeded")
	}
}
