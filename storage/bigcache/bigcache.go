package bigcache

import (
	"time"

	"github.com/allegro/bigcache"
	"github.com/minotar/imgd/storage"
)

// BigcacheCache stores our Redis Pool
type BigcacheCache struct {
	cache *bigcache.BigCache
}

// ensure that the storage.Storage interface is implemented
var _ storage.Storage = new(BigcacheCache)

// Insert will SET the key in Redis with an expiry TTL
func (b *BigcacheCache) Insert(key string, value []byte, ttl time.Duration) error {
	return b.cache.Set(key, value)
}

func (b *BigcacheCache) Retrieve(key string) ([]byte, error) {
	return b.cache.Get(key)
}

func (b *BigcacheCache) Flush() error {
	return b.cache.Reset()
}

func (b *BigcacheCache) Len() uint {
	return uint(b.cache.Len())
}

func (b *BigcacheCache) Size() uint64 {
	return 0
}

func (b *BigcacheCache) Close() {
	b.cache = nil
}

// Creates a new Redis cache instance, connecting to the given server
// and AUTHing with the provided password. If the password is an
// empty string, the AUTH command will not be run.
func New() (*BigcacheCache, error) {
	cache, err := bigcache.NewBigCache(bigcache.DefaultConfig(10 * time.Minute))
	return &BigcacheCache{cache}, err
}
