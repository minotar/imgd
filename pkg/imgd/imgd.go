package imgd

import (
	"flag"
	"time"

	cache_config "github.com/minotar/imgd/pkg/cache/util/config"
	"github.com/minotar/imgd/pkg/mcclient"
	"github.com/minotar/imgd/pkg/processd"
	"github.com/minotar/imgd/pkg/skind"
	"github.com/minotar/imgd/pkg/util/log"
	"github.com/minotar/imgd/pkg/util/route_helpers"

	"github.com/weaveworks/common/server"
)

type Config struct {
	Server   server.Config   `yaml:"server,omitempty"`
	McClient mcclient.Config `yaml:"mcclient,omitempty"`
	Logger   log.Logger
	// Add open CORS headers to easch response
	CorsAllowAll bool
	// Return an ETag based on the texture ID
	UseETags bool
	// Return a 302 redirect for Username requests to their related UUID
	RedirectUsername bool
	CacheControlTTL  time.Duration
}

// RegisterFlags registers flag.
func (c *Config) RegisterFlags(f *flag.FlagSet) {
	//c.Server.ExcludeRequestInLog = true

	f.BoolVar(&c.CorsAllowAll, "imgd.cors-allow-all", true, "Permissive CORS policy")
	f.BoolVar(&c.UseETags, "imgd.use-etags", true, "Use etags to skip re-processing")
	f.BoolVar(&c.RedirectUsername, "imgd.redirect-username", true, "Redirect username requests to the UUID variant")
	f.DurationVar(&c.CacheControlTTL, "imgd.cache-control-ttl", time.Duration(6)*time.Hour, "Cache TTL returned to clients")

	c.Server.RegisterFlags(f)
	c.McClient.RegisterFlags(f)

}

type Imgd struct {
	Cfg Config

	Server        *server.Server
	McClient      *mcclient.McClient
	ProcessRoutes map[string]skind.SkinProcessor
}

func New(cfg Config) (*Imgd, error) {
	// Set namespace for all metrics
	cfg.Server.MetricsNamespace = "imgd"
	// Set the GRPC to localhost only
	cfg.Server.GRPCListenAddress = "127.0.0.4"

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

	imgd := &Imgd{
		Cfg:           cfg,
		McClient:      mcclient.NewMcClient(&cfg.McClient),
		ProcessRoutes: processd.DefaultProcessRoutes,
	}

	imgd.McClient.Caches.UUID = cacheUUID
	imgd.McClient.Caches.UserData = cacheUserData
	imgd.McClient.Caches.Textures = cacheTextures

	return imgd, nil
}

func (i *Imgd) Run() error {
	//t.Server.HTTP.Handle("/services", http.HandlerFunc(t.servicesHandler))
	if err := i.initServer(); err != nil {
		return err
	}
	// init other bits

	return i.Server.Run()

	//return nil
}

func (i *Imgd) initServer() error {
	serv, err := server.New(i.Cfg.Server)
	if err != nil {
		return err
	}

	//serv.HTTP.Use(route_helpers.LoggingMiddleware(i.Cfg.Logger))

	if i.Cfg.CorsAllowAll {
		serv.HTTP.Use(route_helpers.CorsHandler)
	}

	i.Server = serv
	i.routes()
	return nil

}
