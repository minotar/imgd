package cache_converter

import (
	"strings"
	"time"

	"github.com/minotar/imgd/pkg/cache_converter/legacy_mcuser"
	"github.com/minotar/imgd/pkg/mcclient/mcuser"
	"github.com/minotar/imgd/pkg/mcclient/status"
	"github.com/minotar/imgd/pkg/mcclient/uuid"

	"github.com/minotar/imgd/pkg/minecraft"
	"github.com/minotar/imgd/pkg/util/log"
	"github.com/minotar/imgd/pkg/util/tinytime"
)

const ()

func processUUIDv3(logger log.Logger, inserter CacheInsertProcessor) IteratingProcessor {

	return IteratingProcessor(func(username string, v []byte, ttl time.Duration) {
		if !minecraft.RegexUsername.MatchString(username) {
			logger.Warnf("Username did not validate: %s", username)
			return
		}

		uuidEntry := uuid.UUIDEntry{
			UUID: string(v),
			// Timestamp doesn't really matter for UUIDEntry
			Timestamp: tinytime.NewTinyTime(time.Now()),
		}

		switch uuidEntry.UUID {
		case metaUnknownCode:
			uuidEntry.UUID = ""
			uuidEntry.Status = status.StatusErrorUnknownUser
		case metaRateLimitCode:
			uuidEntry.UUID = ""
			uuidEntry.Status = status.StatusErrorRateLimit
		case metaErrorCode, "":
			uuidEntry.UUID = ""
			uuidEntry.Status = status.StatusErrorGeneric
		default:
			if !uuidEntry.IsValid() {
				logger.Warnf("Invalid UUID for %s", username)
				return
			}
			uuidEntry.Status = status.StatusOk
		}

		logger.Debugf("UUID Entry for %s is %s with TTL %s", username, uuidEntry, ttl)

		uuidBytes := uuidEntry.Encode()

		err := inserter(username, uuidBytes, ttl)

		if err != nil {
			logger.Warnf("Erroring inserting %s with TTL %s: %v", username, ttl, err)
		}

	})
}

func processUserDatav3(logger log.Logger, inserter CacheInsertProcessor) IteratingProcessor {

	return IteratingProcessor(func(uuid string, v []byte, ttl time.Duration) {
		if !minecraft.RegexUUIDPlain.MatchString(uuid) {
			logger.Warnf("UUID did not validate: %s", uuid)
			return
		}

		legacyUser, err := legacy_mcuser.DecodeMcUser(v)
		if err != nil {
			logger.Warnf("Erroring decoding %s: %v", uuid, err)
		}

		mcUser := mcuser.McUser{
			User:      legacyUser.User,
			Timestamp: tinytime.NewTinyTime(legacyUser.Timestamp),
			Textures:  mcuser.Textures{},
		}

		switch legacyUser.Textures.SkinPath {
		case metaUnknownCode:
			mcUser.Status = status.StatusErrorUnknownUser
		case metaRateLimitCode:
			mcUser.Status = status.StatusErrorRateLimit
		case metaErrorCode, "":
			mcUser.Status = status.StatusErrorGeneric
		default:
			// v3 was always True
			mcUser.Textures.TexturesMcNet = true
			mcUser.Textures.SkinPath = strings.TrimPrefix(legacyUser.Textures.SkinPath, "/texture/")

			if !mcUser.IsValid() {
				logger.Warnf("Invalid UserData for %s", uuid)
				return
			}
			mcUser.Status = status.StatusOk

		}

		logger.Debugf("UserData for %s is %s with TTL %s", uuid, mcUser, ttl)

		packedUserBytes, err := mcUser.Compress()
		if err != nil {
			logger.Errorf("Failed to pack UUID:user ready to cache: %v", err)
		}

		err = inserter(uuid, packedUserBytes, ttl)

		if err != nil {
			logger.Warnf("Erroring inserting %s with TTL %s: %v", uuid, ttl, err)
		}

	})
}
