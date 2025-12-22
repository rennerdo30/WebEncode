package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rennerdo30/webencode/internal/api/middleware"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// helper to set user context for testing
func withUserContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, middleware.UserContextKey, &middleware.UserContext{
		UserID: userID,
	})
}

func TestNewNotificationsHandler(t *testing.T) {
	handler := NewNotificationsHandler(nil, nil)
	assert.NotNil(t, handler)
}

func TestNotificationsHandler_Register(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewNotificationsHandler(mockStore, logger.New("test"))

	r := chi.NewRouter()
	handler.Register(r)

	assert.NotNil(t, r)
}

func TestToNotificationResponse(t *testing.T) {
	now := time.Now()
	n := store.Notification{
		ID:        pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true},
		UserID:    pgtype.UUID{Bytes: [16]byte{2, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true},
		Title:     "Test Notification",
		Message:   "This is a test",
		Link:      pgtype.Text{String: "/jobs/123", Valid: true},
		Type:      pgtype.Text{String: "info", Valid: true},
		IsRead:    pgtype.Bool{Bool: false, Valid: true},
		CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
	}

	resp := toNotificationResponse(n)

	assert.Equal(t, "Test Notification", resp.Title)
	assert.Equal(t, "This is a test", resp.Message)
	assert.Equal(t, "/jobs/123", resp.Link)
	assert.Equal(t, "info", resp.Type)
	assert.False(t, resp.IsRead)
}

func TestNotificationsHandler_ListNotifications(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewNotificationsHandler(mockStore, logger.New("test"))

	userIDStr := "01020304-0506-0708-090a-0b0c0d0e0f10"

	notifications := []store.Notification{
		{
			ID:        pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
			UserID:    pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true},
			Title:     "Job Completed",
			Message:   "Your job has completed",
			CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		},
	}

	mockStore.On("ListNotifications", mock.Anything, mock.MatchedBy(func(arg store.ListNotificationsParams) bool {
		return arg.Limit == 20 && arg.Offset == 0
	})).Return(notifications, nil)

	req := httptest.NewRequest("GET", "/v1/notifications", nil)
	req = req.WithContext(withUserContext(req.Context(), userIDStr))
	w := httptest.NewRecorder()

	handler.ListNotifications(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockStore.AssertExpectations(t)

	var resp []NotificationResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Len(t, resp, 1)
}

