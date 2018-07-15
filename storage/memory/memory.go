package memory

import (
	"sync"
	"time"

	"github.com/minotar/imgd/storage"
	"github.com/minotar/imgd/storage/util/expiry"
)

// Duration between times when we clear out all old data from
// the memory cache.
const COMPACTION_INTERVAL = 5 * time.Second

type memoryCacher interface {
	Insert(path string, ptr []byte)
	Retrieve(path string) ([]byte, error)
	Delete(path string)
	Len() uint
}

// ensure that the storage.Storage interface is implemented
var _ storage.Storage = new(MemoryCache)

// This is a simple in-memory cache that uses prefix trees to insert
// and retrieve cache items from memory.
type MemoryCache struct {
	mu     sync.Mutex
	cache  memoryCacher
	size   int
	expiry *expiry.Expiry
	closer chan bool
}

func (m *MemoryCache) Insert(key string, value []byte, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cache.Insert(key, value)
	m.expiry.Add(key, ttl)
	return nil
}

func (m *MemoryCache) Retrieve(key string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.cache.Retrieve(key)
}

func (m *MemoryCache) Flush() error {
	m.cache = newMemoryMap(m.size)
	return nil
}

func (m *MemoryCache) Len() uint {
	return m.cache.Len()
}

// Size will not be accurate for an in-memory Cache
func (m *MemoryCache) Size() uint64 {
	return 0
}

func (m *MemoryCache) Close() {
	m.closer <- true
}

func New(maxEntries int) (*MemoryCache, error) {
	mc := &MemoryCache{
		cache:  newMemoryMap(maxEntries),
		size:   maxEntries,
		expiry: expiry.NewExpiry(),
		closer: make(chan bool),
	}
	go mc.runCompactor()

	return mc, nil
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
