package lru_cache

import (
	"testing"

	"github.com/minotar/imgd/pkg/cache"
	"github.com/minotar/imgd/pkg/cache/util/test_helpers"
	"github.com/minotar/imgd/pkg/util/log"
)

func newCacheTester(t *testing.T, size int) test_helpers.CacheTester {
	logger := &log.DummyLogger{}
	logger.Named("LruTest")
	cache, err := NewLruCache(&LruCacheConfig{
		size: size,
		CacheConfig: cache.CacheConfig{
			Name:   "LruTest",
			Logger: logger,
		},
	})
	if err != nil {
		t.Fatalf("Error creating LruCache: %s", err)
	}

	clock := test_helpers.MockedUTC()
	cache.MemoryExpiry.Clock = clock
	cache.Start()

	return test_helpers.CacheTester{
		Tester:        t,
		Cache:         cache,
		RemoveExpired: cache.RemoveExpired,
		Clock:         clock,
		Iterations:    size,
	}
}

func TestInsertAndRetrieve(t *testing.T) {
	cacheTester := newCacheTester(t, 500)
	defer cacheTester.Cache.Close()

	test_helpers.InsertAndRetrieve(cacheTester)
}

func TestInsertTTLAndRetrieve(t *testing.T) {
	cacheTester := newCacheTester(t, 500)
	defer cacheTester.Cache.Close()

	test_helpers.InsertTTLAndRetrieve(cacheTester)
}

func TestLRUInsertTTLAndRetrieve(t *testing.T) {
	cacheTester := newCacheTester(t, 500)
	defer cacheTester.Cache.Close()

	test_helpers.LRUInsertTTLAndRetrieve(cacheTester)
}

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

// Todo: test filled cached

/*
func TestInsertAndRetrieve(t *testing.T) {
	cache := freshCache(10)
	defer cache.Close()

	for i := 0; i < 10; i++ {
		key := test_helpers.RandString(32)
		cache.InsertTTL(key, []byte(strconv.Itoa(i)), time.Minute)
		item, _ := cache.Retrieve(key)
		if string(item) != strconv.Itoa(i) {
			t.Fail()
		}
	}
}

func TestHousekeeping(t *testing.T) {
	cache := freshCache(5)
	defer cache.Close()

	for i := 0; i < 10; i++ {
		key := test_helpers.RandString(32)
		cache.InsertTTL(key, []byte("foobar"), time.Minute)
	}

	cacheLen := cache.Len()
	expiryLen := cache.MemoryExpiry.Len()

	if cacheLen != expiryLen || cacheLen != 5 {
		t.Errorf("Cache Length %d and Expiry Length %d should be 5", cacheLen, expiryLen)
	}

	cache.Flush()
	cacheLen = cache.Len()
	expiryLen = cache.MemoryExpiry.Len()
	if cacheLen != expiryLen || cacheLen != 0 {
		t.Errorf("Cache Length %d and Expiry Length %d should be 0", cacheLen, expiryLen)
	}
}

var largeBucket = test_store.NewTestStoreBench()

func insertTTL(cache *LruCache) func(string, []byte) error {
	return func(s string, b []byte) error {
		return cache.InsertTTL(s, b, time.Minute)
	}
}

func BenchmarkInsert(b *testing.B) {
	cache := freshCache(b.N)
	defer cache.Close()

	largeBucket.MinSize(b.N)
	b.ResetTimer()

	largeBucket.FillStore(insertTTL(cache), b.N)
}

func BenchmarkRemove(b *testing.B) {
	cache := freshCache(b.N)
	defer cache.Close()

	// Set TestBucket and Cache based on a static size (b.N should only affect loop)
	largeBucket.MinSize(b.N)
	largeBucket.FillStore(insertTTL(cache), b.N)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cache.Remove(largeBucket.Keys[i])
	}
}


*/
