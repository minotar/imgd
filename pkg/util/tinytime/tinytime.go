// TinyTime sacrifices sub-second precision and longevity for minimal time storage
// One aim is to avoid storing pointers to timezones, so simply UTC
// Uses just 4 bytes when Encoded, also low memory
package tinytime

import (
	"encoding/binary"
	"time"
)

type TinyTime uint32

func NewTinyTime(t time.Time) TinyTime {
	seconds := t.Unix()
	return TinyTime(seconds)
}

// Time returns the time.Time based on the TinyTime
func (tt TinyTime) Time() time.Time {
	return time.Unix(int64(tt), 0).UTC()
}

// IsZero confirms if the TinyTime is uninitialized or Unix Epoch
func (tt TinyTime) IsZero() bool {
	return tt == 0
}

// Encode takes a byte slice of at least length 4 and sets the first 4 bytes as the TinyTime
func (tt TinyTime) Encode(b []byte) {
	binary.BigEndian.PutUint32(b[:4], uint32(tt))
}

// Decode takes a byte slice of at least 4 length and takes the first bytes as the TinyTime
func Decode(b []byte) TinyTime {
	return TinyTime(binary.BigEndian.Uint32(b[:4]))
}
