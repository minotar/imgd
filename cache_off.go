package main

import (
	"github.com/minotar/minecraft"
)

type CacheOff struct {
}

func (c *CacheOff) setup() {
	log.Info("Loaded without cache")
}

func (c *CacheOff) has(username string) bool {
	return false
}

// Should never be called.
func (c *CacheOff) pull(username string) minecraft.Skin {
	char, _ := minecraft.FetchSkinForChar()
	return char
}

func (c *CacheOff) add(username string, skin minecraft.Skin) {
}

func (c *CacheOff) memory() uint64 {
	return 0
}
