package bolt_cache

import (
	"encoding/binary"
	"time"
)

// Todo: if the Value will also have an expiry/freshness in it, we would do
// better to have a single uint32 vs. 2 of them.

type BoltCacheEntry struct {
	Key   string
	Value []byte
	// An unsigned int32 is good until 2100...
	ExpirySeconds uint32
}

func (bc *BoltCache) NewBoltCacheEntry(key string, value []byte, ttl time.Duration) *BoltCacheEntry {
	var expiry uint32
	if ttl == time.Duration(0) {
		// We hardcode a 0 ttl to the int32 epoch - otherwise we get a int64 epoch...
		expiry = 0
	} else {
		expiry = uint32(bc.clock.Now().Add(ttl).Unix())
	}

	return &BoltCacheEntry{
		Key:           string(key),
		Value:         value,
		ExpirySeconds: expiry,
	}
}

func DecodeBoltCacheEntry(key, value []byte) *BoltCacheEntry {
	return &BoltCacheEntry{
		Key:           string(key),
		Value:         value[4:],
		ExpirySeconds: getExpirySeconds(value[:4]),
	}
}

func getExpirySeconds(buf []byte) uint32 {
	return binary.BigEndian.Uint32(buf[:4])
}

func getExpiry(expirySeconds uint32) (t time.Time) {
	return time.Unix(int64(expirySeconds), 0).UTC()
}

func HasExpired(buf []byte, now time.Time) bool {
	if expirySeconds := getExpirySeconds(buf[:4]); expirySeconds == 0 {
		// 0 seconds is "no expiry"
		return false
	} else if expiry := getExpiry(expirySeconds); expiry.Before(now) {
		// If Expiry is _before_ now, then it's expired
		return true
	}
	return false
}

// Super simple format
//  |--------------------|
//  | timestamp | value  |
//  |--------------------|
//  | uint32    | []byte |
//  |--------------------|

func (e *BoltCacheEntry) Encode() (key, value []byte) {
	// Uint32 takes up 4 bytes
	buf := make([]byte, 4+len(e.Value))

	// boltdb uses BigEndian in places, set the first 4 bytes as expiry
	binary.BigEndian.PutUint32(buf[:4], e.ExpirySeconds)
	// Fill remaining slice with the Value
	copy(buf[4:], e.Value)

	return []byte(e.Key), buf
}

func (e *BoltCacheEntry) Expiry() time.Time {
	return time.Unix(int64(e.ExpirySeconds), 0).UTC()
}
