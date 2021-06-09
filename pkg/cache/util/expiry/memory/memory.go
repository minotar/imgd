// A thread safe library for handling TTL of keys
package memory

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/minotar/imgd/pkg/cache/util/expiry"
)

// Duration between times when we clear out all old data from
// the memory cache.
const COMPACTOR_INTERVAL = 5 * time.Second

// Handles tracking of expiration times.
type MemoryExpiry struct {
	mu         sync.Mutex
	tuples     []expiry.ExpiryRecord
	removeFunc func(key string) error

	*expiry.Expiry
}

func NewMemoryExpiry(removeFunc func(key string) error) (*MemoryExpiry, error) {
	m := &MemoryExpiry{removeFunc: removeFunc}

	expiryOptions := expiry.DefaultOptions
	expiryOptions.CompactorFunc = m.RemoveExpired
	expiryOptions.CompactorInterval = COMPACTOR_INTERVAL
	e, err := expiry.NewExpiry(expiryOptions)
	if err != nil {
		return nil, err
	}

	m.Expiry = e

	return m, nil
}

// Adds a key and associated TTL to the expiry records.
// A TTL here can be 0 and it is added to the expiry list as normal
func (m *MemoryExpiry) AddExpiry(key string, ttl time.Duration) {
	tuple := expiry.NewExpiryRecordTTL(key, m.Clock, ttl)
	expires := tuple.Expiry()

	m.mu.Lock()
	defer m.mu.Unlock()

	list := m.tuples

	l := len(list)
	if l == 0 || list[l-1].Expiry().Before(expires) {
		// Special case: if the ttl is later than every other element,
		// just append it to the end.
		list = append(list, tuple)
	} else {
		// Otherwise, just do a binary search and insert it.
		idx := sort.Search(l, func(i int) bool {
			return !list[i].Expiry().Before(expires)
		})

		list = append(list, expiry.ExpiryRecord{})
		copy(list[idx+1:], list[idx:])
		list[idx] = tuple
	}

	m.tuples = list
}

// RemoveExpiry allows the storage backend to tell us it removed a key and we can stop tracking it
func (m *MemoryExpiry) RemoveExpiry(key interface{}, _ interface{}) {
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

	m.mu.Lock()
	defer m.mu.Unlock()

	// We loop through from oldest/soonest to expire as this is most likely where eviction was
	for i, val := range m.tuples {
		if val.Key == keyStr {
			// take the first i tuples and combine with the ones _after_ i
			m.tuples = append(m.tuples[:i], m.tuples[i+1:]...)
			return
		}

	}
}

func (m *MemoryExpiry) GetExpiry(key string) (time.Time, error) {
	// We loop through every key
	for _, val := range m.tuples {
		if val.Key == key {
			return val.Expiry(), nil
		}
	}
	return time.Time{}, fmt.Errorf("No expiry value found for %s", key)
}

// Return the TTL of the key (always >0), or 0 if not present/no expiry
func (m *MemoryExpiry) GetTTL(key string) time.Duration {
	expiry, err := m.GetExpiry(key)
	if err != nil {
		// An error means it was not in the Expiry list
		return 0
	}

	ttl := expiry.Sub(m.Clock.Now())
	if ttl < time.Duration(1) {
		// Technically, we could get back a 0 or less Duration - but 0 is "no expiry"
		ttl = time.Duration(1)
	}
	return ttl
}

func (m *MemoryExpiry) Len() uint {
	return uint(len(m.tuples))
}

// Returns all keys that have expired, removing them from the
// expiry's records afterwards.
func (m *MemoryExpiry) Compact() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := m.Clock.Now()
	idx := 0
	removed := []string{}
	for idx < len(m.tuples) && m.tuples[idx].Expiry().Before(now) {
		removed = append(removed, m.tuples[idx].Key)
		idx++
	}

	m.tuples = m.tuples[idx:]
	return removed
}

func (m *MemoryExpiry) RemoveExpired() {
	expired := m.Compact()

	for _, key := range expired {
		err := m.removeFunc(key)
		if err != nil {
			fmt.Printf("Error removing expired key \"%s\": %s", key, err)
		}
	}
}
