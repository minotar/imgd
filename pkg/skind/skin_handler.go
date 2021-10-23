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

// SkinProcessor *MUST* call mcuser.TextureIO.Close() before completing
type SkinProcessor func(log.Logger, mcuser.TextureIO) http.HandlerFunc

type SkinWrapper func(SkinProcessor) http.HandlerFunc

// Requires "uuid" or "username" vars
func NewSkinWrapper(logger log.Logger, mc *mcclient.McClient, useEtags bool, cacheControlTTL time.Duration) SkinWrapper {
	return func(processFunc SkinProcessor) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {

			userReq := route_helpers.MuxToUserReq(r)
			logger, skinIO := mc.GetSkinBufferFromReq(logger, userReq)
			defer skinIO.Close()

			//logger.Infof("Texture ID is: %s", skinIO.TextureID)

			w.Header().Add("Cache-Control", fmt.Sprintf("public, max-age=%d", int(cacheControlTTL.Seconds())))

			// Todo: Technically, this ETag handling is _before_ Content* headers are set, so the 304 will be missing them
			if useEtags {
				// ETag is always included (even for 304 responses)
				w.Header().Set("ETag", skinIO.TextureID)

				reqETag := r.Header.Get("If-None-Match")
				// If the ETag matches the TextureID, then no need to process
				if reqETag == skinIO.TextureID {
					w.WriteHeader(http.StatusNotModified)
					return
				}
			}

			handler := processFunc(logger, skinIO)
			handler.ServeHTTP(w, r)
		}
	}
}

// SkinPageProcessor simply copies the TextureIO to the ResponseWriter
func SkinPageProcessor(logger log.Logger, skinIO mcuser.TextureIO) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "image/png")
		io.Copy(w, skinIO)
		skinIO.Close()
	}
}

// SkinDownloadProcessor Uses the TextureIO to set the downloaded filename of the skin
func SkinDownloadProcessor(logger log.Logger, skinIO mcuser.TextureIO) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Disposition", "attachment; filename=\""+skinIO.TextureID+".png\"")
		SkinPageProcessor(logger, skinIO)(w, r)
	}
}
