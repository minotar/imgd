package bolt_store

import (
	"fmt"
	"os"
	"time"

	bolt "github.com/boltdb/bolt"
	"github.com/minotar/imgd/pkg/storage"
)

type BoltStore struct {
	DB     *bolt.DB
	path   string
	Bucket string
}

// ensure that the storage.Storage interface is implemented
var _ storage.Storage = new(BoltStore)

func NewBoltStore(path, bucketname string) (*BoltStore, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketname))
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("Unable to create bucket \"%s\": %s", bucketname, err)
	}

	bs := &BoltStore{
		DB:     db,
		path:   path,
		Bucket: bucketname,
	}

	return bs, nil
}

func (bs *BoltStore) Insert(key string, value []byte) error {
	err := bs.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bs.Bucket))
		return b.Put([]byte(key), value)
	})
	if err != nil {
		return fmt.Errorf("inserting \"%s\" into \"%s\": %s", key, bs.Bucket, err)
	}
	return nil
}

// InsertBatch seems like it's not worth it...
func (bs *BoltStore) InsertBatch(key string, value []byte) error {
	err := bs.DB.Batch(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bs.Bucket))
		return b.Put([]byte(key), value)
	})
	if err != nil {
		return fmt.Errorf("inserting \"%s\" into \"%s\": %s", key, bs.Bucket, err)
	}
	return nil
}

func (bs *BoltStore) Retrieve(key string) ([]byte, error) {
	var data []byte

	err := bs.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bs.Bucket))
		v := b.Get([]byte(key))
		if v == nil {
			return storage.ErrNotFound
		}
		// Set byte slice length for copy
		data = make([]byte, len(v))
		copy(data, v)
		return nil
	})
	if err == storage.ErrNotFound {
		return nil, storage.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("retrieving \"%s\" from \"%s\": %s", key, bs.Bucket, err)
	}

	return data, nil
}

func (bs *BoltStore) Remove(key string) error {
	err := bs.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bs.Bucket))
		return b.Delete([]byte(key))
	})
	if err != nil {
		return fmt.Errorf("removing \"%s\" from \"%s\": %s", key, bs.Bucket, err)
	}
	return nil
}

func (bs *BoltStore) Flush() error {
	err := bs.DB.Update(func(tx *bolt.Tx) error {
		if err := tx.DeleteBucket([]byte(bs.Bucket)); err != nil {
			return err
		}
		_, err := tx.CreateBucketIfNotExists([]byte(bs.Bucket))
		return err
	})
	if err != nil {
		return fmt.Errorf("flushing \"%s\": %s", bs.Bucket, err)
	}
	return nil
}

func (bs *BoltStore) Len() uint {
	var keyCount uint

	bs.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bs.Bucket))
		stats := b.Stats()
		keyCount = uint(stats.KeyN)
		return nil
	})

	return keyCount
}

func (bs *BoltStore) Size() uint64 {
	fileInfo, err := os.Stat(bs.path)
	if err != nil {
		return 0
	}
	return uint64(fileInfo.Size())
}

func (bs *BoltStore) Close() {
	bs.DB.Close()
}
