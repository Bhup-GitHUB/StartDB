package storage

import (
	"fmt"
	"sync"
)

type IndexManager struct {
	indexes map[string]*BTree
	mu      sync.RWMutex
}

func NewIndexManager() *IndexManager {
	return &IndexManager{
		indexes: make(map[string]*BTree),
	}
}

func (im *IndexManager) CreateIndex(name string, minDegree int) error {
	im.mu.Lock()
	defer im.mu.Unlock()
	
	if _, exists := im.indexes[name]; exists {
		return fmt.Errorf("index '%s' already exists", name)
	}
	
	im.indexes[name] = NewBTree(minDegree)
	return nil
}

func (im *IndexManager) DropIndex(name string) error {
	im.mu.Lock()
	defer im.mu.Unlock()
	
	if _, exists := im.indexes[name]; !exists {
		return fmt.Errorf("index '%s' does not exist", name)
	}
	
	delete(im.indexes, name)
	return nil
}

func (im *IndexManager) Insert(indexName, key string, value []byte) error {
	im.mu.RLock()
	index, exists := im.indexes[indexName]
	im.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("index '%s' does not exist", indexName)
	}
	
	index.Insert(key, value)
	return nil
}

func (im *IndexManager) Search(indexName, key string) ([]byte, bool) {
	im.mu.RLock()
	index, exists := im.indexes[indexName]
	im.mu.RUnlock()
	
	if !exists {
		return nil, false
	}
	
	return index.Search(key)
}

func (im *IndexManager) Delete(indexName, key string) error {
	im.mu.RLock()
	index, exists := im.indexes[indexName]
	im.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("index '%s' does not exist", indexName)
	}
	
	index.Delete(key)
	return nil
}

func (im *IndexManager) Range(indexName, start, end string) ([]KeyValue, error) {
	im.mu.RLock()
	index, exists := im.indexes[indexName]
	im.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("index '%s' does not exist", indexName)
	}
	
	return index.Range(start, end), nil
}

func (im *IndexManager) GetAll(indexName string) ([]KeyValue, error) {
	im.mu.RLock()
	index, exists := im.indexes[indexName]
	im.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("index '%s' does not exist", indexName)
	}
	
	return index.GetAll(), nil
}

func (im *IndexManager) ListIndexes() []string {
	im.mu.RLock()
	defer im.mu.RUnlock()
	
	indexes := make([]string, 0, len(im.indexes))
	for name := range im.indexes {
		indexes = append(indexes, name)
	}
	
	return indexes
}

func (im *IndexManager) GetIndexInfo(indexName string) (map[string]interface{}, error) {
	im.mu.RLock()
	index, exists := im.indexes[indexName]
	im.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("index '%s' does not exist", indexName)
	}
	
	info := map[string]interface{}{
		"name":       indexName,
		"size":       index.Size,
		"min_degree": index.MinDegree,
		"is_empty":   index.Root == nil,
	}
	
	return info, nil
}

func (im *IndexManager) Exists(indexName string) bool {
	im.mu.RLock()
	defer im.mu.RUnlock()
	
	_, exists := im.indexes[indexName]
	return exists
}

func (im *IndexManager) ClearIndex(indexName string) error {
	im.mu.Lock()
	defer im.mu.Unlock()
	
	if _, exists := im.indexes[indexName]; !exists {
		return fmt.Errorf("index '%s' does not exist", indexName)
	}
	
	im.indexes[indexName] = NewBTree(im.indexes[indexName].MinDegree)
	return nil
}

func (im *IndexManager) GetIndexStats() map[string]map[string]interface{} {
	im.mu.RLock()
	defer im.mu.RUnlock()
	
	stats := make(map[string]map[string]interface{})
	for name, index := range im.indexes {
		stats[name] = map[string]interface{}{
			"size":       index.Size,
			"min_degree": index.MinDegree,
			"is_empty":   index.Root == nil,
		}
	}
	
	return stats
}