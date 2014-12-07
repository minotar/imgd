package main

import (
	"github.com/fzzy/radix/redis"
	"github.com/minotar/minecraft"
)

// Cache object that stores skins in memory.
type CacheRedis struct {
	Client *redis.Client
}

func (c *CacheRedis) setup() {
	conn, err := redis.Dial("tcp",  "127.0.0.1:6379")
	if err != nil {
		log.Error("Error connecting to redis database")
	}

	c.Client = conn
}

// Returns whether the item exists in the cache.
func (c *CacheRedis) has(username string) bool {
	res := c.Client.Cmd("GET", "skins:" + username)
	if res != nil {
		return true
	} 
	
	return false
}

// Retrieves the item from the cache. We'll promote it to the "top" of the
// cache, effectively updating its expiry time.
func (c *CacheRedis) pull(username string) minecraft.Skin {
	//_:= c.Client.Cmd("GET", "skins:" + username)
	return minecraft.Skin{}
}

// Adds the skin to the cache, remove the oldest, expired skin if the cache
// list is full.
func (c *CacheRedis) add(username string, skin minecraft.Skin) {
	_ = c.Client.Cmd("SET", "skins:" + username, skin.Image)
}
