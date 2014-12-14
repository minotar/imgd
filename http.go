package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/minotar/minecraft"
	"net/http"
	"strconv"
	"strings"
)

type Router struct {
	Mux *mux.Router
}

type NotFoundHandler struct{}

// Handles 404 errors
func (h NotFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	fmt.Fprintf(w, "404 not found")
}

// Converts and sanitizes the string for the avatar size.
func (r *Router) GetSize(inp string) uint {
	out64, err := strconv.ParseUint(inp, 10, 0)
	out := uint(out64)
	if err != nil {
		return DefaultSize
	} else if out > MaxSize {
		return MaxSize
	} else if out < MinSize {
		return MinSize
	}
	return out

}

// Shows only the user's skin.
func (router *Router) SkinPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	username := vars["username"]

	skin := fetchSkin(username)

	w.Header().Add("Content-Type", "image/png")
	w.Header().Add("X-Requested", "skin")
	w.Header().Add("X-Result", "ok")

	skin.WriteSkin(w)
}

// Shows the skin and tells the browser to attempt to download it.
func (router *Router) DownloadPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Disposition", "attachment; filename=\"skin.png\"")
	router.SkinPage(w, r)
}

// Shows the skin and tells the browser to attempt to download it.
func (router *Router) CacheRemovePage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	username := vars["username"]

	cache.remove(strings.ToLower(username))

	w.WriteHeader(204)
}

// Pull the Get<resource> method from the skin. Originally this used
// reflection, but that was slow.
func (router *Router) ResolveMethod(skin *mcSkin, resource string) func(int) error {
	switch resource {
	case "Helm":
		return skin.GetHelm
	case "Avatar":
		return skin.GetHead
	case "Cube":
		return skin.GetCube
	case "Body":
		return skin.GetBody
	case "Bust":
		return skin.GetBust
	default:
		return skin.GetHelm
	}
}

// Binds the route and makes a handler function for the requested resource.
func (router *Router) Serve(resource string) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		size := router.GetSize(vars["size"])
		skin := fetchSkin(vars["username"])

		err := router.ResolveMethod(skin, resource)(int(size))
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "500 internal server error")
			return
		}

		w.Header().Add("Content-Type", "image/png")
		w.Header().Add("X-Skin-Hash", skin.Hash)
		w.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d", TimeoutActualSkin))
		skin.WritePNG(w)
	}

	router.Mux.HandleFunc("/"+strings.ToLower(resource)+"/{username:"+minecraft.ValidUsernameRegex+"}{extension:(.png)?}", fn)
	router.Mux.HandleFunc("/"+strings.ToLower(resource)+"/{username:"+minecraft.ValidUsernameRegex+"}/{size:[0-9]+}{extension:(.png)?}", fn)
}

// Binds routes to the ServerMux.
func (router *Router) Bind() {

	router.Mux.NotFoundHandler = NotFoundHandler{}

	router.Serve("Avatar")
	router.Serve("Helm")
	router.Serve("Body")
	router.Serve("Bust")
	router.Serve("Cube")

	router.Mux.HandleFunc("/download/{username:"+minecraft.ValidUsernameRegex+"}{extension:(.png)?}", router.DownloadPage)
	router.Mux.HandleFunc("/skin/{username:"+minecraft.ValidUsernameRegex+"}{extension:(.png)?}", router.SkinPage)
	router.Mux.HandleFunc("/remove/{username:"+minecraft.ValidUsernameRegex+"}", router.CacheRemovePage)

	router.Mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s", MinotarVersion)
	})

	router.Mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(stats.ToJSON())
	})

	router.Mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://minotar.net/", 302)
	})
}

func fetchSkin(username string) *mcSkin {
	if cache.has(strings.ToLower(username)) {
		stats.HitCache()
		return &mcSkin{Processed: nil, Skin: cache.pull(strings.ToLower(username))}
	}

	skin, err := minecraft.FetchSkinFromMojang(username)
	if err != nil {
		log.Error("Failed Skin Mojang: " + username + " (" + err.Error() + ")")
		// Let's fallback to S3 and try and serve at least an old skin...
		skin, err = minecraft.FetchSkinFromS3(username)
		if err != nil {
			log.Error("Failed Skin S3: " + username + " (" + err.Error() + ")")
			// Well, looks like they don't exist after all.
			skin, _ = minecraft.FetchSkinForChar()
		}
	}

	stats.MissCache()
	cache.add(strings.ToLower(username), skin)

	return &mcSkin{Processed: nil, Skin: skin}
}
