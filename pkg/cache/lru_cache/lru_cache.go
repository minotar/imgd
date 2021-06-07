package lru_cache

import (
	"time"

	"github.com/minotar/imgd/pkg/cache"
	expiry "github.com/minotar/imgd/pkg/cache/util/memory_expiry"
	"github.com/minotar/imgd/pkg/storage/lru_store"
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

	return lc, nil
}

func (lc *LruCache) InsertTTL(key string, value []byte, ttl time.Duration) error {
	err := lc.Insert(key, value)
	if err != nil {
		return err
	}

	// TTL of 0 has no expiry
	if ttl > time.Duration(0) {
		lc.Expiry.AddExpiry(key, ttl)
	}
	return nil
}

func (lc *LruCache) TTL(key string) (time.Duration, error) {
	_, err := lc.LruStore.Retrieve(key)
	if err != nil {
		return 0, nil
	}

	return lc.Expiry.GetTTL(key), nil

}

func (lc *LruCache) Start() {
	// Todo: Add running check?
	// Start the Expiry monitor
	lc.Expiry.Start()
}

func (lc *LruCache) Stop() {
	// Todo: Add running check?
	lc.Expiry.Stop()
}

func (lc *LruCache) Close() {
	lc.Stop()
	return
}
