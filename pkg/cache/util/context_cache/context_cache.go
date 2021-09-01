// context_cache is a magic wrapper around a cache.Cache object
// These allows us to add metric/tracing code when needed, via calling the
// WithContext method on the Cache object.

package context_cache

import (
	"context"
	"time"

	"github.com/minotar/imgd/pkg/cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel/trace"
)

type ContextCache struct {
	cache.Cache
	Context context.Context
}

var _ cache.Cache = new(ContextCache)

var DefBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}

var (
	insertHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "cache",
			Subsystem: "api",
			Name:      "get_duration_seconds",
			Help:      "Histogram of the time (in seconds) external API Requests took.",
			Buckets:   DefBuckets,
		}, []string{"cacheName"},
	)
)

// Insert a new value into the store
func (cc *ContextCache) Insert(key string, value []byte) error {

	curSpan := trace.SpanContextFromContext(cc.Context)

	insertHist := insertHistogram.WithLabelValues(cc.Name())

	observeFunc := func(value float64) {

		if exemplarObserver, ok := insertHist.(prometheus.ExemplarObserver); ok && curSpan.IsValid() {
			exemplarObserver.ObserveWithExemplar(value, prometheus.Labels{
				"TraceID": curSpan.TraceID().String(),
			})
		}
	}

	timer := prometheus.NewTimer(prometheus.ObserverFunc(observeFunc))
	defer timer.ObserveDuration()

	return cc.Cache.Insert(key, value)
}

// Retrieve will attempt to find the key in the store. Returns
// nil if it does not exist with an ErrNotFound
func (cc *ContextCache) Retrieve(key string) ([]byte, error) {
	return cc.Cache.Retrieve(key)
}

// Remove will silently attempt to delete the key from the store
func (cc *ContextCache) Remove(key string) error {
	return cc.Cache.Remove(key)
}

// Flush will empty the store
func (cc *ContextCache) Flush() error {
	return cc.Cache.Flush()
}

// InsertTTL inserts a new value into the store with the given expiry
func (cc *ContextCache) InsertTTL(key string, value []byte, ttl time.Duration) error {
	return cc.Cache.InsertTTL(key, value, ttl)
}

func (cc *ContextCache) WithContext(ctx context.Context) cache.Cache {
	return &ContextCache{
		Cache:   cc.Cache,
		Context: ctx,
	}
}
