package status

import (
	"strings"
	"time"

	"github.com/minotar/imgd/pkg/util/log"
)

const (
	day = 24 * time.Hour

	uuidTTL = 60 * day
	// Detect sooner if a username has changed hands?
	//uuidFresh        = 30 * day
	uuidUnknownTTL   = 14 * day
	uuidRateLimitTTL = 2 * time.Hour
	uuidErrorTTL     = 1 * time.Hour

	userTTL = 60 * day
	// Detect sooner when a skin changes
	UserFreshTTL     = 2 * time.Hour
	userUnknownTTL   = 7 * day
	userRateLimitTTL = 1 * time.Hour
	userErrorTTL     = 30 * time.Minute
)

// Todo: Username vs. UUID logic??
const (
	StatusUnSet Status = iota
	StatusOk
	StatusErrorGeneric
	StatusErrorUnknownUser
	StatusErrorRateLimit
)

// Status is for recording the API response status for a specific request
// It does not correspond to the data validity - simple a record of how the API responded
type Status uint8

// Implements the error interface
var _ error = new(Status)

func (s Status) Byte() byte {
	return byte(s)
}

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

// Creates an `error` based on the previous API response status
// This is for raising an error already recorded as a Status
func (s Status) GetError() error {
	if s == StatusOk {
		return nil
	}
	return s
}

func (s Status) DurationUUID() time.Duration {
	switch s {
	case StatusOk:
		return uuidTTL
	case StatusErrorUnknownUser:
		return uuidUnknownTTL
	case StatusErrorRateLimit:
		return uuidRateLimitTTL
	default:
		// StatusUnSet, StatusErrorGeneric, Others
		return uuidErrorTTL
	}
}

func (s Status) DurationUser() time.Duration {
	switch s {
	case StatusOk:
		return userTTL
	case StatusErrorUnknownUser:
		return userUnknownTTL
	case StatusErrorRateLimit:
		return userRateLimitTTL
	default:
		// StatusUnSet, StatusErrorGeneric, Others
		return userErrorTTL
	}
}

// Todo: remove the `query` here as it should already be tagged on the logger
// NOTE: This will record metrics based on the given error
func NewStatusFromError(logger log.Logger, query string, err error) Status {
	if err == nil {
		return StatusOk
	}

	errMsg := err.Error()
	switch {

	// Todo: We should have already tagged the logger with the UUID/Username
	// Do we need to specify it in the message??
	case errMsg == "unable to GetAPIProfile: user not found":
		logger.Infof("No UUID found for: %s", query)
		// Previously named "UnknownUsername"
		// stats.Errored("APIProfileUnknown")
		//return metaUnknownCode, usernameUnknownTTL
		apiGetErrors.WithLabelValues("GetAPIProfile", "UnknownUser").Inc()
		return StatusErrorUnknownUser

	case errMsg == "unable to GetSessionProfile: user not found":
		logger.Infof("No User found for: %s", query)
		// Previously named "UnknownUsername"
		// stats.Errored("SessionProfileUnknown")
		//return metaUnknownCode, uuidUnknownTTL
		apiGetErrors.WithLabelValues("GetSessionProfile", "UnknownUser").Inc()
		return StatusErrorUnknownUser

	case errMsg == "unable to GetAPIProfile: rate limited":
		logger.Warnf("Rate limited looking up UUID for: %s", query)
		// Previously named "LookupUUIDRateLimit"
		// stats.Errored("APIProfileRateLimit")
		//return metaRateLimitCode, usernameRateLimitTTL
		apiGetErrors.WithLabelValues("GetAPIProfile", "RateLimit").Inc()
		return StatusErrorRateLimit

	case errMsg == "unable to GetSessionProfile: rate limited":
		logger.Warnf("Rate limited looking up User for: %s", query)
		// Previously named "LookupUUIDRateLimit"
		// stats.Errored("SessionProfileRateLimit")
		//return metaRateLimitCode, uuidRateLimitTTL
		apiGetErrors.WithLabelValues("GetSessionProfile", "RateLimit").Inc()
		return StatusErrorRateLimit

	case strings.HasPrefix(errMsg, "unable to GetAPIProfile"):
		logger.Errorf("Failed UUID lookup for \"%s\": %s", query, errMsg)
		// Previously named "LookupUUID"``
		// stats.Errored("APIProfileGeneric")
		//return metaErrorCode, usernameErrorTTL
		apiGetErrors.WithLabelValues("GetAPIProfile", "Generic").Inc()
		return StatusErrorGeneric

	case strings.HasPrefix(errMsg, "unable to GetSessionProfile"):
		logger.Errorf("Failed SessionProfile lookup for \"%s\": %s", query, errMsg)
		// Previously named "LookupUUID"
		// stats.Errored("SessionProfileGeneric")
		//return metaErrorCode, uuidErrorTTL
		apiGetErrors.WithLabelValues("GetSessionProfile", "Generic").Inc()
		return StatusErrorGeneric

	default:
		// Todo: Probably a DPanicf preferred
		logger.Errorf("Unknown lookup error occured for \"%s\": %s", query, errMsg)
		// Stat GenericLookup Error
		//return metaErrorCode, uuidErrorTTL
		apiGetErrors.WithLabelValues("Unknown", "Generic").Inc()
		return StatusErrorGeneric

	}

}

func MetricTextureFetchError() {
	apiGetErrors.WithLabelValues("TextureFetch", "Generic").Inc()
}
