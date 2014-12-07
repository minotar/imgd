package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/minotar/minecraft"
	"github.com/op/go-logging"
	"image"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/debug"
	"strconv"
	"time"
)

const (
	DefaultSize = uint(180)
	MaxSize     = uint(300)
	MinSize     = uint(8)

	SkinCache

	Minutes            uint = 60
	Hours                   = 60 * Minutes
	Days                    = 24 * Hours
	TimeoutActualSkin       = 2 * Days
	TimeoutFailedFetch      = 15 * Minutes

	MinotarVersion = "2.1"
)

var (
	config = &Configuration{}
)

type NotFoundHandler struct{}

func (h NotFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	fmt.Fprintf(w, "404 not found")
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

		skin := fetchSkin(username)
		var err error

		timeFetch := time.Now()

		img, err := callback(skin)
		if err != nil {
			serverErrorPage(w, r)
			return
		}
		timeProcess := time.Now()

		imgResized := Resize(size, img)
		timeResize := time.Now()

		w.Header().Add("Content-Type", "image/png")
		w.Header().Add("X-Requested", "processed")
		w.Header().Add("X-Skin-Hash", skin.Hash)
		var timeout uint
		if ok {
			w.Header().Add("X-Result", "ok")
			timeout = TimeoutActualSkin
		} else {
			w.Header().Add("X-Result", "failed")
			timeout = TimeoutFailedFetch
		}

		timing := fmt.Sprintf("%d+%d+%d=%dms", timeBetween(timeReqStart, timeFetch), timeBetween(timeFetch, timeProcess), timeBetween(timeProcess, timeResize), timeBetween(timeReqStart, timeResize))

		w.Header().Add("X-Timing", timing)
		addCacheTimeoutHeader(w, timeout)
		WritePNG(w, imgResized)

		log.Info("Serving skin for " + username + " (" + timing + ") md5: " + skin.Hash)
	}
}
func skinPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	username := vars["username"]

	skin := fetchSkin(username)

	w.Header().Add("Content-Type", "image/png")
	w.Header().Add("X-Requested", "skin")
	w.Header().Add("X-Result", "ok")

	WritePNG(w, skin.Image)
}
func downloadPage(w http.ResponseWriter, r *http.Request) {
	headers := w.Header()
	headers.Add("Content-Disposition", "attachment; filename=\"skin.png\"")
	skinPage(w, r)
}

func fetchSkin(username string) minecraft.Skin {
	skin, err := minecraft.FetchSkinFromUrl(username)
	if err != nil {
		log.Error("Failed to get skin for " + username + " from Mojang (" + err.Error() + ")")
		skin, _ = minecraft.FetchSkinForChar()
	}

	return skin

	/* We're not using this for now due to rate limiting restrictions
	skin, err := minecraft.GetSkin(minecraft.User{Name: username})
	if err != nil {
		// Problem with the returned image, probably means we have an incorrect username
		// Hit the accounts api
		user, err := minecraft.GetUser(username)

		if err != nil {
			// There's no account for this person, serve char
			skin, _ = minecraft.FetchSkinForChar()
		} else {
			// Get valid skin
			skin, err = minecraft.GetSkin(user)
			if err != nil {
				// Their skin somehow errored, fallback
				skin, _ = minecraft.FetchSkinForChar()
			}
		}
	}

	return skin
	*/
}

var log = logging.MustGetLogger("imgd")
var format = "[%{time:15:04:05.000000}] %{level:.4s} %{message}"

func main() {
	err := config.load()
	if err != nil {
		fmt.Printf("Error loading config: %s\n", err)
		return
	}

	logBackend := logging.NewLogBackend(os.Stdout, "", 0)
	logging.SetBackend(logBackend)
	logging.SetFormatter(logging.MustStringFormatter(format))

	debug.SetGCPercent(10)

	avatarPage := fetchImageProcessThen(func(skin minecraft.Skin) (image.Image, error) {
		return GetHead(skin)
	})
	helmPage := fetchImageProcessThen(func(skin minecraft.Skin) (image.Image, error) {
		return GetHelm(skin)
	})
	bodyPage := fetchImageProcessThen(func(skin minecraft.Skin) (image.Image, error) {
		return GetBody(skin)
	})
	bustPage := fetchImageProcessThen(func(skin minecraft.Skin) (image.Image, error) {
		return GetBust(skin)
	})

	r := mux.NewRouter()
	r.NotFoundHandler = NotFoundHandler{}

	r.HandleFunc("/avatar/{username:"+minecraft.ValidUsernameRegex+"}{extension:(.png)?}", avatarPage)
	r.HandleFunc("/avatar/{username:"+minecraft.ValidUsernameRegex+"}/{size:[0-9]+}{extension:(.png)?}", avatarPage)

	r.HandleFunc("/helm/{username:"+minecraft.ValidUsernameRegex+"}{extension:(.png)?}", helmPage)
	r.HandleFunc("/helm/{username:"+minecraft.ValidUsernameRegex+"}/{size:[0-9]+}{extension:(.png)?}", helmPage)

	r.HandleFunc("/body/{username:"+minecraft.ValidUsernameRegex+"}{extension:(.png)?}", bodyPage)
	r.HandleFunc("/body/{username:"+minecraft.ValidUsernameRegex+"}/{size:[0-9]+}{extension:(.png)?}", bodyPage)

	r.HandleFunc("/bust/{username:"+minecraft.ValidUsernameRegex+"}{extension:(.png)?}", bustPage)
	r.HandleFunc("/bust/{username:"+minecraft.ValidUsernameRegex+"}/{size:[0-9]+}{extension:(.png)?}", bustPage)

	r.HandleFunc("/download/{username:"+minecraft.ValidUsernameRegex+"}{extension:(.png)?}", downloadPage)

	r.HandleFunc("/skin/{username:"+minecraft.ValidUsernameRegex+"}{extension:(.png)?}", skinPage)

	r.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s", MinotarVersion)
	})

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "https://minotar.net/")
		w.WriteHeader(302)
	})

	http.Handle("/", r)
	err = http.ListenAndServe(config.Address, nil)
	log.Critical(err.Error())
}
