package tiered_cache

import (
	"testing"

	"github.com/minotar/imgd/pkg/cache"
	"github.com/minotar/imgd/pkg/cache/lru_cache"
	"github.com/minotar/imgd/pkg/cache/util/test_helpers"
)

func newBackendCache(t *testing.T, clock *test_helpers.MockClock, size int) *lru_cache.LruCache {
	cache, err := lru_cache.NewLruCache(size)
	if err != nil {
		t.Fatalf("Error creating LruCache: %s", err)
	}

	cache.MemoryExpiry.Clock = clock
	return cache
}

func newCacheTester(t *testing.T, size int) test_helpers.CacheTester {
	clock := test_helpers.MockedUTC()

	c1 := newBackendCache(t, clock, size/2)
	c2 := newBackendCache(t, clock, size)

	cache, err := NewTieredCache([]cache.Cache{c1, c2})
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
