package skind

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/minotar/imgd/pkg/mcclient"
	"github.com/minotar/imgd/pkg/mcclient/mcuser"
	"github.com/minotar/imgd/pkg/util/log"
	"github.com/minotar/imgd/pkg/util/route_helpers"
)

// CapeProcessor *MUST* call mcuser.TextureIO.Close() before completing
type CapeProcessor func(log.Logger, mcuser.TextureIO) http.HandlerFunc

type CapeWrapper func(CapeProcessor) http.HandlerFunc

// Requires "uuid" or "username" vars
func NewCapeWrapper(logger log.Logger, mc *mcclient.McClient, useEtags bool, redirectUsernames bool, cacheControlTTL time.Duration) CapeWrapper {
	return func(processFunc CapeProcessor) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {

			userReq := route_helpers.MuxToUserReq(r)

			// Todo: Instead set headers based on lookup results
			w.Header().Add("Cache-Control", fmt.Sprintf("public, max-age=%d", int(cacheControlTTL.Seconds())))

			if redirectUsernames && userReq.Username != "" {
				// Redirect Usernames is enabled, and a Username was given
				logger, uuid, err := userReq.GetUUID(logger, mc)
				if err != nil {
					logger.Debugf("User not found, cannot redirect: %v", err)
					http.NotFound(w, r)
					return
				}

				http.Redirect(w, r, uuid, http.StatusFound)
				return
			}

			logger, capeIO, hasCape := mc.GetCapeBufferFromReq(logger, userReq)
			if !hasCape {
				// User has no cape, return 404
				logger.Debugf("No cape found for user")
				http.NotFound(w, r)
				return
			}
			defer capeIO.Close()

			// Todo: Technically, this ETag handling is _before_ Content* headers are set, so the 304 will be missing them
			if useEtags {
				// ETag is always included (even for 304 responses)
				w.Header().Set("ETag", capeIO.TextureID)

				reqETag := r.Header.Get("If-None-Match")
				// If the ETag matches the TextureID, then no need to process
				if reqETag == capeIO.TextureID {
					w.WriteHeader(http.StatusNotModified)
					return
				}
			}

			handler := processFunc(logger, capeIO)
			handler.ServeHTTP(w, r)
		}
	}
}

// CapePageProcessor simply copies the TextureIO to the ResponseWriter
func CapePageProcessor(logger log.Logger, capeIO mcuser.TextureIO) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "image/png")
		io.Copy(w, capeIO)
		capeIO.Close()
	}
}
