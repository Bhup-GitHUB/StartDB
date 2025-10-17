package storage

import (
	"fmt"
	"path/filepath"
)

type WALStorage struct {
	engine Engine
	wal    *WAL
}

func NewWALStorage(engine Engine, walPath string) (*WALStorage, error) {
	wal, err := NewWAL(walPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create WAL: %w", err)
	}

	if err := wal.Replay(engine); err != nil {
		wal.Close()
		return nil, fmt.Errorf("failed to replay WAL: %w", err)
	}

	return &WALStorage{
		engine: engine,
		wal:    wal,
	}, nil
}

func NewWALStorageWithEngine(engine Engine, walPath string) (*WALStorage, error) {
	wal, err := NewWAL(walPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create WAL: %w", err)
	}

	return &WALStorage{
		engine: engine,
		wal:    wal,
	}, nil
}

func (ws *WALStorage) Get(key string) ([]byte, error) {
	return ws.engine.Get(key)
}

func (ws *WALStorage) Put(key string, value []byte) error {
	if err := ws.wal.LogPut(key, value); err != nil {
		return fmt.Errorf("failed to log PUT operation: %w", err)
	}

	if err := ws.engine.Put(key, value); err != nil {
		return fmt.Errorf("failed to apply PUT operation: %w", err)
	}

	return nil
}

func (ws *WALStorage) Delete(key string) error {
	if err := ws.wal.LogDelete(key); err != nil {
		return fmt.Errorf("failed to log DELETE operation: %w", err)
	}

	if err := ws.engine.Delete(key); err != nil {
		return fmt.Errorf("failed to apply DELETE operation: %w", err)
	}

	return nil
}

func (ws *WALStorage) Exists(key string) (bool, error) {
	return ws.engine.Exists(key)
}

func (ws *WALStorage) Keys() ([]string, error) {
	return ws.engine.Keys()
}

func (ws *WALStorage) Close() error {
	var errs []error

	if err := ws.engine.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close engine: %w", err))
	}

	if err := ws.wal.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close WAL: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}

	return nil
}

func (ws *WALStorage) Checkpoint() error {
	return ws.wal.Truncate()
}

func (ws *WALStorage) Recover() error {
	return ws.wal.Replay(ws.engine)
}

func (ws *WALStorage) GetWALPath() string {
	return ws.wal.filePath
}

func NewWALMemoryEngine(walPath string) (*WALStorage, error) {
	engine := NewMemoryEngine()
	return NewWALStorage(engine, walPath)
}

func NewWALDiskEngine(dataPath, walPath string) (*WALStorage, error) {
	engine, err := NewDiskEngine(dataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create disk engine: %w", err)
	}

	return NewWALStorage(engine, walPath)
}

func NewWALDiskEngineWithAutoPath(dataPath string) (*WALStorage, error) {
	dir := filepath.Dir(dataPath)
	filename := filepath.Base(dataPath)
	ext := filepath.Ext(filename)
	name := filename[:len(filename)-len(ext)]
	walPath := filepath.Join(dir, name+".wal")

	return NewWALDiskEngine(dataPath, walPath)
}
