package main

import (
	"github.com/minotar/minecraft"
)

const (
	// Get the skin size in bytes. Skins have an 8-bit channel depth and are
	// 64x32 pixels. Maximum of 8*64*32 plus 13 bytes for png metadata. They'll
	// rarely (never) be that large due to compression, but we'll leave some
	// extra wiggle to account for map overhead.
	SKIN_SIZE = 8 * (64 * 32)

	// Define a 64 MB cache size.
	CACHE_SIZE = 2 << 25

	// Based off those, calculate the maximum number of skins we'll store
	// in memory.
	SKIN_NUMBER = CACHE_SIZE / SKIN_SIZE
)

// Cache object that stores skins in memory.
type CacheMemory struct {
	// Map of usernames to minecraft skins. Lookups here are O(1), so that
	// makes my happy.
	Skins map[string]minecraft.Skin
	// Additionally keep a *slice* of usernames which we can update
	Usernames []string
}


// Find the position of a string in a slice. Returns -1 on failure.
func indexOf(str string, list []string) int {
	for index, value := range list {
		if value == str {
			return index
		}
	}

	return -1
}

func (c *CacheMemory) setup() {
	c.Skins = map[string]minecraft.Skin{}
	c.Usernames = []string{}
}

// Returns whether the item exists in the cache.
func (c *CacheMemory) has(username string) bool {
	if _, exists := c.Skins[username]; exists {
		return true
	} else {
		return false
	}
}

// Retrieves the item from the cache. We'll promote it to the "top" of the
// cache, effectively updating its expiry time.
func (c *CacheMemory) pull(username string) minecraft.Skin {
	index := indexOf(username, c.Usernames)
	c.Usernames = append(c.Usernames, username)
	c.Usernames = append(c.Usernames[:index], c.Usernames[index+1:]...)

	return c.Skins[username]
}

// Adds the skin to the cache, remove the oldest, expired skin if the cache
// list is full.
func (c *CacheMemory) add(username string, skin minecraft.Skin) {
	if len(c.Usernames) >= SKIN_NUMBER {
		first := c.Usernames[0]
		delete(c.Skins, first)
		c.Usernames = append(c.Usernames[1:], username)
	} else {
		c.Usernames = append(c.Usernames, username)
	}

	c.Skins[username] = skin
}
