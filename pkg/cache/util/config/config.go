package config

import (
	"flag"
	"fmt"
	"strings"

	"github.com/minotar/imgd/pkg/cache"
	"github.com/minotar/imgd/pkg/cache/badger_cache"
	"github.com/minotar/imgd/pkg/cache/bolt_cache"
	"github.com/minotar/imgd/pkg/util/log"
)

const (
	CACHE_LIST    = "{bolt|badger}"
	CACHE_DEFAULT = "bolt"
)

type Config struct {
	CacheType string
	Logger    log.Logger
	cache.CacheConfig
	bolt_cache.BoltCacheConfig
	badger_cache.BadgerCacheConfig
}

func (c *Config) RegisterFlags(f *flag.FlagSet, cacheID string) {

	f.StringVar(&c.CacheType, strings.ToLower("cache."+cacheID+".backend"), CACHE_DEFAULT, "Backend cache to use "+CACHE_LIST)
	c.CacheConfig.RegisterFlags(f, cacheID)

	c.BoltCacheConfig.RegisterFlags(f, cacheID)
	c.BadgerCacheConfig.RegisterFlags(f, cacheID)
}

func NewCache(cfg *Config) (cache.Cache, error) {
	cfg.CacheConfig.Logger = cfg.Logger

	switch strings.ToLower(cfg.CacheType) {
	case "bolt":
		cfg.BoltCacheConfig.CacheConfig = cfg.CacheConfig
		return bolt_cache.NewBoltCache(&cfg.BoltCacheConfig)
	case "badger":
		cfg.BadgerCacheConfig.CacheConfig = cfg.CacheConfig
		return badger_cache.NewBadgerCache(&cfg.BadgerCacheConfig)
	default:
		//return bolt_cache.NewBoltCache(&cfg.BoltCacheConfig)
		return nil, fmt.Errorf("no cache was specififed")
	}
}
