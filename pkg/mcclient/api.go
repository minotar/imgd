// Usernames and UUIDs should be normalized before calling mcclient (eg. lowercase / no-dashes)
package mcclient

import (
	"github.com/minotar/imgd/pkg/util/log"
	"github.com/minotar/minecraft"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/minotar/imgd/pkg/mcclient/mcuser"
	"github.com/minotar/imgd/pkg/mcclient/status"
	mc_uuid "github.com/minotar/imgd/pkg/mcclient/uuid"
)

func (mc *McClient) RequestUUIDEntry(logger log.Logger, username string, uuidEntry mc_uuid.UUIDEntry) mc_uuid.UUIDEntry {
	// GetUUID uses the GetAPIProfile which would also pull the Username (not wanted)
	apiTimer := prometheus.NewTimer(apiGetDuration.WithLabelValues("GetAPIProfile"))
	uuidFresh, err := mc.API.GetUUID(username)
	apiTimer.ObserveDuration()
	uuidEntryFresh := mc_uuid.NewUUIDEntry(logger, username, uuidFresh, err)

	if !uuidEntryFresh.IsValid() && uuidEntry.IsValid() {
		// New result errored, but the original/stale Entry was already valid - Don't cache!
		// Todo: stat this?
		return uuidEntry
	}

	logger.With("uuid", uuidEntryFresh.UUID)
	// Todo: goroutine?
	mc.CacheInsertUUIDEntry(logger, username, uuidEntryFresh)
	return uuidEntryFresh
}

func (mc *McClient) RequestMcUser(logger log.Logger, uuid string, mcUser mcuser.McUser) mcuser.McUser {
	apiTimer := prometheus.NewTimer(apiGetDuration.WithLabelValues("GetSessionProfile"))
	sessionProfile, err := mc.API.GetSessionProfile(uuid)
	apiTimer.ObserveDuration()

	mcUserFresh := mcuser.NewMcUser(logger, uuid, sessionProfile, err)

	if !mcUserFresh.IsValid() && mcUser.IsValid() {
		// New result errored, but the original/stale Entry was already valid - Don't cache!
		return mcUser
	}

	// Todo: Add username to logger With() field?
	// Todo: goroutine?
	logger.With("username", mc)
	if mcUserFresh.IsValid() {
		username := mcUserFresh.Username
		logger = logger.With("username", username)
		// Cache the Username -> UUID mapping
		// Todo: Is it okay to copy these values to new object? Status?
		go mc.CacheInsertUUIDEntry(logger, username, mc_uuid.UUIDEntry{
			UUID:      mcUserFresh.UUID,
			Timestamp: mcUserFresh.Timestamp,
			Status:    mcUserFresh.Status,
		})
	}
	mc.CacheInsertMcUser(logger, uuid, mcUserFresh)
	return mcUserFresh
}

func (mc *McClient) RequestTexture(logger log.Logger, textureKey string, textureURL string) (texture minecraft.Texture, err error) {
	// Use our API object for the request
	texture.Mc = mc.API
	texture.URL = textureURL

	// Retry logic?

	apiTimer := prometheus.NewTimer(apiGetDuration.WithLabelValues("TextureFetch"))
	err = texture.Fetch()
	apiTimer.ObserveDuration()

	if err != nil {
		status.MetricTextureFetchError()
		return
	}

	mc.CacheInsertTexture(logger, textureKey, texture)
	return
}
