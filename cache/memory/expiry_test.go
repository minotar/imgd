package memory

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockClock struct {
	now time.Time
}

func (m *mockClock) Now() time.Time {
	return m.now
}

func TestExpiry(t *testing.T) {
	clock := &mockClock{time.Unix(0, 0)}
	ex := &expiry{clock: clock}

	number := 5000
	sorted := make([]string, number)
	for _, offset := range rand.Perm(number) {
		str := randString(32)
		sorted[offset] = str
		ex.Add(str, time.Second*time.Duration(offset))
	}

	for len(sorted) > 0 {
		n := rand.Intn(number / 10)
		clock.now = clock.now.Add(time.Duration(n) * time.Second)

		sl := min(len(sorted), n)
		assert.Equal(t, sorted[:sl], ex.Compact())
		sorted = sorted[sl:]
	}

	assert.Zero(t, len(ex.Compact()))
}
