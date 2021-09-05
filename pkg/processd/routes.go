package processd

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/minotar/imgd/pkg/util/route_helpers"
	"github.com/minotar/minecraft"
)

type SkinProcessor func(minecraft.Skin) http.Handler

type SkinWrapper func(SkinProcessor) http.Handler

// skinWrapper would eg. be SkinLookupWrapper
func RegisterRoutes(m *mux.Router, skinWrapper SkinWrapper, processRoutes map[string]SkinProcessor) {

	for resource, processor := range processRoutes {
		mainPath := "/{resource:" + strings.ToLower(resource) + "}" + route_helpers.UserPath
		m.Path(mainPath).Handler(skinWrapper(processor)).Name(resource)
		m.Path(mainPath + "/{width:[0-9]+}").Handler(skinWrapper(processor)).Name(resource)
	}

}
