package storage

// Engine defines the interface for storage operations
type Engine interface {
	Get(key string) ([]byte, error)
	Put(key string, value []byte) error
	Delete(key string) error
	Exists(key string) (bool, error)
	Close() error
}
