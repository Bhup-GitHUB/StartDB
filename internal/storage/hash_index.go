package storage

import (
	"hash/fnv"
	"sync"
)

// HashIndex implements a hash-based index for fast equality lookups
type HashIndex struct {
	buckets []map[string][]byte
	mu      sync.RWMutex
	size    int
}

// NewHashIndex creates a new hash index with the specified number of buckets
func NewHashIndex(bucketCount int) *HashIndex {
	if bucketCount <= 0 {
		bucketCount = 16 // Default bucket count
	}
	return &HashIndex{
		buckets: make([]map[string][]byte, bucketCount),
		size:    0,
	}
}

// hash computes the hash value for a key
func (hi *HashIndex) hash(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

// getBucket returns the bucket index for a given key
func (hi *HashIndex) getBucket(key string) int {
	return int(hi.hash(key)) % len(hi.buckets)
}

// Insert inserts a key-value pair into the hash index
func (hi *HashIndex) Insert(key string, value []byte) {
	hi.mu.Lock()
	defer hi.mu.Unlock()

	bucketIdx := hi.getBucket(key)
	if hi.buckets[bucketIdx] == nil {
		hi.buckets[bucketIdx] = make(map[string][]byte)
	}

	// Check if key already exists
	if _, exists := hi.buckets[bucketIdx][key]; !exists {
		hi.size++
	}

	hi.buckets[bucketIdx][key] = value
}

// Search searches for a key in the hash index
func (hi *HashIndex) Search(key string) ([]byte, bool) {
	hi.mu.RLock()
	defer hi.mu.RUnlock()

	bucketIdx := hi.getBucket(key)
	bucket := hi.buckets[bucketIdx]
	if bucket == nil {
		return nil, false
	}

	value, exists := bucket[key]
	return value, exists
}

// Delete deletes a key from the hash index
func (hi *HashIndex) Delete(key string) bool {
	hi.mu.Lock()
	defer hi.mu.Unlock()

	bucketIdx := hi.getBucket(key)
	bucket := hi.buckets[bucketIdx]
	if bucket == nil {
		return false
	}

	if _, exists := bucket[key]; exists {
		delete(bucket, key)
		hi.size--
		return true
	}

	return false
}

// Size returns the number of entries in the hash index
func (hi *HashIndex) Size() int {
	hi.mu.RLock()
	defer hi.mu.RUnlock()
	return hi.size
}

// GetAll returns all key-value pairs in the hash index
func (hi *HashIndex) GetAll() []KeyValue {
	hi.mu.RLock()
	defer hi.mu.RUnlock()

	var result []KeyValue
	for _, bucket := range hi.buckets {
		if bucket != nil {
			for key, value := range bucket {
				result = append(result, KeyValue{
					Key:   key,
					Value: value,
				})
			}
		}
	}

	return result
}

// Clear removes all entries from the hash index
func (hi *HashIndex) Clear() {
	hi.mu.Lock()
	defer hi.mu.Unlock()

	for i := range hi.buckets {
		hi.buckets[i] = nil
	}
	hi.size = 0
}

