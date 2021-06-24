package lru_cache

import (
	"fmt"
	"time"

	"github.com/minotar/imgd/pkg/cache"
	memory_expiry "github.com/minotar/imgd/pkg/cache/util/expiry/memory"
	"github.com/minotar/imgd/pkg/storage/lru_store"
)

type LruCache struct {
	*lru_store.LruStore
	*memory_expiry.MemoryExpiry
}

var _ cache.Cache = new(LruCache)

func NewLruCache(maxEntries int) (*LruCache, error) {
	// Start with empty struct we can pass around
	lc := &LruCache{}
	// Pass in the lruRemove function which it will call whenever an item expires
	// This cannot be the lru_store Remove func directly as that is currently nil (🐣)
	me, err := memory_expiry.NewMemoryExpiry(lc.lruRemove)
	if err != nil {
		return nil, err
	}

	lc.MemoryExpiry = me

	// Pass the Expiry special Function to the LRU initilization
	ls, err := lru_store.NewLruStoreWithEvict(maxEntries, lc.MemoryExpiry.RemoveExpiry)
	if err != nil {
		return nil, err
	}

	lc.LruStore = ls

	return lc, nil
}

// lruRemove allows us to pass a reference to the LruStore Remove function before we instatiate it
func (lc *LruCache) lruRemove(key string) error {
	return lc.LruStore.Remove(key)
}

func (lc *LruCache) InsertTTL(key string, value []byte, ttl time.Duration) error {
	err := lc.Insert(key, value)
	if err != nil {
		// LruCache won't throw an error ever, so :shrugging:
		return err
	}

	// TTL of 0 has no expiry
	if ttl > time.Duration(0) {
		lc.MemoryExpiry.AddExpiry(key, ttl)
	}
	return nil
}

// TTL returns an error if the key does not exist, or it has no expiry
// Otherwise return a TTL (always at least 1 Second per `MemoryExpiry`)
func (lc *LruCache) TTL(key string) (time.Duration, error) {
	_, err := lc.LruStore.Retrieve(key)
	// Return 0 and error when the key is not in the store
	if err != nil {
		return 0, err
	}

	ttl, err := lc.MemoryExpiry.GetTTL(key)
	if err != nil {
		// Todo: use global Error in cache package? Instead log something with key name?
		return 0, fmt.Errorf("No expiry set for key \"%s\"", key)
	}
	return ttl, nil
}

// Both MemoryExpiry and LruStore have Len methods - we need to specify which here
func (lc *LruCache) Len() uint {
	return lc.LruStore.Len()
}

func (lc *LruCache) Start() {
	// Start the Expiry monitor/compactor
	lc.MemoryExpiry.Start()
}

func (lc *LruCache) Stop() {
	// Stop the Expiry monitor/compactor
	lc.MemoryExpiry.Stop()
}

func (lc *LruCache) Close() {
	lc.Stop()
}
