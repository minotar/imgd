package processd

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	metricsNamespace = "processd"
)

var (
	requestedUserType = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: "request",
			Name:      "user_type",
			Help:      "Type of processd User requested.",
		}, []string{"type"},
	)
)
