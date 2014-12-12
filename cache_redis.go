package main

import (
	"bytes"
	"fmt"
	"github.com/fzzy/radix/extra/pool"
	"github.com/fzzy/radix/redis"
	"github.com/minotar/minecraft"
	"image/png"
)

type CacheRedis struct {
	Client *redis.Client
	Pool   *pool.Pool
}

func dialFunc(network, addr string) (*redis.Client, error) {
	client, err := redis.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	if config.Redis.Auth != "" {
		if err := client.Cmd("AUTH", config.Redis.Auth).Err; err != nil {
			client.Close()
			return nil, err
		}
	}
	return client, nil
}

func (c *CacheRedis) setup() {
	pool, err := pool.NewCustomPool(
		"tcp",
		config.Redis.Address,
		config.Redis.PoolSize,
		dialFunc,
	)
	if err != nil {
		log.Error("Error connecting to redis database")
		return
	}

	c.Pool = pool

	log.Info("Loaded Redis cache (pool: " + fmt.Sprintf("%v", config.Redis.PoolSize) + ")")
}

func (c *CacheRedis) getFromPool() *redis.Client {
	client, err := c.Pool.Get()
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	return client
}

func (c *CacheRedis) has(username string) bool {
	var err error
	client := c.getFromPool()
	if client == nil {
		return false
	}
	defer c.Pool.CarefullyPut(client, &err)

	var exists bool
	exists, err = client.Cmd("EXISTS", config.Redis.Prefix+username).Bool()
	if err != nil {
		log.Error(err.Error())
		return false
	}

	return exists
}

// What to do when failing to pull a skin from redis
func (c *CacheRedis) pullFailed(username string) minecraft.Skin {
	c.remove(username)
	char, _ := minecraft.FetchSkinForChar()
	return char
}

func (c *CacheRedis) pull(username string) minecraft.Skin {
	var err error
	client := c.getFromPool()
	if client == nil {
		return c.pullFailed(username)
	}
	defer c.Pool.CarefullyPut(client, &err)

	resp := client.Cmd("GET", config.Redis.Prefix+username)
	skin, err := getSkinFromReply(resp)
	if err != nil {
		log.Error(err.Error())
		return c.pullFailed(username)
	}

	return skin
}

func (c *CacheRedis) add(username string, skin minecraft.Skin) {
	var err error
	client := c.getFromPool()
	if client == nil {
		return
	}
	defer c.Pool.CarefullyPut(client, &err)

	skinBuf := new(bytes.Buffer)
	_ = png.Encode(skinBuf, skin.Image)

	// read into err so that it's set for the defer
	err = client.Cmd("SETEX", "skins:"+username, config.Redis.Ttl, skinBuf.Bytes()).Err
}

func (c *CacheRedis) remove(username string) {
	var err error
	client := c.getFromPool()
	if client == nil {
		return
	}
	defer c.Pool.CarefullyPut(client, &err)

	// read into err so that it's set for the defer
	err = client.Cmd("DEL", config.Redis.Prefix+username).Err
}

func getSkinFromReply(resp *redis.Reply) (minecraft.Skin, error) {
	respBytes, respErr := resp.Bytes()
	if respErr != nil {
		return minecraft.Skin{}, respErr
	}

	imgBuf := bytes.NewBuffer(respBytes)

	skin, skinErr := minecraft.DecodeSkin(imgBuf)
	if skinErr != nil {
		return minecraft.Skin{}, skinErr
	}

	return skin, nil
}
