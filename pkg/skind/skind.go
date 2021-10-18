package skind

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/felixge/fgprof"
	cache_config "github.com/minotar/imgd/pkg/cache/util/config"
	"github.com/minotar/imgd/pkg/mcclient"
	"github.com/minotar/imgd/pkg/util/log"
	"github.com/minotar/imgd/pkg/util/route_helpers"

	"github.com/weaveworks/common/server"
)

type Config struct {
	Server          server.Config   `yaml:"server,omitempty"`
	McClient        mcclient.Config `yaml:"mcclient,omitempty"`
	Logger          log.Logger
	CorsAllowAll    bool
	UseETags        bool
	CacheControlTTL time.Duration
}

// RegisterFlags registers flag.
func (c *Config) RegisterFlags(f *flag.FlagSet) {
	//c.Server.ExcludeRequestInLog = true

	f.BoolVar(&c.CorsAllowAll, "skind.cors-allow-all", true, "Permissive CORS policy")
	f.BoolVar(&c.UseETags, "skind.use-etags", true, "Use etags to skip re-processing")
	f.DurationVar(&c.CacheControlTTL, "skind.cache-control-ttl", time.Duration(6)*time.Hour, "Cache TTL returned to clients")

	c.Server.RegisterFlags(f)
	c.McClient.RegisterFlags(f)

}

type Skind struct {
	Cfg Config

	Server   *server.Server
	McClient *mcclient.McClient
}

func New(cfg Config) (*Skind, error) {
	// Set namespace for all metrics
	cfg.Server.MetricsNamespace = "skind"
	// Set the GRPC to localhost only
	cfg.Server.GRPCListenAddress = "127.0.0.2"

	cfg.McClient.CacheUUID.Logger = cfg.Logger
	cacheUUID, err := cache_config.NewCache(cfg.McClient.CacheUUID)
	if err != nil {
		cfg.Logger.Panicf("Unable to create cache UUID: %v", err)
	}
	cacheUUID.Start()

	cfg.McClient.CacheUserData.Logger = cfg.Logger
	cacheUserData, _ := cache_config.NewCache(cfg.McClient.CacheUserData)
	if err != nil {
		cfg.Logger.Panicf("Unable to create cache UserData: %v", err)
	}
	cacheUserData.Start()

	cfg.McClient.CacheTextures.Logger = cfg.Logger
	cacheTextures, _ := cache_config.NewCache(cfg.McClient.CacheTextures)
	if err != nil {
		cfg.Logger.Panicf("Unable to create cache Textures: %v", err)
	}
	cacheTextures.Start()

	skind := &Skind{
		Cfg:      cfg,
		McClient: mcclient.NewMcClient(&cfg.McClient),
	}

	skind.McClient.Caches.UUID = cacheUUID
	skind.McClient.Caches.UserData = cacheUserData
	skind.McClient.Caches.Textures = cacheTextures

	return skind, nil
}

// Requires "uuid" or "username" vars
func (s *Skind) SkinPageHandler() http.Handler {
	logger := s.Cfg.Logger

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userReq := route_helpers.MuxToUserReq(r)
		logger, skinIO := s.McClient.GetSkinBufferFromReq(logger, userReq)
		defer skinIO.Close()

		logger.Infof("Texture ID is: %s", skinIO.TextureID)

		reqETag := r.Header.Get("If-None-Match")
		if s.Cfg.UseETags {
			// If the response was a StatusNotModified (it should be as we already sent the If-None-Match!)
			// If the ETag matches from request to response, then no need to process
			if reqETag == skinIO.TextureID {
				w.WriteHeader(http.StatusNotModified)
				return
			}
			// Need to unset ETag/cache if we later have an issue!
			w.Header().Set("ETag", skinIO.TextureID)
			w.Header().Add("Cache-Control", fmt.Sprintf("public, max-age=%d", int(s.Cfg.CacheControlTTL.Seconds())))
		}

		w.Header().Add("Content-Type", "image/png")
		// No more header changes after writing
		io.Copy(w, skinIO)
	})
}

func (s *Skind) Run() error {
	//t.Server.HTTP.Handle("/services", http.HandlerFunc(t.servicesHandler))
	if err := s.initServer(); err != nil {
		return err
	}
	// init other bits

	return s.Server.Run()

	//return nil
}

func (s *Skind) initServer() error {
	serv, err := server.New(s.Cfg.Server)
	if err != nil {
		return err
	}

	serv.HTTP.Use(route_helpers.LoggingMiddleware(s.Cfg.Logger))

	serv.HTTP.Path("/debug/fgprof").Handler(fgprof.Handler())
	serv.HTTP.Path("/healthcheck").Handler(HealthcheckHandler(s.McClient))
	serv.HTTP.Path("/dbsize").Handler(SizecheckHandler(s.McClient))

	if s.Cfg.CorsAllowAll {
		serv.HTTP.Use(route_helpers.CorsHandler)
	}

	RegisterRoutes(serv.HTTP, s.SkinPageHandler())

	s.Server = serv
	return nil

}
