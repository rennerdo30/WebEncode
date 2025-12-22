package handlers

import (
	"encoding/json"
	"net/http"
	"runtime"

	"github.com/go-chi/chi/v5"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/logger"
)

type SystemHandler struct {
	db     store.Querier
	logger *logger.Logger
}

func NewSystemHandler(db store.Querier, l *logger.Logger) *SystemHandler {
	return &SystemHandler{db: db, logger: l}
}

func (h *SystemHandler) Register(r chi.Router) {
	r.Get("/v1/system/health", h.HealthCheck)
	r.Get("/v1/system/stats", h.GetStats)
}

// HealthResponse represents the system health status
type HealthResponse struct {
	Status   string            `json:"status"`
	Version  string            `json:"version"`
	Services map[string]string `json:"services"`
}

func (h *SystemHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	health := HealthResponse{
		Status:  "healthy",
		Version: "1.0.0",
		Services: map[string]string{
			"database": "healthy",
			"nats":     "healthy",
			"storage":  "healthy",
		},
	}

	// Check database connectivity
	if _, err := h.db.CountJobs(r.Context()); err != nil {
		health.Services["database"] = "unhealthy"
		health.Status = "degraded"
	}

	if err := json.NewEncoder(w).Encode(health); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

// SystemStats represents overall system statistics
type SystemStats struct {
	Jobs struct {
		Total      int64 `json:"total"`
		Queued     int64 `json:"queued"`
		Processing int64 `json:"processing"`
		Completed  int64 `json:"completed"`
		Failed     int64 `json:"failed"`
	} `json:"jobs"`
	Workers struct {
		Total   int `json:"total"`
		Healthy int `json:"healthy"`
	} `json:"workers"`
	Streams struct {
		Total int `json:"total"`
		Live  int `json:"live"`
	} `json:"streams"`
	System struct {
		GoVersion    string `json:"go_version"`
		NumGoroutine int    `json:"num_goroutine"`
		NumCPU       int    `json:"num_cpu"`
	} `json:"system"`
}

func (h *SystemHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	stats := SystemStats{}

	// Job counts
	stats.Jobs.Total, _ = h.db.CountJobs(ctx)
	stats.Jobs.Queued, _ = h.db.CountJobsByStatus(ctx, store.JobStatusQueued)
	stats.Jobs.Processing, _ = h.db.CountJobsByStatus(ctx, store.JobStatusProcessing)
	stats.Jobs.Completed, _ = h.db.CountJobsByStatus(ctx, store.JobStatusCompleted)
	stats.Jobs.Failed, _ = h.db.CountJobsByStatus(ctx, store.JobStatusFailed)

	// Worker counts
	workers, _ := h.db.ListWorkers(ctx)
	stats.Workers.Total = len(workers)
	for _, w := range workers {
		if w.IsHealthy.Bool {
			stats.Workers.Healthy++
		}
	}

	// Stream counts
	streams, _ := h.db.ListStreams(ctx, store.ListStreamsParams{Limit: 1000, Offset: 0})
	stats.Streams.Total = len(streams)
	for _, s := range streams {
		if s.IsLive {
			stats.Streams.Live++
		}
	}

	// System info
	stats.System.GoVersion = runtime.Version()
	stats.System.NumGoroutine = runtime.NumGoroutine()
	stats.System.NumCPU = runtime.NumCPU()

	if err := json.NewEncoder(w).Encode(stats); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}