func TestNotificationsHandler_ListNotifications_WithPagination(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewNotificationsHandler(mockStore, logger.New("test"))

	userIDStr := "01020304-0506-0708-090a-0b0c0d0e0f10"

	mockStore.On("ListNotifications", mock.Anything, mock.MatchedBy(func(arg store.ListNotificationsParams) bool {
		return arg.Limit == 10 && arg.Offset == 5
	})).Return([]store.Notification{}, nil)

	req := httptest.NewRequest("GET", "/v1/notifications?limit=10&offset=5", nil)
	req = req.WithContext(withUserContext(req.Context(), userIDStr))
	w := httptest.NewRecorder()

	handler.ListNotifications(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockStore.AssertExpectations(t)
}

func TestNotificationsHandler_ListNotifications_InvalidUserID(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewNotificationsHandler(mockStore, logger.New("test"))

	req := httptest.NewRequest("GET", "/v1/notifications", nil)
	req = req.WithContext(withUserContext(req.Context(), "not-a-uuid"))
	w := httptest.NewRecorder()

	handler.ListNotifications(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestNotificationsHandler_ListNotifications_DBError(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewNotificationsHandler(mockStore, logger.New("test"))

	userIDStr := "01020304-0506-0708-090a-0b0c0d0e0f10"

	mockStore.On("ListNotifications", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	req := httptest.NewRequest("GET", "/v1/notifications", nil)
	req = req.WithContext(withUserContext(req.Context(), userIDStr))
	w := httptest.NewRecorder()

	handler.ListNotifications(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockStore.AssertExpectations(t)
}

func TestNotificationsHandler_MarkRead(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewNotificationsHandler(mockStore, logger.New("test"))

	userIDStr := "01020304-0506-0708-090a-0b0c0d0e0f10"

	mockStore.On("MarkNotificationRead", mock.Anything, mock.Anything).Return(nil)

	r := chi.NewRouter()
	r.Put("/v1/notifications/{id}/read", func(w http.ResponseWriter, req *http.Request) {
		handler.MarkRead(w, req.WithContext(withUserContext(req.Context(), userIDStr)))
	})

	req := httptest.NewRequest("PUT", "/v1/notifications/01020304-0506-0708-090a-0b0c0d0e0f10/read", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockStore.AssertExpectations(t)
}

func TestNotificationsHandler_MarkRead_InvalidUserID(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewNotificationsHandler(mockStore, logger.New("test"))

	r := chi.NewRouter()
	r.Put("/v1/notifications/{id}/read", func(w http.ResponseWriter, req *http.Request) {
		handler.MarkRead(w, req.WithContext(withUserContext(req.Context(), "not-a-uuid")))
	})

	req := httptest.NewRequest("PUT", "/v1/notifications/01020304-0506-0708-090a-0b0c0d0e0f10/read", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestNotificationsHandler_MarkRead_InvalidNotificationID(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewNotificationsHandler(mockStore, logger.New("test"))

	userIDStr := "01020304-0506-0708-090a-0b0c0d0e0f10"

	r := chi.NewRouter()
	r.Put("/v1/notifications/{id}/read", func(w http.ResponseWriter, req *http.Request) {
		handler.MarkRead(w, req.WithContext(withUserContext(req.Context(), userIDStr)))
	})

	req := httptest.NewRequest("PUT", "/v1/notifications/not-a-uuid/read", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestNotificationsHandler_MarkRead_DBError(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewNotificationsHandler(mockStore, logger.New("test"))

	userIDStr := "01020304-0506-0708-090a-0b0c0d0e0f10"

	mockStore.On("MarkNotificationRead", mock.Anything, mock.Anything).Return(assert.AnError)

	r := chi.NewRouter()
	r.Put("/v1/notifications/{id}/read", func(w http.ResponseWriter, req *http.Request) {
		handler.MarkRead(w, req.WithContext(withUserContext(req.Context(), userIDStr)))
	})

	req := httptest.NewRequest("PUT", "/v1/notifications/01020304-0506-0708-090a-0b0c0d0e0f10/read", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockStore.AssertExpectations(t)
}

func TestNotificationsHandler_ClearAll(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewNotificationsHandler(mockStore, logger.New("test"))

	userIDStr := "01020304-0506-0708-090a-0b0c0d0e0f10"

	mockStore.On("MarkAllNotificationsRead", mock.Anything, mock.Anything).Return(nil)

	req := httptest.NewRequest("POST", "/v1/notifications/clear", nil)
	req = req.WithContext(withUserContext(req.Context(), userIDStr))
	w := httptest.NewRecorder()

	handler.ClearAll(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockStore.AssertExpectations(t)
}

func TestNotificationsHandler_ClearAll_InvalidUserID(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewNotificationsHandler(mockStore, logger.New("test"))

	req := httptest.NewRequest("POST", "/v1/notifications/clear", nil)
	req = req.WithContext(withUserContext(req.Context(), "not-a-uuid"))
	w := httptest.NewRecorder()

	handler.ClearAll(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestNotificationsHandler_ClearAll_DBError(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewNotificationsHandler(mockStore, logger.New("test"))

	userIDStr := "01020304-0506-0708-090a-0b0c0d0e0f10"

	mockStore.On("MarkAllNotificationsRead", mock.Anything, mock.Anything).Return(assert.AnError)

	req := httptest.NewRequest("POST", "/v1/notifications/clear", nil)
	req = req.WithContext(withUserContext(req.Context(), userIDStr))
	w := httptest.NewRecorder()

	handler.ClearAll(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockStore.AssertExpectations(t)
}
