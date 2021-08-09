package skind

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/minotar/imgd/pkg/mcclient"
)

// Requires "uuid" or "username" vars
func SkinPageHandler(skind *Skind) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := skind.Cfg.Logger
		logger.Info("Incoming Request!!!!")

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

		w.WriteHeader(200)
	})
}
