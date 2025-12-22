package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rennerdo30/webencode/internal/orchestrator"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/errors"
	"github.com/rennerdo30/webencode/pkg/logger"
)

type RestreamResponse struct {
	ID                 string                   `json:"id"`
	Title              string                   `json:"title,omitempty"`
	Description        string                   `json:"description,omitempty"`
	InputType          string                   `json:"input_type,omitempty"`
	InputURL           string                   `json:"input_url,omitempty"`
	OutputDestinations []map[string]interface{} `json:"output_destinations"`
	Status             string                   `json:"status"`
	CreatedAt          time.Time                `json:"created_at"`
	UpdatedAt          time.Time                `json:"updated_at"`
	StartedAt          *time.Time               `json:"started_at,omitempty"`
	StoppedAt          *time.Time               `json:"stopped_at,omitempty"`
}

func toRestreamResponse(r store.RestreamJob) RestreamResponse {
	var dests []map[string]interface{}
	json.Unmarshal(r.OutputDestinations, &dests)

	var start, stop *time.Time
	if r.StartedAt.Valid {
		t := r.StartedAt.Time
		start = &t
	}
	if r.StoppedAt.Valid {
		t := r.StoppedAt.Time
		stop = &t
	}

	return RestreamResponse{
		ID:                 r.ID.String(),
		Title:              r.Title.String,
		Description:        r.Description.String,
		InputType:          r.InputType.String,
		InputURL:           r.InputUrl.String,
		OutputDestinations: dests,
		Status:             r.Status.String,
		CreatedAt:          r.CreatedAt.Time,
		UpdatedAt:          r.UpdatedAt.Time,
		StartedAt:          start,
		StoppedAt:          stop,
	}
}

type RestreamsHandler struct {
	db           store.Querier
	orchestrator orchestrator.Service
	logger       *logger.Logger
}

func NewRestreamsHandler(db store.Querier, svc orchestrator.Service, l *logger.Logger) *RestreamsHandler {
	return &RestreamsHandler{db: db, orchestrator: svc, logger: l}
}

func (h *RestreamsHandler) Register(r chi.Router) {
	r.Get("/v1/restreams", h.ListRestreams)
	r.Post("/v1/restreams", h.CreateRestream)
	r.Get("/v1/restreams/{id}", h.GetRestream)
	r.Delete("/v1/restreams/{id}", h.DeleteRestream)
	r.Post("/v1/restreams/{id}/start", h.StartRestream)
	r.Post("/v1/restreams/{id}/stop", h.StopRestream)
}

func (h *RestreamsHandler) ListRestreams(w http.ResponseWriter, r *http.Request) {
	limit := 20
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil {
			offset = v
		}
	}

	restreams, err := h.db.ListRestreamJobs(r.Context(), store.ListRestreamJobsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		h.logger.Error("Failed to list restreams", "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	resp := make([]RestreamResponse, 0, len(restreams))
	for _, r := range restreams {
		resp = append(resp, toRestreamResponse(r))
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

func (h *RestreamsHandler) GetRestream(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var uid pgtype.UUID
	if err := uid.Scan(id); err != nil {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	restream, err := h.db.GetRestreamJob(r.Context(), uid)
	if err != nil {
		errors.Response(w, r, errors.ErrNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(toRestreamResponse(restream)); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

type CreateRestreamRequest struct {
	Title              string                   `json:"title"`
	Description        string                   `json:"description"`
	InputType          string                   `json:"input_type"` // rtmp, file, vod_job
	InputURL           string                   `json:"input_url"`
	OutputDestinations []map[string]interface{} `json:"output_destinations"`
	ScheduleType       string                   `json:"schedule_type"` // immediate, scheduled, recurring
	LoopEnabled        bool                     `json:"loop_enabled"`
	SimulateLive       bool                     `json:"simulate_live"`
}

func (h *RestreamsHandler) CreateRestream(w http.ResponseWriter, r *http.Request) {
	var req CreateRestreamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	userID := "00000000-0000-0000-0000-000000000001"
	var uid pgtype.UUID
	uid.Scan(userID)

	destBytes, _ := json.Marshal(req.OutputDestinations)

	restream, err := h.db.CreateRestreamJob(r.Context(), store.CreateRestreamJobParams{
		UserID:             uid,
		Title:              pgtype.Text{String: req.Title, Valid: true},
		Description:        pgtype.Text{String: req.Description, Valid: true},
		InputType:          pgtype.Text{String: req.InputType, Valid: true},
		InputUrl:           pgtype.Text{String: req.InputURL, Valid: true},
		OutputDestinations: destBytes,
		ScheduleType:       pgtype.Text{String: req.ScheduleType, Valid: req.ScheduleType != ""},
		LoopEnabled:        pgtype.Bool{Bool: req.LoopEnabled, Valid: true},
		SimulateLive:       pgtype.Bool{Bool: req.SimulateLive, Valid: true},
	})
	if err != nil {
		h.logger.Error("Failed to create restream", "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(toRestreamResponse(restream)); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

func (h *RestreamsHandler) DeleteRestream(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var uid pgtype.UUID
	if err := uid.Scan(id); err != nil {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	if err := h.db.DeleteRestreamJob(r.Context(), uid); err != nil {
		h.logger.Error("Failed to delete restream", "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *RestreamsHandler) StartRestream(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var uid pgtype.UUID
	if err := uid.Scan(id); err != nil {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	err := h.db.UpdateRestreamJobStatus(r.Context(), store.UpdateRestreamJobStatusParams{
		ID:     uid,
		Status: pgtype.Text{String: "streaming", Valid: true},
	})
	if err != nil {
		h.logger.Error("Failed to update status", "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	if err := h.orchestrator.SubmitRestream(r.Context(), id); err != nil {
		h.logger.Error("Failed to dispatch restream task", "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "streaming"}); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

func (h *RestreamsHandler) StopRestream(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var uid pgtype.UUID
	if err := uid.Scan(id); err != nil {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	err := h.db.UpdateRestreamJobStatus(r.Context(), store.UpdateRestreamJobStatusParams{
		ID:     uid,
		Status: pgtype.Text{String: "stopped", Valid: true},
	})
	if err != nil {
		h.logger.Error("Failed to update status", "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	if err := h.orchestrator.StopRestream(r.Context(), id); err != nil {
		h.logger.Error("Failed to stop restream task", "error", err)
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "stopped"}); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}
