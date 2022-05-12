package lru_store

import (
	lru "github.com/hashicorp/golang-lru"
	"github.com/minotar/imgd/pkg/storage"
)

type LruStore struct {
	store *lru.Cache
}

// ensure that the storage.Storage interface is implemented
var _ storage.Storage = new(LruStore)

func NewLruStore(maxEntries int) (*LruStore, error) {
	return NewLruStoreWithEvict(maxEntries, nil)
}

func NewLruStoreWithEvict(maxEntries int, onEvicted func(key interface{}, value interface{})) (*LruStore, error) {
	freshStore, err := lru.NewWithEvict(maxEntries, onEvicted)
	if err != nil {
		return nil, err
	}
	ls := &LruStore{
		store: freshStore,
	}

	return ls, nil
}

func (ls *LruStore) Insert(key string, value []byte) error {
	ls.store.Add(key, value)
	return nil
}

func (ls *LruStore) Retrieve(key string) ([]byte, error) {
	if value, ok := ls.store.Get(key); ok {
		return value.([]byte), nil
	}
	return nil, storage.ErrNotFound
}

func (ls *LruStore) Remove(key string) error {
	ls.store.Remove(key)
	return nil
}

func (ls *LruStore) Flush() error {
	ls.store.Purge()
	return nil
}

func (ls *LruStore) Len() uint {
	return uint(ls.store.Len())
}

func (ls *LruStore) Size() uint64 {
	return 0
}

func (ls *LruStore) Close() {
	return
}
