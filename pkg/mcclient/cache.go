package mcclient

import (
	"bytes"
	"context"
	"fmt"
	"image/png"
	"strings"
	"time"

	"github.com/minotar/imgd/pkg/cache"
	"github.com/minotar/imgd/pkg/mcclient/mcuser"
	"github.com/minotar/imgd/pkg/mcclient/uuid"
	"github.com/minotar/imgd/pkg/util/log"
	"github.com/minotar/minecraft"
)

const (
	skinTTL = 1 * time.Hour
)

// Todo: metrics / tracing / timing

// Todo: Also, add function context to logger
// User cache name and UUID in logger.With()?

// CacheRetrieveUUIDEntry searches the cache based on Username, expecting a UUID in response
func (mc *McClient) CacheRetrieveUUIDEntry(logger log.Logger, ctx context.Context, username string) (uuidEntry uuid.UUIDEntry, err error) {
	// logger should already be With() the username
	username = strings.ToLower(username)
	// Metrics timer / tracing
	// Though - is this useless when using a TieredCache which is inherentantly varied?
	cacheCtx := mc.Caches.UUID.WithContext(ctx)
	uuidBytes, err := cacheCtx.Retrieve(username)
	// Observe Cache retrieve
	if err != nil {
		// Return an error (and log based on severity)
		if err == cache.ErrNotFound {
			// Metrics stat "Miss"
			logger.Debugf("Did not find username in %s", cacheCtx.Name())
			return
		} else {
			// Metrics stat Cache RetrieveError
			logger.Errorf("Failed to lookup up username in %s: %v", cacheCtx.Name(), err)
			return
		}
	}

	if len(uuidBytes) < 5 {
		// 4 bytes or less would be an invalid status/timestamp
		// Metrics stats Cache Empty Error
		logger.Errorf("Null UUID returned for username in %s: %v", cacheCtx.Name(), uuidBytes)
		return uuid.UUIDEntry{}, fmt.Errorf("cache returned null UUID")
	}

	uuidEntry = uuid.DecodeUUIDEntry(uuidBytes)

	// Metrics stat Hit
	logger.Debugf("Found username in %s", cacheCtx.Name())
	return
}

// CacheInsertUUIDEntry takes a valid UUIDEntry and caches it
// There is no sanity checking on the input
// The item is cached witha TTL assuming it's brand-new
func (mc *McClient) CacheInsertUUIDEntry(logger log.Logger, username string, uuidEntry uuid.UUIDEntry) (err error) {
	// logger should already be With() the username and UUID
	username = strings.ToLower(username)
	// Technically this could be empty / nil???
	uuidBytes := uuidEntry.Encode()

	// Metrics timer / tracing
	// Though - is this useless when using a TieredCache which is inherentantly varied?
	err = mc.Caches.UUID.InsertTTL(username, uuidBytes, uuidEntry.TTL())
	// Observe Cache insert
	if err != nil {
		// stats.CacheUUID("error")
		logger.Errorf("Failed Insert into cache %s: %v", mc.Caches.UUID.Name(), err)
	}
	return
}

func (mc *McClient) CacheRetrieveMcUser(logger log.Logger, ctx context.Context, uuid string) (user mcuser.McUser, err error) {
	// logger should already be With() the UUID (and maybe username)
	uuid = strings.ToLower(uuid)
	// Metrics timer / tracing
	// Though - is this useless when using a TieredCache which is inherentantly varied?
	cacheCtx := mc.Caches.UserData.WithContext(ctx)
	userBytes, err := cacheCtx.Retrieve(uuid)
	// Observe Cache retrieve
	if err != nil {
		// Return an error (and log based on severity)
		if err == cache.ErrNotFound {
			// Metrics stat "Miss"
			logger.Debugf("Did not find uuid in %s", cacheCtx.Name())
			return
		} else {
			// Metrics stat Cache RetrieveError
			logger.Errorf("Failed to lookup up uuid in %s: %v", cacheCtx.Name(), err)
			return
		}
	}

	user, err = mcuser.DecompressMcUser(userBytes)
	if err != nil {
		logger.Errorf("Failed to decode McUser from uuid in %s: %v", cacheCtx.Name(), err)
		// Metrics stats Cache Decode Error
		return
	}
	// Metrics stat Hit
	logger.Debugf("Found uuid in %s", cacheCtx.Name())
	return
}

func (mc *McClient) CacheInsertMcUser(logger log.Logger, uuid string, user mcuser.McUser) (err error) {
	// logger should already be With() the UUID (and maybe username)
	uuid = strings.ToLower(uuid)
	packedUserBytes, err := user.Compress()
	if err != nil {
		// stats.CacheUser("pack_error")
		logger.Errorf("Failed to pack UUID:user ready to cache: %v", err)
	}

	logger.Info("McUser Data is packed")

	// Metrics timer / tracing
	// Though - is this useless when using a TieredCache which is inherentantly varied?
	err = mc.Caches.UUID.InsertTTL(uuid, packedUserBytes, user.TTL())
	// Observe Cache insert
	if err != nil {
		// stats.CacheUser("insert_error")
		logger.Errorf("Failed Insert UUID:user into cache %s: %v", mc.Caches.UserData.Name(), err)
	}

	logger.Info("apparently data was cached")

	return
}

//

func (mc *McClient) CacheRetrieveTexture(logger log.Logger, ctx context.Context, textureKey string) (texture minecraft.Texture, err error) {
	// logger should already be With() the skinPath/texturePath (and UUID and Username)
	textureKey = strings.ToLower(textureKey)
	// Metrics timer / tracing
	// Though - is this useless when using a TieredCache which is inherentantly varied?
	cacheCtx := mc.Caches.Textures.WithContext(ctx)
	textureBytes, err := cacheCtx.Retrieve(textureKey)
	// Observe Cache retrieve
	if err != nil {
		// Return an error (and log based on severity)
		if err == cache.ErrNotFound {
			// Metrics stat "Miss"
			logger.Debugf("Did not find texture in %s", cacheCtx.Name())
			return
		} else {
			// Metrics stat Cache RetrieveError
			logger.Errorf("Failed to lookup up texture in %s: %v", cacheCtx.Name(), err)
			return
		}
	}

	textureReader := bytes.NewReader(textureBytes)

	err = texture.Decode(textureReader)
	if err != nil {
		logger.Errorf("Failed to decode texture from %s: %v", cacheCtx.Name(), err)
		// Metrics stats Cache Decode Error
		return
	}
	// Metrics stat Hit
	logger.Debugf("Found texture in %s", cacheCtx.Name())
	return
}

func (mc *McClient) CacheInsertTexture(logger log.Logger, textureKey string, texture minecraft.Texture) (err error) {
	// logger should already be With() the skinPath/texturePath (and UUID and Username)
	textureKey = strings.ToLower(textureKey)

	textureBuf := new(bytes.Buffer)
	// Todo: research better lossless encoding - or just try the png.Encoder with higher compression
	// bimg/vips ? Maybe cache as a different format if more efficient
	err = png.Encode(textureBuf, texture.Image)
	if err != nil {
		// stats.CacheUser("pack_error")
		logger.Errorf("Failed to PNG encode texture for cache: %v", err)
	}

	// Metrics timer / tracing
	// Though - is this useless when using a TieredCache which is inherentantly varied?
	err = mc.Caches.UUID.InsertTTL(textureKey, textureBuf.Bytes(), skinTTL)
	// Observe Cache insert
	if err != nil {
		// stats.CacheUser("insert_error")
		logger.Errorf("Failed Insert texture PNG into cache %s: %v", mc.Caches.Textures.Name(), err)
	}
	return
}
