package main

import (
	"bytes"
	"fmt"
	"github.com/fzzy/radix/redis"
	"github.com/minotar/minecraft"
	"image/png"
)

type CacheRedis struct {
	Client *redis.Client
	Pool   *RedisPool
}

func (c *CacheRedis) setup() {
	c.Pool = CreateRedisPool(config.Redis.Address, config.Redis.PoolSize, config.Redis.Auth)

	log.Info("Loaded Redis cache (pool: " + fmt.Sprintf("%v", config.Redis.PoolSize) + ")")
}

func (c *CacheRedis) has(username string) bool {
	client := c.Pool.Get()

	res := client.Cmd("EXISTS", config.Redis.Prefix+username)

	exists, err := res.Bool()
	if err != nil {
		log.Error(err.Error())
		return false
	}

	return exists
}

func (c *CacheRedis) pull(username string) minecraft.Skin {
	client := c.Pool.Get()

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
	client := c.Pool.Get()

	skinBuf := new(bytes.Buffer)
	_ = png.Encode(skinBuf, skin.Image)

	_ = client.Cmd("SETEX", "skins:"+username, config.Redis.Ttl, skinBuf.Bytes())
}

func (c *CacheRedis) remove(username string) {
	client := c.Pool.Get()

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
