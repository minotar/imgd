package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/prometheus/common/version"

	_ "github.com/minotar/imgd/pkg/build"
	"github.com/minotar/imgd/pkg/processd"
	"github.com/minotar/imgd/pkg/util/cfg"
	"github.com/minotar/imgd/pkg/util/log"
)

func init() {
	prometheus.MustRegister(version.NewCollector("processd"))
}

type Config struct {
	processd.Config `yaml:",inline"`
	printVersion    bool
}

func (c *Config) RegisterFlags(f *flag.FlagSet) {
	f.BoolVar(&c.printVersion, "version", false, "Print this builds version information")
	c.Config.RegisterFlags(f)
}

func main() {

	mainLogger, _ := zap.NewDevelopment()
	defer mainLogger.Sync() // flushes buffer, if any
	logger := log.NewZapLogger(mainLogger)

	var config Config
	cfg.Parse(&config, "PROCESSD")
	logger.Infof("Config: %+v\n", config)

	config.Logger = logger
	//config.Config.CorsAllowAll = true

	switch {
	case config.printVersion:
		fmt.Println(version.Print("processd"))
		os.Exit(0)
	}

	// Start processd
	s, err := processd.New(config.Config)
	if err != nil {
		logger.Errorf("Error initialising processd: %v", err)
	}

	logger.Infof("Starting processd %s", version.Info())

	err = s.Run()

	if err != nil {
		logger.Errorf("Error running processd : %v", err)
		os.Exit(1)
	}
}
