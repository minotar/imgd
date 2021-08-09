package skind

import (
	"flag"

	"github.com/felixge/fgprof"
	"github.com/minotar/imgd/pkg/cache"
	"github.com/minotar/imgd/pkg/cache/bolt_cache"
	"github.com/minotar/imgd/pkg/mcclient"
	"github.com/minotar/imgd/pkg/util/log"
	"github.com/minotar/minecraft"

	"github.com/weaveworks/common/server"
)

type Config struct {
	Server server.Config `yaml:"server,omitempty"`
	Logger log.Logger
}

// RegisterFlags registers flag.
func (c *Config) RegisterFlags(f *flag.FlagSet) {
	c.Server.MetricsNamespace = "skind"
	//c.Server.ExcludeRequestInLog = true

	c.Server.RegisterFlags(f)

}

type Skind struct {
	Cfg Config

	Server   *server.Server
	McClient *mcclient.McClient
}

func New(cfg Config) (*Skind, error) {

	cacheConfig := cache.CacheConfig{
		Name:   "BoltAll",
		Logger: cfg.Logger,
	}
	bc_cfg := bolt_cache.NewBoltCacheConfig(cacheConfig, "/tmp/bolt_cache_skind.db", "skind")

	bc, _ := bolt_cache.NewBoltCache(bc_cfg)

	skind := &Skind{
		Cfg: cfg,
		McClient: &mcclient.McClient{
			API: minecraft.NewMinecraft(),
		},
	}

	skind.McClient.Caches.UUID = bc
	skind.McClient.Caches.UserData = bc
	skind.McClient.Caches.Textures = bc

	return skind, nil
}

func (s *Skind) Run() error {
	//t.Server.HTTP.Handle("/services", http.HandlerFunc(t.servicesHandler))
	if err := s.initServer(); err != nil {
		return err
	}
	// init other bits

	s.Server.HTTP.Path("/debug/fgprof").Handler(fgprof.Handler())

	s.Server.HTTP.Path("/skin/{uuid:" + minecraft.ValidUUIDPlainRegex + "}").Handler(SkinPageHandler(s))
	s.Server.HTTP.Path("/skin/{username:" + minecraft.ValidUsernameRegex + "}").Handler(SkinPageHandler(s))

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
