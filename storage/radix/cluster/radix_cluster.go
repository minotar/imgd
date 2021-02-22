package cluster

import (
	"time"

	"github.com/mediocregopher/radix.v2/cluster"
	"github.com/mediocregopher/radix.v2/redis"
	"github.com/minotar/imgd/storage"
	"github.com/minotar/imgd/storage/radix"
	"github.com/minotar/imgd/storage/util/redisinfo"
)

// RedisClusterCache stores our Redis Cluster
type RedisClusterCache struct {
	cluster *cluster.Cluster
}

// ensure that the storage.Storage interface is implemented
var _ storage.Storage = new(RedisClusterCache)

// Insert will SET the key in Redis with an expiry TTL
func (r *RedisClusterCache) Insert(key string, value []byte, ttl time.Duration) error {
	resp := r.cluster.Cmd("SET", key, value, "EX", ttl.Seconds())
	return resp.Err
}

func (r *RedisClusterCache) Retrieve(key string) ([]byte, error) {
	resp := r.cluster.Cmd("GET", key)
	if resp.Err != nil {
		return nil, resp.Err
	}

	bytes, err := resp.Bytes()
	if err != nil && err.Error() == "response is nil" {
		return bytes, storage.ErrNotFound
	}
	return bytes, err
}

func (r *RedisClusterCache) Flush() error {
	resp := r.cluster.Cmd("FLUSHDB")
	return resp.Err
}

func (r *RedisClusterCache) Len() uint {
	resp := r.cluster.Cmd("DBSIZE")
	if resp.Err != nil {
		return 0
	}

	size, err := resp.Int()
	if err != nil {
		return 0
	}

	return uint(size)
}

func (r *RedisClusterCache) Size() uint64 {
	resp := r.cluster.Cmd("INFO")
	if resp.Err != nil {
		return 0
	}

	info, err := resp.Bytes()
	if err != nil {
		return 0
	}

	return redisinfo.ParseUsedMemory(info)
}

func (r *RedisClusterCache) Close() {
	r.cluster.Close()
}

// Creates a new Redis cache instance, connecting to the given server
// and AUTHing with the provided password. If the password is an
// empty string, the AUTH command will not be run.
func New(config radix.RedisConfig) (*RedisClusterCache, error) {
	cluster, err := makeCluster(config)
	return &RedisClusterCache{cluster}, err
}

func makeCluster(config radix.RedisConfig) (*cluster.Cluster, error) {

	df := func(net, addr string) (*redis.Client, error) {
		// net must always be "tcp" for cluster
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

	return cluster.NewWithOpts(cluster.Opts{
		Addr:     config.Address,
		PoolSize: config.Size,
		Dialer:   df,
	})
}
