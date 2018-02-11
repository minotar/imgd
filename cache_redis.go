package main

import (
	"bytes"
	"errors"
	"image/png"
	"strconv"
	"strings"

	"github.com/fzzy/radix/extra/pool"
	"github.com/fzzy/radix/redis"
	"github.com/minotar/minecraft"
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
		r := client.Cmd("AUTH", config.Redis.Auth)
		if r.Err != nil {
			client.Close()
			return nil, err
		}
	}

	// Select the DB within Redis
	r := client.Cmd("SELECT", config.Redis.DB)
	if r.Err != nil {
		client.Close()
		return nil, err
	}

	return client, nil
}

func (c *CacheRedis) setup() error {
	pool, err := pool.NewCustomPool(
		"tcp",
		config.Redis.Address,
		config.Redis.PoolSize,
		dialFunc,
	)
	if err != nil {
		log.Error("Error connecting to redis database")
		return err
	}

	c.Pool = pool

	log.Noticef("Loaded Redis cache (address: %s, db: %v, prefix: \"%s\", pool: %v)", config.Redis.Address, config.Redis.DB, config.Redis.Prefix, config.Redis.PoolSize)
	return nil
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
	err = client.Cmd("SETEX", "skins:"+username, strconv.Itoa(config.Server.Ttl), skinBuf.Bytes()).Err
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

func (c *CacheRedis) size() uint {
	var err error
	client := c.getFromPool()
	if client == nil {
		return 0
	}
	defer c.Pool.CarefullyPut(client, &err)

	resp := client.Cmd("DBSIZE")
	size, err := resp.Int()
	if err != nil {
		log.Error(err.Error())
		return 0
	}

	return uint(size)
}

func (c *CacheRedis) memory() uint64 {
	var err error
	client := c.getFromPool()
	if client == nil {
		return 0
	}
	defer c.Pool.CarefullyPut(client, &err)

	data, err := parseStats(client.Cmd("INFO"))
	if err != nil {
		return 0
	}

	mem, _ := strconv.ParseUint(data["used_memory"], 10, 64)
	return mem
}

func getSkinFromReply(resp *redis.Reply) (minecraft.Skin, error) {
	respBytes, respErr := resp.Bytes()
	if respErr != nil {
		return minecraft.Skin{}, respErr
	}

	imgBuf := bytes.NewReader(respBytes)

	skin, skinErr := minecraft.DecodeSkin(imgBuf)
	if skinErr != nil {
		return minecraft.Skin{}, skinErr
	}

	return skin, nil
}

// Parses a reply from redis INFO into a nice map.
func parseStats(resp *redis.Reply) (map[string]string, error) {
	r, err := resp.Bytes()
	if err != nil {
		return nil, err
	}

	raw := strings.Split(string(r), "\r\n")
	output := map[string]string{}

	for _, line := range raw {
		// Skip blank lines or comment lines
		if len(line) == 0 || string(line[0]) == "#" {
			continue
		}
		// Get the position the seperator breaks
		sep := strings.Index(line, ":")
		if sep == -1 {
			return nil, errors.New("Invalid line: " + line)
		}
		output[line[:sep]] = line[sep+1:]
	}

	return output, nil
}
