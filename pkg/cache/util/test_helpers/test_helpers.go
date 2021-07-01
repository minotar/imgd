package test_helpers

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/minotar/imgd/pkg/cache"
	"github.com/minotar/imgd/pkg/storage/util/test_helpers"
	store_test_helpers "github.com/minotar/imgd/pkg/storage/util/test_helpers"
)

var (
	Min        = store_test_helpers.Min
	RandString = store_test_helpers.RandString
)

type MockClock struct {
	time time.Time
}

func (m *MockClock) Now() time.Time {
	return m.time
}

func (m *MockClock) Add(t time.Duration) {
	m.time = m.time.Add(t)
}

// Mocked time (which isn't epoch as that's special)
func TimeUTC() time.Time {
	specificTime, _ := time.Parse(time.RFC3339, "2021-05-19T00:00:00Z")
	return specificTime.UTC()
}

func MockedUTC() *MockClock {
	return &MockClock{TimeUTC()}
}

type CacheTester struct {
	Tester        *testing.T
	Cache         cache.Cache
	RemoveExpired func()
	Clock         *MockClock
	Iterations    int
}

func (ct *CacheTester) RetrieveKey(i int, key string) {
	value, err := ct.Cache.Retrieve(key)
	if err != nil {
		ct.Tester.Errorf("Key %s (%d) had an error: %s", key, i, err)
	}
	if string(value) != fmt.Sprint("value_", key) {
		ct.Tester.Errorf("Key %s (%d) was not the expected value: %s", key, i, value)
	}
}

func (ct *CacheTester) RetrieveDeletedKey(i int, key string) {
	value, err := ct.Cache.Retrieve(key)
	if value != nil {
		ct.Tester.Errorf("Key %s (%d) after expiry/removal should have a nil value: %s", key, i, value)
	}
	if err != cache.ErrNotFound {
		ct.Tester.Errorf("Key %s (%d) after expiry/removal should have been a cache.ErrNotFound. Error was: %s", key, i, err)
	}
}

func AddSortedString(insertFunc func(string, time.Duration), iterationCount int) []string {
	sorted := make([]string, iterationCount)
	for _, offset := range rand.Perm(iterationCount) {
		key := RandString(32)
		// Insert key into our slice at offset position (making it sorted)
		sorted[offset] = key
		// Add values unordered into the test function
		insertFunc(key, time.Duration(offset)*time.Second)
	}
	return sorted
}

// Using the cache.Cache InsertTTL function, create a dynamic []byte value for the AddSortedString
func DebugInsertTTL(cache cache.Cache) func(string, time.Duration) {
	return func(key string, ttl time.Duration) {
		value := []byte(fmt.Sprint("value_", key))
		cache.InsertTTL(key, value, ttl)
	}
}

// Using the cache.Cache Insert function, create a dynamic []byte value for the AddSortedString
func DebugInsert(cache cache.Cache) func(string, time.Duration) {
	return func(key string, _ time.Duration) {
		value := []byte(fmt.Sprint("value_", key))
		cache.Insert(key, value)
	}
}

func InsertAndRetrieve(cacheTester CacheTester) {
	sorted := AddSortedString(DebugInsert(cacheTester.Cache), cacheTester.Iterations)

	for i, key := range sorted {
		cacheTester.RetrieveKey(i, key)
	}

	if cacheLen := int(cacheTester.Cache.Len()); cacheLen != cacheTester.Iterations {
		cacheTester.Tester.Errorf("Cache length should have been %d, not: %d", cacheTester.Iterations, cacheLen)
	}
}

func InsertTTLAndRetrieve(cacheTester CacheTester) {
	sorted := AddSortedString(DebugInsertTTL(cacheTester.Cache), cacheTester.Iterations)

	for i, key := range sorted {
		cacheTester.RetrieveKey(i, key)
	}

	if cacheLen := int(cacheTester.Cache.Len()); cacheLen != cacheTester.Iterations {
		cacheTester.Tester.Errorf("Cache length should have been %d, not: %d", cacheTester.Iterations, cacheLen)
	}
}

// LRU size should be same as Iterations
func LRUInsertTTLAndRetrieve(cacheTester CacheTester) []string {
	sorted := AddSortedString(DebugInsertTTL(cacheTester.Cache), cacheTester.Iterations)

	// Use a subset of the keys which we'll keep requesting
	hotKeyCount := cacheTester.Iterations / 10

	for i := 0; i < 3; i++ {
		// Request the keys to keep them "recently used"
		for i, key := range sorted[:hotKeyCount] {
			cacheTester.RetrieveKey(i, key)
		}
		// Add new keys that will evict the cold keys
		AddSortedString(DebugInsertTTL(cacheTester.Cache), cacheTester.Iterations-hotKeyCount)
	}

	// Hot keys should be present
	for i, key := range sorted[:hotKeyCount] {
		cacheTester.RetrieveKey(i, key)
	}
	// Original keys should have been evicted
	for i, key := range sorted[hotKeyCount:] {
		cacheTester.RetrieveDeletedKey(i, key)
	}

	if cacheLen := int(cacheTester.Cache.Len()); cacheLen != cacheTester.Iterations {
		cacheTester.Tester.Errorf("Cache length should have been %d, not: %d", cacheTester.Iterations, cacheLen)
	}

	return sorted[:hotKeyCount]
}

