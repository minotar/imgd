package bolt_cache

import (
	"testing"

	"github.com/minotar/imgd/pkg/cache/util/test_helpers"
)

const (
	TestBoltPath       = "/tmp/bolt_cache_test.db"
	TestBoltBucketName = "bolt_test"
)

func newCacheTester(t *testing.T) test_helpers.CacheTester {
	cache, err := NewBoltCache(TestBoltPath, TestBoltBucketName)
	if err != nil {
		t.Fatalf("Error creating BoltCache: %s", err)
	}

	clock := test_helpers.MockedUTC()
	cache.StoreExpiry.Clock = clock
	cache.Flush()
	cache.Start()

	return test_helpers.CacheTester{
		Tester:        t,
		Cache:         cache,
		RemoveExpired: cache.ExpiryScan,
		Clock:         clock,
		Iterations:    100,
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

/*

func addSortedTestData(t *testing.T, cache *BoltCache, iterationCount int) []string {
	sorted := make([]string, iterationCount)
	for i, offset := range rand.Perm(iterationCount) {
		key := test_helpers.RandString(32)
		// Insert key into our slice at offset position (making it sorted)
		sorted[offset] = key
		// Add Expiry values unordered into the expiry listing
		cache.InsertTTL(key, []byte("foobar"), time.Second*time.Duration(offset+1))

		expectedLen := i + 1
		if dbLength := cache.Len(); dbLength != uint(expectedLen) {
			t.Errorf("DB Length %d should be %d", dbLength, expectedLen)
		}
	}
	return sorted
}

func TestExpiryScanIteration(t *testing.T) {
	cache := freshCache()
	defer cache.Close()
	clock := &mockClock{timeUTC()}
	cache.clock = clock
	// Normally set by Start() - but we want to control run
	cache.running = true

	// A iteration number larger than 10 and divisible by 10
	iterationCount := 100
	sorted := addSortedTestData(t, cache, iterationCount)

	// DEBUG
	cache.DB.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(cache.Name))

		fmt.Printf("Keys:\n")
		b.ForEach(func(k, v []byte) error {
			fmt.Printf("key=%s ", k)
			return nil
		})
		fmt.Printf("\n")
		return nil
	})

	reviewTime := cache.clock.Now().Add(time.Duration(1) * time.Second)

	for len(sorted) > 0 {
		aheadSize := rand.Intn(iterationCount/10) + 1
		reviewTime = reviewTime.Add(time.Duration(aheadSize) * time.Second)
		// Debug
		fmt.Printf("Mocked time is: %s\n", reviewTime)

		// chunkSize can't be larger than the keys left in the sorted list
		chunkSize := test_helpers.Min(len(sorted), aheadSize)

		// Remove expired keys
		cache.expiryScan(reviewTime, 3)

		for _, k := range sorted[:chunkSize] {
			_, err := cache.Retrieve(k)
			if err != storage.ErrNotFound {
				t.Fail()
			}
		}

		sorted = sorted[chunkSize:]

		if cacheLen := int(cache.Len()); cacheLen != len(sorted) {
			t.Errorf("Cache Length %d should be %d", cacheLen, len(sorted))
		}
	}

	if dbLength := cache.Len(); dbLength != 0 {
		t.Errorf("DB length should have been 0, not %d", dbLength)
	}

	// Must be set to false otherwise the Close() will fail
	cache.running = false
}

func TestExpiryScan(t *testing.T) {
	cache := freshCache()
	defer cache.Close()
	clock := &mockClock{timeUTC()}
	cache.clock = clock
	// Normally set by Start() - but we want to control run
	cache.running = true

	// A iteration number larger than 10 and divisible by 10
	iterationCount := 100
	sorted := addSortedTestData(t, cache, iterationCount)

	// DEBUG
	cache.DB.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(cache.Name))

		fmt.Printf("Keys:\n")
		b.ForEach(func(k, v []byte) error {
			fmt.Printf("key=%s ", k)
			return nil
		})
		fmt.Printf("\n")
		return nil
	})

	// addSortedTestData adds the keys with a TTL of expiry+1. We add that offset here.
	clock.Add(time.Duration(1) * time.Second)

	for len(sorted) > 0 {
		aheadSize := rand.Intn(iterationCount/10) + 1
		// Advance the clock by a set amount to then verify the expected keys expired
		clock.Add(time.Duration(aheadSize) * time.Second)

		// Debug
		currentTime := cache.clock.Now().String()
		fmt.Printf("Mocked time is: %s\n", currentTime)

		// chunkSize can't be larger than the keys left in the sorted list
		chunkSize := test_helpers.Min(len(sorted), aheadSize)

		// Remove expired keys
		cache.ExpiryScan()

		for _, k := range sorted[:chunkSize] {
			_, err := cache.Retrieve(k)
			if err != storage.ErrNotFound {
				t.Fail()
			}
		}

		sorted = sorted[chunkSize:]

		if cacheLen := int(cache.Len()); cacheLen != len(sorted) {
			t.Errorf("Cache Length %d should be %d", cacheLen, len(sorted))
		}
	}

	if dbLength := cache.Len(); dbLength != 0 {
		t.Errorf("DB length should have been 0, not %d", dbLength)
	}

	// Must be set to false otherwise the Close() will fail
	cache.running = false
}


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
