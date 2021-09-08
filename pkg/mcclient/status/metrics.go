package status

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	metricsNamespace = "imgd"
)

var (
	apiGetErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: "mcclient",
			Name:      "api_get_errors",
			Help:      "Error events from external API Requests.",
		}, []string{"source", "event"},
	)
)
