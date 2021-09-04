package skind

import (
	"image/png"
	"net/http"

	"github.com/minotar/imgd/pkg/util/route_helpers"
	"github.com/minotar/minecraft"
)

// Requires "uuid" or "username" vars
func SkinPageHandler(skind *Skind) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := skind.Cfg.Logger

		userReq := route_helpers.MuxToUserReq(r)
		skin := skind.McClient.GetSkinFromReq(logger, userReq)

		logger.Infof("User hash is: %s", skin.Hash)

		// No more header changes after writing
		WriteSkin(w, skin)
		logger.Debug(w.Header())
	})
}

func WriteSkin(w http.ResponseWriter, skin minecraft.Skin) {
	// Todo: do we still want to use Skin Hash
	eTag := skin.Hash
	w.Header().Add("ETag", eTag)
	w.Header().Add("Content-Type", "image/png")
	png.Encode(w, skin.Image)
}
