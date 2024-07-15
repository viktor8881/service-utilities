package simplehttp

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type MetricRoundTripper struct {
	Proxied         http.RoundTripper
	requestDuration *prometheus.HistogramVec
	requestCounter  *prometheus.CounterVec
}

func NewMetricsRoundTripper(proxied http.RoundTripper) *MetricRoundTripper {
	return &MetricRoundTripper{
		Proxied: proxied,
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "http_request_duration_seconds",
				Help: "Duration of HTTP requests in seconds.",
			},
			[]string{"method", "url", "error"},
		),
		requestCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests.",
			},
			[]string{"method", "url", "status", "error"},
		),
	}
}

func (lrt *MetricRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	resp, err := lrt.Proxied.RoundTrip(req)
	duration := time.Since(start).Seconds()

	errorLabel := "false"
	if err != nil {
		errorLabel = "true"
		lrt.requestDuration.WithLabelValues(req.Method, req.URL.String(), errorLabel).Observe(duration)
		lrt.requestCounter.WithLabelValues(req.Method, req.URL.String(), "error", errorLabel).Inc()
		return nil, err
	}

	lrt.requestDuration.WithLabelValues(req.Method, req.URL.String(), errorLabel).Observe(duration)
	lrt.requestCounter.WithLabelValues(req.Method, req.URL.String(), resp.Status, errorLabel).Inc()

	return resp, nil
}

func (lrt *MetricRoundTripper) RegisterMetrics() {
	prometheus.MustRegister(lrt.requestDuration, lrt.requestCounter)
}
