package nutsdb_store

import (
	"fmt"

	"github.com/minotar/imgd/pkg/storage"
	"github.com/xujiajun/nutsdb"
)

type NutsDBStore struct {
	DB   *nutsdb.DB
	dir  string
	Name string
}

// ensure that the storage.Storage interface is implemented
var _ storage.Storage = new(NutsDBStore)

func NewNutsDBStore(dir, name string) (*NutsDBStore, error) {
	opt := nutsdb.DefaultOptions
	opt.Dir = dir
	// Todo: test this!
	opt.EntryIdxMode = nutsdb.HintBPTSparseIdxMode

	db, err := nutsdb.Open(opt)
	if err != nil {
		return nil, err
	}

	ns := &NutsDBStore{
		DB:   db,
		dir:  dir,
		Name: name,
	}

	return ns, nil
}

func (ns *NutsDBStore) Insert(key string, value []byte) error {
	err := ns.DB.Update(func(tx *nutsdb.Tx) error {
		return tx.Put(ns.Name, []byte(key), value, 0)
	})
	if err != nil {
		return fmt.Errorf("Inserting \"%s\" into \"%s\": %s", key, ns.Name, err)
	}
	return nil
}

func (ns *NutsDBStore) Retrieve(key string) ([]byte, error) {
	var data []byte

	err := ns.DB.View(func(tx *nutsdb.Tx) error {
		if res, err := tx.Get(ns.Name, []byte(key)); err != nil {
			return err
		} else {
			// this copy might not be needed?
			data = make([]byte, len(res.Value))
			copy(data, res.Value)
		}
		return nil
	})
	if err == nutsdb.ErrNotFoundKey {
		return nil, storage.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("Retrieving \"%s\" from \"%s\": %s", key, ns.Name, err)
	}

	return data, nil
}

func (ns *NutsDBStore) Remove(key string) error {
	err := ns.DB.Update(func(tx *nutsdb.Tx) error {
		return tx.Delete(ns.Name, []byte(key))
	})
	if err != nil {
		return fmt.Errorf("Removing \"%s\" from \"%s\": %s", key, ns.Name, err)
	}
	return nil
}

func (ns *NutsDBStore) Flush() error {
	// https://github.com/xujiajun/nutsdb/issues/68
	err := ns.DB.Update(func(tx *nutsdb.Tx) error {
		entries, err := tx.GetAll(ns.Name)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			if err := tx.Delete(ns.Name, entry.Key); err != nil {
				fmt.Printf("Removing \"%s\" from \"%s\": %s", string(entry.Key), ns.Name, err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("Flushing \"%s\": %s", ns.Name, err)
	}
	return nil
}

func (ns *NutsDBStore) Len() uint {
	return 0
}

func (ns *NutsDBStore) Size() uint64 {
	return 0
}

func (ns *NutsDBStore) Close() {
	ns.DB.Close()
}
