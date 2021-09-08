package processd

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	metricsNamespace = "imgd"
)

var (
	requestedUserType = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: "processd",
			Name:      "request_user_type",
			Help:      "Type of processd User requested.",
		}, []string{"type"},
	)
)
