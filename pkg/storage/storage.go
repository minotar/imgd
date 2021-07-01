package storage

import (
	"bytes"
	"encoding/gob"
	"errors"
)

type Storage interface {
	// Insert a new value into the store
	Insert(key string, value []byte) error
	// Retrieve will attempt to find the key in the store. Returns
	// nil if it does not exist with an ErrNotFound
	Retrieve(key string) ([]byte, error)
	// Remove will silently attempt to delete the key from the store
	Remove(key string) error
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
	ErrNotFound = errors.New("key does not exist")
)

func InsertKV(store Storage, key, value string) error {
	return store.Insert(key, []byte(value))
}

func RetrieveKV(store Storage, key string) (string, error) {
	respBytes, err := store.Retrieve(key)
	return string(respBytes), err
}

func InsertGob(store Storage, key string, e interface{}) error {
	var bytes *bytes.Buffer

	bytes, err := EncodeGob(e)
	if err != nil {
		return err
	}

	return store.Insert(key, bytes.Bytes())
}

func RetrieveGob(store Storage, key string, e interface{}) error {
	respBytes, err := store.Retrieve(key)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(respBytes)
	dec := gob.NewDecoder(reader)
	if err = dec.Decode(e); err != nil {
		return errors.New("RetrieveGob: " + err.Error())
	}
	return nil
}

func EncodeGob(e interface{}) (*bytes.Buffer, error) {
	var bytes bytes.Buffer
	enc := gob.NewEncoder(&bytes)

	if err := enc.Encode(e); err != nil {
		return nil, errors.New("InsertGob: " + err.Error())
	}

	return &bytes, nil
}
