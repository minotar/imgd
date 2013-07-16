package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/Axxim/Minotar"
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
	DEFAULT_SIZE = uint(180)
	MAX_SIZE     = uint(300)
	MIN_SIZE     = uint(8)

	STATIC_LOCATION = "static"

	LISTEN_ON = ":9999"

	MINUTES              uint = 60
	HOURS                     = 60 * MINUTES
	DAYS                      = 24 * HOURS
	TIMEOUT_ACTUAL_SKIN       = 2 * HOURS
	TIMEOUT_FAILED_FETCH      = 15 * MINUTES

	SERVICE_VERSION = "0.2"
)

func serveStatic(w http.ResponseWriter, r *http.Request, inpath string) error {
	inpath = path.Clean(inpath)
	r.URL.Path = inpath

	if !strings.HasPrefix(inpath, "/") {
		inpath = "/" + inpath
		r.URL.Path = inpath
	}
	path := STATIC_LOCATION + inpath

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
		return DEFAULT_SIZE
	} else if out > MAX_SIZE {
		return MAX_SIZE
	} else if out < MIN_SIZE {
		return MIN_SIZE
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

func fetchImageProcessThen(callback func(minotar.Skin) (image.Image, error)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		timeReqStart := time.Now()

		vars := mux.Vars(r)

		username := vars["username"]
		size := rationalizeSize(vars["size"])
		ok := true

		var skin minotar.Skin
		var err error
		if username == "char" {
			if skin, err = minotar.FetchSkinForSteve(); err != nil {
				serverErrorPage(w, r)
				return
			}
		} else {
			if skin, err = minotar.FetchSkinForUser(username); err != nil {
				ok = false
				if skin, err = minotar.FetchSkinForSteve(); err != nil {
					serverErrorPage(w, r)
					return
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

		imgResized := minotar.Resize(size, size, img)
		timeResize := time.Now()

		w.Header().Add("Content-Type", "image/png")
		w.Header().Add("X-Requested", "processed")
		var timeout uint
		if ok {
			w.Header().Add("X-Result", "ok")
			timeout = TIMEOUT_ACTUAL_SKIN
		} else {
			w.Header().Add("X-Result", "failed")
			timeout = TIMEOUT_FAILED_FETCH
		}
		w.Header().Add("X-Timing", fmt.Sprintf("%d+%d+%d=%d", timeBetween(timeReqStart, timeFetch), timeBetween(timeFetch, timeProcess), timeBetween(timeProcess, timeResize), timeBetween(timeReqStart, timeResize)))
		addCacheTimeoutHeader(w, timeout)
		minotar.WritePNG(w, imgResized)
	}
}
func skinPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	username := vars["username"]

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
}
func downloadPage(w http.ResponseWriter, r *http.Request) {
	headers := w.Header()
	headers.Add("Content-Disposition", "attachment; filename=\"skin.png\"")
	skinPage(w, r)
}

func main() {
	avatarPage := fetchImageProcessThen(func(skin minotar.Skin) (image.Image, error) {
		return skin.Head()
	})
	helmPage := fetchImageProcessThen(func(skin minotar.Skin) (image.Image, error) {
		return skin.Helm()
	})

	r := mux.NewRouter()
	r.NotFoundHandler = NotFoundHandler{}

	r.HandleFunc("/avatar/{username:"+minotar.VALID_USERNAME_REGEX+"}{extension:(.png)?}", avatarPage)
	r.HandleFunc("/avatar/{username:"+minotar.VALID_USERNAME_REGEX+"}/{size:[0-9]+}{extension:(.png)?}", avatarPage)

	r.HandleFunc("/helm/{username:"+minotar.VALID_USERNAME_REGEX+"}{extension:(.png)?}", helmPage)
	r.HandleFunc("/helm/{username:"+minotar.VALID_USERNAME_REGEX+"}/{size:[0-9]+}{extension:(.png)?}", helmPage)

	r.HandleFunc("/download/{username:"+minotar.VALID_USERNAME_REGEX+"}{extension:(.png)?}", downloadPage)

	r.HandleFunc("/skin/{username:"+minotar.VALID_USERNAME_REGEX+"}{extension:(.png)?}", skinPage)

	r.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s", SERVICE_VERSION)
	})

	r.HandleFunc("/", indexPage)

	http.Handle("/", r)
	http.HandleFunc("/assets/", serveAssetPage)
	log.Fatalln(http.ListenAndServe(LISTEN_ON, nil))
}
