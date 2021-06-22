package expiry

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/minotar/imgd/pkg/cache/util/test_helpers"
)

func unixUTC(n int) time.Time {
	return time.Unix(int64(n), 0).UTC()
}

func TestNewExpiryRecord(t *testing.T) {
	iterationCount := 500
	for i := 0; i < iterationCount; i++ {

		keyName := test_helpers.RandString(32)
		r := NewExpiryRecord(keyName, unixUTC(i))

		if r.Key != keyName {
			t.Errorf("Key should be \"%s\": %s", keyName, r.Key)
		}
		if r.ExpirySeconds != uint32(i) {
			t.Errorf("ExpirySeconds did not match %d: %d", i, r.ExpirySeconds)
		}

		if expectedTime := unixUTC(i); r.Expiry() != expectedTime {
			t.Errorf("Expected Time %+v did not Expiry Time %+v", expectedTime, r.Expiry())
		}
	}
}

func TestNewNoExpiryRecord(t *testing.T) {
	r := NewExpiryRecordTTL("foo", test_helpers.MockedUTC(), 0)
	if r.ExpirySeconds != 0 {
		t.Errorf("An Expiry Record with TTL 0 should have ExpirySeconds 0: %d", r.ExpirySeconds)
	}
	if r.HasExpiry() == true {
		t.Error("An Expiry Record with TTL 0 should not expire")
	}
}

func TestExpiryRecordHasExpired(t *testing.T) {
	clock := test_helpers.MockedUTC()

	expiryOptions := DefaultOptions
	expiryOptions.CompactorFunc = func() { return }
	expiryOptions.Clock = clock
	expiry, _ := NewExpiry(DefaultOptions)

	// A iteration number larger than 10 and divisible by 10
	iterationCount := 500
	sorted := make([]ExpiryRecord, iterationCount)

	for i, offset := range rand.Perm(iterationCount) {
		key := fmt.Sprintf("key_%d", i)
		// Insert key into our slice at offset position (making it sorted)
		sorted[offset] = expiry.NewExpiryRecordTTL(key, time.Duration(offset+1)*time.Second)
	}

	clock.Add(time.Second)

	for len(sorted) > 0 {
		aheadSize := rand.Intn(iterationCount/10) + 1
		// chunkSize can't be larger than the keys left in the sorted list
		chunkSize := test_helpers.Min(len(sorted), aheadSize)

		testGroup := sorted[:chunkSize]

		for _, r := range testGroup {
			if r.HasExpired(expiry.Clock.Now()) != false {
				t.Fail()
			}
		}
		// Advance the clock by a set amount to then verify the expected keys expired
		clock.Add(time.Duration(aheadSize) * time.Second)

		for _, r := range testGroup {
			if r.HasExpired(expiry.Clock.Now()) == false {
				t.Fail()
			}
		}

		// Re-slice ready for next loop
		sorted = sorted[chunkSize:]
	}

}
