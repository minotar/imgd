// Implement expiry within the the actual value storage
// The underlying store will need to use this when Inserting/Retrieving
package store

import (
	"encoding/binary"
	"time"

	"github.com/minotar/imgd/pkg/cache/util/expiry"
)

// Todo: if the Value will also have an expiry/freshness in it, we would do
// better to have a single uint32 vs. 2 of them.

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

// Todo: Does this need to be a pointer??? Less GC if not....
func (s *StoreExpiry) NewStoreEntry(key string, value []byte, ttl time.Duration) *StoreEntry {
	e := &StoreEntry{Value: value}
	// Need to add value data into the entry??
	e.ExpiryRecord = s.NewExpiryRecordTTL(key, ttl)
	return e
}

func getBytesExpirySeconds(buf []byte) uint32 {
	return binary.BigEndian.Uint32(buf[:4])
}

func HasBytesExpired(buf []byte, now time.Time) bool {
	if expirySeconds := getBytesExpirySeconds(buf[:4]); expirySeconds == 0 {
		// 0 seconds is "no expiry"
		return false
	} else if expiry := expiry.GetTimeFromEpoch32(expirySeconds); expiry.Before(now) {
		// If Expiry is _before_ now, then it's expired
		return true
	}
	return false
}

// Decode the raw bytes into the StoreEntry type
// Todo: Does this need to be a pointer??? Less GC if not....
func DecodeStoreEntry(key, value []byte) *StoreEntry {
	return &StoreEntry{
		ExpiryRecord: expiry.ExpiryRecord{
			Key:           string(key),
			ExpirySeconds: getBytesExpirySeconds(value[:4]),
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
	buf := make([]byte, 4+len(e.Value))

	// Todo: should we make this as an empty slice of length X
	// buf := make([]byte, 0,  4+len(e.Value))
	// buf = append(buf[4:], e.Value) ??

	// boltdb uses BigEndian in places, set the first 4 bytes as expiry
	binary.BigEndian.PutUint32(buf[:4], e.ExpirySeconds)
	// Fill remaining slice with the Value
	copy(buf[4:], e.Value)

	return []byte(e.Key), buf
}
