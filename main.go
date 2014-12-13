package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/minotar/minecraft"
	"github.com/op/go-logging"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
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

	MinotarVersion = "2.6"
)

var (
	config = &Configuration{}
	cache  Cache
	stats  *StatusCollector
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

func fetchImageProcessThen(callback func(*mcSkin, int) error) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		timeReqStart := time.Now()

		vars := mux.Vars(r)

		username := vars["username"]
		size := rationalizeSize(vars["size"])
		ok := true

		skin := fetchSkin(username)
		var err error

		timeFetch := time.Now()

		err = callback(skin, int(size))
		if err != nil {
			serverErrorPage(w, r)
			return
		}
		timeProcess := time.Now()

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

		timing := fmt.Sprintf("%d+%d=%dms", timeBetween(timeReqStart, timeFetch), timeBetween(timeFetch, timeProcess), timeBetween(timeReqStart, timeProcess))

		w.Header().Add("X-Timing", timing)
		addCacheTimeoutHeader(w, timeout)
		skin.WritePNG(w)

		if skin.Source == "" {
			skin.Source = "Cached"
		}

		log.Info("Serving " + skin.Source + " skin for " + username + " (" + timing + ") md5: " + skin.Hash)
	}
}

func skinPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	username := vars["username"]

	skin := fetchSkin(username)

	w.Header().Add("Content-Type", "image/png")
	w.Header().Add("X-Requested", "skin")
	w.Header().Add("X-Result", "ok")

	skin.WriteSkin(w)
}

func downloadPage(w http.ResponseWriter, r *http.Request) {
	headers := w.Header()
	headers.Add("Content-Disposition", "attachment; filename=\"skin.png\"")
	skinPage(w, r)
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

var log = logging.MustGetLogger("imgd")
var format = "[%{time:15:04:05.000000}] %{level:.4s} %{message}"

func setupConfig() {
	err := config.load()
	if err != nil {
		fmt.Printf("Error loading config: %s\n", err)
		return
	}
}

func setupCache() {
	cache = MakeCache(config.Server.Cache)
	cache.setup()
}

func setupLog(logBackend *logging.LogBackend) {
	logging.SetBackend(logBackend)
	logging.SetFormatter(logging.MustStringFormatter(format))
}

func main() {
	logBackend := logging.NewLogBackend(os.Stdout, "", 0)
	stats = MakeStatsCollector()
	setupConfig()
	setupLog(logBackend)
	setupCache()

	debug.SetGCPercent(10)

	avatarPage := fetchImageProcessThen(func(skin *mcSkin, width int) error {
		stats.Served("avatar")
		return skin.GetHead(width)
	})
	helmPage := fetchImageProcessThen(func(skin *mcSkin, width int) error {
		stats.Served("helm")
		return skin.GetHelm(width)
	})
	bodyPage := fetchImageProcessThen(func(skin *mcSkin, width int) error {
		stats.Served("body")
		return skin.GetBody(width)
	})
	bustPage := fetchImageProcessThen(func(skin *mcSkin, width int) error {
		stats.Served("bust")
		return skin.GetBust(width)
	})
	cubePage := fetchImageProcessThen(func(skin *mcSkin, width int) error {
		stats.Served("cube")
		return skin.GetCube(width)
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

	r.HandleFunc("/cube/{username:"+minecraft.ValidUsernameRegex+"}{extension:(.png)?}", cubePage)
	r.HandleFunc("/cube/{username:"+minecraft.ValidUsernameRegex+"}/{size:[0-9]+}{extension:(.png)?}", cubePage)

	r.HandleFunc("/download/{username:"+minecraft.ValidUsernameRegex+"}{extension:(.png)?}", downloadPage)

	r.HandleFunc("/skin/{username:"+minecraft.ValidUsernameRegex+"}{extension:(.png)?}", skinPage)

	r.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s", MinotarVersion)
	})

	r.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(stats.ToJSON())
	})

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://minotar.net/", 302)
	})

	http.Handle("/", r)
	err := http.ListenAndServe(config.Server.Address, nil)
	log.Critical(err.Error())
}
