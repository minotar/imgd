package config

import (
	"flag"
	"fmt"
	"strings"

	"github.com/minotar/imgd/pkg/cache"
	"github.com/minotar/imgd/pkg/cache/badger_cache"
	"github.com/minotar/imgd/pkg/cache/bolt_cache"
	"github.com/minotar/imgd/pkg/cache/migrate_cache"
	"github.com/minotar/imgd/pkg/util/log"
)

const (
	CACHE_LIST    = "{bolt|badger|migrate}"
	CACHE_DEFAULT = "bolt"
)

type Config struct {
	CacheType string
	Logger    log.Logger
	cache.CacheConfig

	bolt_cache.BoltCacheConfig
	badger_cache.BadgerCacheConfig
	migrate_cache.MigrateCacheConfig
}

func (c *Config) RegisterFlags(f *flag.FlagSet, cacheID string) {

	f.StringVar(&c.CacheType, strings.ToLower("cache."+cacheID+".backend"), CACHE_DEFAULT, "Backend cache to use "+CACHE_LIST)
	c.CacheConfig.RegisterFlags(f, cacheID)

	c.BoltCacheConfig.RegisterFlags(f, cacheID)
	c.BadgerCacheConfig.RegisterFlags(f, cacheID)
}

func NewCache(cfg *Config) (cache.Cache, error) {
	cfg.CacheConfig.Logger = cfg.Logger
	cfg.BoltCacheConfig.CacheConfig = cfg.CacheConfig
	cfg.BadgerCacheConfig.CacheConfig = cfg.CacheConfig
	cfg.MigrateCacheConfig.CacheConfig = cfg.CacheConfig

	switch strings.ToLower(cfg.CacheType) {
	case "bolt":
		return bolt_cache.NewBoltCache(&cfg.BoltCacheConfig)
	case "badger":
		return badger_cache.NewBadgerCache(&cfg.BadgerCacheConfig)
	case "migrate":
		cfg.MigrateCacheConfig.BoltCacheConfig = cfg.BoltCacheConfig
		cfg.MigrateCacheConfig.BadgerCacheConfig = cfg.BadgerCacheConfig
		return migrate_cache.NewMigrateCache(&cfg.MigrateCacheConfig)
	default:
		//return bolt_cache.NewBoltCache(&cfg.BoltCacheConfig)
		return nil, fmt.Errorf("no cache was specififed")
	}
}
