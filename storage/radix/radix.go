package radix

import (
	"time"

	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/redis"
	"github.com/minotar/imgd/storage"
	"github.com/minotar/imgd/storage/util/redisinfo"
)

// RedisCache stores our Redis Pool
type RedisCache struct {
	pool *pool.Pool
}

// RedisConfig contains the configuration for the redis storage
type RedisConfig struct {
	Network string
	Address string
	Auth    string
	DB      int
	Size    int
}

// ensure that the storage.Storage interface is implemented
var _ storage.Storage = new(RedisCache)

// Insert will SET the key in Redis with an expiry TTL
func (r *RedisCache) Insert(key string, value []byte, ttl time.Duration) error {
	resp := r.pool.Cmd("SET", key, value, "EX", ttl.Seconds())
	return resp.Err
}

func (r *RedisCache) Retrieve(key string) ([]byte, error) {
	resp := r.pool.Cmd("GET", key)
	if resp.Err != nil {
		return nil, resp.Err
	}

	bytes, err := resp.Bytes()
	if err != nil && err.Error() == "response is nil" {
		return bytes, storage.ErrNotFound
	}
	return bytes, err
}

func (r *RedisCache) Flush() error {
	resp := r.pool.Cmd("FLUSHDB")
	return resp.Err
}

func (r *RedisCache) Len() uint {
	resp := r.pool.Cmd("DBSIZE")
	if resp.Err != nil {
		return 0
	}

	size, err := resp.Int()
	if err != nil {
		return 0
	}

	return uint(size)
}

func (r *RedisCache) Size() uint64 {
	resp := r.pool.Cmd("INFO")
	if resp.Err != nil {
		return 0
	}

	info, err := resp.Bytes()
	if err != nil {
		return 0
	}

	return redisinfo.ParseUsedMemory(info)
}

func (r *RedisCache) Close() {
	r.pool.Empty()
}

// Creates a new Redis cache instance, connecting to the given server
// and AUTHing with the provided password. If the password is an
// empty string, the AUTH command will not be run.
func New(config RedisConfig) (*RedisCache, error) {
	pool, err := makePool(config)
	return &RedisCache{pool}, err
}

func makePool(config RedisConfig) (*pool.Pool, error) {

	df := func(net, addr string) (*redis.Client, error) {
		client, err := redis.Dial(net, addr)
		if err != nil {
			return nil, err
		}

		if config.Auth != "" {
			if err = client.Cmd("AUTH", config.Auth).Err; err != nil {
				client.Close()
				return nil, err
			}
		}

		if err = client.Cmd("SELECT", config.DB).Err; err != nil {
			client.Close()
			return nil, err
		}

		return client, nil
	}

	return pool.NewCustom(config.Network, config.Address, config.Size, df)
}
