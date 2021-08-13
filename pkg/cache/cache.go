// Provide a standardised interface for using our backend caches
// Insert/InsertTTL should be ran in goroutines to avoid waiting for results
// Retrieve/TTL functions should return the same error `ErrNotFound` when the key does not exist
// TTL functions should return `ErrNoExpiry` if they key was found, but no expiry info was found
package cache

import (
	"bytes"
	"errors"
	"flag"
	"time"

	"github.com/minotar/imgd/pkg/storage"
	util_log "github.com/minotar/imgd/pkg/util/log"
)

type Cache interface {
	// Name returns the CacheConfig Name
	Name() string
	// Implements storage.Storage interface
	storage.Storage
	// InsertTTL inserts a new value into the store with the given expiry
	InsertTTL(key string, value []byte, ttl time.Duration) error
	// TTL returns the existing TTL for a key and an optional error
	// The expectation is that the Cache will handle whether this is a
	TTL(key string) (time.Duration, error)
	// Start the cache / expiry tracker
	Start()
	// Stop the cache / expiry tracker
	Stop()
}

type CacheConfig struct {
	Name   string
	Logger util_log.Logger
}

func (c *CacheConfig) RegisterFlags(f *flag.FlagSet, cacheID string) {
	f.StringVar(&c.Name, cacheID+".name", "Cache"+cacheID, "Internal name of cache")
}

// Errors
var (
	ErrNoExpiry = errors.New("key does not have an associated Expiry/TTL")
)

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
	ErrNotFound = storage.ErrNotFound
)
