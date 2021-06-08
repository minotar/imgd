package memory

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/minotar/imgd/pkg/storage/util/test_helpers"
	"github.com/minotar/imgd/pkg/storage/util/test_store"
)

type mockClock struct {
	time time.Time
}

func (m *mockClock) Now() time.Time {
	return m.time
}

func (m *mockClock) Add(t time.Duration) {
	m.time = m.time.Add(t)
}

func mockExpiry() (*mockClock, *Expiry) {
	clock := &mockClock{time.Unix(0, 0)}
	return clock, &Expiry{clock: clock}
}

func TestExpiryRemove(t *testing.T) {
	_, ex := mockExpiry()

	// A iteration number larger than 10 and divisible by 10
	iterationCount := 500
	sorted := make([]string, iterationCount)
	for i, offset := range rand.Perm(iterationCount) {
		key := fmt.Sprintf("key_%d", i)
		// Insert key into our slice at offset position (making it sorted)
		sorted[offset] = key
		// Add Expiry values unordered into the expiry listing
		ex.AddExpiry(key, time.Second*time.Duration(offset))
	}

	for _, key := range sorted {
		// Directly remove the first key and reslice sorted without it
		ex.RemoveExpiry(key, nil)
		sorted = sorted[1:]

		// Make a slice of the remaining keys in the Expiry list
		tupleSlice := make([]string, len(ex.tuples))
		for i, val := range ex.tuples {
			tupleSlice[i] = val.key
		}

		// Verify the slices match
		if !reflect.DeepEqual(tupleSlice, sorted) {
			t.Fail()
		}
	}
}

func TestExpiryOrdering(t *testing.T) {
	clock, ex := mockExpiry()

	// Use an unordered slice of durations
	for i, ttl := range []int{1337, 512, 2021} {
		key := fmt.Sprintf("key%d", i)

		// Add Expiry values unordered into the expiry listing
		ex.AddExpiry(key, time.Duration(ttl)*time.Second)
	}

	for i, expectedKey := range []string{"key1", "key0", "key2"} {
		if actualKey := ex.tuples[i].key; actualKey != expectedKey {
			t.Errorf("Expected \"%s\" but found \"%s\" at position %d", expectedKey, actualKey, i)
		}
	}

	for _, shouldCompact := range [][]string{{}, {"key1"}, {"key0"}} {
		// Add 500s to the clock to simulate the keys now being past expiry
		clock.Add(time.Duration(500) * time.Second)
		if compacted := ex.Compact(); !reflect.DeepEqual(compacted, shouldCompact) {
			t.Errorf("Should have compacted %+v but instead: %+v", shouldCompact, compacted)
		}
	}

	ex.Compact()
	if !(len(ex.tuples) == 1 && ex.tuples[0].key == "key2") {
		t.Errorf("The last key in the expiry list should be \"key2\": %+v", ex.tuples)
	}
}

func TestExpiryExtended(t *testing.T) {
	clock, ex := mockExpiry()

	// A iteration number larger than 10 and divisible by 10
	iterationCount := 5000
	sorted := make([]string, iterationCount)
	for i, offset := range rand.Perm(iterationCount) {
		key := test_helpers.RandString(32)
		// Insert key into our slice at offset position (making it sorted)
		sorted[offset] = key
		// Add Expiry values unordered into the expiry listing
		ex.AddExpiry(key, time.Second*time.Duration(offset))

		expectedLen := i + 1
		if expiryLen := int(ex.LenExpiry()); expiryLen != expectedLen {
			t.Errorf("Expiry Length %d should be %d", expiryLen, expectedLen)
		}
	}

	for len(sorted) > 0 {
		aheadSize := rand.Intn(iterationCount/10) + 1
		// Advance the clock by a set amount to then verify the expected keys expired
		clock.Add(time.Duration(aheadSize) * time.Second)

		// chunkSize can't be larger than the keys left in the sorted list
		chunkSize := test_helpers.Min(len(sorted), aheadSize)

		// Compare the expired keys vs. the first X keys from the sorted list
		if compacted := ex.Compact(); !reflect.DeepEqual(sorted[:chunkSize], compacted) {
			t.Errorf("Compacted was not as expected. Length %d vs. %d", len(sorted[:chunkSize]), len(compacted))
		}
		// Re-slice ready for next loop
		sorted = sorted[chunkSize:]

		if expiryLen := int(ex.LenExpiry()); expiryLen != len(sorted) {
			t.Errorf("Expiry Length %d should be %d", expiryLen, len(sorted))
		}
	}

	if compacted := len(ex.Compact()); compacted != 0 {
		t.Errorf("Compacted length should have been 0, not %d", compacted)
	}
}

type TestCache struct {
	*test_store.TestStorage
	*Expiry
}

func mockCacheExpiry() (*mockClock, *TestCache) {
	clock, _ := mockExpiry()
	store := &TestCache{
		TestStorage: test_store.NewTestStorage(),
	}
	store.Expiry = NewExpiry(store.Remove)
	store.clock = clock
	return clock, store
}

func TestExpiryStoreRemove(t *testing.T) {
	clock, cache := mockCacheExpiry()
	iterationCount := 10

	for i := 0; i < iterationCount; i++ {
		key := test_helpers.RandString(32)
		cache.Insert(key, []byte("foobar"))
		cache.Expiry.AddExpiry(key, time.Duration(i)*time.Second)
	}
	clock.Add(time.Nanosecond)
	for i := 0; i < iterationCount; i++ {
		cache.RemoveExpired()

		cacheLen := cache.Len()
		expiryLen := cache.LenExpiry()
		expectedLen := iterationCount - (i + 1)
		if cacheLen != expiryLen || expiryLen != uint(expectedLen) {
			t.Errorf("Cache Length %d and Expiry Length %d should be %d", cacheLen, expiryLen, expectedLen)
		}

		clock.Add(time.Duration(1) * time.Second)
	}
}

func TestGoCompactor(t *testing.T) {
	clock, cache := mockCacheExpiry()
	cache.Expiry.compaction_interval = 5 * time.Millisecond
	iterationCount := 10

	for i := 0; i < iterationCount; i++ {
		key := test_helpers.RandString(32)
		cache.Insert(key, []byte("foobar"))
		cache.Expiry.AddExpiry(key, time.Minute)
	}
	clock.Add(time.Duration(5) * time.Minute)

	cache.Expiry.Start()
	time.Sleep(time.Duration(10) * time.Millisecond)
	cache.Expiry.Stop()

	cacheLen := cache.Len()
	expiryLen := cache.LenExpiry()
	if cacheLen != expiryLen || expiryLen != uint(0) {
		t.Errorf("Cache Length %d and Expiry Length %d should be 0", cacheLen, expiryLen)
	}
}
