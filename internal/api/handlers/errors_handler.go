package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/errors"
	"github.com/rennerdo30/webencode/pkg/logger"
)

type ErrorsHandler struct {
	db     store.Querier
	logger *logger.Logger
}

func NewErrorsHandler(db store.Querier, logger *logger.Logger) *ErrorsHandler {
	return &ErrorsHandler{
		db:     db,
		logger: logger,
	}
}

// ReportError handles ingesting errors from clients (frontend, plugins, etc.)
func (h *ErrorsHandler) ReportError(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Source      string                 `json:"source"`
		Severity    string                 `json:"severity"`
		Message     string                 `json:"message"`
		StackTrace  string                 `json:"stack_trace"`
		ContextData map[string]interface{} `json:"context_data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	// Validate severity
	severity := store.ErrorSeverityError
	switch req.Severity {
	case "warning":
		severity = store.ErrorSeverityWarning
	case "critical":
		severity = store.ErrorSeverityCritical
	case "fatal":
		severity = store.ErrorSeverityFatal
	}

	// Convert context data to JSONB bytes
	contextDataBytes, err := json.Marshal(req.ContextData)
	if err != nil {
		contextDataBytes = []byte("{}")
	}

	event, err := h.db.CreateErrorEvent(r.Context(), store.CreateErrorEventParams{
		SourceComponent: req.Source,
		Column2:         severity, // sqlc generated this as Column2 due to ::error_severity cast
		Message:         req.Message,
		StackTrace:      pgtype.Text{String: req.StackTrace, Valid: req.StackTrace != ""},
		ContextData:     contextDataBytes,
	})

	if err != nil {
		h.logger.Error("Failed to persist error event", "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(event); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

// ListErrors returns a paginated list of error events
func (h *ErrorsHandler) ListErrors(w http.ResponseWriter, r *http.Request) {
	limit := 50
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = val
		}
	}

	source := r.URL.Query().Get("source")
	var events []store.ErrorEvent
	var err error

	if source != "" {
		events, err = h.db.ListErrorEventsBySource(r.Context(), store.ListErrorEventsBySourceParams{
			SourceComponent: source,
			Limit:           int32(limit),
			Offset:          int32(offset),
		})
	} else {
		events, err = h.db.ListErrorEvents(r.Context(), store.ListErrorEventsParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		})
	}

	if err != nil {
		h.logger.Error("Failed to list error events", "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(events); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

// ResolveError marks an error event as resolved
func (h *ErrorsHandler) ResolveError(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	var id pgtype.UUID
	if err := id.Scan(idStr); err != nil {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	if err := h.db.ResolveErrorEvent(r.Context(), id); err != nil {
		h.logger.Error("Failed to resolve error event", "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ErrorsHandler) Register(r *chi.Mux) {
	r.Route("/v1/errors", func(r chi.Router) {
		r.Post("/", h.ReportError)
		r.Get("/", h.ListErrors)
		r.Patch("/{id}/resolve", h.ResolveError)
	})
}
