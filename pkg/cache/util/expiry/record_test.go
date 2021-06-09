package expiry

import (
	"testing"
	"time"

	"github.com/minotar/imgd/pkg/storage/util/test_helpers"
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

func timeUTC() time.Time {
	mockedTime, _ := time.Parse(time.RFC3339, "2021-05-19T00:00:00Z")
	return mockedTime.UTC()
}

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
