package lru_cache

import (
	"time"

	"github.com/minotar/imgd/pkg/cache"
	memory_expiry "github.com/minotar/imgd/pkg/cache/util/expiry/memory"
	"github.com/minotar/imgd/pkg/storage/lru_store"
)

type LruCache struct {
	*lru_store.LruStore
	*memory_expiry.MemoryExpiry
	*LruCacheConfig
}

type LruCacheConfig struct {
	cache.CacheConfig
	size int
}

func NewLruCacheConfig(size int, cacheCfg cache.CacheConfig) *LruCacheConfig {
	return &LruCacheConfig{
		size:        size,
		CacheConfig: cacheCfg,
	}
}

var _ cache.Cache = new(LruCache)

func NewLruCache(cfg *LruCacheConfig) (*LruCache, error) {
	cfg.Logger.Infof("initializing LruCache with size %d", cfg.size)
	// Start with empty struct we can pass around
	lc := &LruCache{LruCacheConfig: cfg}

	// Pass in the lruRemove function which it will call whenever an item expires
	// This cannot be the lru_store Remove func directly as that is currently nil (🐣)
	me, err := memory_expiry.NewMemoryExpiry(lc.lruRemove)
	if err != nil {
		return nil, err
	}
	lc.MemoryExpiry = me

	// Pass the Expiry special Function to the LRU initilization
	ls, err := lru_store.NewLruStoreWithEvict(lc.LruCacheConfig.size, lc.MemoryExpiry.RemoveExpiry)
	if err != nil {
		return nil, err
	}
	lc.LruStore = ls

	cfg.Logger.Infof("initialized LruCache \"%s\"", lc.Name())
	return lc, nil
}

func (lc *LruCache) Name() string {
	return lc.CacheConfig.Name
}

// lruRemove allows us to pass a reference to the LruStore Remove function before we instatiate it
func (lc *LruCache) lruRemove(key string) error {
	lc.Logger.Debug("LruStore is evicting ", key)
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
		return 0, cache.ErrNoExpiry
	}
	return ttl, nil
}

// Both MemoryExpiry and LruStore have Len methods - we need to specify which here
func (lc *LruCache) Len() uint {
	return lc.LruStore.Len()
}

func (lc *LruCache) Start() {
	lc.Logger.Info("starting LruCache")
	// Start the Expiry monitor/compactor
	lc.MemoryExpiry.Start()
}

func (lc *LruCache) Stop() {
	lc.Logger.Info("stopping LruCache")
	// Stop the Expiry monitor/compactor
	lc.MemoryExpiry.Stop()
}

func (lc *LruCache) Close() {
	lc.Logger.Debug("closing LruCache")
	lc.Stop()
}
