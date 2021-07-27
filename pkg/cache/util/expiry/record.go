// Common methods for Expiry packages
package expiry

import (
	"time"

	"github.com/minotar/imgd/pkg/cache"
	"github.com/minotar/imgd/pkg/util/tinytime"
)

// ExpiryRecord is an efficient way to encode the Expiry time in a uint32
type ExpiryRecord struct {
	Key string
	// An unsigned int32 is good until 2100...
	Expiry tinytime.TinyTime
}

func NewExpiryRecord(key string, expires time.Time) ExpiryRecord {
	return ExpiryRecord{
		Key:    key,
		Expiry: tinytime.NewTinyTime(expires),
	}
}

func NewExpiryRecordTTL(key string, clock clock, ttl time.Duration) ExpiryRecord {
	var expiry time.Time
	if ttl == time.Duration(0) {
		// a Duration of 0 means does not expire and we set the epoch to 0
		expiry = time.Unix(0, 0).UTC()
	} else {
		expiry = clock.Now().Add(ttl)
	}

	return NewExpiryRecord(key, expiry)
}

// HasExpiry determines where the key has an expiry value
func (r *ExpiryRecord) HasExpiry() bool {
	// 0 seconds is "no expiry"
	// if ExpirySeconds is not 0, then it has an Expiry (true)
	// if ExpirySeconds is 0, then it has no Expiry (false)
	return !r.Expiry.IsZero()
}

// HasExpired uses the given time.Time to determine if the key is valid
func (r *ExpiryRecord) HasExpired(now time.Time) bool {
	if r.HasExpiry() && r.Expiry.Time().Before(now) {
		// If Expiry is _before_ now, then it's expired
		return true
	}
	// Either no Expiry, or it's not expired
	return false
}

// TTL uses a specific time.Time ("now") and works out the TTL of the key (always >=1s), or 0 if no expiry
// An error is returned if the key does not exist in the expiry records (no expiry)
func (r *ExpiryRecord) TTL(now time.Time) (time.Duration, error) {
	if r.HasExpiry() {
		ttl := r.Expiry.Time().Sub(now)
		if ttl < time.Duration(time.Second) {
			// Technically, we could get back a 0 or less Duration - but 0 is "no expiry"
			ttl = time.Duration(time.Second)
		}
		return ttl, nil
	}
	// No expiry is a 0 TTL
	return 0, cache.ErrNoExpiry
}
