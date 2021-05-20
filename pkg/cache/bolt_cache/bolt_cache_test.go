package bolt_cache

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/minotar/imgd/pkg/storage"
	"github.com/minotar/imgd/pkg/storage/util/test_helpers"
	"github.com/minotar/imgd/pkg/storage/util/test_store"
)

const (
	TestBoltPath       = "/tmp/bolt_cache_test.db"
	TestBoltBucketName = "bolt_test"
)

func freshCache() *BoltCache {
	cache, _ := NewBoltCache(TestBoltPath, TestBoltBucketName)
	cache.Flush()
	return cache
}

func TestRetrieveMiss(t *testing.T) {
	cache := freshCache()
	defer cache.Close()

	v, err := cache.Retrieve(test_helpers.RandString(32))
	if v != nil {
		t.Errorf("Retrieve Miss should return a nil value, not: %+v", v)
	}
	if err != storage.ErrNotFound {
		t.Errorf("Retrieve Miss should return a storage.ErrNotFound, not: %s", err)
	}
}

func TestInsertAndRetrieve(t *testing.T) {
	cache := freshCache()
	defer cache.Close()

	for i := 0; i < 10; i++ {
		str := test_helpers.RandString(32)
		cache.Insert(str, []byte(strconv.Itoa(i)))
		item, err := cache.Retrieve(str)
		if err != nil {
			t.Errorf("Retrieve should not be an error: %s", err)
		}
		if string(item) != strconv.Itoa(i) {
			t.Errorf("%+v did not match %d", item, i)
		}
	}
}

func TestInsertAndDelete(t *testing.T) {
	cache := freshCache()
	defer cache.Close()

	for i := 0; i < 10; i++ {
		key := test_helpers.RandString(32)
		cache.Insert(key, []byte(strconv.Itoa(i)))
		cache.Remove(key)
		_, err := cache.Retrieve(key)
		if err != storage.ErrNotFound {
			t.Errorf("Key should have been removed: %s", key)
		}
	}
}

func TestExpiryRemove(t *testing.T) {
	cache := freshCache()
	defer cache.Close()
	clock := &mockClock{timeUTC()}
	cache.clock = clock
	// Normally set by Start() - but we want to control run
	cache.running = true

	// A iteration number larger than 10 and divisible by 10
	iterationCount := 20
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

	for len(sorted) > 0 {
		aheadSize := rand.Intn(iterationCount/10) + 1
		// Advance the clock by a set amount to then verify the expected keys expired
		clock.Add(time.Duration(aheadSize+1) * time.Second)

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
	}

	if dbLength := cache.Len(); dbLength != 0 {
		t.Errorf("DB length should have been 0, not %d", dbLength)
	}

	// Must be set to false otherwise the Close() will fail
	cache.running = false
}

func TestHousekeeping(t *testing.T) {
	cache := freshCache()
	defer cache.Close()

	if len := cache.Len(); len != 0 {
		t.Errorf("Initialized cache should be length 0, not %d", len)
	}

	for i := 0; i < 10; i++ {
		str := test_helpers.RandString(32)
		cache.Insert(str, []byte("var"))
	}

	if len := cache.Len(); len != 10 {
		t.Errorf("Part filled cache should be length 10, not %d", len)
	}

	cache.Flush()

	if len := cache.Len(); len != 0 {
		t.Errorf("Flushed cache should be length 0, not %d", len)
	}
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
	largeBucket.FillStore(insertTTL(cache, 0), b.N-expiredCount)
	largeBucket2.MinSize(expiredCount)
	largeBucket2.FillStore(insertTTL(cache, time.Minute), expiredCount)
	clock.Add(time.Hour)

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
