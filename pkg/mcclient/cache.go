package mcclient

import (
	"fmt"

	"github.com/minotar/imgd/pkg/cache"
	"github.com/minotar/imgd/pkg/mcclient/mcuser"
	"github.com/minotar/imgd/pkg/mcclient/uuid"
	"github.com/minotar/imgd/pkg/util/log"
)

// Todo: metrics / tracing / timing

// Todo: these should all be lowercasing the Username/UUID before retrieving/inserting caches

// Todo: Also, add function context to logger
// User cache name and UUID in logger.With()?

// CacheRetrieveUUIDEntry searches the cache based on Username, expecting a UUID in response
func (mc *McClient) CacheRetrieveUUIDEntry(logger log.Logger, username string) (uuidEntry uuid.UUIDEntry, err error) {
	// Metrics timer / tracing
	// Though - is this useless when using a TieredCache which is inherentantly varied?
	uuidBytes, err := mc.Caches.UUID.Retrieve(username)
	// Observe Cache retrieve
	if err != nil {
		// Return an error (and log based on severity)
		if err == cache.ErrNotFound {
			// Metrics stat "Miss"
			logger.Debugf("Did not find username \"%s\" in %s", username, mc.Caches.UUID.Name())
			return
		} else {
			// Metrics stat Cache RetrieveError
			logger.Errorf("Failed to lookup up username \"%s\" in %s: %v", username, mc.Caches.UUID.Name(), err)
			return
		}
	}

	if len(uuidBytes) < 5 {
		// 4 bytes or less would be an invalid status/timestamp
		// Metrics stats Cache Empty Error
		logger.Errorf("Null UUID returned for username \"%s\" in %s: %v", username, mc.Caches.UUID.Name(), uuidBytes)
		return uuid.UUIDEntry{}, fmt.Errorf("cache returned null UUID")
	}

	uuidEntry = uuid.DecodeUUIDEntry(uuidBytes)

	// Metrics stat Hit
	logger.Debugf("Found username \"%s\" in %s: %v", username, mc.Caches.UUID.Name(), uuidEntry)
	return
}

// CacheInsertUUIDEntry takes a valid UUIDEntry and caches it
// There is no sanity checking on the input
// The item is cached witha TTL assuming it's brand-new
func (mc *McClient) CacheInsertUUIDEntry(logger log.Logger, username string, uuidEntry uuid.UUIDEntry) (err error) {
	// Technically this could be empty / nil???
	uuidBytes := uuidEntry.Encode()

	// Metrics timer / tracing
	// Though - is this useless when using a TieredCache which is inherentantly varied?
	err = mc.Caches.UUID.InsertTTL(username, uuidBytes, uuidEntry.TTL())
	// Observe Cache insert
	if err != nil {
		// stats.CacheUUID("error")
		logger.Errorf("Failed Insert username:UUID into cache (%s:%s): %v", username, uuidEntry, err)
	}
	return
}

func (mc *McClient) CacheRetrieveMcUser(logger log.Logger, uuid string) (user mcuser.McUser, err error) {
	// Metrics timer / tracing
	// Though - is this useless when using a TieredCache which is inherentantly varied?
	userBytes, err := mc.Caches.UserData.Retrieve(uuid)
	// Observe Cache retrieve
	if err != nil {
		// Return an error (and log based on severity)
		if err == cache.ErrNotFound {
			// Metrics stat "Miss"
			logger.Debugf("Did not find uuid \"%s\" in %s", uuid, mc.Caches.UserData.Name())
			return
		} else {
			// Metrics stat Cache RetrieveError
			logger.Errorf("Failed to lookup up uuid \"%s\" in %s: %v", uuid, mc.Caches.UserData.Name(), err)
			return
		}
	}

	user, err = mcuser.DecompressMcUser(userBytes)
	if err != nil {
		logger.Errorf("Failed to decode user from uuid \"%s\" in %s: %v", uuid, mc.Caches.UserData.Name(), err)
		// Metrics stats Cache Decode Error
		return
	}
	// Metrics stat Hit
	logger.Debugf("Found uuid \"%s\" in %s: %s", uuid, mc.Caches.UUID.Name(), uuid)
	return
}

func (mc *McClient) CacheInsertMcUser(logger log.Logger, uuid string, user mcuser.McUser) (err error) {
	packedUserBytes, err := user.Compress()
	if err != nil {
		// stats.CacheUser("pack_error")
		logger.Errorf("Failed to pack UUID:user ready to cache (%s:%s): %v", uuid, user.Username, err)
	}

	// Metrics timer / tracing
	// Though - is this useless when using a TieredCache which is inherentantly varied?
	err = mc.Caches.UUID.InsertTTL(uuid, packedUserBytes, user.TTL())
	// Observe Cache insert
	if err != nil {
		// stats.CacheUser("insert_error")
		logger.Errorf("Failed Insert UUID:user into cache (%s:%s): %v", uuid, user.Username, err)
	}
	return
}
