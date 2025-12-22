package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rennerdo30/webencode/internal/api/middleware"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/errors"
	"github.com/rennerdo30/webencode/pkg/logger"
)

type NotificationsHandler struct {
	db     store.Querier
	logger *logger.Logger
}

func NewNotificationsHandler(db store.Querier, l *logger.Logger) *NotificationsHandler {
	return &NotificationsHandler{db: db, logger: l}
}

func (h *NotificationsHandler) Register(r chi.Router) {
	r.Get("/v1/notifications", h.ListNotifications)
	r.Put("/v1/notifications/{id}/read", h.MarkRead)
	r.Post("/v1/notifications/clear", h.ClearAll)
}

type NotificationResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Link      string    `json:"link,omitempty"`
	Type      string    `json:"type"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

func toNotificationResponse(n store.Notification) NotificationResponse {
	return NotificationResponse{
		ID:        n.ID.String(),
		UserID:    n.UserID.String(),
		Title:     n.Title,
		Message:   n.Message,
		Link:      n.Link.String,
		Type:      n.Type.String,
		IsRead:    n.IsRead.Bool,
		CreatedAt: n.CreatedAt.Time,
	}
}

func (h *NotificationsHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	var userID pgtype.UUID
	if err := userID.Scan(userIDStr); err != nil {
		h.logger.Error("Invalid user ID in context", "user_id_str", userIDStr, "error", err)
		errors.Response(w, r, errors.ErrUnauthorized)
		return
	}

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

	params := store.ListNotificationsParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	notifications, err := h.db.ListNotifications(r.Context(), params)
	if err != nil {
		h.logger.Error("Failed to list notifications", "user_id", userIDStr, "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	resp := make([]NotificationResponse, 0, len(notifications))
	for _, n := range notifications {
		resp = append(resp, toNotificationResponse(n))
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

func (h *NotificationsHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	var userID pgtype.UUID
	if err := userID.Scan(userIDStr); err != nil {
		errors.Response(w, r, errors.ErrUnauthorized)
		return
	}

	notificationID := chi.URLParam(r, "id")
	nUUID := pgtype.UUID{}
	if err := nUUID.Scan(notificationID); err != nil {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	params := store.MarkNotificationReadParams{
		ID:     nUUID,
		UserID: userID,
	}

	if err := h.db.MarkNotificationRead(r.Context(), params); err != nil {
		h.logger.Error("Failed to mark notification read", "id", notificationID, "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *NotificationsHandler) ClearAll(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	var userID pgtype.UUID
	if err := userID.Scan(userIDStr); err != nil {
		errors.Response(w, r, errors.ErrUnauthorized)
		return
	}

	if err := h.db.MarkAllNotificationsRead(r.Context(), userID); err != nil {
		h.logger.Error("Failed to clear notifications", "user_id", userIDStr, "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
