package tiered_cache

import (
	"fmt"
	"time"

	"github.com/minotar/imgd/pkg/cache"
)

type TieredCache struct {
	Caches []*cache.Cache
}

var _ cache.Cache = new(TieredCache)

func NewTieredCache(caches []*cache.Cache) (*TieredCache, error) {
	// Start with empty struct we can pass around
	tc := &TieredCache{
		Caches: caches,
	}

	return tc, nil
}

func (tc *TieredCache) Insert(key string, value []byte) error {
	// InsertTTL with a TTL of 0 is added with no expiry
	return tc.InsertTTL(key, value, 0)
}

func (tc *TieredCache) InsertTTL(key string, value []byte, ttl time.Duration) error {
	var errors []error
	for i, c := range tc.Caches {
		fmt.Printf("InsertingTTL into cache %d", i)
		err := (*c).InsertTTL(key, value, ttl)
		if err != nil {
			errors = append(errors, err)
		}
	}
	if errors != nil {
		return fmt.Errorf("Error inserting: %+v", errors)
	}
	return nil
}

func (tc *TieredCache) Retrieve(key string) ([]byte, error) {
	// Todo: LOGIC HERE
	//return (*tc.Caches[0]).Retrieve(key)
	var errors []error
	for i, c := range tc.Caches {
		fmt.Printf("Retrieving \"%s\"from cache %d", key, i)
		value, err := (*c).Retrieve(key)
		if err == cache.ErrNotFound {
			continue
		} else if err != nil {
			// Todo: Probably just print here?
			errors = append(errors, err)
		}
		// We had a hit - we should update the earlier caches
		go func() {
			for i, c := range tc.Caches[:i] {
				//stuff
			}
		}

	}

}

func (tc *TieredCache) Remove(key string) error {
	var errors []error
	for i, c := range tc.Caches {
		fmt.Printf("Removing %s from cache %d", key, i)
		err := (*c).Remove(key)
		if err != nil {
			errors = append(errors, err)
		}
	}
	if errors != nil {
		return fmt.Errorf("Error removing key %s: %+v", key, errors)
	}
	return nil
}

func (tc *TieredCache) Flush() error {
	var errors []error
	for i, c := range tc.Caches {
		fmt.Printf("Flushing cache %d", i)
		err := (*c).Flush()
		if err != nil {
			errors = append(errors, err)
		}
	}
	if errors != nil {
		return fmt.Errorf("Error flushing: %+v", errors)
	}
	return nil
}

func (tc *TieredCache) Len() uint {
	var maxLen uint
	for i, c := range tc.Caches {
		fmt.Printf("Getting length from cache %d", i)
		cacheLen := (*c).Len()
		fmt.Printf("Cache length of %d is %d", i, cacheLen)
		if cacheLen > maxLen {
			maxLen = cacheLen
		}
	}
	return maxLen
}

func (tc *TieredCache) Size() uint64 {
	var maxSize uint64
	for i, c := range tc.Caches {
		fmt.Printf("Getting size of cache %d", i)
		cacheSize := (*c).Size()
		fmt.Printf("Cache size of %d is %d", i, cacheSize)
		if cacheSize > maxSize {
			maxSize = cacheSize
		}
	}
	return maxSize
}

func (tc *TieredCache) Start() {
	for i, c := range tc.Caches {
		fmt.Printf("Starting cache %d", i)
		(*c).Start()
	}
	return
}

func (tc *TieredCache) Stop() {
	for i, c := range tc.Caches {
		fmt.Printf("Stopping cache %d", i)
		(*c).Stop()
	}
	return
}

func (tc *TieredCache) Close() {
	for i, c := range tc.Caches {
		fmt.Printf("Closing cache %d", i)
		(*c).Close()
	}
	return
}
