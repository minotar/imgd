package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	metricsNamespace = "imgd"
)

var DefBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30}

// Todo: We could record cache hit/miss state on the timings?

var (
	cacheOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Subsystem: "cache",
			Name:      "operation_duration_seconds",
			Help:      "Time (in seconds) cache opertations took.",
			Buckets:   DefBuckets,
		}, []string{"type", "cache", "operation"},
	)
	cacheExpiredCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: "cache",
			Name:      "expired_total",
			Help:      "Total number of expired records.",
		}, []string{"type", "cache"},
	)
)

func NewCacheOperationDuration(cacheType, cacheName string) prometheus.ObserverVec {
	return cacheOperationDuration.MustCurryWith(prometheus.Labels{
		"type":  cacheType,
		"cache": cacheName,
	})
}

func NewCacheExpiredCounter(cacheType, cacheName string) *prometheus.CounterVec {
	return cacheExpiredCounter.MustCurryWith(prometheus.Labels{
		"type":  cacheType,
		"cache": cacheName,
	})
}

func NewCacheSizeGauge(cacheType, cacheName string, f func() uint64) {
	gaugeFunc := func() float64 {
		return float64(f())
	}

	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: metricsNamespace,
		Subsystem: "cache",
		Name:      "size",
		Help:      "Size (in bytes) of the cache",
		ConstLabels: prometheus.Labels{
			"type":  cacheType,
			"cache": cacheName,
		},
	}, gaugeFunc)
}

func NewCacheLenGauge(cacheType, cacheName string, f func() uint) {
	gaugeFunc := func() float64 {
		return float64(f())
	}

	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: metricsNamespace,
		Subsystem: "cache",
		Name:      "length",
		Help:      "Size (in bytes) of the cache",
		ConstLabels: prometheus.Labels{
			"type":  cacheType,
			"cache": cacheName,
		},
	}, gaugeFunc)
}
