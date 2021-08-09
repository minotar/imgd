// Usernames and UUIDs should be normalized before calling mcclient (eg. lowercase / no-dashes)
package mcclient

import (
	"github.com/minotar/imgd/pkg/util/log"
	"github.com/minotar/minecraft"

	"github.com/minotar/imgd/pkg/mcclient/mcuser"
	mc_uuid "github.com/minotar/imgd/pkg/mcclient/uuid"
)

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

func (mc *McClient) RequestTexture(logger log.Logger, textureKey string, textureURL string) (texture minecraft.Texture, err error) {
	// Use our API object for the request
	texture.Mc = mc.API
	texture.URL = textureURL

	// Retry logic?

	// Metrics timer / tracing
	err = texture.Fetch()
	// Observe Texture Fetch() timer

	if err != nil {
		return
	}

	mc.CacheInsertTexture(logger, textureKey, texture)
	return
}
