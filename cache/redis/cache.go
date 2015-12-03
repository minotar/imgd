package redis

import (
	"github.com/garyburd/redigo/redis"
	"time"
)

type RedisCache struct {
	pool *redis.Pool
}

func (r *RedisCache) Insert(key string, value []byte, ttl time.Duration) error {
	cnx := r.pool.Get()
	defer cnx.Close()

	_, err := cnx.Do("SET", key, value, "EX", ttl.Seconds())
	return err
}

func (r *RedisCache) Find(key string) ([]byte, error) {
	cnx := r.pool.Get()
	defer cnx.Close()

	return redis.Bytes(cnx.Do("GET", key))
}

func (r *RedisCache) Flush() error {
	cnx := r.pool.Get()
	defer cnx.Close()

	_, err := cnx.Do("FLUSHALL")
	return err
}

func (r *RedisCache) Close() {
	r.pool.Close()
}

// Creates a new Redis cache instance, connecting to the given server
// and AUTHing with the provided password. If the password is an
// empty string, the AUTH command will not be run.
func New(server, password string) *RedisCache {
	return &RedisCache{makePool(server, password)}
}

// Almost straight from the Redigo docs
func makePool(server, password string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil || password == "" {
				return c, err
			}

			if _, err := c.Do("AUTH", password); err != nil {
				c.Close()
				return nil, err
			}

			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			// Only bother testing stale connections.
			if time.Now().Sub(t) < 10*time.Second {
				_, err := c.Do("PING")
				return err
			}

			return nil
		},
	}

}
