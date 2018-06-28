package storage

import (
	"errors"
	"time"
)

const (
	// Get the skin size in bytes. Stored as a []uint8, one byte each,
	// plus bounces. So 64 * 64 bytes and we'll throw in an extra 16
	// bytes of overhead.
	skinSize = (64 * 64) + 16

	// Define a 64 MB cache size.
	cacheSize = 2 << 25

	// Based off those, calculate the maximum number of skins we'll store
	// in memory.
	skinCount = cacheSize / skinSize
)

// Storage is the basic interface implemented by associated stores
type Storage interface {
	// Insert a new entry into the store
	Insert(key string, value []byte, ttl time.Duration) error
	// Retrieve will attempt to find a value in the store. Returns
	// nil if it does not exist.
	Retrieve(key string) ([]byte, error)
	// Flush will empty the store
	Flush() error
	// Len returns the number of keys in the store (eg. Length of the cache/Count of items)
	Len() uint
	// Size returns the bytes used to store the keys
	Size() uint64
	// Close the cache store and any associated resources.
	Close()
}

// Errors
var (
	ErrNotFound = errors.New("Key does not exist")
)
