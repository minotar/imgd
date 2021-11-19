// Usernames and UUIDs should be normalized before calling mcclient (eg. lowercase / no-dashes)
package mcclient

import (
	"bytes"
	"context"
	"io"

	"github.com/minotar/imgd/pkg/minecraft"
	"github.com/minotar/imgd/pkg/util/log"

	"github.com/minotar/imgd/pkg/mcclient/mcuser"
	"github.com/minotar/imgd/pkg/mcclient/status"
	mc_uuid "github.com/minotar/imgd/pkg/mcclient/uuid"
)

func (mc *McClient) RequestUUIDEntry(logger log.Logger, username string, uuidEntry mc_uuid.UUIDEntry) mc_uuid.UUIDEntry {
	// GetUUID uses the GetAPIProfile which would also pull the Username (not wanted)
	uuidFresh, err := mc.API.GetUUID(username)
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
	sessionProfile, err := mc.API.GetSessionProfile(uuid)

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

// Remember to close the mcuser.TextureIO.ReadCloser if error is nil
func (mc *McClient) RequestTexture(logger log.Logger, textureKey string, textureURL string) (textureIO mcuser.TextureIO, err error) {
	// Use our API object for the request
	textureIO.TextureID = textureKey

	// Todo: Retry logic?

	// Set Ctx Source for metrics
	ctx := minecraft.CtxWithSource(context.Background(), "TextureFetch")
	respBody, err := mc.API.ApiRequestCtx(ctx, textureURL)

	if err != nil {
		logger.Warnf("Texture fetch failed: %v", err)
		status.MetricTextureFetchError()
		return
	}
	// The respBody is used here - we create a new ReadCloser for the mcuser.TextureIO
	defer respBody.Close()

	// Todo: verify this isn't super inefficient..!

	// Read the bytes so we can then send to cache
	textureBytes, err := io.ReadAll(respBody)
	mc.CacheInsertTexture(logger, textureKey, textureBytes)

	// Put the bytes back into a ReadCloser so we can use them later
	textureIO.ReadCloser = io.NopCloser(bytes.NewReader(textureBytes))
	return
}
