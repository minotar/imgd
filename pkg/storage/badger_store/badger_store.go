package badger_store

import (
	"fmt"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/minotar/imgd/pkg/storage"
	"github.com/minotar/imgd/pkg/util/log"
)

// BadgerStore does not handle any GC (eg. BadgerStore.DB.RunValueLogGC())
type BadgerStore struct {
	DB   *badger.DB
	path string
}

// ensure that the storage.Storage interface is implemented
var _ storage.Storage = new(BadgerStore)

func NewBadgerStore(path string, logger log.Logger) (*BadgerStore, error) {
	loggerWithWarning := log.NewShimLoggerWarning(logger)
	opts := badger.DefaultOptions(path)

	// Tuning ideas from https://github.com/dgraph-io/badger/issues/1304#issuecomment-630078745
	// Default 5
	opts = opts.WithNumMemtables(2)
	// Default 5
	opts = opts.WithNumLevelZeroTables(3)
	// Default 10
	opts = opts.WithNumLevelZeroTablesStall(6)
	opts = opts.WithLogger(loggerWithWarning)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	bs := &BadgerStore{
		DB:   db,
		path: path,
	}

	return bs, nil
}

func (bs *BadgerStore) Insert(key string, value []byte) error {
	err := bs.DB.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), value)
	})
	if err != nil {
		return fmt.Errorf("inserting \"%s\": %s", key, err)
	}
	return nil
}

func (bs *BadgerStore) Retrieve(key string) ([]byte, error) {
	var data []byte

	err := bs.DB.View(func(txn *badger.Txn) error {
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
		return nil, fmt.Errorf("retrieving \"%s\": %s", key, err)
	}

	return data, nil
}

func (bs *BadgerStore) Key(key string) (bool, error) {
	err := bs.DB.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return nil
	})

	if err == badger.ErrKeyNotFound {
		return false, storage.ErrNotFound
	} else if err != nil {
		return false, fmt.Errorf("retrieving \"%s\": %s", key, err)
	}
	// Key is present
	return true, nil
}

func (bs *BadgerStore) Remove(key string) error {
	err := bs.DB.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
	if err != nil {
		return fmt.Errorf("removing \"%s\": %s", key, err)
	}

	return nil
}

func (bs *BadgerStore) Flush() error {
	err := bs.DB.DropAll()
	if err != nil {
		return fmt.Errorf("flushing: %s", err)
	}
	return nil
}

func (bs *BadgerStore) Len() uint {
	iterOpts := badger.DefaultIteratorOptions
	iterOpts.PrefetchValues = false

	var len uint
	err := bs.DB.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(iterOpts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			len++
		}
		return nil
	})
	if err != nil {
		bs.DB.Opts().Logger.Errorf("Encountered an error counting keys: %v", err)
		return 0
	}
	return len
}

func (bs *BadgerStore) Size() uint64 {
	lsm, vlog := bs.DB.Size()
	return uint64(lsm + vlog)
}

func (bs *BadgerStore) Close() {
	bs.DB.Close()
}
