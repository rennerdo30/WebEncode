package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Job metrics
	JobsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "webencode_jobs_total",
			Help: "Total number of jobs by status",
		},
		[]string{"status"},
	)

	JobDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "webencode_job_duration_seconds",
			Help:    "Job processing duration in seconds",
			Buckets: prometheus.ExponentialBuckets(1, 2, 15), // 1s to ~9h
		},
		[]string{"profile"},
	)

	// Task metrics
	TasksTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "webencode_tasks_total",
			Help: "Total number of tasks by type and status",
		},
		[]string{"type", "status"},
	)

	TaskDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "webencode_task_duration_seconds",
			Help:    "Task processing duration in seconds",
			Buckets: prometheus.ExponentialBuckets(0.1, 2, 12), // 100ms to ~7min
		},
		[]string{"type"},
	)

	// Worker metrics
	WorkersActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "webencode_workers_active",
			Help: "Number of healthy workers",
		},
	)

	WorkerCapacity = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "webencode_worker_capacity",
			Help: "Worker capacity metrics",
		},
		[]string{"worker_id", "metric"},
	)

	// API metrics
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "webencode_http_requests_total",
			Help: "Total HTTP requests by method and path",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "webencode_http_request_duration_seconds",
			Help:    "HTTP request latency",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Plugin metrics
	PluginRPCDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "webencode_plugin_rpc_duration_seconds",
			Help:    "Plugin RPC call duration",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"plugin", "method"},
	)

	PluginErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "webencode_plugin_errors_total",
			Help: "Plugin error count",
		},
		[]string{"plugin", "error_type"},
	)

	// Stream metrics
	StreamsActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "webencode_streams_active",
			Help: "Number of active live streams",
		},
	)

	StreamViewers = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "webencode_stream_viewers",
			Help: "Current viewers per stream",
		},
		[]string{"stream_id"},
	)
)

func init() {
	// Register all metrics
	prometheus.MustRegister(
		JobsTotal,
		JobDuration,
		TasksTotal,
		TaskDuration,
		WorkersActive,
		WorkerCapacity,
		HTTPRequestsTotal,
		HTTPRequestDuration,
		PluginRPCDuration,
		PluginErrors,
		StreamsActive,
		StreamViewers,
	)
}

// Handler returns the Prometheus HTTP handler
func Handler() http.Handler {
	return promhttp.Handler()
}

// InstrumentHandler wraps an http.Handler with prometheus metrics
func InstrumentHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start).Seconds()
		path := normalizePath(r.URL.Path)

		HTTPRequestsTotal.WithLabelValues(r.Method, path, strconv.Itoa(wrapped.statusCode)).Inc()
		HTTPRequestDuration.WithLabelValues(r.Method, path).Observe(duration)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// normalizePath reduces cardinality by removing IDs from paths
func normalizePath(path string) string {
	// Common patterns: /v1/jobs/{id} -> /v1/jobs/:id
	// This is a simple implementation - could be more sophisticated
	parts := []string{}
	for _, part := range splitPath(path) {
		if isUUID(part) || isNumeric(part) {
			parts = append(parts, ":id")
		} else {
			parts = append(parts, part)
		}
	}
	result := "/" + joinPath(parts)
	if result == "/" {
		return "/"
	}
	return result
}

func splitPath(path string) []string {
	result := []string{}
	current := ""
	for _, c := range path {
		if c == '/' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func joinPath(parts []string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += "/"
		}
		result += p
	}
	return result
}

func isUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	// Check for UUID format: 8-4-4-4-12
	if s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		return false
	}
	return true
}

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}
