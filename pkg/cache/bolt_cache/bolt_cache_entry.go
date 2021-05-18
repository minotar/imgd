package bolt_cache

import (
	"encoding/binary"
	"time"
)

type BoltCacheEntry struct {
	Key    string
	Value  []byte
	Expiry time.Time
}

func NewBoltCacheEntry(key, value []byte, ttl time.Duration) *BoltCacheEntry {
	var expiry time.Time
	if ttl.Nanoseconds() == 0 {
		expiry = time.Unix(0, 0)
	} else {
		expiry = time.Now().Add(ttl)
	}

	return &BoltCacheEntry{
		Key:    string(key),
		Value:  value,
		Expiry: expiry,
	}
}

func DecodeBoltCacheEntry(key, value []byte) *BoltCacheEntry {
	return &BoltCacheEntry{
		Key:    string(key),
		Value:  value[4:],
		Expiry: getExpiry(value[:4]),
	}
}

func getExpiry(buf []byte) time.Time {
	decodedEpoch := binary.BigEndian.Uint32(buf[:4])
	return time.Unix(int64(decodedEpoch), 0)
}

func (e *BoltCacheEntry) Encode() (key, value []byte) {
	// Uint32 takes up 4 bytes
	buf := make([]byte, 4+len(e.Value))

	// An unsigned int32 is good until 2100...
	epochExpiry := uint32(e.Expiry.Unix())

	// boltdb uses BigEndian in places, set the first 4 bytes as expiry
	binary.BigEndian.PutUint32(buf[:4], epochExpiry)
	// Fill remaining slice with the Value
	copy(buf[4:], e.Value)

	return []byte(e.Key), buf
}
