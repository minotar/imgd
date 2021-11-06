// Migrate from bolt to badger
package migrate_cache

import (
	"fmt"
	"time"

	"github.com/boltdb/bolt"
	"github.com/minotar/imgd/pkg/cache"
	"github.com/minotar/imgd/pkg/cache/badger_cache"
	"github.com/minotar/imgd/pkg/cache/bolt_cache"
	store_expiry "github.com/minotar/imgd/pkg/cache/util/expiry/store"
)

const MIN_RECACHE_TTL = time.Duration(1) * time.Minute

type MigrateCacheConfig struct {
	cache.CacheConfig
	bolt_cache.BoltCacheConfig
	badger_cache.BadgerCacheConfig
	OldCache *bolt_cache.BoltCache
	NewCache *badger_cache.BadgerCache
}

type MigrateCache struct {
	*MigrateCacheConfig
}

var _ cache.Cache = new(MigrateCache)

func NewMigrateCache(cfg *MigrateCacheConfig) (*MigrateCache, error) {
	cfg.Logger = cfg.Logger.With(
		"cacheName", cfg.Name,
		"cacheType", "MigrateCache",
		"cacheParent", "MigrateCache", // The caches will overwrite the cacheType
	)
	cfg.Logger.Infof("initializing MigrateCache")
	mc := &MigrateCache{MigrateCacheConfig: cfg}

	cfg.BoltCacheConfig.Logger = cfg.Logger
	cfg.BadgerCacheConfig.Logger = cfg.Logger

	bolt, err := bolt_cache.NewBoltCache(&cfg.BoltCacheConfig)
	if err != nil {
		return nil, err
	}

	badger, err := badger_cache.NewBadgerCache(&cfg.BadgerCacheConfig)
	if err != nil {
		return nil, err
	}

	mc.OldCache = bolt
	mc.NewCache = badger
	cfg.Logger.Infof("initialized MigrateCache \"%s\"", mc.Name())
	return mc, nil
}

func (mc *MigrateCache) Name() string {
	return mc.CacheConfig.Name
}

func (mc *MigrateCache) Insert(key string, value []byte) error {
	return mc.NewCache.Insert(key, value)
}

func (mc *MigrateCache) InsertTTL(key string, value []byte, ttl time.Duration) error {
	return mc.NewCache.InsertTTL(key, value, ttl)
}

func (mc *MigrateCache) Retrieve(key string) ([]byte, error) {
	var errors []error
	for i, c := range []cache.Cache{mc.NewCache, mc.OldCache} {
		mc.Logger.Debugf("Retrieving \"%s\" from cache %d \"%s\"", key, i, c.Name())

		value, err := c.Retrieve(key)
		if err == cache.ErrNotFound {
			// errors logic at end handles ErrNotFound
			continue
		} else if err != nil {
			// This is a cache related error (vs. a missing key)
			mc.Logger.Errorf("Error retrieving key \"%s\" from cache %d (%s): %s", key, i, c.Name(), err)
			errors = append(errors, err)
			continue
		}

		return value, nil
	}

	// If we had a genuine error, `errors` would be populated, otherwise, it must just be ErrNotFound
	if errors != nil {
		return nil, fmt.Errorf("error(s) retrieving \"%s\" from cache(s): %+v", key, errors)
	}
	return nil, cache.ErrNotFound
}

// Probably won't be used too much
func (mc *MigrateCache) TTL(key string) (time.Duration, error) {
	var errors []error
	for i, c := range []cache.Cache{mc.NewCache, mc.OldCache} {
		mc.Logger.Debugf("Getting TTL of key \"%s\" from cache %d (%s)", key, i, c.Name())

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

func (mc *MigrateCache) Remove(key string) error {
	var errors []error
	for i, c := range []cache.Cache{mc.NewCache, mc.OldCache} {
		mc.Logger.Debugf("Removing key \"%s\" from cache %d (%s)", key, i, c.Name())

		err := c.Remove(key)
		if err != nil {
			mc.Logger.Errorf("Error removing key \"%s\" from cache %d (%s): %s", key, i, c.Name(), err)
			errors = append(errors, err)
		}
	}

	if errors != nil {
		return fmt.Errorf("error(s) removing \"%s\" from cache(s): %+v", key, errors)
	}
	return nil
}

func (mc *MigrateCache) Flush() error {
	var errors []error
	for i, c := range []cache.Cache{mc.NewCache, mc.OldCache} {
		mc.Logger.Debugf("Flushing cache %d (%s)", i, c.Name())

		err := c.Flush()
		if err != nil {
			mc.Logger.Errorf("Error flushing cache %d (%s): %s", i, c.Name(), err)
			errors = append(errors, err)
		}
	}
	if errors != nil {
		return fmt.Errorf("error(s) flushing cache(s): %+v", errors)
	}
	return nil
}

func (mc *MigrateCache) Len() uint {
	var maxLen uint
	for i, c := range []cache.Cache{mc.NewCache, mc.OldCache} {
		mc.Logger.Debugf("Getting length of cache %d (%s)", i, c.Name())

		cacheLen := c.Len()
		mc.Logger.Debugf("Length of cache %d (%s) is %d", i, c.Name(), cacheLen)
		if cacheLen > maxLen {
			maxLen = cacheLen
		}
	}
	return maxLen
}

func (mc *MigrateCache) Size() uint64 {
	var maxSize uint64
	for i, c := range []cache.Cache{mc.NewCache, mc.OldCache} {
		mc.Logger.Debugf("Getting size of cache %d (%s)", i, c.Name())

		cacheSize := c.Size()
		mc.Logger.Debugf("Size of cache %d (%s) is %d", i, c.Name(), cacheSize)
		if cacheSize > maxSize {
			maxSize = cacheSize
		}
	}
	return maxSize
}

func (mc *MigrateCache) Migrate() {
	var scannedCount, errorCount int

	dbLength := int(mc.OldCache.Len())
	logger := mc.Logger.With("boltLength", dbLength)
	logger.Info("Starting migration from Bolt -> Badger")
	start := time.Now()

	mc.OldCache.DB.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(mc.OldCache.Bucket)).Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			scannedCount++
			key := string(k)

			present, _ := mc.NewCache.Key(key)
			if !present {
				bse := store_expiry.DecodeStoreEntry(k, v)
				ttl, _ := bse.TTL(mc.OldCache.StoreExpiry.Clock.Now())
				err := mc.NewCache.InsertTTL(key, bse.Value, ttl)
				if err != nil {
					errorCount++
				}
			}
		}

		return nil
	})

	dur := time.Since(start)
	logger = logger.With(
		"scannedCount", scannedCount,
		"errorCount", errorCount,
		"duration", dur,
	)
	logger.Info("Key Migration finished")
}

func (mc *MigrateCache) Start() {
	mc.Logger.Info("starting MigrateCache")
	mc.NewCache.Start()
	mc.Migrate()
}

func (mc *MigrateCache) Stop() {
	mc.NewCache.Stop()
}

func (mc *MigrateCache) Close() {
	mc.Logger.Debug("closing MigrateCache")
	mc.Stop()
	mc.OldCache.Close()
	mc.NewCache.Close()
}
