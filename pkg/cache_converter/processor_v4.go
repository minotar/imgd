package cache_converter

import (
	"time"

	"github.com/minotar/imgd/pkg/cache_converter/legacy_mcuser"
	"github.com/minotar/imgd/pkg/mcclient/mcuser"
	"github.com/minotar/imgd/pkg/mcclient/status"
	"github.com/minotar/imgd/pkg/mcclient/uuid"
	"github.com/minotar/imgd/pkg/util/log"
	"github.com/minotar/minecraft"
)

const (

	// Legacy codes
	metaUnknownCode   = "204"
	metaRateLimitCode = "429"
	metaErrorCode     = "0"
)

func processUUIDv4(logger log.Logger, inserter CacheInsertProcessor) IteratingProcessor {

	return IteratingProcessor(func(k, v []byte, ttl time.Duration) {
		username := string(k)

		if !minecraft.RegexUsername.MatchString(username) {
			logger.Warn("Username did not validate")
			return
		}

		uuidEntry := uuid.DecodeUUIDEntry(v)

		logger.Debugf("UUID Entry for %s is %s with TTL %s", username, uuidEntry, ttl)

		var err error
		switch uuidEntry.Status {
		case status.StatusOk:
			err = inserter.Insert(username, []byte(uuidEntry.UUID), ttl)
		case status.StatusErrorGeneric:
			err = inserter.Insert(username, []byte(metaErrorCode), ttl)
		case status.StatusErrorUnknownUser:
			err = inserter.Insert(username, []byte(metaUnknownCode), ttl)
		case status.StatusErrorRateLimit:
			err = inserter.Insert(username, []byte(metaRateLimitCode), ttl)
		default:
			logger.Warnf("Invalid status for UUID Entry for %s is %s with TTL %s: %v", username, uuidEntry, ttl, uuidEntry.Status)
			return
		}

		if err != nil {
			logger.Warnf("Erroring inserting %s with TTL %s: %v", username, ttl, err)
		}

	})
}

func processUserDatev4(logger log.Logger, inserter CacheInsertProcessor) IteratingProcessor {

	return IteratingProcessor(func(k, v []byte, ttl time.Duration) {
		uuid := string(k)

		if !minecraft.RegexUUIDPlain.MatchString(uuid) {
			logger.Warn("UUID did not validate")
			return
		}

		mcUser, err := mcuser.DecompressMcUser(v)
		if err != nil {
			logger.Warnf("Unable to decode McUser from UUID %s", uuid)
		}

		logger.Debugf("UserData for %s is %s with TTL %s", uuid, mcUser, ttl)

		skinPath := "/texture/" + mcUser.Textures.SkinPath
		legacyUser := legacy_mcuser.NewMcUser(mcUser.User, skinPath, mcUser.Timestamp.Time())

		var err2 error
		var data []byte
		switch mcUser.Status {
		case status.StatusOk:
			data, err2 = legacyUser.Encode()
		case status.StatusErrorGeneric:
			legacyUser.Textures.SkinPath = metaErrorCode
			data, err2 = legacyUser.Encode()
		case status.StatusErrorUnknownUser:
			legacyUser.Textures.SkinPath = metaUnknownCode
			data, err2 = legacyUser.Encode()
		case status.StatusErrorRateLimit:
			legacyUser.Textures.SkinPath = metaRateLimitCode
			data, err2 = legacyUser.Encode()
		default:
			logger.Debugf("Invalid status for UserData for %s is %s with TTL %s: %v", uuid, mcUser, ttl, mcUser.Status)
			return
		}

		if err2 != nil {
			logger.Warnf("Erroring encoding %s: %v", uuid, err)
		}

		err = inserter.Insert(uuid, data, ttl)
		if err != nil {
			logger.Warnf("Erroring inserting %s with TTL %s: %v", uuid, ttl, err)
		}

	})
}
