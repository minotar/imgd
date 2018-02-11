package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/gorilla/mux"
	"github.com/op/go-logging"
)

// Set the default, min and max width to resize processed images to.
const (
	DefaultWidth = uint(180)
	MinWidth     = uint(8)
	MaxWidth     = uint(300)

	ImgdVersion = "2.9.4"
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
		log.Criticalf("Unable to setup Cache. (%v)", err)
		os.Exit(1)
	}
}

func setupLog(logBackend *logging.LogBackend) {
	logging.SetBackend(logBackend)
	logging.SetFormatter(logging.MustStringFormatter(format))
	logLevel, err := logging.LogLevel(config.Server.Logging)
	logging.SetLevel(logLevel, "")
	if err != nil {
		log.Errorf("Invalid log type: %s", config.Server.Logging)
		// If error it sets the logging to ERROR, let's change it to INFO
		logging.SetLevel(4, "")
	}
	log.Noticef("Log level set to %s", logging.GetLevel(""))
}

func startServer() {
	r := Router{Mux: mux.NewRouter()}
	r.Bind()
	http.Handle("/", imgdHandler(r.Mux))
	log.Noticef("imgd %s starting on %s", ImgdVersion, config.Server.Address)
	err := http.ListenAndServe(config.Server.Address, nil)
	if err != nil {
		log.Criticalf("ListenAndServe: \"%s\"", err.Error())
		os.Exit(1)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	logBackend := logging.NewLogBackend(os.Stdout, "", 0)

	signalHandler = MakeSignalHandler()
	stats = MakeStatsCollector()
	setupConfig()
	setupLog(logBackend)
	setupCache()
	startServer()
}
