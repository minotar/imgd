// cache_converter is the go between to migrate to/from v3/4
// To read the source data, it needs to loop through all it's keys
// For this reason, it will likely only ever support Redis (on v3) and BoltCache on v4
package cache_converter

import (
	"flag"
	"strings"
	"time"

	"github.com/minotar/imgd/pkg/util/log"

	radix_util "github.com/mediocregopher/radix.v2/util"
	"github.com/minotar/imgd/pkg/cache_converter/legacy_storage"
	"github.com/minotar/imgd/pkg/cache_converter/legacy_storage/radix"

	"github.com/boltdb/bolt"
	"github.com/minotar/imgd/pkg/cache"
	"github.com/minotar/imgd/pkg/cache/bolt_cache"
	cache_config "github.com/minotar/imgd/pkg/cache/util/config"
	store_expiry "github.com/minotar/imgd/pkg/cache/util/expiry/store"
)

type IteratingProcessor func(k string, v []byte, ttl time.Duration)

type CacheInsertProcessor func(string, []byte, time.Duration) error

type RedisConfig radix.RedisConfig

func (rc *RedisConfig) RegisterFlags(f *flag.FlagSet, cacheID string) {
	flagPath := "cache." + strings.ToLower(cacheID) + "."
	f.StringVar(&rc.Address, flagPath+"address", "localhost:6379", "Redis host:port")
	f.StringVar(&rc.Auth, flagPath+"auth", "", "Redis Authentication")
	f.IntVar(&rc.DB, flagPath+"db", 0, "Redis Database")
}

type Config struct {
	Logger log.Logger
	MinTTL time.Duration

	CacheUUIDv4     *cache_config.Config `yaml:"cache_uuid"`
	CacheUserDatav4 *cache_config.Config `yaml:"cache_userdata"`

	CacheUUIDv3     *RedisConfig `yaml:"cachev3_uuid"`
	CacheUserDatav3 *RedisConfig `yaml:"cachev3_userdata"`
}

func (c *Config) RegisterFlags(f *flag.FlagSet) {

	f.DurationVar(&c.MinTTL, "cacheconv.min-ttl", 1*time.Hour, "Min TTL to consider converting")

	c.CacheUUIDv4 = &cache_config.Config{}
	c.CacheUserDatav4 = &cache_config.Config{}

	c.CacheUUIDv3 = &RedisConfig{}
	c.CacheUserDatav3 = &RedisConfig{}

	c.CacheUUIDv4.RegisterFlags(f, "UUID")
	c.CacheUserDatav4.RegisterFlags(f, "UserData")

	c.CacheUUIDv3.RegisterFlags(f, "Legacy-UUID")
	c.CacheUserDatav3.RegisterFlags(f, "Legacy-UserData")
}

type CacheConverter struct {
	Cfg      Config
	Cachesv4 struct {
		UUID     cache.Cache
		UserData cache.Cache
	}
	Cachesv3 struct {
		UUID     legacy_storage.Storage
		UserData legacy_storage.Storage
	}
	Now time.Time
}

func New(cfg Config) (*CacheConverter, error) {

	// Caches v3
	cacheUUIDv3, err := radix.New(radix.RedisConfig{
		Network: "tcp",
		Address: cfg.CacheUUIDv3.Address,
		Auth:    cfg.CacheUUIDv3.Auth,
		DB:      cfg.CacheUUIDv3.DB,
		Size:    10,
	})
	if err != nil {
		cfg.Logger.Errorf("Unable to create cache UUIDv3: %v", err)
	}

	cacheUserDatav3, err := radix.New(radix.RedisConfig{
		Network: "tcp",
		Address: cfg.CacheUserDatav3.Address,
		Auth:    cfg.CacheUserDatav3.Auth,
		DB:      cfg.CacheUserDatav3.DB,
		Size:    10,
	})
	if err != nil {
		cfg.Logger.Errorf("Unable to create cache UserDatav3: %v", err)
	}

	// Caches v4
	cfg.CacheUUIDv4.Logger = cfg.Logger
	cacheUUIDv4, err := cache_config.NewCache(cfg.CacheUUIDv4)
	if err != nil {
		cfg.Logger.Panicf("Unable to create cache UUIDv4: %v", err)
	}
	// Skip running the compactor
	//cacheUUIDv4.Start()

	cfg.CacheUserDatav4.Logger = cfg.Logger
	cacheUserDatav4, _ := cache_config.NewCache(cfg.CacheUserDatav4)
	if err != nil {
		cfg.Logger.Panicf("Unable to create cache UserDatav4: %v", err)
	}
	// Skip running the compactor
	//cacheUserDatav4.Start()

	cacheConverter := &CacheConverter{
		Cfg: cfg,
		Now: time.Now(),
	}

	cacheConverter.Cachesv3.UUID = cacheUUIDv3
	cacheConverter.Cachesv3.UserData = cacheUserDatav3
	cacheConverter.Cachesv4.UUID = cacheUUIDv4
	cacheConverter.Cachesv4.UserData = cacheUserDatav4

	return cacheConverter, nil
}

