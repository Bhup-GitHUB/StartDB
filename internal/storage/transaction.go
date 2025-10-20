package storage

import (
	"fmt"
	"sync"
	"time"
)

// Transaction represents a database transaction
type Transaction struct {
	ID        string
	StartTime time.Time
	ReadSet   map[string][]byte // Keys read during transaction
	WriteSet  map[string][]byte // Keys written during transaction
	Deleted   map[string]bool   // Keys deleted during transaction
	mu        sync.RWMutex
	committed bool
	aborted   bool
}

// TransactionManager manages concurrent transactions
type TransactionManager struct {
	transactions map[string]*Transaction
	mu           sync.RWMutex
	nextID       int64
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager() *TransactionManager {
	return &TransactionManager{
		transactions: make(map[string]*Transaction),
	}
}

// BeginTransaction starts a new transaction
func (tm *TransactionManager) BeginTransaction() *Transaction {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.nextID++
	tx := &Transaction{
		ID:        fmt.Sprintf("tx_%d", tm.nextID),
		StartTime: time.Now(),
		ReadSet:   make(map[string][]byte),
		WriteSet:  make(map[string][]byte),
		Deleted:   make(map[string]bool),
	}

	tm.transactions[tx.ID] = tx
	return tx
}

// GetTransaction retrieves a transaction by ID
func (tm *TransactionManager) GetTransaction(id string) (*Transaction, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	tx, exists := tm.transactions[id]
	return tx, exists
}

// CommitTransaction commits a transaction
func (tm *TransactionManager) CommitTransaction(id string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tx, exists := tm.transactions[id]
	if !exists {
		return ErrTransactionNotFound
	}

	if tx.aborted {
		return ErrTransactionAborted
	}

	if tx.committed {
		return ErrTransactionAlreadyCommitted
	}

	tx.mu.Lock()
	tx.committed = true
	tx.mu.Unlock()

	delete(tm.transactions, id)
	return nil
}

// AbortTransaction aborts a transaction
func (tm *TransactionManager) AbortTransaction(id string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tx, exists := tm.transactions[id]
	if !exists {
		return ErrTransactionNotFound
	}

	if tx.committed {
		return ErrTransactionAlreadyCommitted
	}

	tx.mu.Lock()
	tx.aborted = true
	tx.mu.Unlock()

	delete(tm.transactions, id)
	return nil
}

// Transaction methods
func (tx *Transaction) Get(key string) ([]byte, error) {
	tx.mu.RLock()
	defer tx.mu.RUnlock()

	if tx.aborted {
		return nil, ErrTransactionAborted
	}

	if tx.committed {
		return nil, ErrTransactionAlreadyCommitted
	}

	// Check if we've already read this key in this transaction
	if value, exists := tx.ReadSet[key]; exists {
		return value, nil
	}

	// Check if we've written this key in this transaction
	if value, exists := tx.WriteSet[key]; exists {
		// Add to read set
		tx.ReadSet[key] = make([]byte, len(value))
		copy(tx.ReadSet[key], value)
		return value, nil
	}

	// Check if we've deleted this key in this transaction
	if tx.Deleted[key] {
		return nil, ErrKeyNotFound
	}

	return nil, ErrKeyNotFound
}

func (tx *Transaction) Put(key string, value []byte) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.aborted {
		return ErrTransactionAborted
	}

	if tx.committed {
		return ErrTransactionAlreadyCommitted
	}

	if key == "" {
		return ErrInvalidKey
	}

	if value == nil {
		return ErrInvalidValue
	}

	// Add to write set
	tx.WriteSet[key] = make([]byte, len(value))
	copy(tx.WriteSet[key], value)

	// Remove from deleted set if it was deleted
	delete(tx.Deleted, key)

	return nil
}

func (tx *Transaction) Delete(key string) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.aborted {
		return ErrTransactionAborted
	}

	if tx.committed {
		return ErrTransactionAlreadyCommitted
	}

	if key == "" {
		return ErrInvalidKey
	}

	// Mark as deleted
	tx.Deleted[key] = true

	// Remove from write set if it was written
	delete(tx.WriteSet, key)

	return nil
}

func (tx *Transaction) Exists(key string) (bool, error) {
	tx.mu.RLock()
	defer tx.mu.RUnlock()

	if tx.aborted {
		return false, ErrTransactionAborted
	}

	if tx.committed {
		return false, ErrTransactionAlreadyCommitted
	}

	// Check if we've deleted this key
	if tx.Deleted[key] {
		return false, nil
	}

	// Check if we've written this key
	if _, exists := tx.WriteSet[key]; exists {
		return true, nil
	}

	// Check if we've read this key
	if _, exists := tx.ReadSet[key]; exists {
		return true, nil
	}

	return false, nil
}

func (tx *Transaction) Keys() ([]string, error) {
	tx.mu.RLock()
	defer tx.mu.RUnlock()

	if tx.aborted {
		return nil, ErrTransactionAborted
	}

	if tx.committed {
		return nil, ErrTransactionAlreadyCommitted
	}

	keys := make([]string, 0)

	// Add keys from write set that aren't deleted
	for key := range tx.WriteSet {
		if !tx.Deleted[key] {
			keys = append(keys, key)
		}
	}

	// Add keys from read set that aren't deleted and weren't written
	for key := range tx.ReadSet {
		if !tx.Deleted[key] && tx.WriteSet[key] == nil {
			keys = append(keys, key)
		}
	}

	return keys, nil
}

// IsCommitted checks if the transaction is committed
func (tx *Transaction) IsCommitted() bool {
	tx.mu.RLock()
	defer tx.mu.RUnlock()
	return tx.committed
}

// IsAborted checks if the transaction is aborted
func (tx *Transaction) IsAborted() bool {
	tx.mu.RLock()
	defer tx.mu.RUnlock()
	return tx.aborted
}

// GetWriteSet returns the write set (for engine implementation)
func (tx *Transaction) GetWriteSet() map[string][]byte {
	tx.mu.RLock()
	defer tx.mu.RUnlock()
	
	result := make(map[string][]byte)
	for key, value := range tx.WriteSet {
		if !tx.Deleted[key] {
			result[key] = make([]byte, len(value))
			copy(result[key], value)
		}
	}
	return result
}

// GetDeletedSet returns the deleted set (for engine implementation)
func (tx *Transaction) GetDeletedSet() map[string]bool {
	tx.mu.RLock()
	defer tx.mu.RUnlock()
	
	result := make(map[string]bool)
	for key, deleted := range tx.Deleted {
		if deleted {
			result[key] = true
		}
	}
	return result
}
