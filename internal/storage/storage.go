package storage

type Storage struct {
	engine Engine
	txManager *TransactionManager
	indexManager *IndexManager
}

func New(engine Engine) *Storage {
	return &Storage{
		engine: engine,
		txManager: NewTransactionManager(),
		indexManager: NewIndexManager(),
	}
}

func (s *Storage) Get(key string) ([]byte, error) {
	return s.engine.Get(key)
}

func (s *Storage) Put(key string, value []byte) error {
	return s.engine.Put(key, value)
}

func (s *Storage) Delete(key string) error { 
	return s.engine.Delete(key)
}

func (s *Storage) Exists(key string) (bool, error) {
	return s.engine.Exists(key)
}

func (s *Storage) Keys() ([]string, error) {
	return s.engine.Keys()
}

func (s *Storage) Close() error {
	return s.engine.Close()
}

func (s *Storage) BeginTransaction() *Transaction {
	return s.txManager.BeginTransaction()
}

func (s *Storage) CommitTransaction(tx *Transaction) error {
	if err := s.engine.CommitTransaction(tx); err != nil {
		return err
	}
	return s.txManager.CommitTransaction(tx.ID)
}

func (s *Storage) AbortTransaction(tx *Transaction) error {
	if err := s.engine.AbortTransaction(tx); err != nil {
		return err
	}
	return s.txManager.AbortTransaction(tx.ID)
}

func (s *Storage) GetIndexManager() *IndexManager {
	return s.indexManager
}
