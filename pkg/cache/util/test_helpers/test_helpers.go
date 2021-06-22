package test_helpers

import (
	"math/rand"
	"testing"
	"time"

	"github.com/minotar/imgd/pkg/cache"
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
	return &MockClock{
		TimeUTC(),
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

func InsertTTLAndRetrieve(c *cache.Cache, t *testing.T) {

}
