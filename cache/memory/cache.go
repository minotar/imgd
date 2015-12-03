package memory

import (
	"sync"
	"time"
)

// Duration between times when we clear out all old data from
// the memory cache.
const COMPACTION_INTERVAL = 5 * time.Second

type memoryCacher interface {
	Find(path string) []byte
	Delete(path string)
	Insert(path string, ptr []byte)
}

// This is a simple in-memory cache that uses prefix trees to insert
// and retrieve cache items from memory.
type MemoryCache struct {
	mu     sync.Mutex
	cache  memoryCacher
	expiry *expiry
	closer chan bool
}

func New() *MemoryCache {
	mc := &MemoryCache{
		cache:  newMemoryMap(),
		expiry: newExpiry(),
		closer: make(chan bool),
	}
	go mc.runCompactor()

	return mc
}

func (m *MemoryCache) Insert(key string, value []byte, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cache.Insert(key, value)
	m.expiry.Add(key, ttl)
	return nil
}

func (m *MemoryCache) Find(key string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.cache.Find(key), nil
}

func (m *MemoryCache) Flush() error {
	m.cache = newMemoryMap()
	return nil
}

func (m *MemoryCache) Close() {
	m.closer <- true
}

func (m *MemoryCache) compact() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, str := range m.expiry.Compact() {
		m.cache.Delete(str)
	}
}

func (m *MemoryCache) runCompactor() {
	ticker := time.NewTicker(COMPACTION_INTERVAL)

COMPACT:
	for {
		select {
		case <-m.closer:
			break COMPACT
		case <-ticker.C:
			m.compact()
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	ticker.Stop()
	m.cache = nil
}
