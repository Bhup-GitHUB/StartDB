package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type DiskEngine struct {
	data     map[string][]byte
	mutex    sync.RWMutex
	closed   bool
	filePath string
}

type DiskData struct {
	Data map[string][]byte `json:"data"`
}

func NewDiskEngine(filePath string) (*DiskEngine, error) {
	engine := &DiskEngine{
		data:     make(map[string][]byte),
		filePath: filePath,
	}

	if err := engine.load(); err != nil {
		return nil, fmt.Errorf("failed to load data: %w", err)
	}

	return engine, nil
}

func (d *DiskEngine) load() error {
	if _, err := os.Stat(d.filePath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(d.filePath)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	var diskData DiskData
	if err := json.Unmarshal(data, &diskData); err != nil {
		return fmt.Errorf("corrupted data file: %w", err)
	}

	d.data = diskData.Data
	return nil
}

func (d *DiskEngine) save() error {
	if d.closed {
		return ErrStorageClosed
	}

	dir := filepath.Dir(d.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	diskData := DiskData{Data: d.data}
	data, err := json.Marshal(diskData)
	if err != nil {
		return err
	}

	tempFile := d.filePath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return err
	}

	return os.Rename(tempFile, d.filePath)
}

func (d *DiskEngine) Get(key string) ([]byte, error) {
	if d.closed {
		return nil, ErrStorageClosed
	}

	if key == "" {
		return nil, ErrInvalidKey
	}

	d.mutex.RLock()
	defer d.mutex.RUnlock()

	value, exists := d.data[key]
	if !exists {
		return nil, ErrKeyNotFound
	}

	result := make([]byte, len(value))
	copy(result, value)
	return result, nil
}

func (d *DiskEngine) Put(key string, value []byte) error {
	if d.closed {
		return ErrStorageClosed
	}

	if key == "" {
		return ErrInvalidKey
	}

	if value == nil {
		return ErrInvalidValue
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.data[key] = make([]byte, len(value))
	copy(d.data[key], value)

	return d.save()
}

func (d *DiskEngine) Delete(key string) error {
	if d.closed {
		return ErrStorageClosed
	}

	if key == "" {
		return ErrInvalidKey
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	if _, exists := d.data[key]; !exists {
		return ErrKeyNotFound
	}

	delete(d.data, key)
	return d.save()
}

func (d *DiskEngine) Exists(key string) (bool, error) {
	if d.closed {
		return false, ErrStorageClosed
	}

	if key == "" {
		return false, ErrInvalidKey
	}

	d.mutex.RLock()
	defer d.mutex.RUnlock()

	_, exists := d.data[key]
	return exists, nil
}

func (d *DiskEngine) Keys() ([]string, error) {
	if d.closed {
		return nil, ErrStorageClosed
	}

	d.mutex.RLock()
	defer d.mutex.RUnlock()

	keys := make([]string, 0, len(d.data))
	for key := range d.data {
		keys = append(keys, key)
	}

	return keys, nil
}

func (d *DiskEngine) Close() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.closed {
		return nil
	}

	d.closed = true
	return d.save()
}
