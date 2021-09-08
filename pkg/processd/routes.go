package processd

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/minotar/imgd/pkg/util/route_helpers"
	"github.com/minotar/minecraft"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type SkinProcessor func(minecraft.Skin) http.Handler

type SkinWrapper func(SkinProcessor) http.Handler

// skinWrapper would eg. be SkinLookupWrapper
func RegisterRoutes(m *mux.Router, skinWrapper SkinWrapper, processRoutes map[string]SkinProcessor) {
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
