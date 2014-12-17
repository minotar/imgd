package main

import (
	"github.com/minotar/minecraft"
)

type Cache interface {
	setup() error
	has(username string) bool
	pull(username string) minecraft.Skin
	add(username string, skin minecraft.Skin)
	memory() uint64
}

func MakeCache(cacheType string) Cache {
	if cacheType == "redis" {
		return &CacheRedis{}
	} else if cacheType == "memory" {
		return &CacheMemory{}
	} else {
		return &CacheOff{}
	}
}
