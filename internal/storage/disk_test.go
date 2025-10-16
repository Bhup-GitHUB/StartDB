package storage

import (
	"os"
	"testing"
)

func TestDiskEngine(t *testing.T) {
	tempFile := "test_data.json"
	defer os.Remove(tempFile)

	engine, err := NewDiskEngine(tempFile)
	if err != nil {
		t.Fatalf("Failed to create disk engine: %v", err)
	}
	defer engine.Close()

	err = engine.Put("key1", []byte("value1"))
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	value, err := engine.Get("key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(value) != "value1" {
		t.Fatalf("Expected 'value1', got '%s'", string(value))
	}

	exists, err := engine.Exists("key1")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Fatal("Key should exist")
	}

	err = engine.Delete("key1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = engine.Get("key1")
	if err != ErrKeyNotFound {
		t.Fatalf("Expected ErrKeyNotFound, got %v", err)
	}
}

func TestDiskEnginePersistence(t *testing.T) {
	tempFile := "test_persistence.json"
	defer os.Remove(tempFile)

	engine1, err := NewDiskEngine(tempFile)
	if err != nil {
		t.Fatalf("Failed to create disk engine: %v", err)
	}

	err = engine1.Put("persistent_key", []byte("persistent_value"))
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	engine1.Close()

	engine2, err := NewDiskEngine(tempFile)
	if err != nil {
		t.Fatalf("Failed to create second disk engine: %v", err)
	}
	defer engine2.Close()

	value, err := engine2.Get("persistent_key")
	if err != nil {
		t.Fatalf("Get failed after restart: %v", err)
	}

	if string(value) != "persistent_value" {
		t.Fatalf("Expected 'persistent_value', got '%s'", string(value))
	}
}

func TestDiskEngineErrors(t *testing.T) {
	tempFile := "test_errors.json"
	defer os.Remove(tempFile)

	engine, err := NewDiskEngine(tempFile)
	if err != nil {
		t.Fatalf("Failed to create disk engine: %v", err)
	}
	defer engine.Close()

	err = engine.Put("", []byte("value"))
	if err != ErrInvalidKey {
		t.Fatalf("Expected ErrInvalidKey, got %v", err)
	}

	err = engine.Put("key", nil)
	if err != ErrInvalidValue {
		t.Fatalf("Expected ErrInvalidValue, got %v", err)
	}

	engine.Close()
	err = engine.Put("key", []byte("value"))
	if err != ErrStorageClosed {
		t.Fatalf("Expected ErrStorageClosed, got %v", err)
	}
}
