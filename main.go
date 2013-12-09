package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/minotar/minecraft"
	"image"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultSize = uint(180)
	MaxSize     = uint(300)
	MinSize     = uint(8)

	StaticLocation = "www"
	SkinCache

	ListenOn = ":9999"

	Minutes            uint = 60
	Hours                   = 60 * Minutes
	Days                    = 24 * Hours
	TimeoutActualSkin       = 2 * Days
	TimeoutFailedFetch      = 15 * Minutes

	MinotarVersion = "1.0"
)

func serveStatic(w http.ResponseWriter, r *http.Request, inpath string) error {
	inpath = path.Clean(inpath)
	r.URL.Path = inpath

	if !strings.HasPrefix(inpath, "/") {
		inpath = "/" + inpath
		r.URL.Path = inpath
	}
	path := StaticLocation + inpath

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		return err
	}

	http.ServeContent(w, r, d.Name(), d.ModTime(), f)
	return nil
}

func serveAssetPage(w http.ResponseWriter, r *http.Request) {
	err := serveStatic(w, r, r.URL.Path)
	if err != nil {
		notFoundPage(w, r)
	}
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	err := serveStatic(w, r, "index.html")
	if err != nil {
		notFoundPage(w, r)
	}
}

type NotFoundHandler struct{}

func (h NotFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)

	f, err := os.Open("static/404.html")
	if err != nil {
		fmt.Fprintf(w, "404 file not found")
		return
	}
	defer f.Close()

	io.Copy(w, f)
}

func notFoundPage(w http.ResponseWriter, r *http.Request) {
	nfh := NotFoundHandler{}
	nfh.ServeHTTP(w, r)
}
func serverErrorPage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(500)
	fmt.Fprintf(w, "500 internal server error")
}

func rationalizeSize(inp string) uint {
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

func addCacheTimeoutHeader(w http.ResponseWriter, timeout uint) {
	w.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d", timeout))
}

func timeBetween(timeA time.Time, timeB time.Time) int64 {
	// millis between two timestamps

	if timeB.Before(timeA) {
		timeA, timeB = timeB, timeA
	}
	return timeB.Sub(timeA).Nanoseconds() / 1000000
}

func fetchImageProcessThen(callback func(minecraft.Skin) (image.Image, error)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		timeReqStart := time.Now()

		vars := mux.Vars(r)

		username := vars["username"]
		size := rationalizeSize(vars["size"])
		ok := true

		var skin minecraft.Skin
		var err error

		// Fetch image using username
		skin, err = minecraft.GetSkin(minecraft.User{Name: username})
		if err != nil {
			// Problem with the returned image, probably means we have an incorrect username
			// Hit the accounts api
			user, err := minecraft.GetUser(username)

			if err != nil {
				// There's no account for this person, serve char
				skin, _ = minecraft.GetSkin(minecraft.User{Name: "char"})
			} else {
				// Get valid skin
				skin, err = minecraft.GetSkin(user)
				if err != nil {
					// Their skin somehow errored, fallback
					skin, _ = minecraft.GetSkin(minecraft.User{Name: "char"})
				}
			}
		}

		timeFetch := time.Now()

		img, err := callback(skin)
		if err != nil {
			serverErrorPage(w, r)
			return
		}
		timeProcess := time.Now()

		imgResized := Resize(size, size, img)
		timeResize := time.Now()

		w.Header().Add("Content-Type", "image/png")
		w.Header().Add("X-Requested", "processed")
		var timeout uint
		if ok {
			w.Header().Add("X-Result", "ok")
			timeout = TimeoutActualSkin
		} else {
			w.Header().Add("X-Result", "failed")
			timeout = TimeoutFailedFetch
		}
		w.Header().Add("X-Timing", fmt.Sprintf("%d+%d+%d=%dms", timeBetween(timeReqStart, timeFetch), timeBetween(timeFetch, timeProcess), timeBetween(timeProcess, timeResize), timeBetween(timeReqStart, timeResize)))
		addCacheTimeoutHeader(w, timeout)
		WritePNG(w, imgResized)
	}
}
func skinPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	username := vars["username"]

	user, _ := minecraft.GetUser(username)
	skin, _ := minecraft.GetSkin(user)

	w.Header().Add("Content-Type", "image/png")
	w.Header().Add("X-Requested", "skin")
	w.Header().Add("X-Result", "ok")

	WritePNG(w, skin.Image)

	/*
		userSkinURL := minotar.URLForUser(username)
		resp, err := http.Get(userSkinURL)
		if err != nil {
			notFoundPage(w, r)
			return
		}
		w.Header().Add("Content-Type", "image/png")
		w.Header().Add("X-Requested", "skin")
		w.Header().Add("X-Result", "ok")
		addCacheTimeoutHeader(w, TIMEOUT_ACTUAL_SKIN)
		defer resp.Body.Close()
		io.Copy(w, resp.Body)
	*/
}
func downloadPage(w http.ResponseWriter, r *http.Request) {
	headers := w.Header()
	headers.Add("Content-Disposition", "attachment; filename=\"skin.png\"")
	skinPage(w, r)
}

func main() {
	avatarPage := fetchImageProcessThen(func(skin minecraft.Skin) (image.Image, error) {
		return GetHead(skin)
	})
	helmPage := fetchImageProcessThen(func(skin minecraft.Skin) (image.Image, error) {
		return GetHelm(skin)
	})

	r := mux.NewRouter()
	r.NotFoundHandler = NotFoundHandler{}

	r.HandleFunc("/avatar/{username:"+minecraft.ValidUsernameRegex+"}{extension:(.png)?}", avatarPage)
	r.HandleFunc("/avatar/{username:"+minecraft.ValidUsernameRegex+"}/{size:[0-9]+}{extension:(.png)?}", avatarPage)

	r.HandleFunc("/helm/{username:"+minecraft.ValidUsernameRegex+"}{extension:(.png)?}", helmPage)
	r.HandleFunc("/helm/{username:"+minecraft.ValidUsernameRegex+"}/{size:[0-9]+}{extension:(.png)?}", helmPage)

	r.HandleFunc("/download/{username:"+minecraft.ValidUsernameRegex+"}{extension:(.png)?}", downloadPage)

	r.HandleFunc("/skin/{username:"+minecraft.ValidUsernameRegex+"}{extension:(.png)?}", skinPage)

	r.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s", MinotarVersion)
	})

	r.HandleFunc("/", indexPage)

	http.Handle("/", r)
	http.HandleFunc("/assets/", serveAssetPage)
	log.Fatalln(http.ListenAndServe(ListenOn, nil))
}
