package storage

type Storage struct {
	engine Engine
}

func New(engine Engine) *Storage {
	return &Storage{
		engine: engine,
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
