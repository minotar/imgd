package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"runtime"

	"github.com/minotar/imgd/storage"
	"github.com/minotar/imgd/storage/lru"
	"github.com/minotar/imgd/storage/memory"
	"github.com/minotar/imgd/storage/radix"
	rcluster "github.com/minotar/imgd/storage/radix/cluster"
	"github.com/minotar/minecraft"

	"github.com/gorilla/mux"
	"github.com/op/go-logging"
)

// Set the default, min and max width to resize processed images to.
const (
	DefaultWidth = uint(180)
	MinWidth     = uint(8)
	MaxWidth     = uint(300)

	ImgdVersion = "3.0.1"
)

var (
	config        = &Configuration{}
	cache         map[string]storage.Storage
	mcClient      *minecraft.Minecraft
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

// Todo: Do we want to rely on config for which caches to set?
func setupCache() {

	if cache == nil {
		cache = make(map[string]storage.Storage)
	}

	for t, conf := range config.Cache {
		var err error
		switch conf.Storage {
		case "memory":
			cache[t], err = memory.New(conf.Size)
		case "lru":
			cache[t], err = lru.New(conf.Size)
		case "redis":
			cache[t], err = radix.New(radix.RedisConfig{
				Network: "tcp",
				Address: config.Redis[t].Address,
				Auth:    config.Redis[t].Auth,
				DB:      config.Redis[t].DB,
				Size:    config.Redis[t].PoolSize,
			})
		case "redis-cluster":
			cache[t], err = rcluster.New(radix.RedisConfig{
				Address: config.Redis[t].Address,
				Auth:    config.Redis[t].Auth,
				DB:      config.Redis[t].DB,
				Size:    config.Redis[t].PoolSize,
			})
		default:
			err = errors.New("No cache selected")
		}
		if err != nil {
			log.Criticalf("Unable to setup Cache. (%v)", err)
			os.Exit(1)
		}
	}
}

func setupMcClient() {
	mcClient = &minecraft.Minecraft{
		Client:    minecraft.NewHTTPClient(),
		UserAgent: config.Minecraft.UserAgent,
		UUIDAPI: minecraft.UUIDAPI{
			SessionServerURL: config.Minecraft.SessionServerURL,
			ProfileURL:       config.Minecraft.ProfileURL,
		},
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

func startProfilingServer() {
	if config.Server.ProfilerAddress != "" {
		go func() {
			profiler := mux.NewRouter()

			profiler.HandleFunc("/debug/pprof/", pprof.Index)
			profiler.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
			profiler.HandleFunc("/debug/pprof/profile", pprof.Profile)
			profiler.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
			profiler.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
			profiler.Handle("/debug/pprof/heap", pprof.Handler("heap"))
			profiler.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
			profiler.Handle("/debug/pprof/block", pprof.Handler("block"))

			log.Noticef("imgd Profliler starting on %s", config.Server.ProfilerAddress)
			err := http.ListenAndServe(config.Server.ProfilerAddress, profiler)
			if err != nil {
				log.Criticalf("imgd Profiler ListenAndServe: \"%s\"", err.Error())
				os.Exit(1)
			}
		}()
	}
}

func startServer() {
	r := Router{Mux: mux.NewRouter()}
	r.Bind()
	http.Handle("/", imgdHandler(r.Mux))
	log.Noticef("imgd %s starting on %s", ImgdVersion, config.Server.Address)
	err := http.ListenAndServe(config.Server.Address, nil)
	if err != nil {
		log.Criticalf("imgd ListenAndServe: \"%s\"", err.Error())
		os.Exit(1)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	logBackend := logging.NewLogBackend(os.Stdout, "", 0)

	signalHandler = MakeSignalHandler()
	setupCache()
	stats = MakeStatsCollector()
	setupConfig()
	setupLog(logBackend)
	startProfilingServer()
	setupCache()
	setupMcClient()
	startServer()
}
