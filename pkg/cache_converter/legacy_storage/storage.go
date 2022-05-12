package legacy_storage

import (
	"bytes"
	"encoding/gob"
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
// Todo: Standardise on how errors are returned when key is not in store
type Storage interface {
	// Insert a new value into the store
	Insert(key string, value []byte, ttl time.Duration) error
	// Retrieve will attempt to find the key in the store. Returns
	// nil if it does not exist with an ErrNotFound
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

func InsertKV(cache Storage, key, value string, ttl time.Duration) error {
	return cache.Insert(key, []byte(value), ttl)
}

func RetrieveKV(cache Storage, key string) (string, error) {
	respBytes, err := cache.Retrieve(key)
	return string(respBytes), err
}

func InsertGob(cache Storage, key string, e interface{}, ttl time.Duration) error {
	var bytes bytes.Buffer
	enc := gob.NewEncoder(&bytes)

	err := enc.Encode(e)
	if err != nil {
		return errors.New("InsertGob: " + err.Error())
	}

	return cache.Insert(key, bytes.Bytes(), ttl)
}

func RetrieveGob(cache Storage, key string, e interface{}) error {
	respBytes, err := cache.Retrieve(key)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(respBytes)
	dec := gob.NewDecoder(reader)
	err = dec.Decode(e)
	if err != nil {
		return errors.New("RetrieveGob: " + err.Error())
	}
	return nil
}
