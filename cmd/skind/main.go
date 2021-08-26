package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/prometheus/common/version"

	_ "github.com/minotar/imgd/pkg/build"
	"github.com/minotar/imgd/pkg/skind"
	"github.com/minotar/imgd/pkg/util/cfg"
	"github.com/minotar/imgd/pkg/util/log"
)

const Var1 = "Hi"

func init() {
	prometheus.MustRegister(version.NewCollector("skind"))
}

type Config struct {
	skind.Config `yaml:",inline"`
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
	v := cfg.Parse(&config)
	flaggedVersion := v.GetBool("version")

	fmt.Printf("Config: %+v\n", config)

	config.Logger = logger
	config.Config.CorsAllowAll = true

	switch {
	case flaggedVersion:
		fmt.Println(version.Print("skind"))
		os.Exit(0)
	}

	// Start skind
	s, err := skind.New(config.Config)
	if err != nil {
		logger.Errorf("Error initialising skind: %v", err)
	}

	logger.Infof("Starting skind %s", version.Info())

	err = s.Run()

	if err != nil {
		logger.Errorf("Error running skind : %v", err)
		os.Exit(1)
	}
}
