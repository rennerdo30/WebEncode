package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/errors"
	"github.com/rennerdo30/webencode/pkg/logger"
)

type AuditHandler struct {
	db     store.Querier
	logger *logger.Logger
}

func NewAuditHandler(db store.Querier, l *logger.Logger) *AuditHandler {
	return &AuditHandler{db: db, logger: l}
}

func (h *AuditHandler) Register(r chi.Router) {
	r.Get("/v1/audit", h.ListAuditLogs)
	r.Get("/v1/audit/user/{userId}", h.ListUserAuditLogs)
}

func (h *AuditHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	limit := 50
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

	logs, err := h.db.ListAuditLogs(r.Context(), store.ListAuditLogsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		h.logger.Error("Failed to list audit logs", "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	if logs == nil {
		logs = []store.AuditLog{}
	}
	if err := json.NewEncoder(w).Encode(logs); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

func (h *AuditHandler) ListUserAuditLogs(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	var userID pgtype.UUID
	if err := userID.Scan(userIDStr); err != nil {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	limit := 50
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

	logs, err := h.db.ListAuditLogsByUser(r.Context(), store.ListAuditLogsByUserParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		h.logger.Error("Failed to list user audit logs", "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	if logs == nil {
		logs = []store.AuditLog{}
	}
	if err := json.NewEncoder(w).Encode(logs); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

// LogAction creates an audit log entry (called internally by other handlers)
func (h *AuditHandler) LogAction(ctx context.Context, userID pgtype.UUID, action, resourceType, resourceID string, details map[string]interface{}) {
	detailBytes, _ := json.Marshal(details)

	err := h.db.CreateAuditLog(ctx, store.CreateAuditLogParams{
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   pgtype.Text{String: resourceID, Valid: resourceID != ""},
		Details:      detailBytes,
	})
	if err != nil {
		h.logger.Error("Failed to create audit log", "error", err)
	}
}
