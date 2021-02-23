// A non-thread safe library for handling TTL of keys
package expiry

import (
	"sort"
	"time"
)

type clock interface {
	Now() time.Time
}

type realClock struct{}

func (r realClock) Now() time.Time { return time.Now() }

type expiryTuple struct {
	key     string
	expires time.Time
}

// Handles tracking of expiration times.
type Expiry struct {
	clock  clock
	tuples []expiryTuple
}

func NewExpiry() *Expiry {
	return &Expiry{clock: realClock{}}
}

// Adds a key and associated TTL to the expiry records.
func (e *Expiry) Add(key string, ttl time.Duration) {
	expires := e.clock.Now().Add(ttl)
	tuple := expiryTuple{key, expires}
	list := e.tuples

	l := len(list)
	if l == 0 || list[l-1].expires.Before(expires) {
		// Special case: if the ttl is later than every other element,
		// just append it to the end.
		list = append(list, tuple)
	} else {
		// Otherwise, just do a binary search and insert it.
		idx := sort.Search(l, func(i int) bool {
			return !list[i].expires.Before(expires)
		})

		list = append(list, expiryTuple{})
		copy(list[idx+1:], list[idx:])
		list[idx] = tuple
	}

	e.tuples = list
}

// Returns all keys that have expired, removing them from the
// expiry's records afterwards.
func (e *Expiry) Compact() []string {
	now := e.clock.Now()
	idx := 0
	removed := []string{}
	for idx < len(e.tuples) && e.tuples[idx].expires.Before(now) {
		removed = append(removed, e.tuples[idx].key)
		idx++
	}

	e.tuples = e.tuples[idx:]
	return removed
}
