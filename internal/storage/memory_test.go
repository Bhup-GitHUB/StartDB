package storage

import (
	"testing"
)

func TestMemoryEngine(t *testing.T) {
	engine := NewMemoryEngine()
	defer engine.Close()

	// Test Put and Get
	err := engine.Put("key1", []byte("value1"))
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

	// Test Exists
	exists, err := engine.Exists("key1")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Fatal("Key should exist")
	}

	// Test Delete
	err = engine.Delete("key1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Test Get after delete
	_, err = engine.Get("key1")
	if err != ErrKeyNotFound {
		t.Fatalf("Expected ErrKeyNotFound, got %v", err)
	}

	// Test Exists after delete
	exists, err = engine.Exists("key1")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Fatal("Key should not exist")
	}
}

func TestMemoryEngineErrors(t *testing.T) {
	engine := NewMemoryEngine()
	defer engine.Close()

	// Test invalid key
	err := engine.Put("", []byte("value"))
	if err != ErrInvalidKey {
		t.Fatalf("Expected ErrInvalidKey, got %v", err)
	}

	// Test invalid value
	err = engine.Put("key", nil)
	if err != ErrInvalidValue {
		t.Fatalf("Expected ErrInvalidValue, got %v", err)
	}

	// Test closed engine
	engine.Close()
	err = engine.Put("key", []byte("value"))
	if err != ErrStorageClosed {
		t.Fatalf("Expected ErrStorageClosed, got %v", err)
	}
}
