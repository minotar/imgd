// Usernames and UUIDs should be normalized before calling mcclient (eg. lowercase / no-dashes)
package mcclient

import (
	"github.com/minotar/imgd/pkg/cache"

	"github.com/minotar/minecraft"
)

type McClient struct {
	cache.Cache
	minecraft.Minecraft
}
