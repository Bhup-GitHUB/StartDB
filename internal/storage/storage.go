package storage

// Storage provides a high-level interface for database operations
type Storage struct {
	engine Engine
}

// New creates a new storage instance with the given engine
func New(engine Engine) *Storage {
	return &Storage{
		engine: engine,
	}
}

// Get retrieves a value by key
func (s *Storage) Get(key string) ([]byte, error) {
	return s.engine.Get(key)
}

// Put stores a key-value pair
func (s *Storage) Put(key string, value []byte) error {
	return s.engine.Put(key, value)
}

// Delete removes a key-value pair
func (s *Storage) Delete(key string) error {
	return s.engine.Delete(key)
}

// Exists checks if a key exists
func (s *Storage) Exists(key string) (bool, error) {
	return s.engine.Exists(key)
}

// Close shuts down the storage
func (s *Storage) Close() error {
	return s.engine.Close()
}
