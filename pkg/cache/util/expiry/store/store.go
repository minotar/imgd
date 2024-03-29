// Implement expiry within the the actual value storage
// The underlying store will need to use this when Inserting/Retrieving
package store

import (
	"time"

	"github.com/minotar/imgd/pkg/cache/util/expiry"
	"github.com/minotar/imgd/pkg/util/tinytime"
)

type StoreExpiry struct {
	*expiry.Expiry
}

func NewStoreExpiry(compactorFunc func(), compactionInterval time.Duration) (*StoreExpiry, error) {
	s := &StoreExpiry{}

	expiryOptions := expiry.DefaultOptions
	expiryOptions.CompactorFunc = compactorFunc
	expiryOptions.CompactorInterval = compactionInterval
	e, err := expiry.NewExpiry(expiryOptions)
	if err != nil {
		return nil, err
	}

	s.Expiry = e

	return s, nil
}

func (s *StoreExpiry) NewStoreEntry(key string, value []byte, ttl time.Duration) StoreEntry {
	e := StoreEntry{Value: value}
	// Need to add value data into the entry??
	e.ExpiryRecord = s.NewExpiryRecordTTL(key, ttl)
	return e
}

func HasBytesExpired(buf []byte, now time.Time) bool {
	tt_expiry := tinytime.Decode(buf[:4])

	if tt_expiry.IsZero() {
		// 0 seconds is "no expiry"
		return false
	} else if tt_expiry.Time().Before(now) {
		// If Expiry is _before_ now, then it's expired
		return true
	}
	return false
}

// Decode the raw bytes into the StoreEntry type
func DecodeStoreEntry(key, value []byte) StoreEntry {
	return StoreEntry{
		ExpiryRecord: expiry.ExpiryRecord{
			Key:    string(key),
			Expiry: tinytime.Decode(value[:4]),
		},
		Value: value[4:],
	}
}

// StoreEntry retains the efficiency of ExpiryRecord
// Further encodes the expiry in the the Value
type StoreEntry struct {
	Value []byte
	expiry.ExpiryRecord
}

// Super simple format
//  |--------------------|
//  | timestamp | value  |
//  |--------------------|
//  | uint32    | []byte |
//  |--------------------|

func (e *StoreEntry) Encode() (key, value []byte) {
	// Uint32 takes up 4 bytes
	buf := make([]byte, 4, 4+len(e.Value))

	// boltdb uses BigEndian in places, set the first 4 bytes as expiry
	e.Expiry.Encode(buf[:4])
	// Fill remaining slice with the Value
	buf = append(buf, e.Value...)

	return []byte(e.Key), buf
}
