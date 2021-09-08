package skind

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	metricsNamespace = "skind"
)

var (
	requestedUserType = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: "request",
			Name:      "user_type",
			Help:      "Type of skind User requested.",
		}, []string{"type"},
	)
)
