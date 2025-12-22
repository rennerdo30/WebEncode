package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/errors"
	"github.com/rennerdo30/webencode/pkg/logger"
)

type WorkersHandler struct {
	db     store.Querier
	logger *logger.Logger
}

type WorkerResponse struct {
	ID            string                 `json:"id"`
	Hostname      string                 `json:"hostname"`
	Version       string                 `json:"version"`
	LastHeartbeat time.Time              `json:"last_heartbeat"`
	Status        string                 `json:"status"`
	Capacity      map[string]interface{} `json:"capacity,omitempty"`
	IsHealthy     bool                   `json:"is_healthy"`
	Capabilities  map[string]interface{} `json:"capabilities,omitempty"`
}

func toWorkerResponse(w store.Worker) WorkerResponse {
	var capMap map[string]interface{}
	json.Unmarshal(w.Capacity, &capMap)

	var capsMap map[string]interface{}
	json.Unmarshal(w.Capabilities, &capsMap)

	return WorkerResponse{
		ID:            w.ID,
		Hostname:      w.Hostname,
		Version:       w.Version,
		LastHeartbeat: w.LastSeen.Time,
		Status:        w.Status,
		Capacity:      capMap,
		IsHealthy:     w.IsHealthy.Bool,
		Capabilities:  capsMap,
	}
}

func NewWorkersHandler(db store.Querier, l *logger.Logger) *WorkersHandler {
	return &WorkersHandler{db: db, logger: l}
}

func (h *WorkersHandler) Register(r chi.Router) {
	r.Get("/v1/workers", h.ListWorkers)
	r.Get("/v1/workers/{id}", h.GetWorker)
}

func (h *WorkersHandler) ListWorkers(w http.ResponseWriter, r *http.Request) {
	workers, err := h.db.ListWorkers(r.Context())
	if err != nil {
		h.logger.Error("Failed to list workers", "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	resp := make([]WorkerResponse, 0, len(workers))
	for _, worker := range workers {
		resp = append(resp, toWorkerResponse(worker))
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

func (h *WorkersHandler) GetWorker(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	worker, err := h.db.GetWorker(r.Context(), id)
	if err != nil {
		errors.Response(w, r, errors.ErrNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(toWorkerResponse(worker)); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}
