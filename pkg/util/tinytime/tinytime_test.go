package tinytime_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/minotar/imgd/pkg/util/tinytime"
)

const (
	TimeString  = "2021-05-19T00:00:00Z"
	TimeSeconds = 1621382400
)

var TimeBytes = []byte{96, 164, 85, 0}

// Mocked time (which isn't epoch as that's special)
func TimeUTC() time.Time {
	specificTime, _ := time.Parse(time.RFC3339, TimeString)
	return specificTime.UTC()
}

func TestNewTinyTime(t *testing.T) {
	mockTime := TimeUTC()
	tt := tinytime.NewTinyTime(mockTime)

	if tt != TimeSeconds {
		t.Errorf("Expected %d and not %d for TinyTime", TimeSeconds, tt)
	}

	tt_time := tt.Time()
	if !tt_time.Equal(mockTime) {
		t.Errorf("TinyTime should match original test time: %v", tt_time)
	}
}

func TestZeroTinyTime(t *testing.T) {
	var tt tinytime.TinyTime
	if !tt.IsZero() {
		t.Errorf("Freshly initialised TinyTime should be Zero")
	}

	epochTime := time.Unix(0, 0).UTC()
	tt_time := tt.Time()
	if !tt_time.Equal(epochTime) {
		t.Errorf("Zeroed TinyTime should be the Unix epoch: %v", tt_time)
	}
}

func TestTinyTimeEncode(t *testing.T) {
	mockTime := TimeUTC()
	tt := tinytime.NewTinyTime(mockTime)

	buf := make([]byte, 4)
	tt.Encode(buf)

	if !bytes.Equal(buf, TimeBytes) {
		t.Errorf("TinyTime did not encode to expected bytes: %v", buf)
	}
}

func TestTinyTimeDecode(t *testing.T) {
	buf := make([]byte, 4)
	copy(buf, TimeBytes)
	tt := tinytime.Decode(buf)

	mockTime := TimeUTC()
	tt_time := tt.Time()
	if !tt_time.Equal(mockTime) {
		t.Errorf("TinyTime should match original test time: %v", tt_time)
	}
}