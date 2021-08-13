package skind

import (
	"image/png"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/minotar/imgd/pkg/mcclient"
	"github.com/minotar/minecraft"
)


func CorsHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
		next.ServeHTTP(w, r)
	})
}

func BrowserDownloadHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Todo: pretty download name would be nice...
		w.Header().Add("Content-Disposition", "attachment; filename=\"skin.png\"")
		next.ServeHTTP(w, r)
	})
}

// Requires "uuid" or "username" vars
func SkinPageHandler(skind *Skind) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := skind.Cfg.Logger

		var userReq mcclient.UserReq
		vars := mux.Vars(r)

		if username, name_req := vars["username"]; name_req {
			userReq.Username = username
			logger.Debugf("username: %+v\n", userReq.Username)
		} else {
			userReq.UUID = vars["uuid"]
			logger.Debugf("uuid: %+v\n", userReq.UUID)
		}

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

func DashedRedirectUUIDHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Todo: log stat
		dst := strings.Replace(r.URL.Path, "-", "", 4)
		http.Redirect(w, r, dst, http.StatusMovedPermanently)
	})
}
