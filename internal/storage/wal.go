package storage

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type LogEntryType uint8

const (
	LogEntryPut    LogEntryType = 1
	LogEntryDelete LogEntryType = 2
	LogEntryCommit LogEntryType = 3
)

type LogEntry struct {
	Type      LogEntryType `json:"type"`
	Key       string       `json:"key"`
	Value     []byte       `json:"value,omitempty"`
	Timestamp int64        `json:"timestamp"`
	Checksum  uint32       `json:"checksum"`
}

type WAL struct {
	filePath string
	file     *os.File
	mutex    sync.RWMutex
	closed   bool
}

func NewWAL(filePath string) (*WAL, error) {
	wal := &WAL{
		filePath: filePath,
	}

	if err := wal.open(); err != nil {
		return nil, fmt.Errorf("failed to open WAL: %w", err)
	}

	return wal, nil
}

func (w *WAL) open() error {
	if w.closed {
		return ErrStorageClosed
	}

	dir := filepath.Dir(w.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create WAL directory: %w", err)
	}

	file, err := os.OpenFile(w.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open WAL file: %w", err)
	}

	w.file = file
	return nil
}

func (w *WAL) LogPut(key string, value []byte) error {
	return w.logEntry(LogEntry{
		Type:      LogEntryPut,
		Key:       key,
		Value:     value,
		Timestamp: time.Now().UnixNano(),
	})
}

func (w *WAL) LogDelete(key string) error {
	return w.logEntry(LogEntry{
		Type:      LogEntryDelete,
		Key:       key,
		Timestamp: time.Now().UnixNano(),
	})
}

func (w *WAL) LogCommit() error {
	return w.logEntry(LogEntry{
		Type:      LogEntryCommit,
		Timestamp: time.Now().UnixNano(),
	})
}

func (w *WAL) logEntry(entry LogEntry) error {
	if w.closed {
		return ErrStorageClosed
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()

	entry.Checksum = w.calculateChecksum(entry)

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	length := uint32(len(data))
	if err := binary.Write(w.file, binary.LittleEndian, length); err != nil {
		return fmt.Errorf("failed to write entry length: %w", err)
	}

	if _, err := w.file.Write(data); err != nil {
		return fmt.Errorf("failed to write entry data: %w", err)
	}

	if err := w.file.Sync(); err != nil {
		return fmt.Errorf("failed to sync WAL: %w", err)
	}

	return nil
}

func (w *WAL) Replay(engine Engine) error {
	if w.closed {
		return ErrStorageClosed
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.file != nil {
		w.file.Close()
		w.file = nil
	}

	file, err := os.Open(w.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return w.open()
		}
		return fmt.Errorf("failed to open WAL for replay: %w", err)
	}
	defer file.Close()

	for {
		entry, err := w.readEntry(file)
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read log entry: %w", err)
		}

		if !w.verifyChecksum(entry) {
			return fmt.Errorf("checksum verification failed for entry at key: %s", entry.Key)
		}

		if err := w.applyEntry(engine, entry); err != nil {
			return fmt.Errorf("failed to apply log entry: %w", err)
		}
	}

	return w.open()
}

func (w *WAL) readEntry(file *os.File) (*LogEntry, error) {
	var length uint32
	if err := binary.Read(file, binary.LittleEndian, &length); err != nil {
		return nil, err
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(file, data); err != nil {
		return nil, err
	}

	var entry LogEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}

	return &entry, nil
}

func (w *WAL) applyEntry(engine Engine, entry *LogEntry) error {
	switch entry.Type {
	case LogEntryPut:
		return engine.Put(entry.Key, entry.Value)
	case LogEntryDelete:
		return engine.Delete(entry.Key)
	case LogEntryCommit:
		return nil
	default:
		return fmt.Errorf("unknown log entry type: %d", entry.Type)
	}
}

func (w *WAL) calculateChecksum(entry LogEntry) uint32 {
	checksum := uint32(entry.Type)
	for _, b := range []byte(entry.Key) {
		checksum += uint32(b)
	}
	for _, b := range entry.Value {
		checksum += uint32(b)
	}
	checksum += uint32(entry.Timestamp & 0xFFFFFFFF)
	return checksum
}

func (w *WAL) verifyChecksum(entry *LogEntry) bool {
	expectedChecksum := w.calculateChecksum(*entry)
	return entry.Checksum == expectedChecksum
}

func (w *WAL) Close() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.closed {
		return nil
	}

	w.closed = true

	if w.file != nil {
		if err := w.file.Sync(); err != nil {
			return fmt.Errorf("failed to sync WAL before close: %w", err)
		}
		return w.file.Close()
	}

	return nil
}

func (w *WAL) Truncate() error {
	if w.closed {
		return ErrStorageClosed
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.file != nil {
		w.file.Close()
	}

	if err := os.Truncate(w.filePath, 0); err != nil {
		return fmt.Errorf("failed to truncate WAL: %w", err)
	}

	return w.open()
}
