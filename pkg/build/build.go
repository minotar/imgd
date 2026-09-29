package build

import "github.com/prometheus/common/version"

// Build metadata can be set with go build -ldflags -X.
var (
	Version   = "dev"
	Revision  string
	Branch    string
	BuildUser string
	BuildDate string
)

func init() {
	version.Version = Version
	version.Revision = Revision
	version.Branch = Branch
	version.BuildUser = BuildUser
	version.BuildDate = BuildDate
}
