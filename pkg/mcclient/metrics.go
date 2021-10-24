package mcclient

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	metricsNamespace = "imgd"
)

var DefBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}

var (
	apiGetDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Subsystem: "mcclient",
			Name:      "api_get_duration_seconds",
			Help:      "Time (in seconds) external API Requests took.",
			Buckets:   DefBuckets,
		}, []string{"source"},
	)

	apiClientInflight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: "mcclient_api",
			Name:      "inflight_requests",
			Help:      "Current number of inflight API requests.",
		},
	)

	cacheStatus = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: "mcclient",
			Name:      "cache_status",
			Help:      "Time (in seconds) external API Requests took.",
		}, []string{"cache", "status"},
	)
)

type cacheStatusRecorder struct {
	counter *prometheus.CounterVec
}

func (c *cacheStatusRecorder) Hit() {
	c.counter.WithLabelValues("hit").Inc()
}
func (c *cacheStatusRecorder) Miss() {
	c.counter.WithLabelValues("miss").Inc()
}
func (c *cacheStatusRecorder) Fresh() {
	c.counter.WithLabelValues("fresh").Inc()
}
func (c *cacheStatusRecorder) Stale() {
	c.counter.WithLabelValues("stale").Inc()
}
func (c *cacheStatusRecorder) Error() {
	c.counter.WithLabelValues("error").Inc()
}

var (
	uuidCacheStatus     = cacheStatusRecorder{cacheStatus.MustCurryWith(prometheus.Labels{"cache": "CacheUUID"})}
	userdataCacheStatus = cacheStatusRecorder{cacheStatus.MustCurryWith(prometheus.Labels{"cache": "CacheUserData"})}
	textureCacheStatus  = cacheStatusRecorder{cacheStatus.MustCurryWith(prometheus.Labels{"cache": "CacheTextures"})}
)
