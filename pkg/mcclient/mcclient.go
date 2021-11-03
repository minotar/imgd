// Usernames and UUIDs should be normalized before calling mcclient (eg. lowercase / no-dashes)
package mcclient

import (
	"github.com/minotar/imgd/pkg/cache"
	"github.com/minotar/imgd/pkg/util/log"

	"github.com/minotar/imgd/pkg/mcclient/mcuser"
	mc_uuid "github.com/minotar/imgd/pkg/mcclient/uuid"
	"github.com/minotar/imgd/pkg/minecraft"
)

// Todo: tracing
// Todo: Counters also support exemplars! eg. cache error metric + Request ID

// Todo: Could have a base logger which we then apply context to when needed
type McClient struct {
	Caches struct {
		UUID     cache.Cache
		UserData cache.Cache
		Textures cache.Cache
	}
	API *minecraft.Minecraft
}

// Todo: I need to be providing logging and request context in here
// This Method will decode the buffer into a Texture (fine for processing, but avoid if you are serving the plain skin)
func (mc *McClient) GetSkinFromReq(logger log.Logger, userReq UserReq) minecraft.Skin {
	logger, textureIO := mc.GetSkinBufferFromReq(logger, userReq)

	// Return decoded skin (or Steve)
	return textureIO.MustDecodeSkin(logger)
}

// Remember to close the mcuser.TextureIO.ReadCloser!
func (mc *McClient) GetSkinBufferFromReq(logger log.Logger, userReq UserReq) (log.Logger, mcuser.TextureIO) {
	logger, mcUser, err := mc.GetMcUserFromReq(logger, userReq)
	if err != nil {
		logger.Debugf("Falling back to Steve: %v", err)
		return logger, mcuser.GetSteveTextureIO()
	}

	// We use the SkinPath (which is either just the hash, or a full URL if the base URL changes)
	textureKey := mcUser.Textures.SkinPath
	textureURL := mcUser.Textures.SkinURL()

	textureIO, err := mc.GetTexture(logger, textureKey, textureURL)

	if err != nil {
		logger.Debugf("Falling back to Steve: %v", err)
		return logger, mcuser.GetSteveTextureIO()
	}

	return logger, textureIO
}

func (mc *McClient) GetMcUserFromReq(logger log.Logger, userReq UserReq) (log.Logger, mcuser.McUser, error) {
	logger, uuid, err := userReq.GetUUID(logger, mc)
	if err != nil {
		return logger, mcuser.McUser{}, err
	}

	logger = logger.With("uuid", uuid)
	mcUser, err := mc.GetMcUser(logger, uuid)
	if err != nil {
		return logger, mcuser.McUser{}, err
	}

	// Re-add username (fixes any capitalisation issues as well)
	logger = logger.With(
		"username", mcUser.Username,
		"skinPath", mcUser.Textures.SkinPath,
	)
	return logger, mcUser, nil
}

// Todo: Do we want a WaitGroup here?
// Only real downside is that we can't goroutine to insert into cache?
// Unless we have 2 locks? 1 here, and then one that blocks reads when writing?
func (mc *McClient) GetUUIDEntry(logger log.Logger, username string) (uuidEntry mc_uuid.UUIDEntry, err error) {
	uuidEntry, err = mc.CacheRetrieveUUIDEntry(logger, username)
	if err != nil {
		if err == cache.ErrNotFound {
			// We cache missed (cache.ErrNotFound)
			uuidCacheStatus.Miss()
			// Let's request from API
			uuidEntry = mc.RequestUUIDEntry(logger, username, uuidEntry)
			// We need to generate a new error though
			return uuidEntry, uuidEntry.Status.GetError()
		} else {
			// Cache experieneed a proper error (already would be logged)
			uuidCacheStatus.Error()
			return
		}
	}

	// Cache was a hit (though still might be a bad result)
	uuidCacheStatus.Hit()

	if uuidEntry.IsValid() {
		if uuidEntry.IsFresh() {
			// Great success - we have a cached result
			uuidCacheStatus.Fresh()
			return
		}
		// A stale result should be re-requested
		uuidCacheStatus.Stale()
		logger.Debugf("Stale UUIDEntry was dated: %v", uuidEntry.Timestamp.Time())
		return mc.RequestUUIDEntry(logger, username, uuidEntry), nil
	}

	// A bad result was returned from the cache, generate an error from it
	return uuidEntry, uuidEntry.Status.GetError()
}

func (mc *McClient) GetMcUser(logger log.Logger, uuid string) (mcUser mcuser.McUser, err error) {
	mcUser, err = mc.CacheRetrieveMcUser(logger, uuid)
	if err != nil {
		if err == cache.ErrNotFound {
			// We cache missed (cache.ErrNotFound)
			userdataCacheStatus.Miss()
			// Let's request from API
			mcUser = mc.RequestMcUser(logger, uuid, mcUser)
			// We need to generate a new error though
			return mcUser, mcUser.Status.GetError()
		} else {
			// Cache experieneed a proper error (already would be logged)
			userdataCacheStatus.Error()
			return
		}
	}

	// Cache was a hit (though still might be a bad result)
	userdataCacheStatus.Hit()

	if mcUser.IsValid() {
		if mcUser.IsFresh() {
			// Known positive result
			userdataCacheStatus.Fresh()
			return
		}
		// A stale result should be re-requested
		userdataCacheStatus.Stale()
		logger.Debugf("Stale McUser was dated: %v", mcUser.Timestamp.Time())
		return mc.RequestMcUser(logger, uuid, mcUser), nil
	}

	// A bad result was returned from the cache, generate an error from it
	return mcUser, mcUser.Status.GetError()
}

func (mc *McClient) GetTexture(logger log.Logger, textureKey string, textureURL string) (textureIO mcuser.TextureIO, err error) {
	textureIO, err = mc.CacheRetrieveTexture(logger, textureKey)
	if err == ErrCacheDisabled {
		return mc.RequestTexture(logger, textureKey, textureURL)
	} else if err == cache.ErrNotFound {
		// We cache missed (cache.ErrNotFound)
		textureCacheStatus.Miss()
		// Let's request from API
		return mc.RequestTexture(logger, textureKey, textureURL)
	} else if err != nil {
		// Cache experieneed a proper error (already would be logged)
		textureCacheStatus.Error()
		// Let's re-request anyway - there is no ratelimit
		return mc.RequestTexture(logger, textureKey, textureURL)
	}

	// Cache was a hit (we don't have logic to cache bad textures)
	textureCacheStatus.Hit()
	return
}
