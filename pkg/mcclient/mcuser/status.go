package mcuser

import (
	"strings"

	"github.com/minotar/imgd/pkg/util/log"
)

// Todo: Username vs. UUID logic??
const (
	StatusUnSet Status = iota
	StatusOk
	StatusErrorGeneric
	StatusErrorUnknownUser
	StatusErrorRateLimit
)

type Status uint8

// Implements the error interface
var _ error = new(Status)

func (s Status) Error() string {
	switch s {
	case StatusUnSet:
		return "status unset"
	case StatusOk:
		// This should never be returned an error
		return "NOT AN ERROR"
	case StatusErrorGeneric:
		// A non-specific upstream error
		return "lookup error"
	case StatusErrorUnknownUser:
		return "user not found"
	case StatusErrorRateLimit:
		return "rate limited"
	default:
		return "unknown lookup failure"
	}
}

// Todo: Do I add a TTL method?

func NewStatusFromError(logger log.Logger, query string, err error) Status {
	errMsg := err.Error()

	switch {

	// Todo: We should have already tagged the logger with the UUID/Username
	// Do we need to specify it in the message??
	case errMsg == "unable to GetAPIProfile: user not found":
		logger.Infof("No UUID found for: %s", query)
		// Previously named "UnknownUsername"
		// stats.Errored("APIProfileUnknown")
		//return metaUnknownCode, usernameUnknownTTL
		return StatusErrorUnknownUser

	case errMsg == "unable to GetSessionProfile: user not found":
		logger.Infof("No User found for: %s", query)
		// Previously named "UnknownUsername"
		// stats.Errored("SessionProfileUnknown")
		//return metaUnknownCode, uuidUnknownTTL
		return StatusErrorUnknownUser

	case errMsg == "unable to GetAPIProfile: rate limited":
		logger.Warnf("Rate limited looking up UUID for: %s", query)
		// Previously named "LookupUUIDRateLimit"
		// stats.Errored("APIProfileRateLimit")
		//return metaRateLimitCode, usernameRateLimitTTL
		return StatusErrorRateLimit

	case errMsg == "unable to GetSessionProfile: rate limited":
		logger.Warnf("Rate limited looking up User for: %s", query)
		// Previously named "LookupUUIDRateLimit"
		// stats.Errored("SessionProfileRateLimit")
		//return metaRateLimitCode, uuidRateLimitTTL
		return StatusErrorRateLimit

	case strings.HasPrefix(errMsg, "unable to GetAPIProfile"):
		logger.Errorf("Failed UUID lookup for \"%s\": %s", query, errMsg)
		// Previously named "LookupUUID"``
		// stats.Errored("APIProfileGeneric")
		//return metaErrorCode, usernameErrorTTL
		return StatusErrorGeneric

	case strings.HasPrefix(errMsg, "unable to GetSessionProfile"):
		logger.Errorf("Failed SessionProfile lookup for \"%s\": %s", query, errMsg)
		// Previously named "LookupUUID"
		// stats.Errored("SessionProfileGeneric")
		//return metaErrorCode, uuidErrorTTL
		return StatusErrorGeneric

	default:
		// Todo: Probably a DPanicf preferred
		logger.Errorf("Unknown lookup error occured for \"%s\": %s", query, errMsg)
		// Stat GenericLookup Error
		//return metaErrorCode, uuidErrorTTL
		return StatusErrorGeneric

	}

}
