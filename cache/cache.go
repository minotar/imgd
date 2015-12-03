package cache

import (
	"time"
)

// The cache is the basic interface implemented by associated stores.
type Cache interface {
	// Puts a new entry into the cache.
	Insert(key string, value []byte, ttl time.Duration) error
	// Attempts to find a value in the cache. Returns nil if
	// it does not exist.
	Find(key string) ([]byte, error)
	// Empties the cache
	Flush() error
	// Close the cache store and any associated resources.
	Close()
}
