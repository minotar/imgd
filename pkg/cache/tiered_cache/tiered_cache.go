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
	cache.CacheConfig
	Caches []cache.Cache
}

var _ cache.Cache = new(TieredCache)

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

func (tc *TieredCache) cacheInsert(endCacheID int, key string, value []byte, ttl time.Duration) []error {
	var errors []error
	// endCacheID is how many of the caches to insert into
	for i, c := range tc.Caches[:endCacheID] {
		tc.Logger.Debugf("Inserting key \"%s\" into cache %d (%s)", key, i, c.Name())

		err := c.InsertTTL(key, value, ttl)
		if err != nil {
			tc.Logger.Errorf("Error inserting key \"%s\" into cache %d (%s): %s", key, i, c.Name(), err)
			errors = append(errors, err)
		}
	}
	return errors
}

func (tc *TieredCache) InsertTTL(key string, value []byte, ttl time.Duration) error {
	// We want to insert into all caches, so the end ID is the number of caches
	endCacheID := len(tc.Caches)
	errors := tc.cacheInsert(endCacheID, key, value, ttl)
	if errors != nil {
		return fmt.Errorf("error(s) inserting \"%s\" into cache(s): %+v", key, errors)
	}
	return nil
}

func (tc *TieredCache) updateCaches(cacheID int, key string, value []byte) {
	validCache := tc.Caches[cacheID]
	ttl, err := validCache.TTL(key)
	if err == cache.ErrNotFound {
		// Likely that the key expired within the split second since?
		tc.Logger.Infof("Cache %d (%s) reports key \"%s\" is now a cache.ErrNotFound", cacheID, validCache.Name(), key)
		return
	} else if err == cache.ErrNoExpiry {
		// It's a key which doesn't have an expiry set - possibly badly added to the Cache?
		tc.Logger.Warnf("Cache %d (%s) reports key \"%s\" had no TTL/Expiry - not re-adding", cacheID, validCache.Name(), key)
		return
	} else if err != nil {
		// This is a cache related error (vs. a missing key/expiry)
		tc.Logger.Errorf("Cache %d (%s) reports key \"%s\" with TTL err: %s\n", cacheID, validCache.Name(), key, err)
		return
	}

	if ttl < time.Duration(1)*time.Minute {
		tc.Logger.Debugf("TTL of key \"%s\" was less than a minute - not re-adding", key)
		return
	}

	// cacheID was the cache the data came from, so we insert in the caches before that
	tc.cacheInsert(cacheID, key, value, ttl)
}

func (tc *TieredCache) Retrieve(key string) ([]byte, error) {
	var errors []error
	for i, c := range tc.Caches {
		tc.Logger.Debugf("Retrieving \"%s\" from cache %d \"%s\"", key, i, c.Name())
		value, err := c.Retrieve(key)
		if err == cache.ErrNotFound {
			continue
		} else if err != nil {
			// This is a cache related error (vs. a missing key)
			tc.Logger.Errorf("Error retrieving key \"%s\" from cache %d (%s): %s", key, i, c.Name(), err)
			errors = append(errors, err)
			continue
		}
		// We had a hit - we should update the earlier caches
		go tc.updateCaches(i, key, value)

		return value, nil
	}

	return nil, fmt.Errorf("error(s) retrieving \"%s\" from cache(s): %+v", key, errors)
}

// Probably won't be used too much
func (tc *TieredCache) TTL(key string) (time.Duration, error) {
	var errors []error
	for i, c := range tc.Caches {
		tc.Logger.Debugf("Getting TTL of key \"%s\" from cache %d (%s)", key, i, c.Name())
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
		tc.Logger.Debugf("Removing key \"%s\" from cache %d (%s)", key, i, c.Name())
		err := c.Remove(key)
		if err != nil {
			tc.Logger.Errorf("Error removing key \"%s\" from cache %d (%s): %s", key, i, c.Name(), err)
			errors = append(errors, err)
		}
	}

	if errors != nil {
		return fmt.Errorf("error(s) removing \"%s\" from cache(s): %+v", key, errors)
	}
	return nil
}

func (tc *TieredCache) Flush() error {
	var errors []error
	for i, c := range tc.Caches {
		tc.Logger.Debugf("Flushing cache %d (%s)", i, c.Name())
		err := c.Flush()
		if err != nil {
			tc.Logger.Errorf("Error flushing cache %d (%s): %s", i, c.Name(), err)
			errors = append(errors, err)
		}
	}
	if errors != nil {
		return fmt.Errorf("error(s) flushing cache(s): %+v", errors)
	}
	return nil
}

func (tc *TieredCache) Len() uint {
	var maxLen uint
	for i, c := range tc.Caches {
		tc.Logger.Debugf("Getting length of cache %d (%s)", i, c.Name())
		cacheLen := c.Len()
		tc.Logger.Debugf("Length of cache %d (%s) is %d", i, c.Name(), cacheLen)
		if cacheLen > maxLen {
			maxLen = cacheLen
		}
	}
	return maxLen
}

func (tc *TieredCache) Size() uint64 {
	var maxSize uint64
	for i, c := range tc.Caches {
		tc.Logger.Debugf("Getting size of cache %d (%s)", i, c.Name())
		cacheSize := c.Size()
		tc.Logger.Debugf("Size of cache %d (%s) is %d", i, c.Name(), cacheSize)
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
