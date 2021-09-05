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
	usernamePath := route_helpers.UsernamePath
	uuidPath := route_helpers.UUIDPath
	extPath := route_helpers.ExtensionPath
	widPath := route_helpers.WidthPath

	for resource, processor := range processRoutes {
		resPath := "/{resource:" + strings.ToLower(resource) + "}"

		// Todo: name Routes based on UUID vs. Username?

		// Username
		m.Path(resPath + usernamePath + extPath).Handler(skinWrapper(processor)).Name(resource)
		m.Path(resPath + usernamePath + "/" + widPath + extPath).Handler(skinWrapper(processor)).Name(resource)

		// UUID
		m.Path(resPath + uuidPath + extPath).Handler(skinWrapper(processor)).Name(resource)
		m.Path(resPath + uuidPath + "/" + widPath + extPath).Handler(skinWrapper(processor)).Name(resource)
	}

}
