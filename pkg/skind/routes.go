package skind

import (
	"fmt"
	"net/http"
	"time"

	"github.com/felixge/fgprof"
	"github.com/gorilla/mux"
	"github.com/minotar/imgd/pkg/cache"
	"github.com/minotar/imgd/pkg/mcclient"
	"github.com/minotar/imgd/pkg/util/route_helpers"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// routes registers all the routes
func (s *Skind) routes() {
	s.Server.HTTP.Path("/debug/fgprof").Handler(fgprof.Handler())
	s.Server.HTTP.Path("/healthcheck").Handler(HealthcheckHandler(s.McClient))
	s.Server.HTTP.Path("/dbsize").Handler(SizecheckHandler(s.McClient))

	skinWrapper := NewSkinWrapper(s.Cfg.Logger, s.McClient, s.Cfg.UseETags, s.Cfg.RedirectUsername, s.Cfg.CacheControlTTL)
	RegisterSkinRoutes(s.Server.HTTP, skinWrapper)
}

func RegisterSkinRoutes(m *mux.Router, skinWrapper SkinWrapper) {

	optionalPNG := "{?:(?:\\.png)?}"
	uuidCounter := requestedUserType.MustCurryWith(prometheus.Labels{"type": "UUID"})
	dashedCounter := requestedUserType.MustCurryWith(prometheus.Labels{"type": "DashedUUID"})
	usernameCounter := requestedUserType.MustCurryWith(prometheus.Labels{"type": "Username"})

	skinPageHandler := skinWrapper(SkinPageProcessor)

	skinSR := m.PathPrefix("/skin/").Subrouter()
	skinSR.Path(route_helpers.UUIDPath + optionalPNG).Handler(promhttp.InstrumentHandlerCounter(uuidCounter, skinPageHandler)).Name("skin")
	skinSR.Path(route_helpers.UsernamePath + optionalPNG).Handler(promhttp.InstrumentHandlerCounter(usernameCounter, skinPageHandler)).Name("skin")
	route_helpers.SubRouteDashedRedirect(skinSR, dashedCounter)

	downloadSkinHandler := skinWrapper(SkinDownloadProcessor)

	downloadSR := m.PathPrefix("/download/").Subrouter()
	downloadSR.Path(route_helpers.UUIDPath + optionalPNG).Handler(promhttp.InstrumentHandlerCounter(uuidCounter, downloadSkinHandler)).Name("download")
	downloadSR.Path(route_helpers.UsernamePath + optionalPNG).Handler(promhttp.InstrumentHandlerCounter(usernameCounter, downloadSkinHandler)).Name("download")
	route_helpers.SubRouteDashedRedirect(downloadSR, dashedCounter)
}

func SizecheckHandler(mc *mcclient.McClient) http.HandlerFunc {
	caches := mc.Caches
	return func(w http.ResponseWriter, r *http.Request) {
		uuidSize := caches.UUID.Size()
		userdataSize := caches.UserData.Size()
		texturesSize := caches.Textures.Size()

		w.WriteHeader(200)
		fmt.Fprintf(w, "UUID: %d\nUserdata: %d\nTextures: %d\n", uuidSize, userdataSize, texturesSize)
	}
}

func checkCache(c cache.Cache) (string, error) {
	key := "!!HEALTHCHECK"
	nowStr := time.Now().String()
	err := c.InsertTTL(key, []byte(nowStr), time.Second)
	if err != nil {
		return fmt.Sprintf("%s Errored: %v", c.Name(), err), err
	}
	_, err = c.Retrieve(key)
	if err != nil {
		return fmt.Sprintf("%s Errored: %v", c.Name(), err), err
	}
	return fmt.Sprintf("%s is OK", c.Name()), nil
}

func HealthcheckHandler(mc *mcclient.McClient) http.HandlerFunc {
	caches := mc.Caches
	return func(w http.ResponseWriter, r *http.Request) {
		var message string
		msg, err1 := checkCache(caches.UUID)
		message += msg + "\n"

		msg, err2 := checkCache(caches.UserData)
		message += msg + "\n"

		msg, err3 := checkCache(caches.Textures)
		message += msg + "\n"

		if err1 != nil || err2 != nil || err3 != nil {
			w.WriteHeader(503)
		} else {
			w.WriteHeader(200)
		}
		fmt.Fprint(w, message)
	}
}
