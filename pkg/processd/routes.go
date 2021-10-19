package processd

import (
	"net/http"
	"strings"

	"github.com/felixge/fgprof"
	"github.com/gorilla/mux"
	"github.com/minotar/imgd/pkg/skind"
	"github.com/minotar/imgd/pkg/util/route_helpers"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// routes registers all the routes
func (p *Processd) routes() {
	p.Server.HTTP.Path("/debug/fgprof").Handler(fgprof.Handler())
	p.Server.HTTP.Path("/healthcheck").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	RegisterProcessingRoutes(p.Server.HTTP, p.SkinLookupWrapper, p.ProcessRoutes)
}

func RegisterProcessingRoutes(m *mux.Router, skinWrapper skind.SkinWrapper, processRoutes map[string]skind.SkinProcessor) {
	uuidCounter := requestedUserType.MustCurryWith(prometheus.Labels{"type": "UUID"})
	dashedCounter := requestedUserType.MustCurryWith(prometheus.Labels{"type": "DashedUUID"})
	usernameCounter := requestedUserType.MustCurryWith(prometheus.Labels{"type": "Username"})

	usernamePath := route_helpers.UsernamePath
	uuidPath := route_helpers.UUIDPath
	extPath := route_helpers.ExtensionPath
	widPath := route_helpers.WidthPath

	for resource, processor := range processRoutes {
		handler := skinWrapper(processor)
		usernameHandler := promhttp.InstrumentHandlerCounter(usernameCounter, handler)
		uuidHandler := promhttp.InstrumentHandlerCounter(uuidCounter, handler)

		resPath := "/{resource:" + strings.ToLower(resource) + "}/"
		sr := m.PathPrefix(resPath).Subrouter()

		// Username
		sr.Path(usernamePath + extPath).Handler(usernameHandler).Name(resource)
		sr.Path(usernamePath + "/" + widPath + extPath).Handler(usernameHandler).Name(resource)

		// UUID
		sr.Path(uuidPath + extPath).Handler(uuidHandler).Name(resource)
		sr.Path(uuidPath + "/" + widPath + extPath).Handler(uuidHandler).Name(resource)

		// Dashed Redirect
		route_helpers.SubRouteDashedRedirect(sr, dashedCounter)
	}

}
