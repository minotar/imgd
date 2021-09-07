package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/prometheus/common/version"

	_ "github.com/minotar/imgd/pkg/build"
	"github.com/minotar/imgd/pkg/imgd"
	"github.com/minotar/imgd/pkg/util/cfg"
	"github.com/minotar/imgd/pkg/util/log"
)

func init() {
	prometheus.MustRegister(version.NewCollector("imgd"))
}

type Config struct {
	imgd.Config  `yaml:",inline"`
	printVersion bool
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
	cfg.Parse(&config, "IMGD")
	logger.Infof("Config: %+v\n", config)

	config.Logger = logger

	switch {
	case config.printVersion:
		fmt.Println(version.Print("imgd"))
		os.Exit(0)
	}

	// Start imgd
	s, err := imgd.New(config.Config)
	if err != nil {
		logger.Errorf("Error initialising imgd: %v", err)
	}

	logger.Infof("Starting imgd %s", version.Info())

	err = s.Run()

	if err != nil {
		logger.Errorf("Error running imgd : %v", err)
		os.Exit(1)
	}
}
