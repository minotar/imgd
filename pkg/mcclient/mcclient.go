// Usernames and UUIDs should be normalized before calling mcclient (eg. lowercase / no-dashes)
package mcclient

import (
	"errors"
	"strings"
	"time"

	"github.com/minotar/imgd/pkg/cache"
	"github.com/minotar/imgd/pkg/util/log"

	"github.com/minotar/imgd/pkg/mcclient/mcuser"
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
		fetchedUuid, err := mc.GetUUID(logger, userReq.Username)
		if err != nil {
			skin, _ := minecraft.FetchSkinForSteve()
			return skin
		}
		uuid = fetchedUuid
	} else {
		uuid = userReq.UUID
	}

	logger = logger.With("uuid", uuid)
	user, err := mc.GetUser(logger, uuid)

	_, _ = user, err
	return minecraft.Skin{}
}

// Todo: This should handle all cache / API
// Todo: Do we want a WaitGroup here?
// Only real downside is that we can't goroutine to insert into cache?
// Unless we have 2 locks? 1 here, and then one that blocks reads when writing?
func (mc *McClient) GetUUID(logger log.Logger, username string) (string, error) {
	uuid, err := mc.CacheLookupUsername(logger, username)
	if err != nil && err != cache.ErrNotFound {
		// Cache experieneed a proper error (already would be logged)
		return "", err
	}

	if minecraft.RegexUUIDPlain.MatchString(uuid) {
		// Great success - we have a cached result
		return uuid, nil
	}

	var errCode string
	// Either we cache missed (cache.ErrNotFound), OR, we have a cached error code
	if err == cache.ErrNotFound {
		// Metrics timer / tracing
		// GetUUID uses the GetAPIProfile which would also pull the Username (not wanted)
		uuid, err := mc.API.GetUUID(username)
		// Observe GetUUID
		if minecraft.RegexUUIDPlain.MatchString(uuid) {
			// Great success - we have a fresh result
			// Todo: goroutine?
			mc.CacheAddUsername(logger, username, uuid, usernameTTL)
			return uuid, nil
		}
		// Need to decide on what to do based on returned API errors
		// eg. Cache a meta code for X TTL
		// Set errCode based on this
		var ttl time.Duration
		errCode, ttl = errorToMetaCode(logger, username, err)
		// Todo: goroutine?
		mc.CacheAddUsername(logger, username, errCode, ttl)
	} else {
		// Todo: Convert error code to number
		errCode = uuid
	}

	return "", metaCodeToError(logger, username, errCode)
}

func (mc *McClient) GetUser(logger log.Logger, uuid string) (mcuser.McUser, error) {
	user, err := mc.CacheLookupUUID(logger, uuid)
	if err != nil && err != cache.ErrNotFound {
		// Cache experieneed a proper error (already would be logged)
		return user, err
	}

	if user.Status == StatusOk {
		// Known positive result
		// Todo: Stale checks??
		return user, err
	}

	if err == cache.ErrNotFound {
		// Todo: !!! this
		// Metrics timer / tracing
		// GetUUID uses the GetAPIProfile which would also pull the Username (not wanted)
		apiUser, err := mc.API.GetUUID(username)
		// Observe GetUUID
	}

	// Todo: !!!

	return user, nil
}

// Todo: Counters also support exemplars!

// Takes an error from the GetAPIProfile or GetSessionProfile and returns a metacode and TTL
func errorToMetaCode(logger log.Logger, query string, err error) (string, time.Duration) {
	errMsg := err.Error()

	switch {

	// Todo: We should have already tagged the logger with the UUID/Username
	// Do we need to specify it in the message??
	case errMsg == "unable to GetAPIProfile: user not found":
		logger.Infof("No UUID found for: %s", query)
		// Previously named "UnknownUsername"
		// stats.Errored("APIProfileUnknown")
		return metaUnknownCode, usernameUnknownTTL

	case errMsg == "unable to GetSessionProfile: user not found":
		logger.Infof("No User found for: %s", query)
		// Previously named "UnknownUsername"
		// stats.Errored("SessionProfileUnknown")
		return metaUnknownCode, uuidUnknownTTL

	case errMsg == "unable to GetAPIProfile: rate limited":
		logger.Warnf("Rate limited looking up UUID for: %s", query)
		// Previously named "LookupUUIDRateLimit"
		// stats.Errored("APIProfileRateLimit")
		return metaRateLimitCode, usernameRateLimitTTL

	case errMsg == "unable to GetSessionProfile: rate limited":
		logger.Warnf("Rate limited looking up User for: %s", query)
		// Previously named "LookupUUIDRateLimit"
		// stats.Errored("SessionProfileRateLimit")
		return metaRateLimitCode, uuidRateLimitTTL

	case strings.HasPrefix(errMsg, "unable to GetAPIProfile"):
		logger.Errorf("Failed UUID lookup for \"%s\": %s", query, errMsg)
		// Previously named "LookupUUID"
		// stats.Errored("APIProfileGeneric")
		return metaErrorCode, usernameErrorTTL

	case strings.HasPrefix(errMsg, "unable to GetSessionProfile"):
		logger.Errorf("Failed SessionProfile lookup for \"%s\": %s", query, errMsg)
		// Previously named "LookupUUID"
		// stats.Errored("SessionProfileGeneric")
		return metaErrorCode, uuidErrorTTL

	default:
		// Todo: Probably a DPanicf preferred
		logger.Errorf("Unknown lookup error occured for \"%s\": %s", query, errMsg)
		// Stat GenericLookup Error
		return metaErrorCode, uuidErrorTTL

	}
}

func metaCodeToError(logger log.Logger, query string, errCode string) error {
	switch errCode {

	case metaUnknownCode:
		return ErrUserNotFound
	case metaRateLimitCode:
		return ErrRateLimit
	case metaErrorCode:
		return ErrLookupFailed
	default:
		logger.Errorf("Unexpected Meta Error Code \"%s\" for: %s", errCode, query)
		return ErrUnknownLookupFailure
	}
}
