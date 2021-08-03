package mcclient

import (
	"fmt"
	"time"

	"github.com/minotar/imgd/pkg/cache"
	"github.com/minotar/imgd/pkg/mcclient/mcuser"
	"github.com/minotar/imgd/pkg/util/log"
)

// Todo: metrics / tracing / timing

// Todo: these should all be lowercasing the Username/UUID before retrieving/inserting caches

// Todo: Also, add function context to logger
// User cache name and UUID in logger.With()?

// CacheLookupUsername searches the cache based on Username, expecting a UUID in response
func (mc *McClient) CacheLookupUsername(logger log.Logger, username string) (uuid string, err error) {
	// Metrics timer / tracing
	// Though - is this useless when using a TieredCache which is inherentantly varied?
	uuid, err = cache.RetrieveKV(mc.Caches.UUID, username)
	// Observe Cache retrieve
	if err != nil {
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
	if uuid == "" {
		// Should not be a possible code path? (we shouldn't be adding empty values)
		// Metrics stats Cache Empty Error
		logger.Errorf("Empty UUID returned for username \"%s\" in %s", username, mc.Caches.UUID.Name())
		return uuid, fmt.Errorf("cache returned empty UUID")
	}
	// Metrics stat Hit
	logger.Debugf("Found username \"%s\" in %s: %s", username, mc.Caches.UUID.Name(), uuid)
	return
}

func (mc *McClient) CacheAddUsername(logger log.Logger, username string, uuid string, ttl time.Duration) (err error) {
	// Metrics timer / tracing
	// Though - is this useless when using a TieredCache which is inherentantly varied?
	err = cache.InsertKV(mc.Caches.UUID, username, uuid, ttl)
	// Observe Cache retrieve
	if err != nil {
		// stats.CacheUUID("error")
		logger.Errorf("Failed Insert username:UUID into cache (%s:%s): %v", username, uuid, err)
	}
	return
}

func (mc *McClient) CacheLookupUUID(logger log.Logger, uuid string) (user mcuser.McUser, err error) {
	// Metrics timer / tracing
	// Though - is this useless when using a TieredCache which is inherentantly varied?
	userBytes, err := mc.Caches.UserData.Retrieve(uuid)
	// Observe Cache retrieve
	if err != nil {
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

func (mc *McClient) CacheAddUUID(logger log.Logger, uuid string, user mcuser.McUser, ttl time.Duration) (err error) {
	packedUserBytes, err := user.Compress()
	if err != nil {
		// stats.CacheUser("pack_error")
		logger.Errorf("Failed to pack UUID:user ready to cache (%s:%s): %v", uuid, user.Username, err)
	}

	// Metrics timer / tracing
	// Though - is this useless when using a TieredCache which is inherentantly varied?
	err = mc.Caches.UUID.InsertTTL(uuid, packedUserBytes, ttl)
	// Observe Cache retrieve
	if err != nil {
		// stats.CacheUser("insert_error")
		logger.Errorf("Failed Insert UUID:user into cache (%s:%s): %v", uuid, user.Username, err)
	}
	return
}
