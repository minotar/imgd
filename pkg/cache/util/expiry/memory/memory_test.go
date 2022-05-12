package memory

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/minotar/imgd/pkg/cache"
	"github.com/minotar/imgd/pkg/cache/util/expiry"
	"github.com/minotar/imgd/pkg/cache/util/test_helpers"
)

func unixUTC(n int) time.Time {
	return time.Unix(int64(n), 0).UTC()
}

func freshMemoryExpiry() (*MemoryExpiry, *test_helpers.MockClock) {
	me, _ := NewMemoryExpiry(func(string) error { return nil })
	clock := test_helpers.MockedUTC()
	me.Clock = clock
	return me, clock
}

func TestNewStoreNoCompactor(t *testing.T) {
	_, err := NewMemoryExpiry(nil)
	if err == nil || !strings.Contains(err.Error(), "Remove function") {
		t.Errorf("Lack of specified Compactor function should have raised an error")
	}
}

func TestNewStoreExpiry(t *testing.T) {
	calledCount := 0
	compactorFunc := func() { calledCount++ }
	me, _ := freshMemoryExpiry()

	options := expiryOptions(compactorFunc)
	options.CompactorInterval = 5 * time.Millisecond
	e, err := expiry.NewExpiry(options)
	if err != nil {
		t.Errorf("NewExpiry should not have raised an error: %s", err)
	}
	me.Expiry = e

	if err != nil {
		t.Error("NewMemoryExpiry should not throw an error")
	}

	if calledCount != 0 {
		t.Errorf("compactorFunc shouldn't be called before Start()")
	}
	me.Start()
	time.Sleep(time.Duration(2) * time.Millisecond)
	if calledCount != 1 {
		t.Errorf("compactorFunc should be called once immediately after Start()")
	}
	time.Sleep(time.Duration(10) * time.Millisecond)
	me.Stop()
	time.Sleep(time.Duration(2) * time.Millisecond)

	// It should not be less than 2
	if calledCount < 2 {
		t.Errorf("compactorFunc should be called after ticking")
	}

}

func TestExpiryRemoveError(t *testing.T) {
	me, _ := freshMemoryExpiry()
	me.RemoveExpiry(1337, nil)
	// Todo: check logging occured
}

func TestExpiryRemove(t *testing.T) {
	me, _ := freshMemoryExpiry()

	// A iteration number larger than 10 and divisible by 10
	iterationCount := 500
	sorted := test_helpers.AddSortedString(me.AddExpiry, iterationCount)
	// We compare against "iterationCount-1" as the 0 duration will not be added to the records
	if expiryLen := int(me.Len()); expiryLen != iterationCount-1 {
		t.Errorf("Length of Expiry records should be %d, not %d", iterationCount-1, expiryLen)
	}

	for i, key := range sorted {
		// Alternate between string and byte key types
		if i%2 == 0 {
			// Remove string key
			me.RemoveExpiry(string(key), nil)
		} else {
			// Remove byte key
			me.RemoveExpiry([]byte(key), nil)
		}
		// Directly remove the first key and reslice sorted without it
		sorted = sorted[1:]

		// Make a slice of the remaining keys in the Expiry list
		tupleSlice := make([]string, len(me.records))
		for i, val := range me.records {
			tupleSlice[i] = val.Key
		}

		// Verify the slices match
		if !reflect.DeepEqual(tupleSlice, sorted) {
			t.Fail()
		}
	}
}

// Given a preset list of keys, verify the logic works as expected
func TestExpiryOrdering(t *testing.T) {
	me, clock := freshMemoryExpiry()

	// Use an unordered slice of durations
	for i, ttl := range []int{1337, 512, 2021} {
		key := fmt.Sprintf("key_%d", i)

		// Add Expiry values unordered into the expiry listing
		me.AddExpiry(key, time.Duration(ttl)*time.Second)
	}

	for i, expectedKey := range []string{"key_1", "key_0", "key_2"} {
		if actualKey := me.records[i].Key; actualKey != expectedKey {
			t.Errorf("Expected \"%s\" but found \"%s\" at position %d", expectedKey, actualKey, i)
		}
	}

	for _, shouldCompact := range [][]string{{}, {"key_1"}, {"key_0"}} {
		// Add 500s to the clock to simulate the keys now being past expiry
		clock.Add(time.Duration(500) * time.Second)
		if compacted := me.Compact(); !reflect.DeepEqual(compacted, shouldCompact) {
			t.Errorf("Should have compacted %+v but instead: %+v", shouldCompact, compacted)
		}
	}

	me.Compact()
	if !(len(me.records) == 1 && me.records[0].Key == "key_2") {
		t.Errorf("The last key in the expiry list should be \"key_2\": %+v", me.records)
	}
}

