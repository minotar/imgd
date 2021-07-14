package mcclient

import (
	"fmt"

	"github.com/minotar/imgd/pkg/cache"
	"github.com/minotar/imgd/pkg/util/log"
)

// Todo: metrics / tracing / timing

// CacheLookupUsername searches the cache based on Username, expecting a UUID in response
func (mc *McClient) CacheLookupUsername(logger log.Logger, username string) (uuid string, err error) {
	// Metrics timer / tracing
	// Though - is this useless when using a TieredCache which is inherentantly varied?
	uuid, err = cache.RetrieveKV(mc.Cache, username)
	// Observe Cache retrieve
	if err != nil {
		if err == cache.ErrNotFound {
			// Metrics stat "Miss"
			logger.Debugf("Did not find username \"%s\" in %s", username, mc.Cache.Name())
			return
		} else {
			// Metrics stat Cache Error
			logger.Errorf("Failed to lookup up username \"%s\" in %s: ", username, mc.Cache.Name(), err)
			return
		}
	}
	if uuid == "" {
		// Should not be a possible code path? (we shouldn't be adding empty values)
		// Metrics stats Cache Empty Error
		logger.Errorf("Empty UUID returned for username \"%s\" in %s", username, mc.Cache.Name())
		return uuid, fmt.Errorf("cache returned empty UUID")
	}
	// Metrics stat Hit
	logger.Debugf("Found username \"%s\" in %s: %s", username, mc.Cache.Name(), uuid)
	return
}

func (mc *McClient) CacheLookupUUID(logger log.Logger, uuid string) (user McUser, err error) {
	logger.Debug("debug")
	_ = uuid
	return
}
