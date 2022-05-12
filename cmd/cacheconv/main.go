package main

import (
	"flag"
	"fmt"
	"os"

	_ "github.com/minotar/imgd/pkg/build"
	"github.com/minotar/imgd/pkg/cache_converter"
	"github.com/minotar/imgd/pkg/util/cfg"
	"github.com/minotar/imgd/pkg/util/log"

	"github.com/prometheus/common/version"
	"go.uber.org/zap"
)

type Config struct {
	cache_converter.Config `yaml:",inline"`
	printVersion           bool
	prodLogging            bool
	upgradeV3UUID          bool
	upgradeV3UserData      bool
	downgradeV4UUID        bool
	downgradeV4UserData    bool
}

func (c *Config) RegisterFlags(f *flag.FlagSet) {
	f.BoolVar(&c.printVersion, "version", false, "Print this builds version information")
	f.BoolVar(&c.prodLogging, "prod-logging", false, "Use prod level logs")

	f.BoolVar(&c.upgradeV3UUID, "upgrade-v3-uuid", false, "run the v3 -> v4 UUID conversion")
	f.BoolVar(&c.upgradeV3UserData, "upgrade-v3-userdata", false, "run the v3 -> v4 User Data conversion")

	f.BoolVar(&c.downgradeV4UUID, "downgrade-v4-uuid", false, "run the v4 -> v3 UUID conversion")
	f.BoolVar(&c.downgradeV4UserData, "downgrade-v4-userdata", false, "run the v4 -> v3 User Data conversion")

	c.Config.RegisterFlags(f)
}

func main() {

	var config Config
	cfg.Parse(&config, "CACHECONV")

	var logConfig zap.Config
	if config.prodLogging {
		logConfig = zap.NewProductionConfig()
		logConfig.Encoding = "console"
		logConfig.EncoderConfig = zap.NewDevelopmentEncoderConfig()
	} else {
		logConfig = zap.NewDevelopmentConfig()
	}
	mainLogger, err := logConfig.Build()
	if err != nil {
		fmt.Printf("Logger failed to init: %+v\n", err)
		os.Exit(1)
	}

	logConfig.Build()

	defer mainLogger.Sync() // flushes buffer, if any
	logger := log.NewZapLogger(mainLogger)

	logger.Infof("Config: %+v\n", config)

	config.Logger = logger

	switch {
	case config.printVersion:
		fmt.Println(version.Print("processd"))
		os.Exit(0)
	}

	cc, err := cache_converter.New(config.Config)
	if err != nil {
		logger.Errorf("Error initialising cacheconv: %v", err)
	}

	logger.Infof("Starting cacheconv %s", version.Info())

	switch {
	case config.downgradeV4UUID:
		cc.MigrateUUIDV4toV3()
	case config.downgradeV4UserData:
		cc.MigrateUserDataV4toV3()
	case config.upgradeV3UUID:
		cc.MigrateUUIDV3toV4()
	case config.upgradeV3UserData:
		cc.MigrateUserDataV3toV4()
	}
}