func TestExpiryExtended(t *testing.T) {
	me, clock := freshMemoryExpiry()

	// A iteration number larger than 10 and divisible by 10
	iterationCount := 5000
	sorted := test_helpers.AddSortedString(me.AddExpiry, iterationCount)
	// We compare against "iterationCount-1" as the 0 duration will not be added to the records
	if expiryLen := int(me.Len()); expiryLen != iterationCount-1 {
		t.Errorf("Length of Expiry records should be %d, not %d", iterationCount-1, expiryLen)
	}

	// Skip the first item (0 duration) and advance the clock by 1 second so the offset is corrected
	sorted = sorted[1:]
	clock.Add(time.Duration(1) * time.Second)

	for len(sorted) > 0 {
		aheadSize := rand.Intn(iterationCount/10) + 1
		// chunkSize can't be larger than the keys left in the sorted list
		chunkSize := test_helpers.Min(len(sorted), aheadSize)

		// Advance the clock by a set amount to then verify the expected keys expired
		clock.Add(time.Duration(aheadSize) * time.Second)

		// Compare the expired keys vs. the first X keys from the sorted list
		if compacted := me.Compact(); !reflect.DeepEqual(sorted[:chunkSize], compacted) {
			t.Errorf("Compacted was not as expected. Length %d vs. %d", len(sorted[:chunkSize]), len(compacted))
		}
		// Re-slice ready for next loop
		sorted = sorted[chunkSize:]

		if expiryLen := int(me.Len()); expiryLen != len(sorted) {
			t.Errorf("Expiry Length %d should be %d", expiryLen, len(sorted))
		}
	}

	if compacted := len(me.Compact()); compacted != 0 {
		t.Errorf("Compacted length should have been 0, not %d", compacted)
	}
}

// There are a few behaviours specific for Nonexistent or non-expiring keys
func TestNoExpiryTTL(t *testing.T) {
	me, _ := freshMemoryExpiry()

	// Directly add key with no expiry
	me.AddExpiry("no expiry", 0)
	if me.Len() != 0 {
		t.Errorf("Non-expiring keys should not be added to Expiry records")
	}

	// Check for both a key without expiry, and a key never added (they behave the same)
	for _, keyStr := range []string{"no expiry", "invalid key"} {

		expires := me.GetExpiry(keyStr)

		if !expires.IsZero() {
			t.Errorf("GetExpiry should be zero time, not: %v", expires)
		}

		ttl, err := me.GetTTL("invalid key")
		if err != cache.ErrNoExpiry {
			t.Errorf("GetTTL should return a specific error, not: %s", err)
		}
		if ttl != time.Duration(0) {
			t.Errorf("GetTTL should return a ttl of 0, not: %d", ttl)
		}
	}
}

