package storage

import "errors"

var (
	ErrKeyNotFound    = errors.New("key not found")
	ErrKeyExists      = errors.New("key already exists")
	ErrInvalidKey     = errors.New("invalid key")
	ErrInvalidValue   = errors.New("invalid value")
	ErrStorageClosed  = errors.New("storage is closed")
)
