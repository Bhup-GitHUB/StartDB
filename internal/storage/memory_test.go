package storage

import (
	"testing"
)

func TestMemoryEngine(t *testing.T) {
	engine := NewMemoryEngine()
	defer engine.Close()

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

	err := engine.Put("", []byte("value"))
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