// Verify TTL values are 1 Second or above
func TestExpiryTTL(t *testing.T) {
	me, clock := freshMemoryExpiry()

	// A iteration number larger than 10 and divisible by 10
	iterationCount := 5000
	sorted := test_helpers.AddSortedString(me.AddExpiry, iterationCount)

	// Skip the first item (0 duration) and advance the clock by 1 second so the offset is corrected
	sorted = sorted[1:]
	clock.Add(time.Duration(1) * time.Second)

	for i, key := range sorted {
		expiry := me.GetExpiry(key)
		if !expiry.Equal(clock.Now().Add(time.Duration(i) * time.Second)) {
			t.Errorf("Expiry for key %d was not as expected: %v", i, expiry)
		}
	}

	for len(sorted) > 0 {
		aheadSize := rand.Intn(iterationCount/10) + 1
		// chunkSize can't be larger than the keys left in the sorted list
		chunkSize := test_helpers.Min(len(sorted), aheadSize)

		for i, key := range sorted[:chunkSize] {
			ttl, err := me.GetTTL(key)
			if err != nil {
				t.Errorf("Key %s (%d) had an error: %s", key, i, err)
			}
			if ttl == time.Duration(0) {
				t.Errorf("Key %s (%d) had a TTL of 0 (no expiry)", key, i)
			}

			// Key 0 (which has technically just expired) should return to 1 Second TTL
			expectedTTL := i
			if expectedTTL == 0 {
				expectedTTL = 1
			}
			if ttl != time.Duration(expectedTTL)*time.Second {
				t.Errorf("Key %s (%d) TTL was %v", key, i, ttl)
			}
		}

		// Advance the clock by a set amount to then verify the expected keys TTL is now 1 second
		clock.Add(time.Duration(aheadSize) * time.Second)

		for i, key := range sorted[:chunkSize] {
			ttl, err := me.GetTTL(key)
			if err != nil {
				t.Errorf("Key %s (%d) had an error: %s", key, i, err)
			}
			if ttl != time.Duration(time.Second) {
				t.Errorf("Key %s (%d) had a TTL of %v, not 1 Second", key, i, ttl)
			}
		}
		// Remove the now expired records and update the sorted slice
		me.Compact()
		sorted = sorted[chunkSize:]
	}
}

func TestRemoveExpiredError(t *testing.T) {
	me, clock := freshMemoryExpiry()

	me.removeFunc = func(key string) error {
		return fmt.Errorf("This is a removal error")
	}

	me.AddExpiry("error", time.Second)
	clock.Add(time.Minute)

	me.RemoveExpired()
	// Todo: check logging occured
}

// Could likely fully replace TestExpiryExtended
func TestRemoveExpired(t *testing.T) {
	me, clock := freshMemoryExpiry()

	var removedKeys = []string{}

	me.removeFunc = func(key string) error {
		removedKeys = append(removedKeys, key)
		return nil
	}

	// A iteration number larger than 10 and divisible by 10
	iterationCount := 5000
	sorted := test_helpers.AddSortedString(me.AddExpiry, iterationCount)

	// Skip the first item (0 duration) and advance the clock by 1 second so the offset is corrected
	sorted = sorted[1:]
	clock.Add(time.Duration(1) * time.Second)

	for len(sorted) > 0 {
		aheadSize := rand.Intn(iterationCount/10) + 1
		// chunkSize can't be larger than the keys left in the sorted list
		chunkSize := test_helpers.Min(len(sorted), aheadSize)

		// Advance the clock by a set amount to then verify the expected keys expired
		clock.Add(time.Duration(aheadSize) * time.Second)

		// Simulate the Compactor Function being executed
		me.RemoveExpired()

		// Compare the expired keys vs. the first X keys from the sorted list
		if !reflect.DeepEqual(sorted[:chunkSize], removedKeys) {
			t.Errorf("Removed was not as expected. Length %d vs. %d", len(sorted[:chunkSize]), len(removedKeys))
		}

		// Re-slice ready for next loop
		sorted = sorted[chunkSize:]
		// Clear the removedKeys list
		removedKeys = []string{}

		if expiryLen := int(me.Len()); expiryLen != len(sorted) {
			t.Errorf("Expiry Length %d should be %d", expiryLen, len(sorted))
		}
	}

	// Simulate the Compactor Function being executed
	me.RemoveExpired()

	if len(removedKeys) != 0 {
		t.Errorf("RemoveExpired shouldn't have removed any keys when nohting to expire: %v", removedKeys)
	}

	if compacted := len(me.Compact()); compacted != 0 {
		t.Errorf("Compacted length should have been 0, not %d", compacted)
	}

}

func benchAddExpiry(size int, b *testing.B) {
	// This uses increasing TTLs as the mock Clock would otherwise add each
	// record at the same moment
	me, _ := freshMemoryExpiry()
	keyStr := test_helpers.RandString(size)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		me.AddExpiry(keyStr, time.Duration(i)*time.Second)
	}
}

func BenchmarkAddExpiry32(b *testing.B) {
	benchAddExpiry(32, b)
}

func BenchmarkAddExpiry64(b *testing.B) {
	benchAddExpiry(64, b)
}
