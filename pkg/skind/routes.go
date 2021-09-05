package skind

import (
	"image/png"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/minotar/imgd/pkg/util/route_helpers"
	"github.com/minotar/minecraft"
)

func RegisterRoutes(m *mux.Router, skinHandler http.Handler) {

	optionalPNG := "{?:(?:\\.png)?}"

	skinSR := m.PathPrefix("/skin/").Subrouter()
	skinSR.Path(route_helpers.UUIDPath + optionalPNG).Handler(skinHandler).Name("skinUUID")
	skinSR.Path(route_helpers.UsernamePath + optionalPNG).Handler(skinHandler).Name("skinUsername")
	route_helpers.SubRouteDashedRedirect(skinSR)

	downloadSkinHandler := route_helpers.BrowserDownloadHandler(skinHandler)

	downloadSR := m.PathPrefix("/download/").Subrouter()
	downloadSR.Path(route_helpers.UUIDPath + optionalPNG).Handler(downloadSkinHandler).Name("downloadUUID")
	downloadSR.Path(route_helpers.UsernamePath + optionalPNG).Handler(downloadSkinHandler).Name("downloadUsername")
	route_helpers.SubRouteDashedRedirect(downloadSR)
}

func WriteSkin(w http.ResponseWriter, skin minecraft.Skin) {
	w.Header().Add("Content-Type", "image/png")
	png.Encode(w, skin.Image)
}
