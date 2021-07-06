package tiered_cache

import (
	"fmt"
	"time"

	"github.com/minotar/imgd/pkg/cache"
)

type TieredCache struct {
	*TieredCacheConfig
}

type TieredCacheConfig struct {
	Caches []cache.Cache
	cache.CacheConfig
}

var _ cache.Cache = new(TieredCache)

// Todo: There is a huge amount of debug/fluff/fmt.Print awaiting logging decisions

func NewTieredCache(cfg *TieredCacheConfig) (*TieredCache, error) {
	cfg.Logger.Infof("initializing TieredCache with %d cache(s)", cfg.Name, len(cfg.Caches))
	tc := &TieredCache{TieredCacheConfig: cfg}
	cfg.Logger.Infof("initialized TieredCache \"%s\"", tc.Name())
	return tc, nil
}

func (tc *TieredCache) Name() string {
	return tc.CacheConfig.Name
}

func (tc *TieredCache) Insert(key string, value []byte) error {
	// InsertTTL with a TTL of 0 is added with no expiry
	return tc.InsertTTL(key, value, 0)
}

func (tc *TieredCache) InsertTTL(key string, value []byte, ttl time.Duration) error {
	var errors []error
	for i, c := range tc.Caches {
		tc.Logger.Debugf("InsertingTTL into cache %d (%s)", i, c.Name())
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

func (tc *TieredCache) updateCaches(cacheID int, key string, value []byte) {
	validCache := tc.Caches[cacheID]
	ttl, err := validCache.TTL(key)
	if err == cache.ErrNotFound {
		tc.Logger.Warnf("Cache %d (%s) reports key \"%s\" is now a cache.ErrNotFound", cacheID, validCache.Name(), key)
		return
	} else if err == cache.ErrNoExpiry {
		// Todo: does this logic make sense????
		tc.Logger.Warnf("Cache %d (%s) reports key \"%s\" had no TTL/Expiry - not re-adding", cacheID, validCache.Name(), key)
		return
	} else if err != nil {
		tc.Logger.Warnf("Cache %d (%s) reports key \"%s\"  with TTL err: %s\n", cacheID, validCache.Name(), key, err)
		return
	}

	if ttl < time.Duration(1)*time.Minute {
		tc.Logger.Debugf("TTL of key \"%s\" was less than a minute - not re-adding", key)
		return
	}

	for i, c := range tc.Caches[:cacheID] {
		tc.Logger.Debugf("Inserting key \"%s\" into cache %d (%s)", key, i, c.Name())

		err := c.InsertTTL(key, value, ttl)
		if err != nil {
			tc.Logger.Errorf("Error inserting key \"%s\" into cache %d (%s): %s", key, i, c.Name(), err)
			return
		}
	}

}

func (tc *TieredCache) Retrieve(key string) ([]byte, error) {
	// Todo: LOGIC HERE
	//return (*tc.Caches[0]).Retrieve(key)
	var errors []error
	for i, c := range tc.Caches {
		tc.Logger.Debugf("Retrieving \"%s\" from cache %d \"%s\"", key, i, c.Name())
		value, err := c.Retrieve(key)
		if err == cache.ErrNotFound {
			continue
		} else if err != nil {
			// Todo: Probably just print here?
			errors = append(errors, err)
			continue
		}
		// We had a hit - we should update the earlier caches
		go tc.updateCaches(i, key, value)

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
		fmt.Printf("Removing %s from cache %d\n", key, i)
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
	tc.Logger.Info("starting TieredCache")
	for i, c := range tc.Caches {
		tc.Logger.Infof("starting cache %d \"%s\"", i, c.Name())
		c.Start()
	}
}

func (tc *TieredCache) Stop() {
	tc.Logger.Info("stopping TieredCache")
	for i, c := range tc.Caches {
		tc.Logger.Infof("stopping cache %d \"%s\"", i, c.Name())
		c.Stop()
	}
}

func (tc *TieredCache) Close() {
	tc.Logger.Debug("closing TieredCache")
	tc.Stop()
	for i, c := range tc.Caches {
		tc.Logger.Debugf("closing cache %d \"%s\"", i, c.Name())
		c.Close()
	}
}
