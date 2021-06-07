// A non-thread safe library for handling TTL of keys
package memory_expiry

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// Duration between times when we clear out all old data from
// the memory cache.
const COMPACTION_INTERVAL = 5 * time.Second

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
	mu                  sync.Mutex
	clock               clock
	removeFunc          func(key string) error
	tuples              []expiryTuple
	closer              chan bool
	compaction_interval time.Duration
}

func NewExpiry(removeFunc func(key string) error) *Expiry {
	e := &Expiry{
		clock:               realClock{},
		removeFunc:          removeFunc,
		closer:              make(chan bool),
		compaction_interval: COMPACTION_INTERVAL,
	}

	return e
}

func (e *Expiry) Start() {
	go e.runCompactor()
	return
}

func (e *Expiry) Stop() {
	e.closer <- true
}

// Adds a key and associated TTL to the expiry records.
// A TTL here can be 0 and it is added to the expiry list as normal
func (e *Expiry) AddExpiry(key string, ttl time.Duration) {
	expires := e.clock.Now().Add(ttl)
	tuple := expiryTuple{key, expires}

	e.mu.Lock()
	defer e.mu.Unlock()

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

// RemoveExpiry allows the storage backend to tell us it removed a key and we can stop tracking it
func (e *Expiry) RemoveExpiry(key interface{}, _ interface{}) {
	var keyStr string

	switch detectedType := key.(type) {
	case string:
		keyStr = key.(string)
	case []byte:
		keyStr = string(key.([]byte))
	default:
		// Todo: Log an error!
		fmt.Printf("I don't know about type %T!\n", detectedType)
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// We loop through from oldest/soonest to expire as this is most likely where eviction was
	for i, val := range e.tuples {
		if val.key == keyStr {
			// take the first i tuples and combine with the ones _after_ i
			e.tuples = append(e.tuples[:i], e.tuples[i+1:]...)
			return
		}

	}
}

func (e *Expiry) GetExpiry(key string) (time.Time, error) {
	// We loop through every key
	for _, val := range e.tuples {
		if val.key == key {
			return val.expires, nil
		}
	}
	return time.Time{}, fmt.Errorf("No expiry value found for %s", key)
}

// Return the TTL of the key (always >0), or 0 if not present
func (e *Expiry) GetTTL(key string) time.Duration {
	expiry, err := e.GetExpiry(key)
	if err != nil {
		// An error means it was not in the Expiry list
		return 0
	}

	ttl := expiry.Sub(e.clock.Now())
	if ttl == time.Duration(0) {
		// Technically, we could get back a 0 Duration - but that is ambiguous
		ttl = time.Duration(1)
	}
	return ttl
}

func (e *Expiry) LenExpiry() uint {
	return uint(len(e.tuples))
}

// Returns all keys that have expired, removing them from the
// expiry's records afterwards.
func (e *Expiry) Compact() []string {
	e.mu.Lock()
	defer e.mu.Unlock()

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

func (e *Expiry) RemoveExpired() {
	expired := e.Compact()

	for _, key := range expired {
		err := e.removeFunc(key)
		if err != nil {
			fmt.Printf("Error removing expired key \"%s\": %s", key, err)
		}
	}
}

// runCompactor is in its own goroutine and thus needs the closer to stop
func (e *Expiry) runCompactor() {
	// Todo: manually run first?
	ticker := time.NewTicker(e.compaction_interval)

COMPACT:
	for {
		select {
		case <-e.closer:
			break COMPACT
		case <-ticker.C:
			e.RemoveExpired()
		}
	}

	ticker.Stop()
}
