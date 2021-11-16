package mcclient

import (
	"flag"
	"net/http"
	"time"

	"github.com/minotar/imgd/pkg/cache/util/config"
	"github.com/minotar/imgd/pkg/minecraft"
	"github.com/minotar/imgd/pkg/minecraft/minecraft_trace"
	"github.com/prometheus/client_golang/prometheus"
)

type Config struct {
	UpstreamTimeout  time.Duration  `yaml:"upstream_timeout"`
	UserAgent        string         `yaml:"useragent"`
	SessionServerURL string         `yaml:"sessionserver_url"`
	ProfileURL       string         `yaml:"profile_url"`
	CacheUUID        *config.Config `yaml:"cache_uuid"`
	CacheUserData    *config.Config `yaml:"cache_userdata"`
	CacheTextures    *config.Config `yaml:"cache_textures"`
}

// RegisterFlags registers flag.
func (c *Config) RegisterFlags(f *flag.FlagSet) {

	c.CacheUUID = &config.Config{}
	c.CacheUserData = &config.Config{}
	c.CacheTextures = &config.Config{}

	f.DurationVar(&c.UpstreamTimeout, "mcclient.upstream-timeout", 10*time.Second, "Timeout for Minecraft API Client")
	f.StringVar(&c.UserAgent, "mcclient.useragent", "minotar/imgd (https://github.com/minotar/imgd) - default", "UserAgent for Minecraft API Client")
	f.StringVar(&c.SessionServerURL, "mcclient.sessionserver-url", "https://sessionserver.mojang.com/session/minecraft/profile/", "API for UUID -> Texture Properties")
	f.StringVar(&c.ProfileURL, "mcclient.profile-url", "https://api.mojang.com/users/profiles/minecraft/", "API for Username -> UUID lookups")
	c.CacheUUID.RegisterFlags(f, "UUID")
	c.CacheUserData.RegisterFlags(f, "UserData")
	c.CacheTextures.RegisterFlags(f, "Textures")

}

func NewMcClient(cfg *Config) *McClient {

	minecraftCfg := minecraft.Config{
		UUIDAPIConfig: minecraft.UUIDAPIConfig{
			SessionServerURL: cfg.SessionServerURL,
			ProfileURL:       cfg.ProfileURL,
		},
		UserAgent:      cfg.UserAgent,
		RequestTimeout: cfg.UpstreamTimeout,
	}

	mc := &minecraft.Minecraft{
		Client: &http.Client{
			Timeout: minecraftCfg.RequestTimeout,
			// Transport is set below
		},
		Cfg: minecraftCfg,
	}

	// Todo: put extra traces behind feature flag?

	trace := &minecraft_trace.InstrumentTrace{
		ConnectStart:         apiClientTraceDuration.MustCurryWith(prometheus.Labels{"event": "connectStart"}),
		ConnectDone:          apiClientTraceDuration.MustCurryWith(prometheus.Labels{"event": "connectDone"}),
		GotFirstResponseByte: apiClientTraceDuration.MustCurryWith(prometheus.Labels{"event": "timeToFirstByte"}),
	}

	mc.Client.Transport = minecraft_trace.InstrumentRoundTripperInFlight(apiClientInflight,
		minecraft_trace.InstrumentRoundTripperTrace(trace,
			minecraft_trace.InstrumentRoundTripperDuration(apiClientDuration, http.DefaultTransport),
		),
	)

	return &McClient{
		API: mc,
	}
}
