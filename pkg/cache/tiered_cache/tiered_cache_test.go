package tiered_cache

import (
	"testing"

	"github.com/minotar/imgd/pkg/cache"
	"github.com/minotar/imgd/pkg/cache/lru_cache"
	"github.com/minotar/imgd/pkg/cache/util/test_helpers"
	"github.com/minotar/imgd/pkg/util/log"
)

func newBackendCache(t *testing.T, clock *test_helpers.MockClock, name string, size int) *lru_cache.LruCache {
	logger := &log.DummyLogger{}
	logger.Named(name)
	cache, err := lru_cache.NewLruCache(lru_cache.NewLruCacheConfig(size,
		cache.CacheConfig{
			Name:   name,
			Logger: logger,
		},
	))

	if err != nil {
		t.Fatalf("Error creating LruCache: %s", err)
	}

	cache.MemoryExpiry.Clock = clock
	return cache
}

func newCacheTester(t *testing.T, size int) test_helpers.CacheTester {
	clock := test_helpers.MockedUTC()

	c1 := newBackendCache(t, clock, "cache0", size/2)
	c2 := newBackendCache(t, clock, "cache1", size)

	logger := &log.DummyLogger{}
	logger.Named("TieredCacheTest")

	cache, err := NewTieredCache(&TieredCacheConfig{
		Caches: []cache.Cache{c1, c2},
		CacheConfig: cache.CacheConfig{
			Name:   "TieredCacheTest",
			Logger: logger,
		},
	})
	if err != nil {
		t.Fatalf("Error creating TieredCache: %s", err)
	}

	cache.Start()

	removeExpired := func() {
		c1.RemoveExpired()
		c2.RemoveExpired()
	}

	return test_helpers.CacheTester{
		Tester:        t,
		Cache:         cache,
		RemoveExpired: removeExpired,
		Clock:         clock,
		Iterations:    size,
	}
}

func TestInsertTTLAndRetrieve(t *testing.T) {
	cacheTester := newCacheTester(t, 500)
	defer cacheTester.Cache.Close()

	// Needs a cache at least 500 big
	test_helpers.InsertTTLAndRetrieve(cacheTester)
}

func TestLRUInsertTTLAndRetrieve(t *testing.T) {
	cacheTester := newCacheTester(t, 500)
	defer cacheTester.Cache.Close()

	hotKeys := test_helpers.LRUInsertTTLAndRetrieve(cacheTester)

	// Todo: does it make sense to further check cache1 has the keys present??
	_ = hotKeys

}

// Todo: this is lazy code coverage - these tests need thought for TieredCache..!

func TestInsertTTLAndRemove(t *testing.T) {
	cacheTester := newCacheTester(t, 500)
	defer cacheTester.Cache.Close()

	test_helpers.InsertTTLAndRemove(cacheTester)
}

func TestInsertTTLAndExpiry(t *testing.T) {
	cacheTester := newCacheTester(t, 500)
	defer cacheTester.Cache.Close()

	test_helpers.InsertTTLAndExpiry(cacheTester)
}

func TestInsertTTLAndTTLCheck(t *testing.T) {
	cacheTester := newCacheTester(t, 500)
	defer cacheTester.Cache.Close()

	test_helpers.InsertTTLAndTTLCheck(cacheTester)
}

func TestInsertTTLAndFlush(t *testing.T) {
	cacheTester := newCacheTester(t, 500)
	defer cacheTester.Cache.Close()

	test_helpers.InsertTTLAndFlush(cacheTester)
}
