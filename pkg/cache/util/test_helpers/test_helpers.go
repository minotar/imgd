package test_helpers

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/minotar/imgd/pkg/cache"
	"github.com/minotar/imgd/pkg/storage"
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
func insertTTL(cache cache.Cache) func(string, time.Duration) {
	return func(key string, ttl time.Duration) {
		value := []byte(fmt.Sprint("value_", key))
		cache.InsertTTL(key, value, ttl)
	}
}

func InsertTTLAndRetrieve(cacheTester CacheTester) {
	iterationCount := 500
	sorted := AddSortedString(insertTTL(cacheTester.Cache), iterationCount)

	for i, key := range sorted {
		value, err := cacheTester.Cache.Retrieve(key)
		if err != nil {
			cacheTester.Tester.Errorf("Key %s (%d) had an error: %s", key, i, err)
		}
		if string(value) != fmt.Sprint("value_", key) {
			cacheTester.Tester.Errorf("Key %s (%d) was not the expected value: %s", key, i, value)
		}
	}

	if cacheLen := int(cacheTester.Cache.Len()); cacheLen != iterationCount {
		cacheTester.Tester.Errorf("Cache lenght should have been %d, not: %d", iterationCount, cacheLen)
	}
}

func InsertTTLAndExpiry(cacheTester CacheTester) {
	iterationCount := 500
	sorted := AddSortedString(insertTTL(cacheTester.Cache), iterationCount)

	// Skip the first item (0 duration) and advance the clock by 1 second so the offset is corrected
	sorted = sorted[1:]
	cacheTester.Clock.Add(time.Duration(1) * time.Second)

	for len(sorted) > 0 {
		aheadSize := rand.Intn(iterationCount/10) + 1
		// chunkSize can't be larger than the keys left in the sorted list
		chunkSize := test_helpers.Min(len(sorted), aheadSize)

		cacheTester.Clock.Add(time.Duration(aheadSize) * time.Second)
		cacheTester.RemoveExpired()

		for i, key := range sorted[:chunkSize] {
			_, err := cacheTester.Cache.Retrieve(key)
			if err != storage.ErrNotFound {
				cacheTester.Tester.Errorf("Key %s (%d) after expiry should have been a storage.ErrNotFound. Error was: %s", key, i, err)
			}
		}

		// Re-slice ready for next loop
		sorted = sorted[chunkSize:]
	}

	if cacheLen := int(cacheTester.Cache.Len()); cacheLen != 1 {
		cacheTester.Tester.Errorf("Cache lenght should have been 1, not: %d", cacheLen)
	}
}

func InsertTTLAndTTLCheck(cacheTester CacheTester) {
	iterationCount := 500
	sorted := AddSortedString(insertTTL(cacheTester.Cache), iterationCount)

	// Test No Expiration beahviour
	noExpiry := sorted[0]
	ttl, err := cacheTester.Cache.TTL(noExpiry)
	if !strings.Contains(err.Error(), "No expiry set for key") {
		cacheTester.Tester.Errorf("Non expiring key should return a TTL error: %s", err)
	}
	if ttl != time.Duration(0) {
		cacheTester.Tester.Errorf("Non expiring key should return a TTL of 0: %s", ttl)
	}

	// Skip the first item (0 duration) and advance the clock by 1 second so the offset is corrected
	sorted = sorted[1:]
	cacheTester.Clock.Add(time.Duration(1) * time.Second)

	for len(sorted) > 0 {
		aheadSize := rand.Intn(iterationCount/10) + 1
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
			if err != storage.ErrNotFound {
				cacheTester.Tester.Errorf("Key %s (%d) TTL after expiry should have been a storage.ErrNotFound. Error was: %s", key, i, err)
			}
			if ttl != time.Duration(0) {
				cacheTester.Tester.Errorf("Key %s (%d) TTL was %v", key, i, ttl)
			}
		}

		// Re-slice ready for next loop
		sorted = sorted[chunkSize:]
	}

	if cacheLen := int(cacheTester.Cache.Len()); cacheLen != 1 {
		cacheTester.Tester.Errorf("Cache lenght should have been 1, not: %d", cacheLen)
	}
}
