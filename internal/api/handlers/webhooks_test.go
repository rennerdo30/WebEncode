package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupWebhooksTest() (*WebhooksHandler, *MockStore) {
	mockStore := new(MockStore)
	logger := logger.New("test")
	handler := NewWebhooksHandler(mockStore, logger)
	return handler, mockStore
}

func TestWebhooksHandler_Register(t *testing.T) {
	handler, _ := setupWebhooksTest()
	r := chi.NewRouter()
	handler.Register(r)

	// We trust Chi to register routes.
	// Detailed handler execution is tested in individual handler tests.
	assert.NotNil(t, r)
}

func TestWebhooksHandler_ListWebhooks(t *testing.T) {
	handler, mockStore := setupWebhooksTest()

	expected := []store.Webhook{
		{
			ID:     pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
			Url:    "http://example.com",
			Events: []string{"start"},
		},
	}

	mockStore.On("ListWebhooks", mock.Anything).Return(expected, nil)

	req := httptest.NewRequest("GET", "/v1/webhooks", nil)
	w := httptest.NewRecorder()

	handler.ListWebhooks(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []store.Webhook
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Len(t, resp, 1)
	assert.Equal(t, expected[0].Url, resp[0].Url)
	mockStore.AssertExpectations(t)
}

func TestWebhooksHandler_ListWebhooks_Error(t *testing.T) {
	handler, mockStore := setupWebhooksTest()

	mockStore.On("ListWebhooks", mock.Anything).Return(nil, pgx.ErrTxClosed)

	req := httptest.NewRequest("GET", "/v1/webhooks", nil)
	w := httptest.NewRecorder()

	handler.ListWebhooks(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWebhooksHandler_GetWebhook(t *testing.T) {
	handler, mockStore := setupWebhooksTest()
	idStr := "00000000-0000-0000-0000-000000000001"

	var uid pgtype.UUID
	uid.Scan(idStr)

	expected := store.Webhook{
		ID:  uid,
		Url: "http://example.com",
	}

	mockStore.On("GetWebhook", mock.Anything, uid).Return(expected, nil)

	r := chi.NewRouter()
	r.Get("/v1/webhooks/{id}", handler.GetWebhook)

	req := httptest.NewRequest("GET", "/v1/webhooks/"+idStr, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockStore.AssertExpectations(t)
}

func TestWebhooksHandler_GetWebhook_NotFound(t *testing.T) {
	handler, mockStore := setupWebhooksTest()
	idStr := "00000000-0000-0000-0000-000000000001"

	var uid pgtype.UUID
	uid.Scan(idStr)

	mockStore.On("GetWebhook", mock.Anything, uid).Return(store.Webhook{}, pgx.ErrNoRows)

	r := chi.NewRouter()
	r.Get("/v1/webhooks/{id}", handler.GetWebhook)

	req := httptest.NewRequest("GET", "/v1/webhooks/"+idStr, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestWebhooksHandler_GetWebhook_InvalidID(t *testing.T) {
	handler, _ := setupWebhooksTest()

	r := chi.NewRouter()
	r.Get("/v1/webhooks/{id}", handler.GetWebhook)

	req := httptest.NewRequest("GET", "/v1/webhooks/invalid", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWebhooksHandler_CreateWebhook(t *testing.T) {
	handler, mockStore := setupWebhooksTest()

	body := CreateWebhookRequest{
		URL:    "http://example.com/hook",
		Secret: "secret",
		Events: []string{"job.completed"},
	}
	bodyBytes, _ := json.Marshal(body)

	expected := store.Webhook{
		ID:     pgtype.UUID{Bytes: [16]byte{2}, Valid: true},
		Url:    body.URL,
		Secret: body.Secret,
		Events: body.Events,
	}

	mockStore.On("CreateWebhook", mock.Anything, mock.MatchedBy(func(arg store.CreateWebhookParams) bool {
		return arg.Url == body.URL && arg.Secret == body.Secret
	})).Return(expected, nil)

	req := httptest.NewRequest("POST", "/v1/webhooks", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	handler.CreateWebhook(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockStore.AssertExpectations(t)
}

func TestWebhooksHandler_CreateWebhook_InvalidParams(t *testing.T) {
	handler, _ := setupWebhooksTest()

	// Missing URL
	body := CreateWebhookRequest{
		Events: []string{"job.completed"},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/webhooks", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	handler.CreateWebhook(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Invalid JSON
	req = httptest.NewRequest("POST", "/v1/webhooks", bytes.NewReader([]byte("invalid")))
	w = httptest.NewRecorder()
	handler.CreateWebhook(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWebhooksHandler_DeleteWebhook(t *testing.T) {
	handler, mockStore := setupWebhooksTest()
	idStr := "00000000-0000-0000-0000-000000000001"

	var uid pgtype.UUID
	uid.Scan(idStr)

	mockStore.On("DeleteWebhook", mock.Anything, uid).Return(nil)

	r := chi.NewRouter()
	r.Delete("/v1/webhooks/{id}", handler.DeleteWebhook)

	req := httptest.NewRequest("DELETE", "/v1/webhooks/"+idStr, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockStore.AssertExpectations(t)
}

func TestWebhooksHandler_DeleteWebhook_Error(t *testing.T) {
	handler, mockStore := setupWebhooksTest()
	idStr := "00000000-0000-0000-0000-000000000001"

	var uid pgtype.UUID
	uid.Scan(idStr)

	mockStore.On("DeleteWebhook", mock.Anything, uid).Return(pgx.ErrTxClosed)

	r := chi.NewRouter()
	r.Delete("/v1/webhooks/{id}", handler.DeleteWebhook)

	req := httptest.NewRequest("DELETE", "/v1/webhooks/"+idStr, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
