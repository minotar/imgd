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

type Config struct {
	Logger          log.Logger
	CacheUUIDv4     *cache_config.Config `yaml:"cache_uuid"`
	CacheUserDatav4 *cache_config.Config `yaml:"cache_userdata"`
}

func (c *Config) RegisterFlags(f *flag.FlagSet) {
	c.CacheUUIDv4 = &cache_config.Config{}
	c.CacheUserDatav4 = &cache_config.Config{}

	c.CacheUUIDv4.RegisterFlags(f, "UUID")
	c.CacheUserDatav4.RegisterFlags(f, "UserData")

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
		Address: "localhost:6379",
		Auth:    "",
		DB:      0,
		Size:    10,
	})
	if err != nil {
		cfg.Logger.Panicf("Unable to create cache UUIDv3: %v", err)
	}

	cacheUserDatav3, err := radix.New(radix.RedisConfig{
		Network: "tcp",
		Address: "localhost:6379",
		Auth:    "",
		DB:      1,
		Size:    10,
	})
	if err != nil {
		cfg.Logger.Panicf("Unable to create cache UserDatav3: %v", err)
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

func (cc *CacheConverter) MigrateV3toV4() {

	// UUIDs

	redisCachePool := cc.Cachesv3.UUID.(*radix.RedisCache).Pool()

	scanner := radix_util.NewScanner(redisCachePool, radix_util.ScanOpts{Command: "SCAN"})

	for scanner.HasNext() {
		cc.Cfg.Logger.Infof("next: %q", scanner.Next())
	}
	if err := scanner.Err(); err != nil {
		cc.Cfg.Logger.Fatal(err)
	}

}

func (cc *CacheConverter) boltIterator(bc *bolt_cache.BoltCache, processor IteratingProcessor) {

	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bc.Bucket))

		return b.ForEach(func(k, v []byte) error {
			key := string(k)
			key = strings.ToLower(key)
			data := make([]byte, len(v))
			copy(data, v)
			bse := store_expiry.DecodeStoreEntry(k, data)
			ttl, err := bse.TTL(cc.Now)
			if err != nil {
				cc.Cfg.Logger.Warnf("%s key threw a TTL error: %v", key, err)
			}
			processor(key, bse.Value, ttl)
			return nil
		})
	})
	if err != nil {
		bc.Logger.Errorf("Error iterating through: %v", err)
	}
}

func (cc *CacheConverter) radixIterator(r *radix.RedisCache, processor IteratingProcessor) {

	scanner := radix_util.NewScanner(r.Pool(), radix_util.ScanOpts{Command: "SCAN"})

	for scanner.HasNext() {
		key := scanner.Next()
		key = strings.ToLower(key)
		value, err := r.Retrieve(key)
		if err != nil {
			cc.Cfg.Logger.Warnf("%s key threw a retrieve error: %v", key, err)
		}
		resp := r.Pool().Cmd("TTL", key)
		seconds, err := resp.Int()
		if err != nil {
			cc.Cfg.Logger.Warnf("%s key threw a TTL error: %v", key, err)
		}

		ttl := time.Duration(seconds) * time.Second
		processor(key, value, ttl)
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

// V3 -> V4
func (cc *CacheConverter) MigrateUUIDV3toV4() {

	redisCache := cc.Cachesv3.UUID.(*radix.RedisCache)
	cc.radixIterator(redisCache, processUUIDv3(cc.Cfg.Logger, cc.Cachesv4.UUID.InsertTTL))

}

// V3 -> V4
func (cc *CacheConverter) MigrateUserDataV3toV4() {

	redisCache := cc.Cachesv3.UUID.(*radix.RedisCache)
	cc.radixIterator(redisCache, processUserDatav3(cc.Cfg.Logger, cc.Cachesv4.UUID.InsertTTL))

}
