package expiry

import (
	"math/rand"
	"testing"
	"time"

	"github.com/minotar/imgd/storage/util/helper"
	. "github.com/smartystreets/goconvey/convey"
)

type mockClock struct {
	now time.Time
}

func (m *mockClock) Now() time.Time {
	return m.now
}

func TestExpiry(t *testing.T) {
	Convey("Expiry should expire", t, func() {

		clock := &mockClock{time.Unix(0, 0)}
		ex := &Expiry{clock: clock}

		number := 5000
		sorted := make([]string, number)
		for _, offset := range rand.Perm(number) {
			str := helper.RandString(32)
			sorted[offset] = str
			ex.Add(str, time.Second*time.Duration(offset))
		}

		for len(sorted) > 0 {
			n := rand.Intn(number / 10)
			clock.now = clock.now.Add(time.Duration(n) * time.Second)

			sl := helper.Min(len(sorted), n)
			So(sorted[:sl], ShouldResemble, ex.Compact())
			sorted = sorted[sl:]
		}

		So(len(ex.Compact()), ShouldEqual, 0)
	})
}
