package imgd

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/felixge/fgprof"
	"github.com/gorilla/mux"
	"github.com/minotar/imgd/pkg/frontend"
	"github.com/minotar/imgd/pkg/render"
	"github.com/minotar/imgd/pkg/skin"
	"github.com/minotar/imgd/pkg/util/route_helpers"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Handler serves every public endpoint without opening a listener.
func (i *Imgd) Handler() http.Handler {
	router := mux.NewRouter()
	if i.Cfg.CorsAllowAll {
		router.Use(route_helpers.CorsHandler)
	}
	if i.Cfg.Server.RegisterInstrumentation {
		router.Handle("/metrics", promhttp.Handler())
		router.PathPrefix("/debug/pprof").Handler(http.DefaultServeMux)
	}
	router.Path("/debug/fgprof").Handler(fgprof.Handler())
	router.Path("/healthcheck").Handler(skin.HealthcheckHandler(i.McClient))
	router.Path("/dbsize").Handler(skin.SizecheckHandler(i.McClient))

	skinWrapper := skin.NewSkinWrapper(i.Cfg.Logger, i.McClient, i.Cfg.UseETags, i.Cfg.RedirectUsername, i.Cfg.CacheControlTTL)

	skin.RegisterSkinRoutes(router, skinWrapper)
	render.RegisterProcessingRoutes(router, skinWrapper, i.ProcessRoutes)
	router.PathPrefix("/").Handler(frontend.Handler())
	return router
}
