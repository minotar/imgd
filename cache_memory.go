package main

import (
	"time"

	"github.com/minotar/minecraft"
)

const (
	// Get the skin size in bytes. Stored as a []uint8, one byte each,
	// plus bounces. So 64 * 64 bytes and we'll throw in an extra 16
	// bytes of overhead.
	skinSize = (64 * 64) + 16

	// Define a 64 MB cache size.
	cacheSize = 2 << 25

	// Based off those, calculate the maximum number of skins we'll store
	// in memory.
	skinCount = skinSize / cacheSize
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

func (c *CacheMemory) setup() error {
	c.Skins = map[string]minecraft.Skin{}
	c.Usernames = []string{}

	log.Notice("Loaded Memory cache")
	return nil
}

// Returns whether the item exists in the cache.
func (c *CacheMemory) has(username string) bool {
	if _, exists := c.Skins[username]; exists {
		return true
	}
	return false
}

// Retrieves the item from the cache. We'll promote it to the "top" of the
// cache, effectively updating its expiry time.
func (c *CacheMemory) pull(username string) minecraft.Skin {
	index := indexOf(username, c.Usernames)
	c.Usernames = append(c.Usernames, username)
	c.Usernames = append(c.Usernames[:index], c.Usernames[index+1:]...)

	return c.Skins[username]
}

// Removes the username from the cache
func (c *CacheMemory) remove(username string) {
	index := indexOf(username, c.Usernames)
	if index == -1 {
		return
	}

	key := c.Usernames[index]
	delete(c.Skins, key)
}

// Adds the skin to the cache, remove the oldest, expired skin if the cache
// list is full.
func (c *CacheMemory) add(username string, skin minecraft.Skin) {
	if len(c.Usernames) >= skinCount {
		first := c.Usernames[0]
		delete(c.Skins, first)
		c.Usernames = append(c.Usernames[1:], username)
	} else {
		c.Usernames = append(c.Usernames, username)
	}

	c.Skins[username] = skin

	// After the expiration time, remove the item from the cache.
	time.AfterFunc(time.Duration(config.Server.Ttl)*time.Second, func() {
		c.remove(username)
	})
}

// The exact number of usernames in the map
func (c *CacheMemory) size() uint {
	return uint(len(c.Skins))
}

// The byte size of the cache. Fairly rough... don't really want to venture
// into the land of manual memory management, because there be dragons.
func (c *CacheMemory) memory() uint64 {
	return uint64(len(c.Skins) * skinSize)
}
