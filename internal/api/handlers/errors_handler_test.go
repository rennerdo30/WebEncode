package handlers

import (
	"bytes"
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

func TestNewErrorsHandler(t *testing.T) {
	handler := NewErrorsHandler(nil, nil)
	assert.NotNil(t, handler)
}

func TestErrorsHandler_Register(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewErrorsHandler(mockStore, logger.New("test"))

	r := chi.NewMux()
	handler.Register(r)

	assert.NotNil(t, r)
}

func TestErrorsHandler_ReportError(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewErrorsHandler(mockStore, logger.New("test"))

	reqBody := map[string]interface{}{
		"source":       "frontend",
		"severity":     "error",
		"message":      "Something went wrong",
		"stack_trace":  "Error at line 10",
		"context_data": map[string]string{"page": "dashboard"},
	}
	body, _ := json.Marshal(reqBody)

	mockStore.On("CreateErrorEvent", mock.Anything, mock.MatchedBy(func(arg store.CreateErrorEventParams) bool {
		return arg.SourceComponent == "frontend" &&
			arg.Column2 == store.ErrorSeverityError &&
			arg.Message == "Something went wrong"
	})).Return(store.ErrorEvent{
		ID:              pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
		SourceComponent: "frontend",
		Message:         "Something went wrong",
	}, nil)

	req := httptest.NewRequest("POST", "/v1/errors", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.ReportError(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockStore.AssertExpectations(t)
}

func TestErrorsHandler_ReportError_WarningSeverity(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewErrorsHandler(mockStore, logger.New("test"))

	reqBody := map[string]interface{}{
		"source":   "api",
		"severity": "warning",
		"message":  "Rate limit approaching",
	}
	body, _ := json.Marshal(reqBody)

	mockStore.On("CreateErrorEvent", mock.Anything, mock.MatchedBy(func(arg store.CreateErrorEventParams) bool {
		return arg.Column2 == store.ErrorSeverityWarning
	})).Return(store.ErrorEvent{}, nil)

	req := httptest.NewRequest("POST", "/v1/errors", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.ReportError(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockStore.AssertExpectations(t)
}

func TestErrorsHandler_ReportError_CriticalSeverity(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewErrorsHandler(mockStore, logger.New("test"))

	reqBody := map[string]interface{}{
		"source":   "worker",
		"severity": "critical",
		"message":  "Out of memory",
	}
	body, _ := json.Marshal(reqBody)

	mockStore.On("CreateErrorEvent", mock.Anything, mock.MatchedBy(func(arg store.CreateErrorEventParams) bool {
		return arg.Column2 == store.ErrorSeverityCritical
	})).Return(store.ErrorEvent{}, nil)

	req := httptest.NewRequest("POST", "/v1/errors", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.ReportError(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockStore.AssertExpectations(t)
}

func TestErrorsHandler_ReportError_FatalSeverity(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewErrorsHandler(mockStore, logger.New("test"))

	reqBody := map[string]interface{}{
		"source":   "system",
		"severity": "fatal",
		"message":  "System crash",
	}
	body, _ := json.Marshal(reqBody)

	mockStore.On("CreateErrorEvent", mock.Anything, mock.MatchedBy(func(arg store.CreateErrorEventParams) bool {
		return arg.Column2 == store.ErrorSeverityFatal
	})).Return(store.ErrorEvent{}, nil)

	req := httptest.NewRequest("POST", "/v1/errors", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.ReportError(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockStore.AssertExpectations(t)
}

func TestErrorsHandler_ReportError_InvalidJSON(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewErrorsHandler(mockStore, logger.New("test"))

	req := httptest.NewRequest("POST", "/v1/errors", bytes.NewReader([]byte("invalid-json")))
	w := httptest.NewRecorder()

	handler.ReportError(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestErrorsHandler_ReportError_DBError(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewErrorsHandler(mockStore, logger.New("test"))

	reqBody := map[string]interface{}{
		"source":  "test",
		"message": "Test error",
	}
	body, _ := json.Marshal(reqBody)

	mockStore.On("CreateErrorEvent", mock.Anything, mock.Anything).Return(store.ErrorEvent{}, assert.AnError)

	req := httptest.NewRequest("POST", "/v1/errors", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.ReportError(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockStore.AssertExpectations(t)
}

func TestErrorsHandler_ListErrors(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewErrorsHandler(mockStore, logger.New("test"))

	events := []store.ErrorEvent{
		{
			ID:              pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
			SourceComponent: "frontend",
			Message:         "Test error",
		},
	}

	mockStore.On("ListErrorEvents", mock.Anything, mock.MatchedBy(func(arg store.ListErrorEventsParams) bool {
		return arg.Limit == 50 && arg.Offset == 0
	})).Return(events, nil)

	req := httptest.NewRequest("GET", "/v1/errors", nil)
	w := httptest.NewRecorder()

	handler.ListErrors(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockStore.AssertExpectations(t)
}

func TestErrorsHandler_ListErrors_WithPagination(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewErrorsHandler(mockStore, logger.New("test"))

	mockStore.On("ListErrorEvents", mock.Anything, mock.MatchedBy(func(arg store.ListErrorEventsParams) bool {
		return arg.Limit == 10 && arg.Offset == 20
	})).Return([]store.ErrorEvent{}, nil)

	req := httptest.NewRequest("GET", "/v1/errors?limit=10&offset=20", nil)
	w := httptest.NewRecorder()

	handler.ListErrors(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockStore.AssertExpectations(t)
}

func TestErrorsHandler_ListErrors_BySource(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewErrorsHandler(mockStore, logger.New("test"))

	mockStore.On("ListErrorEventsBySource", mock.Anything, mock.MatchedBy(func(arg store.ListErrorEventsBySourceParams) bool {
		return arg.SourceComponent == "frontend" && arg.Limit == 50
	})).Return([]store.ErrorEvent{}, nil)

	req := httptest.NewRequest("GET", "/v1/errors?source=frontend", nil)
	w := httptest.NewRecorder()

	handler.ListErrors(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockStore.AssertExpectations(t)
}

func TestErrorsHandler_ListErrors_DBError(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewErrorsHandler(mockStore, logger.New("test"))

	mockStore.On("ListErrorEvents", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	req := httptest.NewRequest("GET", "/v1/errors", nil)
	w := httptest.NewRecorder()

	handler.ListErrors(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockStore.AssertExpectations(t)
}

func TestErrorsHandler_ResolveError(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewErrorsHandler(mockStore, logger.New("test"))

	mockStore.On("ResolveErrorEvent", mock.Anything, mock.Anything).Return(nil)

	r := chi.NewRouter()
	r.Patch("/v1/errors/{id}/resolve", handler.ResolveError)

	req := httptest.NewRequest("PATCH", "/v1/errors/01020304-0506-0708-090a-0b0c0d0e0f10/resolve", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockStore.AssertExpectations(t)
}

func TestErrorsHandler_ResolveError_InvalidUUID(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewErrorsHandler(mockStore, logger.New("test"))

	r := chi.NewRouter()
	r.Patch("/v1/errors/{id}/resolve", handler.ResolveError)

	req := httptest.NewRequest("PATCH", "/v1/errors/not-a-uuid/resolve", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestErrorsHandler_ResolveError_DBError(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewErrorsHandler(mockStore, logger.New("test"))

	mockStore.On("ResolveErrorEvent", mock.Anything, mock.Anything).Return(assert.AnError)

	r := chi.NewRouter()
	r.Patch("/v1/errors/{id}/resolve", handler.ResolveError)

	req := httptest.NewRequest("PATCH", "/v1/errors/01020304-0506-0708-090a-0b0c0d0e0f10/resolve", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockStore.AssertExpectations(t)
}
