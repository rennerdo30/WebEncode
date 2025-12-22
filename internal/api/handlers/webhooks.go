package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/errors"
	"github.com/rennerdo30/webencode/pkg/logger"
)

type WebhooksHandler struct {
	db     store.Querier
	logger *logger.Logger
}

func NewWebhooksHandler(db store.Querier, l *logger.Logger) *WebhooksHandler {
	return &WebhooksHandler{db: db, logger: l}
}

func (h *WebhooksHandler) Register(r chi.Router) {
	r.Get("/v1/webhooks", h.ListWebhooks)
	r.Post("/v1/webhooks", h.CreateWebhook)
	r.Get("/v1/webhooks/{id}", h.GetWebhook)
	r.Delete("/v1/webhooks/{id}", h.DeleteWebhook)
}

func (h *WebhooksHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	webhooks, err := h.db.ListWebhooks(r.Context())
	if err != nil {
		h.logger.Error("Failed to list webhooks", "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	if webhooks == nil {
		webhooks = []store.Webhook{}
	}
	if err := json.NewEncoder(w).Encode(webhooks); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

func (h *WebhooksHandler) GetWebhook(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var uid pgtype.UUID
	if err := uid.Scan(id); err != nil {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	webhook, err := h.db.GetWebhook(r.Context(), uid)
	if err != nil {
		errors.Response(w, r, errors.ErrNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(webhook); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

type CreateWebhookRequest struct {
	URL    string   `json:"url"`
	Secret string   `json:"secret"`
	Events []string `json:"events"` // job.completed, job.failed, stream.started, etc.
}

func (h *WebhooksHandler) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	var req CreateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	if req.URL == "" || len(req.Events) == 0 {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	userID := "00000000-0000-0000-0000-000000000001"
	var uid pgtype.UUID
	uid.Scan(userID)

	webhook, err := h.db.CreateWebhook(r.Context(), store.CreateWebhookParams{
		UserID: uid,
		Url:    req.URL,
		Secret: req.Secret,
		Events: req.Events,
	})
	if err != nil {
		h.logger.Error("Failed to create webhook", "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(webhook); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

func (h *WebhooksHandler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var uid pgtype.UUID
	if err := uid.Scan(id); err != nil {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	if err := h.db.DeleteWebhook(r.Context(), uid); err != nil {
		h.logger.Error("Failed to delete webhook", "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
