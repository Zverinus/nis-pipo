package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	reg           = prometheus.NewRegistry()
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

func init() {
	reg.MustRegister(requestsTotal, requestDuration)
}

func Handler() http.Handler {
	return promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		DisableCompression: true,
	})
}

func RecordRequest(r *http.Request, status int, duration time.Duration) {
	path := r.URL.Path
	if ctx := chi.RouteContext(r.Context()); ctx != nil && ctx.RoutePattern() != "" {
		path = ctx.RoutePattern()
	}
	switch path {
	case "/metrics", "/swagger/*", "/favicon.ico":
		return
	}
	statusStr := strconv.Itoa(status)
	requestsTotal.WithLabelValues(r.Method, path, statusStr).Inc()
	requestDuration.WithLabelValues(r.Method, path).Observe(duration.Seconds())
}
