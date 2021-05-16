package lru

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
	freshStore, err := lru.New(maxEntries)
	if err != nil {
		return nil, err
	}
	lc := &LruStore{
		store: freshStore,
	}

	return lc, nil
}

func (l *LruStore) Insert(key string, value []byte) error {
	l.store.Add(key, value)
	return nil
}

func (l *LruStore) Retrieve(key string) ([]byte, error) {
	if value, ok := l.store.Get(key); ok {
		return value.([]byte), nil
	}
	return nil, storage.ErrNotFound
}

func (l *LruStore) Flush() error {
	l.store.Purge()
	return nil
}

func (l *LruStore) Len() uint {
	return uint(l.store.Len())
}

func (l *LruStore) Size() uint64 {
	return 0
}

func (l *LruStore) Close() {
	return
}
