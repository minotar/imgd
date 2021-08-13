package config

import (
	"flag"
	"fmt"

	"github.com/minotar/imgd/pkg/cache"
	"github.com/minotar/imgd/pkg/cache/bolt_cache"
)

type Config struct {
	bolt_cache.BoltCacheConfig
}

func (c *Config) RegisterFlags(f *flag.FlagSet, cacheID string) {

	c.BoltCacheConfig.RegisterFlags(f, cacheID)
}

func NewCache(cfg *Config) (cache.Cache, error) {

	if cfg.BoltCacheConfig.Name != "" {
		return bolt_cache.NewBoltCache(&cfg.BoltCacheConfig)
	}
	return nil, fmt.Errorf("no cache was specififed")
}
