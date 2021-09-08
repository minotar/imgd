package skind

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
			Subsystem: "skind",
			Name:      "request_user_type",
			Help:      "Type of skind User requested.",
		}, []string{"type"},
	)
)
