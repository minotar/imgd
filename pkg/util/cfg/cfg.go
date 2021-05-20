package cfg

import (
	"flag"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Registerer interface {
	RegisterFlags(*flag.FlagSet)
}

func Parse(r Registerer) *viper.Viper {

	fs := flag.CommandLine

	r.RegisterFlags(fs)
	pflag.CommandLine.AddGoFlagSet(fs)
	pflag.Parse()
	v := viper.New()
	v.BindPFlags(pflag.CommandLine)

	return v
}
