package main

import (
	"bytes"
	"github.com/fzzy/radix/redis"
	"github.com/minotar/minecraft"
	"image/png"
	"time"
)

type CacheRedis struct {
	Client *redis.Client
}

func (c *CacheRedis) setup() {
	conn, err := redis.DialTimeout("tcp", config.Redis.Address, time.Duration(10)*time.Second)
	if err != nil {
		log.Error("Error connecting to redis database")
		return
	}

	c.Client = conn
	_ = c.Client.Cmd("AUTH", config.Redis.Auth)

	log.Info("Loaded Redis cache")
}

func (c *CacheRedis) has(username string) bool {
	res := c.Client.Cmd("EXISTS", config.Redis.Prefix+username)
	exists, err := res.Bool()
	if err != nil {
		log.Error(err.Error())
		return false
	}

	return exists
}

func (c *CacheRedis) pull(username string) minecraft.Skin {
	resp := c.Client.Cmd("GET", config.Redis.Prefix+username)

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
	skinBuf := new(bytes.Buffer)
	_ = png.Encode(skinBuf, skin.Image)

	_ = c.Client.Cmd("SETEX", "skins:"+username, config.Redis.Ttl, skinBuf.Bytes())
}

func (c *CacheRedis) remove(username string) {
	_ = c.Client.Cmd("DEL", config.Redis.Prefix+username)
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
