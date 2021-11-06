package badger_cache

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/minotar/imgd/pkg/cache"
	store_expiry "github.com/minotar/imgd/pkg/cache/util/expiry/store"
	cache_metrics "github.com/minotar/imgd/pkg/cache/util/metrics"
	"github.com/minotar/imgd/pkg/storage/badger_store"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Minimum Duration between full bucket scans looking for expired keys
	COMPACTION_SCAN_INTERVAL = 15 * time.Minute
)

func NewBadgerCacheConfig(cacheConfig cache.CacheConfig, path string) *BadgerCacheConfig {
	return &BadgerCacheConfig{
		CacheConfig: cacheConfig,
		path:        path,
	}
}

type BadgerCacheConfig struct {
	opDuration     prometheus.ObserverVec
	expiredCounter *prometheus.CounterVec
	cache.CacheConfig
	path string
}

func (c *BadgerCacheConfig) RegisterFlags(f *flag.FlagSet, cacheID string) {
	defaultPath := strings.ToLower("/tmp/badger_cache_" + cacheID + "/")
	f.StringVar(&c.path, strings.ToLower("cache."+cacheID+".badger-path"), defaultPath, "Badger data folder (cannot be used by other caches)")
}

type BadgerCache struct {
	*badger_store.BadgerStore
	*store_expiry.StoreExpiry
	*BadgerCacheConfig
}

// ensure that the cache.Cache interface is implemented
var _ cache.Cache = new(BadgerCache)

func NewBadgerCache(cfg *BadgerCacheConfig) (*BadgerCache, error) {
	cfg.Logger = cfg.Logger.With(
		"cacheName", cfg.Name,
		"cacheType", "BadgerCache",
	)
	cfg.Logger.Infof("initializing BadgerCache \"%s\"", cfg.path)
	bs, err := badger_store.NewBadgerStore(cfg.path, cfg.Logger)
	if err != nil {
		return nil, err
	}

	bc := &BadgerCache{BadgerStore: bs, BadgerCacheConfig: cfg}
	bc.opDuration = cache_metrics.NewCacheOperationDuration("BadgerCache", bc.Name())
	bc.expiredCounter = cache_metrics.NewCacheExpiredCounter("BadgerCache", bc.Name())
	cache_metrics.NewCacheSizeGauge("BadgerCache", bc.Name(), bc.Size)
	//cache_metrics.NewCacheLenGauge("BadgerCache", bc.Name(), bc.Len)

	// Create a StoreExpiry using the BadgerCache method
	se, err := store_expiry.NewStoreExpiry(bc.ExpiryScan, COMPACTION_SCAN_INTERVAL)
	if err != nil {
		return nil, err
	}
	bc.StoreExpiry = se

	cfg.Logger.Infof("initialized BadgerCache \"%s\"", bc.Name())
	return bc, nil
}

func (bc *BadgerCache) Name() string {
	return bc.CacheConfig.Name
}

func (bc *BadgerCache) Insert(key string, value []byte) error {
	return bc.InsertTTL(key, value, 0)
}

func (bc *BadgerCache) InsertTTL(key string, value []byte, ttl time.Duration) error {
	cacheTimer := prometheus.NewTimer(bc.opDuration.WithLabelValues("insert"))
	defer cacheTimer.ObserveDuration()

	if ttl == 0 {
		return bc.BadgerStore.Insert(key, value)
	}

	err := bc.DB.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(key), value).WithTTL(ttl)
		err := txn.SetEntry(e)
		return err
	})
	if err != nil {
		return fmt.Errorf("inserting \"%s\": %s", key, err)
	}
	return nil
}

func (bc *BadgerCache) Retrieve(key string) ([]byte, error) {
	cacheTimer := prometheus.NewTimer(bc.opDuration.WithLabelValues("retrieve"))
	defer cacheTimer.ObserveDuration()

	return bc.BadgerStore.Retrieve(key)
}

// TTL returns an error if the key does not exist, or it has no expiry
// Otherwise return a TTL (always at least 1 Second per `StoreExpiry`)
func (bc *BadgerCache) TTL(key string) (time.Duration, error) {
	cacheTimer := prometheus.NewTimer(bc.opDuration.WithLabelValues("ttl"))
	defer cacheTimer.ObserveDuration()

	var expiresAt uint64
	err := bc.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		expiresAt = item.ExpiresAt()
		return nil
	})

	if err != nil {
		return 0, err
	}

	if expiresAt == 0 {
		// Not an error, so must be a non-expiring key
		return 0, cache.ErrNoExpiry
	}

	now := bc.StoreExpiry.Clock.Now()

	ttl := time.Unix(int64(expiresAt), 0).UTC().Sub(now) // Could be negative?
	if ttl < time.Duration(time.Second) {
		ttl = time.Duration(time.Second)
	}
	return ttl, nil
}

func (bc *BadgerCache) Remove(key string) error {
	cacheTimer := prometheus.NewTimer(bc.opDuration.WithLabelValues("remove"))
	defer cacheTimer.ObserveDuration()

	return bc.BadgerStore.Remove(key)
}

// Ran on interval by the StoreExpiry
func (bc *BadgerCache) ExpiryScan() {
	cacheTimer := prometheus.NewTimer(bc.opDuration.WithLabelValues("expiryScan"))
	defer cacheTimer.ObserveDuration()

	logger := bc.Logger.With(
		"dbLength", bc.Len(),
		"dbSize", bc.Size(),
	)
	logger.Info("Starting expiryScan")
	start := time.Now()

	err := bc.DB.RunValueLogGC(0.5)
	if err == badger.ErrNoRewrite {
		logger.Debug("No Value Log GC needed")
	} else if err != nil {
		logger.Errorf("BadgerCache had an error GC'ing DB: %v", err)
	}

	dur := time.Since(start)
	logger = bc.Logger.With(
		"dbLength", bc.Len(),
		"dbSize", bc.Size(),
		"duration", dur,
	)
	logger.Info("expiryScan has scanned all keys")
}

func (bc *BadgerCache) Start() {
	bc.Logger.Info("starting BadgerCache")
	// Start the Expiry monitor/compactor
	bc.StoreExpiry.Start()
}

func (bc *BadgerCache) Stop() {
	bc.Logger.Info("stopping BadgerCache")
	// Start the Expiry monitor/compactor
	bc.StoreExpiry.Stop()
}

func (bc *BadgerCache) Close() {
	bc.Logger.Debug("closing BadgerCache")
	bc.Stop()
	bc.BadgerStore.Close()
}
