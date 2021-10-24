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
	apiClientDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Subsystem: "mcclient_api",
			Name:      "duration_seconds",
			Help:      "Time (in seconds) API requests took.",
			Buckets:   DefBuckets,
		}, []string{"source", "code"},
	)

	apiClientInflight = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: "mcclient_api",
			Name:      "inflight_requests",
			Help:      "Current number of inflight API requests.",
		}, []string{"source"},
	)

	apiClientTraceDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Subsystem: "mcclient_api",
			Name:      "trace_duration_seconds",
			Help:      "Time (in seconds) since start of API request for events to occur.",
			Buckets:   DefBuckets,
		}, []string{"source", "event"},
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
