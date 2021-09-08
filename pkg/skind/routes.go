package skind

import (
	"fmt"
	"image/png"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/minotar/imgd/pkg/cache"
	"github.com/minotar/imgd/pkg/mcclient"
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

func SizecheckHandler(mc *mcclient.McClient) http.Handler {
	caches := mc.Caches
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uuidSize := caches.UUID.Size()
		userdataSize := caches.UserData.Size()
		texturesSize := caches.Textures.Size()

		w.WriteHeader(200)
		fmt.Fprintf(w, "UUID: %d\nUserdata: %d\nTextures: %d\n", uuidSize, userdataSize, texturesSize)
	})
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

func HealthcheckHandler(mc *mcclient.McClient) http.Handler {
	caches := mc.Caches
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	})
}
