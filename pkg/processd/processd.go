package processd

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/minotar/imgd/pkg/mcclient/mcuser"
	"github.com/minotar/imgd/pkg/processd/mcskin"
	"github.com/minotar/imgd/pkg/skind"
	"github.com/minotar/imgd/pkg/util/log"
	"github.com/minotar/imgd/pkg/util/route_helpers"

	"github.com/weaveworks/common/server"
)

var (
	DefaultProcessRoutes = map[string]skind.SkinProcessor{
		"Avatar":                 mcskin.HandlerHead,
		"Helm":                   mcskin.HandlerHelm,
		"Cube":                   mcskin.HandlerCube,
		"CubeHelm":               mcskin.HandlerCubeHelm,
		"Bust":                   mcskin.HandlerBust,
		"Body":                   mcskin.HandlerBody,
		"Armor/Bust|Armour/Bust": mcskin.HandlerArmorBust,
		"Armor/Body|Armour/Body": mcskin.HandlerArmorBody,
	}
)

type Config struct {
	Server          server.Config `yaml:"server,omitempty"`
	UpstreamTimeout time.Duration `yaml:"upstream_timeout"`
	SkindURL        string        `yaml:"skind_url,omitempty"`
	Logger          log.Logger
	CorsAllowAll    bool
	UseETags        bool
	CacheControlTTL time.Duration
}

// RegisterFlags registers flag.
func (c *Config) RegisterFlags(f *flag.FlagSet) {
	//c.Server.ExcludeRequestInLog = true

	f.DurationVar(&c.UpstreamTimeout, "processd.upstream-timeout", 10*time.Second, "Timeout for Skin lookup")
	f.StringVar(&c.SkindURL, "processd.skind-url", "http://localhost:4643/skin/", "API for skin lookups")
	f.BoolVar(&c.CorsAllowAll, "processd.cors-allow-all", true, "Permissive CORS policy")
	f.BoolVar(&c.UseETags, "processd.use-etags", true, "Use etags to skip re-processing")
	f.DurationVar(&c.CacheControlTTL, "processd.cache-control-ttl", time.Duration(6)*time.Hour, "Cache TTL returned to clients")

	c.Server.RegisterFlags(f)
}

type Processd struct {
	Cfg Config

	Server        *server.Server
	Client        *http.Client
	UserAgent     string
	SkindURL      string
	ProcessRoutes map[string]skind.SkinProcessor
}

func New(cfg Config) (*Processd, error) {
	// Set namespace for all metrics
	cfg.Server.MetricsNamespace = "processd"
	// Set the GRPC to localhost only
	cfg.Server.GRPCListenAddress = "127.0.0.3"

	processd := &Processd{
		Cfg: cfg,
		Client: &http.Client{
			Timeout: cfg.UpstreamTimeout,
		},
		UserAgent:     "minotar/imgd/processd (https://github.com/minotar/imgd) - default",
		SkindURL:      cfg.SkindURL,
		ProcessRoutes: DefaultProcessRoutes,
	}

	return processd, nil
}

// need some skin lookup wrapper

func handleSkinLookupError(w http.ResponseWriter, r *http.Request, logger log.Logger, processFunc skind.SkinProcessor) {
	skinIO := mcuser.GetSteveTextureIO()

	handler := processFunc(logger, skinIO)
	handler.ServeHTTP(w, r)
}

func (p *Processd) SkinLookupWrapper(processFunc skind.SkinProcessor) http.HandlerFunc {
	logger := p.Cfg.Logger

	return func(w http.ResponseWriter, r *http.Request) {

		userReq := route_helpers.MuxToUserReq(r)
		var userLookup string

		if userReq.UUID != "" {
			userLookup = userReq.UUID
		} else if userReq.Username != "" {
			userLookup = userReq.Username
		} else {
			logger.Errorf("Request came through without Username/UUID: %v", mux.Vars(r))
			handleSkinLookupError(w, r, logger, processFunc)
			return
		}

		skinReq, err := http.NewRequestWithContext(r.Context(), "GET", fmt.Sprint(p.SkindURL, userLookup), nil)
		if err != nil {
			//return nil, fmt.Errorf("unable to create request: %v", err)
			//Use Steve and call original process logic?
			logger.Errorf("It broken: %v", err)
			handleSkinLookupError(w, r, logger, processFunc)
			return
		}

		reqETag := r.Header.Get("If-None-Match")

		skinReq.Header.Set("User-Agent", p.UserAgent)
		if p.Cfg.UseETags {
			skinReq.Header.Set("If-None-Match", reqETag)
		}
		//req.Header.Set("X-Request-ID", "magic-to-use-existing-or-add-new")

		resp, err := p.Client.Do(skinReq)
		if err != nil {
			//return nil, fmt.Errorf("unable to GET URL: %v", err)
			//Use Steve and call original process logic?
			logger.Errorf("It broken: %v", err)
			handleSkinLookupError(w, r, logger, processFunc)
			return
		}
		// The processFunc *MUST* close the resp.Body via the TextureIO object
		//defer resp.Body.Close()

		w.Header().Add("Cache-Control", fmt.Sprintf("public, max-age=%d", int(p.Cfg.CacheControlTTL.Seconds())))

		if p.Cfg.UseETags {
			respETag := resp.Header.Get("ETag")
			// ETag is always included (even for 304 responses)
			w.Header().Set("ETag", respETag)

			// If the response was a StatusNotModified (it should be as we already sent the If-None-Match!)
			// If the ETag matches from request to response, then no need to process
			if resp.StatusCode == http.StatusNotModified || reqETag == respETag {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}

		skinIO := mcuser.TextureIO{
			ReadCloser: resp.Body,
		}

		// Up to this point, the processing could be metric'd "generically" and the type of processing was irrelevant
		handler := processFunc(logger, skinIO)
		handler.ServeHTTP(w, r)
	}
}

func (p *Processd) Run() error {
	//t.Server.HTTP.Handle("/services", http.HandlerFunc(t.servicesHandler))
	if err := p.initServer(); err != nil {
		return err
	}
	// init other bits

	return p.Server.Run()

	//return nil
}

func (p *Processd) initServer() error {
	serv, err := server.New(p.Cfg.Server)
	if err != nil {
		return err
	}

	serv.HTTP.Use(route_helpers.LoggingMiddleware(p.Cfg.Logger))

	if p.Cfg.CorsAllowAll {
		serv.HTTP.Use(route_helpers.CorsHandler)
	}

	p.Server = serv
	p.routes()
	return nil
}