func InsertTTLAndRemove(cacheTester CacheTester) {
	sorted := AddSortedString(DebugInsertTTL(cacheTester.Cache), cacheTester.Iterations)

	for i, key := range sorted {
		err := cacheTester.Cache.Remove(key)
		if err != nil {
			cacheTester.Tester.Errorf("Key %s (%d) should have removed without an error. Error was: %s", key, i, err)
		}

		cacheTester.RetrieveDeletedKey(i, key)
	}

	if cacheLen := int(cacheTester.Cache.Len()); cacheLen != 0 {
		cacheTester.Tester.Errorf("Cache length should have been 0, not: %d", cacheLen)
	}
}

func InsertTTLAndExpiry(cacheTester CacheTester) {
	sorted := AddSortedString(DebugInsertTTL(cacheTester.Cache), cacheTester.Iterations)

	// Skip the first item (0 duration) and advance the clock by 1 second so the offset is corrected
	sorted = sorted[1:]
	cacheTester.Clock.Add(time.Duration(1) * time.Second)

	for len(sorted) > 0 {
		aheadSize := rand.Intn(cacheTester.Iterations/10) + 1
		// chunkSize can't be larger than the keys left in the sorted list
		chunkSize := test_helpers.Min(len(sorted), aheadSize)

		cacheTester.Clock.Add(time.Duration(aheadSize) * time.Second)
		cacheTester.RemoveExpired()

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

func InsertTTLAndTTLCheck(cacheTester CacheTester) {
	sorted := AddSortedString(DebugInsertTTL(cacheTester.Cache), cacheTester.Iterations)

	// Test No Expiration beahviour
	noExpiry := sorted[0]
	ttl, err := cacheTester.Cache.TTL(noExpiry)
	if err != cache.ErrNoExpiry {
		cacheTester.Tester.Errorf("Non expiring key shoulduld have been a cache.ErrNoExpiry. Error was: %s", err)
	}
	if ttl != time.Duration(0) {
		cacheTester.Tester.Errorf("Non expiring key should return a TTL of 0: %s", ttl)
	}

	// Skip the first item (0 duration) and advance the clock by 1 second so the offset is corrected
	sorted = sorted[1:]
	cacheTester.Clock.Add(time.Duration(1) * time.Second)

	for len(sorted) > 0 {
		aheadSize := rand.Intn(cacheTester.Iterations/10) + 1
		// chunkSize can't be larger than the keys left in the sorted list
		chunkSize := test_helpers.Min(len(sorted), aheadSize)

		for i, key := range sorted[:chunkSize] {
			ttl, err := cacheTester.Cache.TTL(key)
			if err != nil {
				cacheTester.Tester.Errorf("Key %s (%d) had an error: %s", key, i, err)
			}

			// Key 0 (which has technically just expired) should return to 1 Second TTL
			expectedTTL := i
			if expectedTTL == 0 {
				expectedTTL = 1
			}
			if ttl != time.Duration(expectedTTL)*time.Second {
				cacheTester.Tester.Errorf("Key %s (%d) TTL was %v", key, i, ttl)
			}
		}

		cacheTester.Clock.Add(time.Duration(aheadSize) * time.Second)
		cacheTester.RemoveExpired()

		for i, key := range sorted[:chunkSize] {
			ttl, err := cacheTester.Cache.TTL(key)
			if err != cache.ErrNotFound {
				cacheTester.Tester.Errorf("Key %s (%d) TTL after expiry should have been a cache.ErrNotFound. Error was: %s", key, i, err)
			}
			if ttl != time.Duration(0) {
				cacheTester.Tester.Errorf("Key %s (%d) TTL was %v", key, i, ttl)
			}
		}

		// Re-slice ready for next loop
		sorted = sorted[chunkSize:]
	}

	if cacheLen := int(cacheTester.Cache.Len()); cacheLen != 1 {
		cacheTester.Tester.Errorf("Cache length should have been 1, not: %d", cacheLen)
	}
}

func InsertTTLAndFlush(cacheTester CacheTester) {
	AddSortedString(DebugInsertTTL(cacheTester.Cache), cacheTester.Iterations)

	err := cacheTester.Cache.Flush()
	if err != nil {
		cacheTester.Tester.Errorf("Cache flushing should not have returned an error: %s", err)
	}

	if cacheLen := int(cacheTester.Cache.Len()); cacheLen != 0 {
		cacheTester.Tester.Errorf("Cache length should have been %d, not: %d", 0, cacheLen)
	}
}
