package skind

import (
	"flag"
	"net/http"

	"github.com/felixge/fgprof"
	cache_config "github.com/minotar/imgd/pkg/cache/util/config"
	"github.com/minotar/imgd/pkg/mcclient"
	"github.com/minotar/imgd/pkg/util/log"
	"github.com/minotar/imgd/pkg/util/route_helpers"

	"github.com/weaveworks/common/server"
)

type Config struct {
	Server       server.Config   `yaml:"server,omitempty"`
	McClient     mcclient.Config `yaml:"mcclient,omitempty"`
	Logger       log.Logger
	CorsAllowAll bool
}

// RegisterFlags registers flag.
func (c *Config) RegisterFlags(f *flag.FlagSet) {
	//c.Server.ExcludeRequestInLog = true

	f.BoolVar(&c.CorsAllowAll, "skind.cors-allow-all", true, "Permissive CORS policy")
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
		skin := s.McClient.GetSkinFromReq(logger, userReq)

		logger.Infof("User hash is: %s", skin.Hash)

		// No more header changes after writing
		WriteSkin(w, skin)
		logger.Debug(w.Header())
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

	if s.Cfg.CorsAllowAll {
		serv.HTTP.Use(route_helpers.CorsHandler)
	}

	RegisterRoutes(serv.HTTP, s.SkinPageHandler())

	s.Server = serv
	return nil

}
