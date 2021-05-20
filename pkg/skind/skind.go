package skind

import (
	"flag"

	"github.com/felixge/fgprof"
	"github.com/minotar/minecraft"

	"github.com/weaveworks/common/server"
)

type Config struct {
	Server server.Config `yaml:"server,omitempty"`
}

// RegisterFlags registers flag.
func (c *Config) RegisterFlags(f *flag.FlagSet) {
	c.Server.MetricsNamespace = "skind"
	//c.Server.ExcludeRequestInLog = true

	c.Server.RegisterFlags(f)

}

type Skind struct {
	Cfg Config

	Server  *server.Server
	Storage map[int]map[string]string
}

func New(cfg Config) (*Skind, error) {
	skind := &Skind{
		Cfg: cfg,
		Storage: map[int]map[string]string{
			1: map[string]string{
				"lukehandle": "5c115ca73efd41178213a0aff8ef11e0",
			},
		},
	}

	return skind, nil
}

func (s *Skind) Run() error {
	//t.Server.HTTP.Handle("/services", http.HandlerFunc(t.servicesHandler))
	if err := s.initServer(); err != nil {
		return err
	}
	// init other bits

	s.Server.HTTP.Path("/debug/fgprof").Handler(fgprof.Handler())

	s.Server.HTTP.Path("/skin/{uuid:" + minecraft.ValidUUIDPlainRegex + "}").Handler(SkinPageHandler(s.Storage))
	s.Server.HTTP.Path("/skin/{username:" + minecraft.ValidUsernameRegex + "}").Handler(SkinPageHandler(s.Storage))

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
