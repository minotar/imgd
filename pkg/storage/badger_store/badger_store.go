package badger_store

import (
	"fmt"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/minotar/imgd/pkg/storage"
)

type BadgerStore struct {
	db   *badger.DB
	path string
}

// ensure that the storage.Storage interface is implemented
var _ storage.Storage = new(BadgerStore)

func NewBadgerStore(path string) (*BadgerStore, error) {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return nil, err
	}

	bs := &BadgerStore{
		db:   db,
		path: path,
	}

	return bs, nil
}

func (bs *BadgerStore) Insert(key string, value []byte) error {
	err := bs.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), value)
	})
	if err != nil {
		return fmt.Errorf("Inserting \"%s\": %s", key, err)
	}
	return nil
}

func (bs *BadgerStore) Retrieve(key string) ([]byte, error) {
	var data []byte

	err := bs.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		data, err = item.ValueCopy(nil)
		return err
	})
	if err == badger.ErrKeyNotFound {
		return nil, storage.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("Retrieving \"%s\": %s", key, err)
	}

	return data, nil
}

func (bs *BadgerStore) Remove(key string) error {
	err := bs.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
	if err != nil {
		return fmt.Errorf("Removing \"%s\": %s", key, err)
	}

	return nil
}

func (bs *BadgerStore) Flush() error {
	err := bs.db.DropAll()
	if err != nil {
		return fmt.Errorf("Flushing: %s", err)
	}
	return nil
}

func (bs *BadgerStore) Len() uint {
	return 0
}

func (bs *BadgerStore) Size() uint64 {
	return 0
}

func (bs *BadgerStore) Close() {
	bs.db.Close()
}
