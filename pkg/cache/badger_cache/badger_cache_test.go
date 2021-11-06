package badger_cache

import (
	"testing"
	"time"

	"github.com/minotar/imgd/pkg/cache"
	store_expiry "github.com/minotar/imgd/pkg/cache/util/expiry/store"
	"github.com/minotar/imgd/pkg/cache/util/test_helpers"
	"github.com/minotar/imgd/pkg/util/log"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	TestBadgerPath = "/tmp/badger_cache_test/"
)

func newCache(t *testing.T) *BadgerCache {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()

	logger := log.NewBuiltinLogger(1)
	logger.Named("BadgerTest")

	cacheConfig := cache.CacheConfig{
		Name:   "BadgerTest",
		Logger: logger,
	}
	badgerCacheConfig := NewBadgerCacheConfig(cacheConfig, TestBadgerPath)

	cache, err := NewBadgerCache(badgerCacheConfig)
	if err != nil {
		t.Fatalf("Error creating BadgerCache: %s", err)
	}
	return cache
}

func newCacheTester(t *testing.T) test_helpers.CacheTester {
	cache := newCache(t)
	return newCacheTesterWithBadgerCache(t, cache)
}

func newCacheTesterWithBadgerCache(t *testing.T, bc *BadgerCache) test_helpers.CacheTester {
	// Create a dummy StoreExpiry to override the builtin scanner / schedule
	// This means we control when the ExpiryScan runs (we aren't trying to test the scheduller)
	se, err := store_expiry.NewStoreExpiry(func() {}, time.Minute)
	if err != nil {
		t.Fatalf("Error creating StoreExpiry: %s", err)
	}
	bc.StoreExpiry = se

	clock := test_helpers.MockedUTC()
	bc.StoreExpiry.Clock = clock
	bc.Flush()
	bc.Start()

	return test_helpers.CacheTester{
		Tester:        t,
		Cache:         bc,
		RemoveExpired: bc.ExpiryScan,
		Clock:         clock,
		// Used for later Iterations tests, so larger than 10 and divisible by 10
		// Controls test speed
		Iterations: 100,
	}
}

func TestInsertAndRetrieve(t *testing.T) {
	cacheTester := newCacheTester(t)
	defer cacheTester.Cache.Close()

	test_helpers.InsertAndRetrieve(cacheTester)
}

func TestInsertTTLAndRetrieve(t *testing.T) {
	cacheTester := newCacheTester(t)
	defer cacheTester.Cache.Close()

	test_helpers.InsertTTLAndRetrieve(cacheTester)
}

func TestInsertTTLAndRemove(t *testing.T) {
	cacheTester := newCacheTester(t)
	defer cacheTester.Cache.Close()

	test_helpers.InsertTTLAndRemove(cacheTester)
}

func TestInsertTTLAndFlush(t *testing.T) {
	cacheTester := newCacheTester(t)
	defer cacheTester.Cache.Close()

	test_helpers.InsertTTLAndFlush(cacheTester)
}

/*
var largeBucket = test_store.NewTestStoreBench()
var largeBucket2 = test_store.NewTestStoreBench()

func BenchmarkInsert(b *testing.B) {
	cache := freshCache()
	defer cache.Close()

	largeBucket.MinSize(b.N)
	b.ResetTimer()

	largeBucket.FillStore(cache.Insert, b.N)
}

func BenchmarkLookup(b *testing.B) {
	cache := freshCache()
	defer cache.Close()

	// Set TestBucket and Cache based on a static size (b.N should only affect loop)
	largeBucket.MinSize(1000)
	largeBucket.FillStore(cache.Insert, 1000)

	// Each operation we will read the same set of keys
	iter := 10
	if b.N < 10 {
		iter = b.N
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for k := 0; k < iter; k++ {
			cache.Retrieve(largeBucket.Keys[k])
		}
	}
	b.StopTimer()
}

func insertTTL(cache *BadgerCache, ttl time.Duration) func(string, []byte) error {
	return func(s string, b []byte) error {
		return cache.InsertTTL(s, b, ttl)
	}
}

func benchExpiryRemove(n float32, b *testing.B) {
	cache := freshCache()
	defer cache.Close()
	clock := &mockClock{timeUTC()}
	cache.clock = clock
	// Normally set by Start() - but we want to control run
	cache.running = true

	// Set TestBucket and Cache based on a static size (b.N should only affect loop)

	var expiredCount int = int(float32(b.N) * n)
	largeBucket.MinSize(b.N - expiredCount)
	largeBucket.FillStore(insertTTL(cache, time.Hour), b.N-expiredCount)
	largeBucket2.MinSize(expiredCount)
	largeBucket2.FillStore(insertTTL(cache, time.Second), expiredCount)
	clock.Add(time.Minute)

	fmt.Printf("Cache length: %d, total run: %d, expiring: %d\n\n", cache.Len(), b.N, expiredCount)

	b.ResetTimer()
	cache.ExpiryScan()
	b.StopTimer()

	cache.running = false
}

func BenchmarkExpiryRemove75(b *testing.B) {
	benchExpiryRemove(0.75, b)
}

func BenchmarkExpiryRemove50(b *testing.B) {
	benchExpiryRemove(0.5, b)
}

func BenchmarkExpiryRemove25(b *testing.B) {
	benchExpiryRemove(0.25, b)
}

func BenchmarkExpiryRemove10(b *testing.B) {
	benchExpiryRemove(0.1, b)
}

*/
