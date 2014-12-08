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

func (c *CacheRedis) setup() {
	pool, err := pool.NewPool("tcp", config.Redis.Address, config.Redis.PoolSize)
	if err != nil {
		log.Error("Error connecting to redis database")
		return
	}

	c.Pool = pool
	client, err := c.Pool.Get()
	if err != nil {
		log.Error(err.Error())
	}
	defer c.Pool.Put(client)

	log.Info("Loaded Redis cache (pool: " + fmt.Sprintf("%v", config.Redis.PoolSize) + ")")
}

func (c *CacheRedis) has(username string) bool {
	client, err := c.Pool.Get()
	if err != nil {
		log.Error(err.Error())
	}
	defer c.Pool.Put(client)

	_ = client.Cmd("AUTH", config.Redis.Auth)
	res := client.Cmd("EXISTS", config.Redis.Prefix+username)

	exists, err := res.Bool()
	if err != nil {
		log.Error(err.Error())
		return false
	}

	return exists
}

func (c *CacheRedis) pull(username string) minecraft.Skin {
	client, err := c.Pool.Get()
	if err != nil {
		log.Error(err.Error())
	}
	defer c.Pool.Put(client)

	_ = client.Cmd("AUTH", config.Redis.Auth)
	resp := client.Cmd("GET", config.Redis.Prefix+username)

	skin, err := getSkinFromReply(resp)
	if err != nil {
		log.Error(err.Error())

		c.remove(username)
		char, _ := minecraft.FetchSkinForChar()

		return char
	}

	return skin
}

func (c *CacheRedis) add(username string, skin minecraft.Skin) {
	client, err := c.Pool.Get()
	if err != nil {
		log.Error(err.Error())
	}
	defer c.Pool.Put(client)

	skinBuf := new(bytes.Buffer)
	_ = png.Encode(skinBuf, skin.Image)

	_ = client.Cmd("AUTH", config.Redis.Auth)
	_ = client.Cmd("SETEX", "skins:"+username, config.Redis.Ttl, skinBuf.Bytes())
}

func (c *CacheRedis) remove(username string) {
	client, err := c.Pool.Get()
	if err != nil {
		log.Error(err.Error())
	}
	defer c.Pool.Put(client)

	_ = client.Cmd("AUTH", config.Redis.Auth)
	_ = client.Cmd("DEL", config.Redis.Prefix+username)
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
