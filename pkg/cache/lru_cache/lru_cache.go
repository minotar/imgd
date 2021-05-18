package lru_cache

import (
	"time"

	"github.com/minotar/imgd/pkg/cache"
	"github.com/minotar/imgd/pkg/storage/lru_store"
	"github.com/minotar/imgd/pkg/storage/util/expiry"
)

type LruCache struct {
	*lru_store.LruStore
	*expiry.Expiry
}

var _ cache.Cache = new(LruCache)

func NewLruCache(maxEntries int) (*LruCache, error) {
	// Start with empty struct we can pass around
	lc := &LruCache{}
	// Pass in the Remove function which it will call whenever an item expires
	lc.Expiry = expiry.NewExpiry(lc.Remove)

	// Pass the Expiry special Function to the LRU initilization
	ls, err := lru_store.NewLruStoreWithEvict(maxEntries, lc.Expiry.RemoveExpiry)
	if err != nil {
		return nil, err
	}

	lc.LruStore = ls
	// Start the Expiry monitor once everything is handled
	lc.Expiry.Start()

	return lc, nil
}

func (lc *LruCache) InsertTTL(key string, value []byte, ttl time.Duration) error {
	err := lc.Insert(key, value)
	if err != nil {
		return err
	}
	lc.Expiry.AddExpiry(key, ttl)
	return nil
}

func (lc *LruCache) Close() {
	lc.Expiry.Stop()
	return
}
