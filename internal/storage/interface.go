package storage

type Engine interface {
	Get(key string) ([]byte, error)
	Put(key string, value []byte) error
	Delete(key string) error
	Exists(key string) (bool, error)
	Keys() ([]string, error)
	Close() error
	BeginTransaction() *Transaction
	CommitTransaction(tx *Transaction) error
	AbortTransaction(tx *Transaction) error
}

type WALEngine interface {
	Engine
	Checkpoint() error
	Recover() error
	GetWALPath() string
}