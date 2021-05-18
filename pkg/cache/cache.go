package cache

import (
	"bytes"
	"time"

	"github.com/minotar/imgd/pkg/storage"
)

type Cache interface {
	// Implements storage.Storage interface
	storage.Storage
	// Insert a new value into the store
	InsertTTL(key string, value []byte, ttl time.Duration) error
}

func InsertKV(cache Cache, key, value string, ttl time.Duration) error {
	return cache.InsertTTL(key, []byte(value), ttl)
}

func InsertGob(cache Cache, key string, e interface{}, ttl time.Duration) error {
	var bytes *bytes.Buffer

	bytes, err := storage.EncodeGob(e)
	if err != nil {
		return err
	}

	return cache.InsertTTL(key, bytes.Bytes(), ttl)
}

var (
	RetrieveKV  = storage.RetrieveKV
	RetrieveGob = storage.RetrieveGob
)
