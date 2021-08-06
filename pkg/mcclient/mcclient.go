// Usernames and UUIDs should be normalized before calling mcclient (eg. lowercase / no-dashes)
package mcclient

import (
	"errors"
	"time"

	"github.com/minotar/imgd/pkg/cache"
	"github.com/minotar/imgd/pkg/util/log"

	"github.com/minotar/imgd/pkg/mcclient/mcuser"
	mc_uuid "github.com/minotar/imgd/pkg/mcclient/uuid"
	"github.com/minotar/minecraft"
)

const (
	day = 24 * time.Hour

	usernameTTL          = 60 * day
	usernameUnknownTTL   = 14 * day
	usernameRateLimitTTL = 2 * time.Hour
	usernameErrorTTL     = 1 * time.Hour

	uuidTTL          = 60 * day
	uuidFreshTTL     = 2 * time.Hour
	uuidUnknownTTL   = 7 * day
	uuidRateLimitTTL = 1 * time.Hour
	uuidErrorTTL     = 30 * time.Minute

	skinTTL      = 1 * time.Hour
	skinErrorTTL = 15 * time.Minute

	metaUnknownCode   = "204"
	metaRateLimitCode = "429"
	metaErrorCode     = "0"
)

// This should match the McUserProto_UserStatus from mcuser_proto
// Todo: we coukd instead import that here
// Todo: could also use a custom type here? With methods for error generation etc. ?
const (
	StatusUnSet uint8 = iota
	StatusOk
	StatusErrorGeneric
	StatusErrorUnknownUser
	StatusErrorRateLimit
)

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrRateLimit            = errors.New("rate limited")
	ErrLookupFailed         = errors.New("lookup failed")
	ErrUnknownLookupFailure = errors.New("unknown lookup failure")
)

// Todo: metrics / tracing / timing

// Todo: Could have a base logger which we then apply context to when needed
type McClient struct {
	Caches struct {
		UUID     cache.Cache
		UserData cache.Cache
		Textures cache.Cache
	}
	API               minecraft.Minecraft
	TexturesMcNetBase string
}

// Todo: Not sure I love this or whether a Context might make more sense
type UserReq struct{ minecraft.User }

// Todo: I need to be providing logging and request context in here
func (mc *McClient) GetSkin(logger log.Logger, userReq UserReq) minecraft.Skin {
	var uuid string
	if userReq.UUID == "" {
		logger = logger.With("username", userReq.Username)
		uuidEntry, err := mc.GetUUIDEntry(logger, userReq.Username)
		if err != nil {
			logger.Debugf("Falling back to Steve: %v", err)
			skin, _ := minecraft.FetchSkinForSteve()
			return skin
		}
		uuid = uuidEntry.UUID
	} else {
		uuid = userReq.UUID
	}

	logger = logger.With("uuid", uuid)
	user, err := mc.GetMcUser(logger, uuid)
	if err != nil {
		logger.Debugf("Falling back to Steve: %v", err)
		skin, _ := minecraft.FetchSkinForSteve()
		return skin
	}

	_ = user
	return minecraft.Skin{}
}

func (mc *McClient) RequestUUIDEntry(logger log.Logger, username string, uuidEntry mc_uuid.UUIDEntry) mc_uuid.UUIDEntry {
	// Metrics timer / tracing
	// GetUUID uses the GetAPIProfile which would also pull the Username (not wanted)
	uuidFresh, err := mc.API.GetUUID(username)
	// Observe GetUUID timer
	uuidEntryFresh := mc_uuid.NewUUIDEntry(logger, username, uuidFresh, err)

	if !uuidEntryFresh.IsValid() && uuidEntry.IsValid() {
		// New result errored, but the original/stale Entry was already valid - Don't cache!
		return uuidEntry
	}

	logger.With("uuid", uuidEntryFresh.UUID)
	// Todo: goroutine?
	mc.CacheInsertUUIDEntry(logger, username, uuidEntryFresh)
	return uuidEntryFresh
}

// Todo: This should handle all cache / API
// Todo: Do we want a WaitGroup here?
// Only real downside is that we can't goroutine to insert into cache?
// Unless we have 2 locks? 1 here, and then one that blocks reads when writing?
func (mc *McClient) GetUUIDEntry(logger log.Logger, username string) (uuidEntry mc_uuid.UUIDEntry, err error) {
	uuidEntry, err = mc.CacheRetrieveUUIDEntry(logger, username)
	if err != nil {
		if err == cache.ErrNotFound {
			// We cache missed (cache.ErrNotFound) so let's request from API
			uuidEntry = mc.RequestUUIDEntry(logger, username, uuidEntry)
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
		// A stale result should be re-requested
		return mc.RequestUUIDEntry(logger, username, uuidEntry), nil
	}

	// A bad result was returned from the cache, generate an error from it
	return uuidEntry, uuidEntry.Status.GetError()
}

func (mc *McClient) RequestMcUser(logger log.Logger, uuid string, mcUser mcuser.McUser) mcuser.McUser {
	// Metrics timer / tracing
	sessionProfile, err := mc.API.GetSessionProfile(uuid)
	// Observe GetUUID timer

	mcUserFresh := mcuser.NewMcUser(logger, uuid, sessionProfile, err)

	if !mcUserFresh.IsValid() && mcUser.IsValid() {
		// New result errored, but the original/stale Entry was already valid - Don't cache!
		return mcUser
	}

	// Todo: Add username to logger With() field?
	// Todo: goroutine?
	mc.CacheInsertMcUser(logger, uuid, mcUserFresh)
	return mcUserFresh
}

func (mc *McClient) GetMcUser(logger log.Logger, uuid string) (mcUser mcuser.McUser, err error) {
	mcUser, err = mc.CacheRetrieveMcUser(logger, uuid)
	if err != nil {
		if err == cache.ErrNotFound {
			// We cache missed (cache.ErrNotFound) so let's request from API
			mcUser = mc.RequestMcUser(logger, uuid, mcUser)
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
		// A stale result should be re-requested
		return mc.RequestMcUser(logger, uuid, mcUser), nil
	}

	// A bad result was returned from the cache, generate an error from it
	return mcUser, mcUser.Status.GetError()
}

// Todo: Counters also support exemplars!
