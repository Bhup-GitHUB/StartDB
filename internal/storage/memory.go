package storage

import (
	"sync"
)

// MemoryEngine implements an in-memory key-value store
type MemoryEngine struct {
	data  map[string][]byte
	mutex sync.RWMutex
	closed bool
}

// NewMemoryEngine creates a new in-memory storage engine
func NewMemoryEngine() *MemoryEngine {
	return &MemoryEngine{
		data: make(map[string][]byte),
	}
}

// Get retrieves a value by key
func (m *MemoryEngine) Get(key string) ([]byte, error) {
	if m.closed {
		return nil, ErrStorageClosed
	}

	if key == "" {
		return nil, ErrInvalidKey
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	value, exists := m.data[key]
	if !exists {
		return nil, ErrKeyNotFound
	}

	// Return a copy to prevent external modification
	result := make([]byte, len(value))
	copy(result, value)
	return result, nil
}

// Put stores a key-value pair
func (m *MemoryEngine) Put(key string, value []byte) error {
	if m.closed {
		return ErrStorageClosed
	}

	if key == "" {
		return ErrInvalidKey
	}

	if value == nil {
		return ErrInvalidValue
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Store a copy to prevent external modification
	m.data[key] = make([]byte, len(value))
	copy(m.data[key], value)
	return nil
}

// Delete removes a key-value pair
func (m *MemoryEngine) Delete(key string) error {
	if m.closed {
		return ErrStorageClosed
	}

	if key == "" {
		return ErrInvalidKey
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.data[key]; !exists {
		return ErrKeyNotFound
	}

	delete(m.data, key)
	return nil
}

// Exists checks if a key exists
func (m *MemoryEngine) Exists(key string) (bool, error) {
	if m.closed {
		return false, ErrStorageClosed
	}

	if key == "" {
		return false, ErrInvalidKey
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	_, exists := m.data[key]
	return exists, nil
}

// Close shuts down the storage engine
func (m *MemoryEngine) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.closed {
		return nil
	}

	m.closed = true
	m.data = nil
	return nil
}
