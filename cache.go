package main

import (
	"github.com/minotar/minecraft"
)

type Cache interface {
	setup() 
	has(username string) bool
	pull(username string) minecraft.Skin
	add(username string, skin minecraft.Skin)
}

func MakeCache(cacheType string) Cache {
	if cacheType == "memory" {
		cache = &CacheMemory{}
	} else if cacheType == "redis" {
		cache = &CacheRedis{}
	} 

	return cache
}
