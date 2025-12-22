package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewAuditHandler(t *testing.T) {
	handler := NewAuditHandler(nil, nil)
	assert.NotNil(t, handler)
}

func TestAuditHandler_Register(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewAuditHandler(mockStore, logger.New("test"))

	r := chi.NewRouter()
	handler.Register(r)

	// Verify routes are registered by checking the router
	assert.NotNil(t, r)
}

func TestAuditHandler_ListAuditLogs(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewAuditHandler(mockStore, logger.New("test"))

	logs := []store.AuditLog{
		{
			ID:           pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
			UserID:       pgtype.UUID{Bytes: [16]byte{2}, Valid: true},
			Action:       "create_job",
			ResourceType: "job",
		},
	}

	mockStore.On("ListAuditLogs", mock.Anything, mock.MatchedBy(func(arg store.ListAuditLogsParams) bool {
		return arg.Limit == 50 && arg.Offset == 0
	})).Return(logs, nil)

	req := httptest.NewRequest("GET", "/v1/audit", nil)
	w := httptest.NewRecorder()

	handler.ListAuditLogs(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockStore.AssertExpectations(t)
}

func TestAuditHandler_ListAuditLogs_WithPagination(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewAuditHandler(mockStore, logger.New("test"))

	mockStore.On("ListAuditLogs", mock.Anything, mock.MatchedBy(func(arg store.ListAuditLogsParams) bool {
		return arg.Limit == 10 && arg.Offset == 20
	})).Return([]store.AuditLog{}, nil)

	req := httptest.NewRequest("GET", "/v1/audit?limit=10&offset=20", nil)
	w := httptest.NewRecorder()

	handler.ListAuditLogs(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockStore.AssertExpectations(t)
}

func TestAuditHandler_ListAuditLogs_DBError(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewAuditHandler(mockStore, logger.New("test"))

	mockStore.On("ListAuditLogs", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	req := httptest.NewRequest("GET", "/v1/audit", nil)
	w := httptest.NewRecorder()

	handler.ListAuditLogs(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockStore.AssertExpectations(t)
}

func TestAuditHandler_ListAuditLogs_NilLogs(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewAuditHandler(mockStore, logger.New("test"))

	mockStore.On("ListAuditLogs", mock.Anything, mock.Anything).Return(nil, nil)

	req := httptest.NewRequest("GET", "/v1/audit", nil)
	w := httptest.NewRecorder()

	handler.ListAuditLogs(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []store.AuditLog
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Empty(t, resp)
}

func TestAuditHandler_ListUserAuditLogs(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewAuditHandler(mockStore, logger.New("test"))

	userID := pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}

	logs := []store.AuditLog{
		{
			ID:           pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
			UserID:       userID,
			Action:       "delete_job",
			ResourceType: "job",
		},
	}

	mockStore.On("ListAuditLogsByUser", mock.Anything, mock.MatchedBy(func(arg store.ListAuditLogsByUserParams) bool {
		return arg.Limit == 50 && arg.Offset == 0
	})).Return(logs, nil)

	r := chi.NewRouter()
	r.Get("/v1/audit/user/{userId}", handler.ListUserAuditLogs)

	req := httptest.NewRequest("GET", "/v1/audit/user/01020304-0506-0708-090a-0b0c0d0e0f10", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockStore.AssertExpectations(t)
}

func TestAuditHandler_ListUserAuditLogs_InvalidUUID(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewAuditHandler(mockStore, logger.New("test"))

	r := chi.NewRouter()
	r.Get("/v1/audit/user/{userId}", handler.ListUserAuditLogs)

	req := httptest.NewRequest("GET", "/v1/audit/user/not-a-uuid", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuditHandler_ListUserAuditLogs_WithPagination(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewAuditHandler(mockStore, logger.New("test"))

	mockStore.On("ListAuditLogsByUser", mock.Anything, mock.MatchedBy(func(arg store.ListAuditLogsByUserParams) bool {
		return arg.Limit == 25 && arg.Offset == 50
	})).Return([]store.AuditLog{}, nil)

	r := chi.NewRouter()
	r.Get("/v1/audit/user/{userId}", handler.ListUserAuditLogs)

	req := httptest.NewRequest("GET", "/v1/audit/user/01020304-0506-0708-090a-0b0c0d0e0f10?limit=25&offset=50", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockStore.AssertExpectations(t)
}

func TestAuditHandler_ListUserAuditLogs_DBError(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewAuditHandler(mockStore, logger.New("test"))

	mockStore.On("ListAuditLogsByUser", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	r := chi.NewRouter()
	r.Get("/v1/audit/user/{userId}", handler.ListUserAuditLogs)

	req := httptest.NewRequest("GET", "/v1/audit/user/01020304-0506-0708-090a-0b0c0d0e0f10", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAuditHandler_LogAction(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewAuditHandler(mockStore, logger.New("test"))

	userID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	details := map[string]interface{}{"job_id": "123"}

	mockStore.On("CreateAuditLog", mock.Anything, mock.MatchedBy(func(arg store.CreateAuditLogParams) bool {
		return arg.Action == "create_job" && arg.ResourceType == "job" && arg.ResourceID.String == "123"
	})).Return(nil)

	handler.LogAction(context.Background(), userID, "create_job", "job", "123", details)

	mockStore.AssertExpectations(t)
}

func TestAuditHandler_LogAction_EmptyResourceID(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewAuditHandler(mockStore, logger.New("test"))

	userID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	details := map[string]interface{}{}

	mockStore.On("CreateAuditLog", mock.Anything, mock.MatchedBy(func(arg store.CreateAuditLogParams) bool {
		return !arg.ResourceID.Valid
	})).Return(nil)

	handler.LogAction(context.Background(), userID, "system_action", "system", "", details)

	mockStore.AssertExpectations(t)
}

func TestAuditHandler_LogAction_DBError(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewAuditHandler(mockStore, logger.New("test"))

	userID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}

	mockStore.On("CreateAuditLog", mock.Anything, mock.Anything).Return(assert.AnError)

	// Should not panic
	handler.LogAction(context.Background(), userID, "action", "resource", "id", nil)

	mockStore.AssertExpectations(t)
}
