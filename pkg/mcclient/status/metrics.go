package status

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	metricsNamespace = "mcclient"
)

var (
	apiGetErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: "api",
			Name:      "get_errors",
			Help:      "Error events from external API Requests.",
		}, []string{"source", "event"},
	)
)
