//
// Heavily influenced by https://github.com/prometheus/client_golang
// https://github.com/prometheus/client_golang/blob/8184d76b3b0bd3b01ed903690431ccb6826bf3e0/prometheus/promhttp/instrument_client.go
//
// Copyright 2017 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Custom version of the promhttp.InstrumentRoundTripper's with labelling for the minecraft API Client "source"

package minecraft_trace

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptrace"
	"time"

	"github.com/minotar/imgd/pkg/minecraft"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func ctxLabels(ctx context.Context) prometheus.Labels {
	source := minecraft.CtxGetSource(ctx)
	if source == "" {
		source = "Unknown"
	}
	return prometheus.Labels{"source": source}
}

// prometheus.GaugeVec *MUST* have a "source" label
func InstrumentRoundTripperInFlight(gaugeVec *prometheus.GaugeVec, next http.RoundTripper) promhttp.RoundTripperFunc {
	return promhttp.RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
		gauge := gaugeVec.With(ctxLabels(r.Context()))

		gauge.Inc()
		defer gauge.Dec()
		return next.RoundTrip(r)
	})
}

// prometheus.ObserverVec *MUST* have a "source". Optional "code" and "method" (though method will always be GET)
func InstrumentRoundTripperDuration(obs prometheus.ObserverVec, next http.RoundTripper) promhttp.RoundTripperFunc {
	return promhttp.RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
		observerVec := obs.MustCurryWith(ctxLabels(r.Context()))
		return promhttp.InstrumentRoundTripperDuration(observerVec, next).RoundTrip(r)
	})
}

// Each prometheus.ObserverVec *MUST* have a "source" label
type InstrumentTrace struct {
	GotFirstResponseByte prometheus.ObserverVec
	DNSStart             prometheus.ObserverVec
	DNSDone              prometheus.ObserverVec
	ConnectStart         prometheus.ObserverVec
	ConnectDone          prometheus.ObserverVec
	TLSHandshakeStart    prometheus.ObserverVec
	TLSHandshakeDone     prometheus.ObserverVec
}

func InstrumentRoundTripperTrace(it *InstrumentTrace, next http.RoundTripper) promhttp.RoundTripperFunc {
	return promhttp.RoundTripperFunc(func(r *http.Request) (*http.Response, error) {

		labels := ctxLabels(r.Context())
		start := time.Now()

		trace := &httptrace.ClientTrace{
			DNSStart: func(_ httptrace.DNSStartInfo) {
				if it.DNSStart != nil {
					it.DNSStart.With(labels).Observe(time.Since(start).Seconds())
				}
			},
			DNSDone: func(_ httptrace.DNSDoneInfo) {
				if it.DNSDone != nil {
					it.DNSDone.With(labels).Observe(time.Since(start).Seconds())
				}
			},
			ConnectStart: func(_, _ string) {
				if it.ConnectStart != nil {
					it.ConnectStart.With(labels).Observe(time.Since(start).Seconds())
				}
			},
			ConnectDone: func(_, _ string, err error) {
				if err != nil {
					return
				}
				if it.ConnectDone != nil {
					it.ConnectDone.With(labels).Observe(time.Since(start).Seconds())
				}
			},
			GotFirstResponseByte: func() {
				if it.GotFirstResponseByte != nil {
					it.GotFirstResponseByte.With(labels).Observe(time.Since(start).Seconds())
				}
			},
			TLSHandshakeStart: func() {
				if it.TLSHandshakeStart != nil {
					it.TLSHandshakeStart.With(labels).Observe(time.Since(start).Seconds())
				}
			},
			TLSHandshakeDone: func(_ tls.ConnectionState, err error) {
				if err != nil {
					return
				}
				if it.TLSHandshakeDone != nil {
					it.TLSHandshakeDone.With(labels).Observe(time.Since(start).Seconds())
				}
			},
		}
		r = r.WithContext(httptrace.WithClientTrace(r.Context(), trace))

		return next.RoundTrip(r)
	})
}
