package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/minotar/imgd/pkg/cache_converter"
	"github.com/minotar/imgd/pkg/util/cfg"
	"github.com/minotar/imgd/pkg/util/log"
	"github.com/prometheus/common/version"
	"go.uber.org/zap"
)

type Config struct {
	cache_converter.Config `yaml:",inline"`
	printVersion           bool
	uuidv3tov4             bool
	userDatav3tov4         bool
}

func (c *Config) RegisterFlags(f *flag.FlagSet) {
	f.BoolVar(&c.printVersion, "version", false, "Print this builds version information")

	f.BoolVar(&c.uuidv3tov4, "uuid-v3-to-v4", false, "run the v3 -> v4 UUID conversion")
	f.BoolVar(&c.userDatav3tov4, "userdata-v3-to-v4", false, "run the v3 -> v4 User Data conversion")

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

	switch {
	case flaggedVersion:
		fmt.Println(version.Print("processd"))
		os.Exit(0)
	}

	cc, err := cache_converter.New(config.Config)
	if err != nil {
		logger.Errorf("Error initialising cacheconv: %v", err)
	}

	logger.Infof("Starting cacheconv %s", version.Info())

	switch {
	case config.uuidv3tov4:
		cc.MigrateUUIDV4toV3()
	case config.userDatav3tov4:
		cc.MigrateUserDataV4toV3()
	}
}
