package storage

import (
	"sync"
)

type MemoryEngine struct {
	data  map[string][]byte
	mutex sync.RWMutex
	closed bool
}

func NewMemoryEngine() *MemoryEngine {
	return &MemoryEngine{
		data: make(map[string][]byte),
	}
}

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

	result := make([]byte, len(value))
	copy(result, value)
	return result, nil
}

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

	m.data[key] = make([]byte, len(value))
	copy(m.data[key], value)
	return nil
}

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

func (m *MemoryEngine) Keys() ([]string, error) {
	if m.closed {
		return nil, ErrStorageClosed
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	keys := make([]string, 0, len(m.data))
	for key := range m.data {
		keys = append(keys, key)
	}

	return keys, nil
}

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
