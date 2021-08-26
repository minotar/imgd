package mcclient

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	metricsNamespace = "mcclient"
)

var DefBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}

var (
	apiGetDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Subsystem: "api",
			Name:      "get_duration_seconds",
			Help:      "Histogram of the time (in seconds) external API Requests took.",
			Buckets:   DefBuckets,
		}, []string{"source"},
	)
)
