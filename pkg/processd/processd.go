package processd

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/felixge/fgprof"
	"github.com/gorilla/mux"
	"github.com/minotar/imgd/pkg/util/log"
	"github.com/minotar/minecraft"

	"github.com/weaveworks/common/server"
)

type Config struct {
	Server          server.Config `yaml:"server,omitempty"`
	UpstreamTimeout time.Duration `yaml:"upstream_timeout"`
	SkindURL        string        `yaml:"skind_url,omitempty"`
	Logger          log.Logger
	CorsAllowAll    bool
	UseETags        bool
}

// RegisterFlags registers flag.
func (c *Config) RegisterFlags(f *flag.FlagSet) {
	//c.Server.ExcludeRequestInLog = true

	c.Server.RegisterFlags(f)

	f.DurationVar(&c.UpstreamTimeout, "processd.upstream-timeout", 10*time.Second, "Timeout for Skin lookup")
	f.StringVar(&c.SkindURL, "processd.skind-url", "http://localhost:4643/skin/", "API for skin lookups")
	f.BoolVar(&c.CorsAllowAll, "processd.cors-allow-all", true, "Permissive CORS policy")
	f.BoolVar(&c.UseETags, "processd.use-etags", true, "Use etags to skip re-processing")

}

type Processd struct {
	Cfg Config

	Server    *server.Server
	Client    *http.Client
	UserAgent string
	SkindURL  string
}

func New(cfg Config) (*Processd, error) {
	// Set namespace for all metrics
	cfg.Server.MetricsNamespace = "processd"
	// Set the GRPC to localhost only
	cfg.Server.GRPCListenAddress = "127.0.0.3"

	processd := &Processd{
		Cfg:      cfg,
		SkindURL: cfg.SkindURL,
		Client: &http.Client{
			Timeout: cfg.UpstreamTimeout,
		},
		UserAgent: "minotar/imgd/processd (https://github.com/minotar/imgd) - default",
	}

	return processd, nil
}

// need some skin lookup wrapper

func handleError(w http.ResponseWriter, r *http.Request) {

}

func (s *Processd) SkinLookupWrapper(processFunc func(minecraft.Skin) http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		var userLookup string
		if user, userGiven := vars["user"]; userGiven {
			userLookup = user
		}

		skinReq, err := http.NewRequestWithContext(r.Context(), "GET", fmt.Sprint(s.SkindURL, userLookup), nil)
		if err != nil {
			//return nil, fmt.Errorf("unable to create request: %v", err)
			//Use Steve and call original process logic?
			return
		}

		reqETag := r.Header.Get("If-None-Match")

		skinReq.Header.Set("User-Agent", s.UserAgent)
		if s.Cfg.UseETags {
			skinReq.Header.Set("If-None-Match", reqETag)
		}
		//req.Header.Set("X-Request-ID", "magic-to-use-existing-or-add-new")

		resp, err := s.Client.Do(skinReq)
		if err != nil {
			//return nil, fmt.Errorf("unable to GET URL: %v", err)
			//Use Steve and call original process logic?
			return
		}
		defer resp.Body.Close()

		if s.Cfg.UseETags {
			// If the response was a StatusNotModified (it should be as we already sent the If-None-Match!)
			// If the ETag matches from request to response, then no need to process
			if resp.StatusCode == http.StatusNotModified || reqETag == resp.Header.Get("ETag") {
				w.WriteHeader(http.StatusNotModified)
				return
			}
			// Todo: Can we add the etag to the responseWriter here?
		}

		skin := minecraft.Skin{}
		skin.Decode(resp.Body)

		// Up to this point, the processing could be metric'd "generically" and the type of processing was irrelevant

		handler := processFunc(skin)
		handler.ServeHTTP(w, r)
	})
}

func (s *Processd) Run() error {
	//t.Server.HTTP.Handle("/services", http.HandlerFunc(t.servicesHandler))
	if err := s.initServer(); err != nil {
		return err
	}
	// init other bits

	s.Server.HTTP.Path("/debug/fgprof").Handler(fgprof.Handler())

	RegisterRoutes(s.Server.HTTP, s.SkinLookupWrapper)

	return s.Server.Run()

	//return nil
}

func (s *Processd) initServer() error {
	serv, err := server.New(s.Cfg.Server)
	if err != nil {
		return err
	}

	s.Server = serv
	return nil

}
