package route_helpers

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/minotar/imgd/pkg/mcclient"
	"github.com/minotar/imgd/pkg/util/log"
	"github.com/minotar/minecraft"
)

const (
	UUIDPath     = "/{uuid:" + minecraft.ValidUUIDPlainRegex + "}"
	DashPath     = "/{dashedUUID:" + minecraft.ValidUUIDDashRegex + "}"
	UsernamePath = "/{username:" + minecraft.ValidUsernameRegex + "}"
	UserPath     = "/{user:" + minecraft.ValidUsernameRegex + "|" + minecraft.ValidUUIDPlainRegex + "}"
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

func DashedRedirectUUIDHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Todo: log stat
		dst := strings.Replace(r.URL.Path, "-", "", 4)
		http.Redirect(w, r, dst, http.StatusMovedPermanently)
	})
}

func SubRouteDashedRedirect(m *mux.Router) {
	m.Path(DashPath + "{?:(?:.[a-z]{3})?}").Handler(DashedRedirectUUIDHandler()).Name("dashedRedirect")
}

// var "username" or "uuid" _MUST_ be present
func MuxToUserReq(r *http.Request) (userReq mcclient.UserReq) {
	vars := mux.Vars(r)

	if username, usernameGiven := vars["username"]; usernameGiven {
		userReq.Username = username
	} else if uuid, uuidGiven := vars["uuid"]; uuidGiven {
		userReq.UUID = uuid
	}
	return
}

func LoggingMiddleware(logger log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.With(
				"method", r.Method,
				"content_length", r.ContentLength,
				"host", r.Host,
				"remote_addr", r.RemoteAddr,
				"url", r.URL,
			).Debug("incoming")
			next.ServeHTTP(w, r)
		})
	}
}
