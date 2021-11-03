package bolt_cache

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/boltdb/bolt"
	"github.com/minotar/imgd/pkg/cache"
	store_expiry "github.com/minotar/imgd/pkg/cache/util/expiry/store"
	"github.com/minotar/imgd/pkg/cache/util/test_helpers"
	"github.com/minotar/imgd/pkg/util/log"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	TestBoltPath       = "/tmp/bolt_cache_test.db"
	TestBoltBucketName = "bolt_test"
)

func newCache(t *testing.T) *BoltCache {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()

	logger := log.NewBuiltinLogger(1)
	logger.Named("BoltTest")

	cacheConfig := cache.CacheConfig{
		Name:   "BoltTest",
		Logger: logger,
	}
	boltCacheConfig := NewBoltCacheConfig(cacheConfig, TestBoltPath, TestBoltBucketName)

	cache, err := NewBoltCache(boltCacheConfig)
	if err != nil {
		t.Fatalf("Error creating BoltCache: %s", err)
	}
	return cache
}

func newCacheTester(t *testing.T) test_helpers.CacheTester {
	cache := newCache(t)
	return newCacheTesterWithBoltCache(t, cache)
}

func newCacheTesterWithBoltCache(t *testing.T, bc *BoltCache) test_helpers.CacheTester {
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

func TestInsertTTLAndExpiry(t *testing.T) {
	cacheTester := newCacheTester(t)
	defer cacheTester.Cache.Close()

	test_helpers.InsertTTLAndExpiry(cacheTester)
}

func TestInsertTTLAndTTLCheck(t *testing.T) {
	cacheTester := newCacheTester(t)
	defer cacheTester.Cache.Close()

	test_helpers.InsertTTLAndTTLCheck(cacheTester)
}

func TestInsertTTLAndFlush(t *testing.T) {
	cacheTester := newCacheTester(t)
	defer cacheTester.Cache.Close()

	test_helpers.InsertTTLAndFlush(cacheTester)
}

// Tests the ExpiryScan with a smaller chunk size / max scan size
// This is because the COMPACTION_MAX_SCAN is much larger than the iteration count
func TestExpiryScanIteration(t *testing.T) {
	cache := newCache(t)
	cacheTester := newCacheTesterWithBoltCache(t, cache)
	defer cacheTester.Cache.Close()

	sorted := test_helpers.AddSortedString(test_helpers.DebugInsertTTL(cacheTester.Cache), cacheTester.Iterations)

	// Skip the first item (0 duration) and advance the clock by 1 second so the offset is corrected
	sorted = sorted[1:]
	cacheTester.Clock.Add(time.Duration(1) * time.Second)

	// DEBUG
	cache.DB.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(cache.Bucket))

		fmt.Printf("Keys:\n")
		b.ForEach(func(k, v []byte) error {
			fmt.Printf("key=%s ", k)
			return nil
		})
		fmt.Printf("\n")
		return nil
	})

	for len(sorted) > 0 {
		aheadSize := rand.Intn(cacheTester.Iterations/10) + 1
		// chunkSize can't be larger than the keys left in the sorted list
		chunkSize := test_helpers.Min(len(sorted), aheadSize)

		cacheTester.Clock.Add(time.Duration(aheadSize) * time.Second)
		// Debug
		fmt.Printf("Mocked time is: %s\n", cacheTester.Clock.Now())
		// Remove expired keys, using a smaller scan size to test logic
		cache.expiryScan(cacheTester.Clock.Now(), 3)

		for i, key := range sorted[:chunkSize] {
			cacheTester.RetrieveDeletedKey(i, key)
		}

		// Re-slice ready for next loop
		sorted = sorted[chunkSize:]
	}

	if cacheLen := int(cacheTester.Cache.Len()); cacheLen != 1 {
		cacheTester.Tester.Errorf("Cache length should have been 1, not: %d", cacheLen)
	}
}

func TestExpiryScanInteruption(t *testing.T) {
	cache := newCache(t)
	cacheTester := newCacheTesterWithBoltCache(t, cache)
	defer cacheTester.Cache.Close()

	// Add a record
	cacheTester.Cache.InsertTTL("key1", []byte{}, 1)

	// Stop the cache so IsRunning will be False
	cacheTester.Cache.Stop()
	cacheTester.Clock.Add(3)

	// Remove expired keys, using a smaller scan size to test logic
	cache.expiryScan(cacheTester.Clock.Now(), 10)

	// Todo: check logged messages?

	if cacheLen := int(cacheTester.Cache.Len()); cacheLen != 1 {
		cacheTester.Tester.Errorf("Cache length should have been 1, not: %d", cacheLen)
	}
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

func insertTTL(cache *BoltCache, ttl time.Duration) func(string, []byte) error {
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
