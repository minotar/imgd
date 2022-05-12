package cfg

import (
	"flag"

	"github.com/kamaln7/envy"
	"github.com/spf13/pflag"
)

type Registerer interface {
	RegisterFlags(*flag.FlagSet)
}

func Parse(r Registerer, envName string) {
	fs := flag.CommandLine
	r.RegisterFlags(fs)

	envy.Parse(envName)
	pflag.CommandLine.AddGoFlagSet(fs)
	pflag.Parse()
}