func (cc *CacheConverter) boltIterator(bc *bolt_cache.BoltCache, processor IteratingProcessor) {

	var count int

	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bc.Bucket))

		return b.ForEach(func(k, v []byte) error {
			count++
			key := string(k)
			key = strings.ToLower(key)
			data := make([]byte, len(v))
			copy(data, v)
			bse := store_expiry.DecodeStoreEntry(k, data)
			ttl, err := bse.TTL(cc.Now)
			if err != nil {
				cc.Cfg.Logger.Warnf("%s key threw a TTL error: %v", key, err)
				return nil
			}
			if ttl < cc.Cfg.MinTTL {
				cc.Cfg.Logger.Debugf("Skipping %s as TTL %s was less than %s", key, ttl, cc.Cfg.MinTTL)
				return nil
			}

			processor(key, bse.Value, ttl)

			if (count % 1000) == 0 {
				cc.Cfg.Logger.Infof("%d keys have been processed out of ~%d", count, bc.Len())
			}
			return nil
		})
	})
	if err != nil {
		bc.Logger.Errorf("Error iterating through: %v", err)
	}
}

func (cc *CacheConverter) radixIterator(r *radix.RedisCache, processor IteratingProcessor) {

	scanner := radix_util.NewScanner(r.Pool(), radix_util.ScanOpts{Command: "SCAN"})

	var count int

	for scanner.HasNext() {
		key := scanner.Next()
		count++
		key = strings.ToLower(key)
		value, err := r.Retrieve(key)
		if err != nil {
			cc.Cfg.Logger.Warnf("%s key threw a retrieve error: %v", key, err)
			continue
		}
		resp := r.Pool().Cmd("TTL", key)
		seconds, err := resp.Int()
		if err != nil {
			cc.Cfg.Logger.Warnf("%s key threw a TTL error: %v", key, err)
			continue
		}

		ttl := time.Duration(seconds) * time.Second
		if ttl < cc.Cfg.MinTTL {
			cc.Cfg.Logger.Debugf("Skipping %s as TTL %s was less than %s", key, ttl, cc.Cfg.MinTTL)
			continue
		}
		processor(key, value, ttl)

		if (count % 1000) == 0 {
			cc.Cfg.Logger.Infof("%d keys have been processed out of ~%d", count, r.Len())
		}
	}
	if err := scanner.Err(); err != nil {
		cc.Cfg.Logger.Fatalf("Error iterating through: %v", err)
	}
}

// IteratingProcessor(func(_, _ []byte) {})

// V4 -> V3
func (cc *CacheConverter) MigrateUUIDV4toV3() {

	boltCache := cc.Cachesv4.UUID.(*bolt_cache.BoltCache)
	cc.boltIterator(boltCache, processUUIDv4(cc.Cfg.Logger, cc.Cachesv3.UUID.Insert))

}

// V4 -> V3
func (cc *CacheConverter) MigrateUserDataV4toV3() {

	boltCache := cc.Cachesv4.UserData.(*bolt_cache.BoltCache)
	cc.boltIterator(boltCache, processUserDatav4(cc.Cfg.Logger, cc.Cachesv3.UserData.Insert))

}

func (cc *CacheConverter) boltSync(bc *bolt_cache.BoltCache) {
	start := time.Now()
	err := bc.DB.Sync()
	dur := time.Now().Sub(start)
	cc.Cfg.Logger.Infof("BoltCache fsync took %s", dur)
	if err != nil {
		cc.Cfg.Logger.Errorf("BoltCache fsync failed: %v", err)
	}
}

// V3 -> V4
func (cc *CacheConverter) MigrateUUIDV3toV4() {

	redisCache := cc.Cachesv3.UUID.(*radix.RedisCache)
	cc.Cfg.Logger.Infof("Size of Redis is %d keys", redisCache.Len())
	boltCache := cc.Cachesv4.UUID.(*bolt_cache.BoltCache)
	boltCache.DB.NoSync = true
	cc.radixIterator(redisCache, processUUIDv3(cc.Cfg.Logger, cc.Cachesv4.UUID.InsertTTL))
	cc.boltSync(boltCache)
}

// V3 -> V4
func (cc *CacheConverter) MigrateUserDataV3toV4() {

	redisCache := cc.Cachesv3.UserData.(*radix.RedisCache)
	cc.Cfg.Logger.Infof("Size of Redis is %d keys", redisCache.Len())
	boltCache := cc.Cachesv4.UUID.(*bolt_cache.BoltCache)
	boltCache.DB.NoSync = true
	cc.radixIterator(redisCache, processUserDatav3(cc.Cfg.Logger, cc.Cachesv4.UserData.InsertTTL))
	cc.boltSync(boltCache)
}
