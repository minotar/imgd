package main

import (
	"bytes"
	"encoding/base64"
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
	res := c.Client.Cmd("EXISTS", "skins:"+username)
	exists, err := res.Bool()
	if err != nil {
		return false
	}

	return exists
}

func (c *CacheRedis) pull(username string) minecraft.Skin {
	resp := c.Client.Cmd("GET", "skins:"+username)

	skin, err := getSkinFromReply(resp)
	if err != nil {
		c.remove(username)

		char, _ := minecraft.FetchSkinForChar()

		return char
	}

	return skin
}

func (c *CacheRedis) add(username string, skin minecraft.Skin) {
	skinBuf := new(bytes.Buffer)
	_ = png.Encode(skinBuf, skin.Image)
	skinStr := base64.StdEncoding.EncodeToString(skinBuf.Bytes())

	resp := c.Client.Cmd("SET", "skins:"+username, skinStr)
	_ = c.Client.Cmd("EXPIRE", "skins:"+username, config.Redis.Ttl)
}

func (c *CacheRedis) remove(username string) {
	_ = c.Client.Cmd("DEL", "skins:"+username)
}

func getSkinFromReply(resp *redis.Reply) (minecraft.Skin, error) {
	respString, respErr := resp.Str()
	if respErr != nil {
		return minecraft.Skin{}, respErr
	}

	imgBytes, decErr := base64.StdEncoding.DecodeString(respString)
	if decErr != nil {
		return minecraft.Skin{}, decErr
	}

	imgBuf := bytes.NewBuffer(imgBytes)

	skin, skinErr := minecraft.DecodeSkin(imgBuf)
	if skinErr != nil {
		return minecraft.Skin{}, skinErr
	}

	return skin, nil
}
