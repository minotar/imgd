package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/gorilla/mux"
	"github.com/op/go-logging"
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

	ImgdVersion = "2.7"
)

var (
	config        = &Configuration{}
	cache         Cache
	stats         *StatusCollector
	signalHandler *SignalHandler
)

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
	err := cache.setup()
	if err != nil {
		log.Critical("Unable to setup Cache. (" + fmt.Sprintf("%v", err) + ")")
		os.Exit(1)
	}
}

func setupLog(logBackend *logging.LogBackend) {
	logging.SetBackend(logBackend)
	logging.SetFormatter(logging.MustStringFormatter(format))
}

func startServer() {
	r := Router{Mux: mux.NewRouter()}
	r.Bind()
	http.Handle("/", r.Mux)
	log.Info("imgd %s starting on %s", ImgdVersion, config.Server.Address)
	err := http.ListenAndServe(config.Server.Address, nil)
	if err != nil {
		log.Critical("ListenAndServe: \"%s\"", err.Error())
		os.Exit(1)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	logBackend := logging.NewLogBackend(os.Stdout, "", 0)

	signalHandler = MakeSignalHandler()
	stats = MakeStatsCollector()
	setupLog(logBackend)
	setupConfig()
	setupCache()
	startServer()
}
