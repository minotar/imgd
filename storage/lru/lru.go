package lru

import (
	"sync"
	"time"

	"github.com/hashicorp/golang-lru"
	"github.com/minotar/imgd/storage"
	"github.com/minotar/imgd/storage/util/expiry"
)

// ensure that the storage.Storage interface is implemented
var _ storage.Storage = new(LruCache)

// Duration between times when we clear out all old data from
// the memory cache.
const COMPACTION_INTERVAL = 5 * time.Second

type LruCache struct {
	mu     sync.Mutex
	cache  *lru.Cache
	expiry *expiry.Expiry
	closer chan bool
}

func New() *LruCache {
	freshCache, err := lru.New(512)
	if err != nil {
		return nil
	}
	lc := &LruCache{
		cache:  freshCache,
		expiry: expiry.NewExpiry(),
		closer: make(chan bool),
	}
	go lc.runCompactor()

	return lc
}

func (l *LruCache) Insert(key string, value []byte, ttl time.Duration) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.cache.Add(key, value)
	l.expiry.Add(key, ttl)
	return nil
}

func (l *LruCache) Retrieve(key string) ([]byte, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	value, ok := l.cache.Get(key)
	if ok {
		return value.([]byte), nil
	}
	return nil, storage.ErrNotFound
}

func (l *LruCache) Flush() error {
	l.cache.Purge()
	return nil
}

func (l *LruCache) Len() uint {
	return uint(l.cache.Len())
}

func (l *LruCache) Size() uint64 {
	return 0
}

func (l *LruCache) Close() {
	l.closer <- true
}

func (l *LruCache) compact() {
	l.mu.Lock()
	expired := l.expiry.Compact()
	l.mu.Unlock()

	for _, str := range expired {
		l.cache.Remove(str)
	}
}

func (l *LruCache) runCompactor() {
	ticker := time.NewTicker(COMPACTION_INTERVAL)

COMPACT:
	for {
		select {
		case <-l.closer:
			break COMPACT
		case <-ticker.C:
			l.compact()
		}
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	ticker.Stop()
	l.cache = nil
}
