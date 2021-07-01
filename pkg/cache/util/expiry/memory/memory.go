// memory is a thread safe library for handling TTL of keys
// as an expiry solution, `memory` is only suitable of in-memory store as the exoiry records do not persist
package memory

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/minotar/imgd/pkg/cache"
	"github.com/minotar/imgd/pkg/cache/util/expiry"
)

// Duration between times when we clear out all old data from
// the memory cache.
const COMPACTOR_INTERVAL = 5 * time.Second

// Handles tracking of expiration times.
type MemoryExpiry struct {
	*expiry.Expiry
	removeFunc func(key string) error
	records    []expiry.ExpiryRecord
	mu         sync.Mutex
}

func expiryOptions(compactorFunc func()) *expiry.Options {
	expiryOptions := expiry.DefaultOptions
	expiryOptions.CompactorFunc = compactorFunc
	expiryOptions.CompactorInterval = COMPACTOR_INTERVAL
	return expiryOptions
}

// NewMemoryExoury takes a `removeFunc` which is the function that will be called when an expiry takes place
func NewMemoryExpiry(removeFunc func(key string) error) (*MemoryExpiry, error) {
	if removeFunc == nil {
		// If function is missing, then throw an error
		return nil, fmt.Errorf("missing Memory Expiry Remove function")
	}

	m := &MemoryExpiry{removeFunc: removeFunc}

	e, err := expiry.NewExpiry(expiryOptions(m.RemoveExpired))
	if err != nil {
		// The only error here would be the m.RemoveExpired function no longer being valid
		return nil, err
	}

	m.Expiry = e

	return m, nil
}

// Adds a key and associated TTL to the expiry records.
// A TTL here can be 0 (no expiry) - though it will _not_ be added to the Expiry records
func (m *MemoryExpiry) AddExpiry(key string, ttl time.Duration) {

	if ttl == time.Duration(0) {
		// Todo: log that no expiry was set for key?
		return
	}

	record := expiry.NewExpiryRecordTTL(key, m.Clock, ttl)
	expires := record.Expiry()

	m.mu.Lock()
	defer m.mu.Unlock()

	list := m.records

	l := len(list)
	if l == 0 || list[l-1].Expiry().Before(expires) {
		// Special case: if the ttl is later than every other element,
		// just append it to the end.
		list = append(list, record)
	} else {
		// Otherwise, just do a binary search and insert it.
		idx := sort.Search(l, func(i int) bool {
			return !list[i].Expiry().Before(expires)
		})

		list = append(list, expiry.ExpiryRecord{})
		copy(list[idx+1:], list[idx:])
		list[idx] = record
	}

	m.records = list
}

// RemoveExpiry allows the storage backend to tell us it removed a key and we can stop tracking it
// This is crucial if the store has performed an eviction to avoid the expiry records needlessly growing/leaking memory
func (m *MemoryExpiry) RemoveExpiry(key interface{}, _ interface{}) {
	var keyStr string

	switch detectedType := key.(type) {
	case string:
		keyStr = key.(string)
	case []byte:
		keyStr = string(key.([]byte))
	default:
		// Todo: Log an error!
		fmt.Printf("RemoveExpiry can't deal with type %T\n", detectedType)
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	list := m.records

	// We loop through from oldest/soonest to expire as this is most likely where eviction was
	for i, val := range list {
		if val.Key == keyStr {
			// take the first i records and combine with the ones _after_ i
			m.records = append(list[:i], list[i+1:]...)
			return
		}
	}
}

func (m *MemoryExpiry) getRecord(key string) (expiry.ExpiryRecord, error) {
	for _, rec := range m.records {
		if rec.Key == key {
			return rec, nil
		}
	}
	return expiry.ExpiryRecord{}, cache.ErrNoExpiry
}

// GetExpiry grabs the Expiry value for a given key
// A "Zero" time is returned if the key is not in the Expiry records
func (m *MemoryExpiry) GetExpiry(key string) time.Time {
	// Todo: I think this should be locking? What if range changes beneath us?
	m.mu.Lock()
	defer m.mu.Unlock()

	rec, err := m.getRecord(key)
	if err != nil {
		return time.Time{}
	}

	return rec.Expiry()
}

// GetTTL grabs the TTL of the key (always >=1s), or 0 if no expiry/not found
// An error is returned if the key does not exist in the expiry records (no expiry/not found)
func (m *MemoryExpiry) GetTTL(key string) (time.Duration, error) {

	rec, err := m.getRecord(key)
	if err != nil {
		return 0, err
	}
	return rec.TTL(m.Clock.Now())
}

func (m *MemoryExpiry) Len() uint {
	return uint(len(m.records))
}

// Returns all keys that have expired, removing them from the
// expiry's records afterwards.
func (m *MemoryExpiry) Compact() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := m.Clock.Now()
	idx := 0
	removed := []string{}
	for idx < len(m.records) && m.records[idx].HasExpired(now) {
		removed = append(removed, m.records[idx].Key)
		idx++
	}

	m.records = m.records[idx:]
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
