// Migrate from bolt to badger
package migrate_cache

import (
	"errors"
	"flag"
	"fmt"
	"strings"
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

	performMigration bool
}

func (c *MigrateCacheConfig) RegisterFlags(f *flag.FlagSet, cacheID string) {
	f.BoolVar(&c.performMigration, strings.ToLower("cache."+cacheID+".migrate"), false, "Peform Bolt -> Badger migration")
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
		if i == 1 {
			// OldCache had the value, update new cache
			ttl, err := mc.OldCache.TTL(key)
			if err == nil {
				mc.Logger.Debugf("Adding \"%s\" to NewCache", key)
				go mc.NewCache.InsertTTL(key, value, ttl)
			}
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

const (
	MIGRATION_FINISHED     = "it has finished"
	MIGRATION_NOT_FINISHED = "not finished"
)

func (mc *MigrateCache) SetMigrationStatus(finished bool, keyMarker string) error {
	bool_value := MIGRATION_NOT_FINISHED
	if finished {
		bool_value = MIGRATION_FINISHED
	}

	err := mc.InsertTTL("MINOTAR_MIGRATION_BOOL", []byte(bool_value), time.Minute*time.Duration(10080))
	if err != nil {
		return err
	}

	err = mc.InsertTTL("MINOTAR_MIGRATION_MARKER", []byte(keyMarker), time.Minute*time.Duration(10080))
	if err != nil {
		return err
	}
	return nil

}

func (mc *MigrateCache) GetMigrationStatus() (finished bool, keyMarker string) {
	bool_value, err := mc.Retrieve("MINOTAR_MIGRATION_BOOL")
	if err != nil {
		mc.Logger.Infof("Failure to get migration status: %v", err)
		return
	}
	if string(bool_value) == MIGRATION_FINISHED {
		return true, ""
	}

	keyMarkerBytes, err := mc.Retrieve("MINOTAR_MIGRATION_MARKER")
	if err != nil {
		mc.Logger.Infof("Failure to get migration marker: %v", err)
		return
	}

	return false, string(keyMarkerBytes)
}

func firstOrSeek(c *bolt.Cursor, keyMarker string) (k, v []byte) {
	if keyMarker == "" {
		return c.First()
	} else {
		return c.Seek([]byte(keyMarker))
	}
}

var ErrCompactionFinished = errors.New("compaction has finished")

func (mc *MigrateCache) Migrate() {

	// keymarker can be an empty string
	migrateCompleted, keyMarker := mc.GetMigrationStatus()
	if migrateCompleted {
		mc.Logger.Info("Migration reports it has already completed")
		return
	}

	var scannedCount, errorCount int

	dbLength := int(mc.OldCache.Len())
	logger := mc.Logger.With("boltLength", dbLength)
	logger.Info("Starting migration from Bolt -> Badger")
	start := time.Now()

	err := mc.OldCache.DB.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(mc.OldCache.Bucket)).Cursor()

		var k []byte
		for k, v := firstOrSeek(c, keyMarker); k != nil; k, v = c.Next() {
			scannedCount++
			key := string(k)

			if scannedCount%1000 == 0 {
				logger.Infof("Marking current keyMarker %s", key)
				mc.SetMigrationStatus(false, key)
			}

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

		if k == nil {
			return ErrCompactionFinished
		}

		return nil
	})

	if err == ErrCompactionFinished {
		logger.Info("Marking migration completion")
		mc.SetMigrationStatus(true, "")
	}

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
	if mc.performMigration {
		mc.Logger.Info("Migration is enabled - starting")
		go mc.Migrate()
	}
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
