package tiered_cache

import (
	"fmt"
	"time"

	"github.com/minotar/imgd/pkg/cache"
)

type TieredCache struct {
	Caches []cache.Cache
}

var _ cache.Cache = new(TieredCache)

// Todo: There is a huge amount of debug/fluff/fmt.Print awaiting logging decisions

func NewTieredCache(caches []cache.Cache) (*TieredCache, error) {
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
		fmt.Printf("InsertingTTL into cache %d\n", i)
		err := c.InsertTTL(key, value, ttl)
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
		fmt.Printf("Retrieving \"%s\"from cache %d\n", key, i)
		value, err := c.Retrieve(key)
		if err == cache.ErrNotFound {
			continue
		} else if err != nil {
			// Todo: Probably just print here?
			errors = append(errors, err)
			continue
		}
		// We had a hit - we should update the earlier caches
		go func() {
			ttl, err := c.TTL(key)
			if err == cache.ErrNotFound {
				fmt.Printf("Cache %d which had returned %s gave TTL a cache.ErrNotFound\n", i, value)
				return
			} else if err == cache.ErrNoExpiry {
				// Todo: does this logic make sense????
				fmt.Printf("Cache %d reports key \"%s\" had no TTL/Expiry - not re-adding\n", i, key)
				return
			} else if err != nil {
				fmt.Printf("Cache %d which had returned %s gave TTL err: %s\n", i, value, err)
				return
			}
			if ttl < time.Duration(1)*time.Minute {
				fmt.Printf("TTL of key was less than a minute - not re-adding\n")
				return
			}
			for i, c := range tc.Caches[:i] {
				//stuff
				fmt.Printf("Inserting %s into cache %d\n", key, i)

				err := c.InsertTTL(key, value, ttl)
				if err != nil {
					fmt.Printf("Error inserting \"%s\" into %d: %s\n", key, i, err)
					return
				}
			}
		}()

		return value, nil

	}

	return nil, nil
}

// Probably won't be used too much
func (tc *TieredCache) TTL(key string) (time.Duration, error) {
	var errors []error
	for i, c := range tc.Caches {
		fmt.Printf("Get TTL of %s from cache %d\n", key, i)
		ttl, err := c.TTL(key)
		if err == cache.ErrNoExpiry {
			// Record has no expiry - we trust the first cache with this response
			// It's important the cache signified between NoExpiry vs. NotFound!
			return ttl, err
		} else if err != nil {
			// Todo: Probably just print here?
			errors = append(errors, err)
			continue
		}
		return ttl, err
	}

	return 0, errors[len(errors)-1]

}

func (tc *TieredCache) Remove(key string) error {
	var errors []error
	for i, c := range tc.Caches {
		fmt.Printf("Removing %s from cache %d\n\n", key, i)
		err := c.Remove(key)
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
		fmt.Printf("Flushing cache %d\n", i)
		err := c.Flush()
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
		fmt.Printf("Getting length from cache %d\n", i)
		cacheLen := c.Len()
		fmt.Printf("Cache length of %d is %d\n", i, cacheLen)
		if cacheLen > maxLen {
			maxLen = cacheLen
		}
	}
	return maxLen
}

func (tc *TieredCache) Size() uint64 {
	var maxSize uint64
	for i, c := range tc.Caches {
		fmt.Printf("Getting size of cache %d\n", i)
		cacheSize := c.Size()
		fmt.Printf("Cache size of %d is %d\n", i, cacheSize)
		if cacheSize > maxSize {
			maxSize = cacheSize
		}
	}
	return maxSize
}

func (tc *TieredCache) Start() {
	for i, c := range tc.Caches {
		fmt.Printf("Starting cache %d\n", i)
		c.Start()
	}
}

func (tc *TieredCache) Stop() {
	for i, c := range tc.Caches {
		fmt.Printf("Stopping cache %d\n", i)
		c.Stop()
	}
}

func (tc *TieredCache) Close() {
	tc.Stop()
	for i, c := range tc.Caches {
		fmt.Printf("Closing cache %d\n", i)
		c.Close()
	}
}
