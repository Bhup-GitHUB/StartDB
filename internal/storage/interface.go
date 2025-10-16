package storage

type Engine interface {
	Get(key string) ([]byte, error)
	Put(key string, value []byte) error
	Delete(key string) error
	Exists(key string) (bool, error)
	Keys() ([]string, error)
	Close() error
}
