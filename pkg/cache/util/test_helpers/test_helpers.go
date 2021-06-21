package test_helpers

import (
	"math/rand"
	"time"

	store_test_helpers "github.com/minotar/imgd/pkg/storage/util/test_helpers"
)

var (
	Min        = store_test_helpers.Min
	RandString = store_test_helpers.RandString
)

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
