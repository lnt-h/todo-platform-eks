package httpapi

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	requestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total HTTP requests by method, route and status code.",
	}, []string{"method", "route", "status"})

	requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request latency by method and route.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route"})
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// instrument records Prometheus metrics and writes one JSON log line per request.
func instrument(mux *http.ServeMux, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		mux.ServeHTTP(rec, r)

		// Use the route pattern ("/todos/{id}"), not the raw path ("/todos/42"),
		// so each todo id does not create a new time series.
		_, route := mux.Handler(r)
		if route == "" {
			route = "unmatched"
		}
		elapsed := time.Since(start)

		requestsTotal.WithLabelValues(r.Method, route, strconv.Itoa(rec.status)).Inc()
		requestDuration.WithLabelValues(r.Method, route).Observe(elapsed.Seconds())

		if route != "GET /healthz" && route != "GET /readyz" && route != "GET /metrics" {
			logger.Info("http request",
				"method", r.Method, "path", r.URL.Path, "status", rec.status,
				"duration_ms", elapsed.Milliseconds())
		}
	})
}
