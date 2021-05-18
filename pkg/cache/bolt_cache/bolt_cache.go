package bolt_cache

import (
	"fmt"
	"time"

	"github.com/boltdb/bolt"
	"github.com/minotar/imgd/pkg/cache"
	"github.com/minotar/imgd/pkg/storage/bolt_store"
)

type BoltCache struct {
	*bolt_store.BoltStore
	nameExpiry string
}

// ensure that the storage.Storage interface is implemented
var _ cache.Cache = new(BoltCache)

func NewBoltCache(path, name string) (*BoltCache, error) {
	nameExpiry := fmt.Sprintf("%s_expiry", name)

	bs, err := bolt_store.NewBoltStore(path, name)
	if err != nil {
		return nil, err
	}

	err = bs.DB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(nameExpiry))
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("Unable to create bucket \"%s\": %s", nameExpiry, err)
	}

	bc := &BoltCache{
		BoltStore:  bs,
		nameExpiry: nameExpiry,
	}

	return bc, nil
}

func (bc *BoltCache) InsertTTL(key string, value []byte, ttl time.Duration) error {
	err := bc.DB.Update(func(tx *boltb.Tx) error {
		keyBytes := []byte(key)
		// Todo
		e := tx.Bucket([]byte(bc.nameExpiry))

		b := tx.Bucket([]byte(bc.Name))
		return b.Put([]byte(key), value)
	})
	if err != nil {
		return fmt.Errorf("Inserting \"%s\" into \"%s\": %s", key, bs.name, err)
	}
	return nil
}

func (bc *BoltCache) Remove(key string) error {
	err := bc.DB.Update(func(tx *boltb.Tx) error {
		b := tx.Bucket([]byte(bc.Name))
		return b.Delete([]byte(key))
	})
	if err != nil {
		return fmt.Errorf("Removing \"%s\" from \"%s\": %s", key, bs.name, err)
	}
	return nil
}
