package imgd

import (
	"github.com/felixge/fgprof"
	"github.com/minotar/imgd/pkg/processd"
	"github.com/minotar/imgd/pkg/skind"
)

// routes registers all the routes
func (i *Imgd) routes() {
	i.Server.HTTP.Path("/debug/fgprof").Handler(fgprof.Handler())
	i.Server.HTTP.Path("/healthcheck").Handler(skind.HealthcheckHandler(i.McClient))
	i.Server.HTTP.Path("/dbsize").Handler(skind.SizecheckHandler(i.McClient))

	skinWrapper := skind.NewSkinWrapper(i.Cfg.Logger, i.McClient, i.Cfg.UseETags, i.Cfg.CacheControlTTL)

	skind.RegisterSkinRoutes(i.Server.HTTP, skinWrapper)
	processd.RegisterProcessingRoutes(i.Server.HTTP, skinWrapper, i.ProcessRoutes)
}
