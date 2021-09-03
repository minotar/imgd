// Usernames and UUIDs should be normalized before calling mcclient (eg. lowercase / no-dashes)
package mcclient

import (
	"context"

	"github.com/minotar/imgd/pkg/cache"
	"github.com/minotar/imgd/pkg/util/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/minotar/imgd/pkg/mcclient/mcuser"
	mc_uuid "github.com/minotar/imgd/pkg/mcclient/uuid"
	"github.com/minotar/minecraft"
)

// Todo: metrics / tracing / timing

// Todo: Could have a base logger which we then apply context to when needed
type McClient struct {
	Caches struct {
		UUID     cache.Cache
		UserData cache.Cache
		Textures cache.Cache
	}
	API *minecraft.Minecraft
	// Todo: I think unused..?
	TexturesMcNetBase string
}

// Todo: I need to be providing logging and request context in here
func (mc *McClient) GetSkinFromReq(logger log.Logger, ctx context.Context, userReq UserReq) minecraft.Skin {

	tracer := otel.Tracer("github.com/minotar/imgd/pkg/mcclient")
	ctx, span := tracer.Start(ctx, "GetSkinFromReq")
	defer span.End()

	logger, uuid, err := userReq.GetUUID(logger, ctx, mc)
	if err != nil {
		logger.Debugf("Falling back to Steve: %v", err)
		skin, _ := minecraft.FetchSkinForSteve()
		return skin
	}

	mcUser, err := mc.GetMcUser(logger, ctx, uuid)
	if err != nil {
		logger.Debugf("Falling back to Steve: %v", err)
		skin, _ := minecraft.FetchSkinForSteve()
		return skin
	}

	// Re-add username (fixes any capitalisation issues as well)
	logger = logger.With(
		"username", mcUser.Username,
		"skinPath", mcUser.Textures.SkinPath,
	)

	// We use the SkinPath (which is either just the hash, or a full URL id the base URL changes)
	textureKey := mcUser.Textures.SkinPath
	textureURL := mcUser.Textures.SkinURL()

	texture, err := mc.GetTexture(logger, ctx, textureKey, textureURL)
	if err != nil {
		logger.Debugf("Falling back to Steve: %v", err)
		skin, _ := minecraft.FetchSkinForSteve()
		return skin
	}

	// Return our Texture in the Skin struct
	return minecraft.Skin{Texture: texture}
}

// Todo: This should handle all cache / API
// Todo: Do we want a WaitGroup here?
// Only real downside is that we can't goroutine to insert into cache?
// Unless we have 2 locks? 1 here, and then one that blocks reads when writing?
func (mc *McClient) GetUUIDEntry(logger log.Logger, ctx context.Context, username string) (uuidEntry mc_uuid.UUIDEntry, err error) {
	uuidEntry, err = mc.CacheRetrieveUUIDEntry(logger, ctx, username)
	if err != nil {
		if err == cache.ErrNotFound {
			// We cache missed (cache.ErrNotFound) so let's request from API
			uuidEntry = mc.RequestUUIDEntry(logger, ctx, username, uuidEntry)
			// We need to generate a new error though
			return uuidEntry, uuidEntry.Status.GetError()
		} else {
			// Cache experieneed a proper error (already would be logged)
			return
		}
	}

	// Cache was a hit (though still might be a bad result)

	if uuidEntry.IsValid() {
		if uuidEntry.IsFresh() {
			// Great success - we have a cached result
			return
		}
		logger.Debugf("Stale UUIDEntry was dated: %v", uuidEntry.Timestamp.Time())
		// A stale result should be re-requested
		return mc.RequestUUIDEntry(logger, ctx, username, uuidEntry), nil
	}

	// A bad result was returned from the cache, generate an error from it
	return uuidEntry, uuidEntry.Status.GetError()
}

func (mc *McClient) GetMcUser(logger log.Logger, ctx context.Context, uuid string) (mcUser mcuser.McUser, err error) {

	tracer := otel.Tracer("github.com/minotar/imgd/pkg/mcclient")
	var span trace.Span
	ctx, span = tracer.Start(ctx, "GetMcUser")
	defer span.End()

	mcUser, err = mc.CacheRetrieveMcUser(logger, ctx, uuid)
	if err != nil {
		if err == cache.ErrNotFound {
			// We cache missed (cache.ErrNotFound) so let's request from API
			mcUser = mc.RequestMcUser(logger, ctx, uuid, mcUser)
			// We need to generate a new error though
			return mcUser, mcUser.Status.GetError()
		} else {
			// Cache experieneed a proper error (already would be logged)
			return
		}
	}

	// Cache was a hit (though still might be a bad result)

	if mcUser.IsValid() {
		if mcUser.IsFresh() {
			// Known positive result
			return
		}
		span.AddEvent("Stale McUser")
		logger.Debugf("Stale McUser was dated: %v", mcUser.Timestamp.Time())
		// A stale result should be re-requested
		return mc.RequestMcUser(logger, ctx, uuid, mcUser), nil
	}

	// A bad result was returned from the cache, generate an error from it
	return mcUser, mcUser.Status.GetError()
}

// Todo: Counters also support exemplars!

func (mc *McClient) GetTexture(logger log.Logger, ctx context.Context, textureKey string, textureURL string) (texture minecraft.Texture, err error) {

	tracer := otel.Tracer("github.com/minotar/imgd/pkg/mcclient")
	var span trace.Span
	ctx, span = tracer.Start(ctx, "GetTexture")
	defer span.End()

	texture, err = mc.CacheRetrieveTexture(logger, ctx, textureKey)
	if err != nil {
		if err == cache.ErrNotFound {
			// We cache missed (cache.ErrNotFound) so let's request from API
			return mc.RequestTexture(logger, ctx, textureKey, textureURL)
		} else {
			// Cache experieneed a proper error (already would be logged)
			return
		}
	}

	// Cache was a hit (we don't have logic to cache bad textures)
	return
}
