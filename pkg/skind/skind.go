package skind

import (
	"flag"

	"github.com/felixge/fgprof"
	cache_config "github.com/minotar/imgd/pkg/cache/util/config"
	"github.com/minotar/imgd/pkg/mcclient"
	"github.com/minotar/imgd/pkg/util/log"
	"github.com/minotar/imgd/pkg/util/route_helpers"
	"github.com/minotar/minecraft"

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

func (s *Skind) Run() error {
	//t.Server.HTTP.Handle("/services", http.HandlerFunc(t.servicesHandler))
	if err := s.initServer(); err != nil {
		return err
	}
	// init other bits

	s.Server.HTTP.Path("/debug/fgprof").Handler(fgprof.Handler())

	skinHandler := SkinPageHandler(s)
	downloadSkinHandler := route_helpers.BrowserDownloadHandler(skinHandler)
	dashedRedirectHandler := route_helpers.DashedRedirectUUIDHandler()

	if s.Cfg.CorsAllowAll {
		skinHandler = route_helpers.CorsHandler(skinHandler)
		downloadSkinHandler = route_helpers.CorsHandler(downloadSkinHandler)
		dashedRedirectHandler = route_helpers.CorsHandler(dashedRedirectHandler)
	}

	s.Server.HTTP.Path("/skin/{uuid:" + minecraft.ValidUUIDPlainRegex + "}").Handler(skinHandler).Name("skinUUID")
	s.Server.HTTP.Path("/skin/{username:" + minecraft.ValidUsernameRegex + "}").Handler(skinHandler).Name("skinUsername")
	s.Server.HTTP.Path("/skin/{uuid:" + minecraft.ValidUUIDDashRegex + "}{extension:(?:.png)?}").Handler(dashedRedirectHandler).Name("skinDashedRedirect")

	s.Server.HTTP.Path("/download/{uuid:" + minecraft.ValidUUIDPlainRegex + "}").Handler(downloadSkinHandler).Name("downloadUUID")
	s.Server.HTTP.Path("/download/{username:" + minecraft.ValidUsernameRegex + "}").Handler(downloadSkinHandler).Name("downloadUsername")
	s.Server.HTTP.Path("/download/{uuid:" + minecraft.ValidUUIDDashRegex + "}{extension:(?:.png)?}").Handler(dashedRedirectHandler).Name("downloadDashedRedirect")

	return s.Server.Run()

	//return nil
}

func (s *Skind) initServer() error {
	serv, err := server.New(s.Cfg.Server)
	if err != nil {
		return err
	}

	s.Server = serv
	return nil

}
