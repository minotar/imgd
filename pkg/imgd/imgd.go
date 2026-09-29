package imgd

import (
	"flag"
	"fmt"
	"sync"
	"time"

	"github.com/minotar/imgd/pkg/cache"
	cache_config "github.com/minotar/imgd/pkg/cache/util/config"
	"github.com/minotar/imgd/pkg/mcclient"
	"github.com/minotar/imgd/pkg/render"
	"github.com/minotar/imgd/pkg/skin"
	"github.com/minotar/imgd/pkg/util/log"
)

type Config struct {
	Server   ServerConfig    `yaml:"server,omitempty"`
	McClient mcclient.Config `yaml:"mcclient,omitempty"`
	Logger   log.Logger
	// Add open CORS headers to each response
	CorsAllowAll bool
	// Return an ETag based on the texture ID
	UseETags bool
	// Return a 302 redirect for Username requests to their related UUID
	RedirectUsername bool
	CacheControlTTL  time.Duration
}

// RegisterFlags registers flag.
func (c *Config) RegisterFlags(f *flag.FlagSet) {
	f.BoolVar(&c.CorsAllowAll, "imgd.cors-allow-all", true, "Permissive CORS policy")
	f.BoolVar(&c.UseETags, "imgd.use-etags", true, "Use etags to skip re-processing")
	f.BoolVar(&c.RedirectUsername, "imgd.redirect-username", true, "Redirect username requests to the UUID variant")
	f.DurationVar(&c.CacheControlTTL, "imgd.cache-control-ttl", time.Duration(6)*time.Hour, "Cache TTL returned to clients")

	c.Server.RegisterFlags(f)
	c.McClient.RegisterFlags(f)

}

type Imgd struct {
	Cfg Config

	closeOnce     sync.Once
	caches        []cache.Cache
	McClient      *mcclient.McClient
	ProcessRoutes map[string]skin.SkinProcessor
}

func New(cfg Config) (*Imgd, error) {
	if cfg.Logger == nil {
		cfg.Logger = log.NewBuiltinLogger(1)
	}
	var caches []cache.Cache
	for index, config := range []*cache_config.Config{cfg.McClient.CacheUUID, cfg.McClient.CacheUserData, cfg.McClient.CacheTextures} {
		var c cache.Cache
		var err error
		if config == nil {
			err = fmt.Errorf("missing cache configuration")
		} else {
			config.Logger = cfg.Logger
			c, err = cache_config.NewCache(config)
			if err == nil && c == nil && index < 2 {
				err = fmt.Errorf("UUID and user data caches cannot be disabled")
			}
		}
		if err != nil {
			for _, opened := range caches {
				if opened != nil {
					opened.Close()
				}
			}
			return nil, fmt.Errorf("initialize cache %d: %w", index, err)
		}
		if c != nil {
			c.Start()
		}
		caches = append(caches, c)
	}

	imgd := &Imgd{
		Cfg:           cfg,
		caches:        caches,
		McClient:      mcclient.NewMcClient(&cfg.McClient),
		ProcessRoutes: render.DefaultProcessRoutes,
	}

	imgd.McClient.Caches.UUID = caches[0]
	imgd.McClient.Caches.UserData = caches[1]
	imgd.McClient.Caches.Textures = caches[2]

	return imgd, nil
}

// Close releases the caches owned by the application.
func (i *Imgd) Close() {
	i.closeOnce.Do(func() {
		for _, c := range i.caches {
			if c != nil {
				c.Close()
			}
		}
	})
}
