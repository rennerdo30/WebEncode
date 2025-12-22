package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExtractStreamKey(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"live prefix", "live/mykey", "mykey"},
		{"no prefix", "mykey", "mykey"},
		{"empty", "", ""},
		{"with spaces", "  key  ", "key"},
		{"live prefix with spaces", "live/  key  ", "key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractStreamKey(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLiveHandler_HandleAuth_Success(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewLiveHandler(logger.New("test"), mockStore)

	body := map[string]string{"path": "live/valid-key"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/live/auth", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mockStore.On("GetStreamByKey", mock.Anything, "valid-key").Return(store.Stream{
		ID: pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
	}, nil)

	handler.HandleAuth(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockStore.AssertExpectations(t)
}

func TestLiveHandler_HandleAuth_InvalidKey(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewLiveHandler(logger.New("test"), mockStore)

	body := map[string]string{"path": "live/invalid-key"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/live/auth", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mockStore.On("GetStreamByKey", mock.Anything, "invalid-key").Return(store.Stream{}, errors.New("not found"))

	handler.HandleAuth(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	mockStore.AssertExpectations(t)
}

func TestLiveHandler_HandleAuth_EmptyKey(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewLiveHandler(logger.New("test"), mockStore)

	body := map[string]string{"path": ""}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/live/auth", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleAuth(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestLiveHandler_HandleAuth_FormData(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewLiveHandler(logger.New("test"), mockStore)

	form := url.Values{}
	form.Add("path", "live/form-key")
	req := httptest.NewRequest("POST", "/live/auth", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	mockStore.On("GetStreamByKey", mock.Anything, "form-key").Return(store.Stream{
		ID: pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
	}, nil)

	handler.HandleAuth(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockStore.AssertExpectations(t)
}

func TestLiveHandler_HandleStart_Success(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewLiveHandler(logger.New("test"), mockStore)

	body := map[string]string{"path": "live/stream-key"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/live/start", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	streamID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	mockStore.On("GetStreamByKey", mock.Anything, "stream-key").Return(store.Stream{
		ID: streamID,
	}, nil)
	mockStore.On("UpdateStreamLive", mock.Anything, store.UpdateStreamLiveParams{
		ID:     streamID,
		IsLive: true,
	}).Return(nil)

	handler.HandleStart(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockStore.AssertExpectations(t)
}

func TestLiveHandler_HandleStart_EmptyKey(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewLiveHandler(logger.New("test"), mockStore)

	body := map[string]string{"path": ""}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/live/start", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleStart(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLiveHandler_HandleStart_KeyNotFound(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewLiveHandler(logger.New("test"), mockStore)

	body := map[string]string{"path": "live/unknown-key"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/live/start", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mockStore.On("GetStreamByKey", mock.Anything, "unknown-key").Return(store.Stream{}, errors.New("not found"))

	handler.HandleStart(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockStore.AssertExpectations(t)
}

func TestLiveHandler_HandleStop_Success(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewLiveHandler(logger.New("test"), mockStore)

	body := map[string]string{"path": "live/stream-key"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/live/stop", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	streamID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	mockStore.On("GetStreamByKey", mock.Anything, "stream-key").Return(store.Stream{
		ID: streamID,
	}, nil)
	mockStore.On("UpdateStreamLive", mock.Anything, store.UpdateStreamLiveParams{
		ID:     streamID,
		IsLive: false,
	}).Return(nil)

	handler.HandleStop(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockStore.AssertExpectations(t)
}

func TestLiveHandler_HandleStop_EmptyKey(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewLiveHandler(logger.New("test"), mockStore)

	body := map[string]string{"path": ""}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/live/stop", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleStop(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLiveHandler_HandleStop_KeyNotFound(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewLiveHandler(logger.New("test"), mockStore)

	body := map[string]string{"path": "live/unknown-key"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/live/stop", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mockStore.On("GetStreamByKey", mock.Anything, "unknown-key").Return(store.Stream{}, errors.New("not found"))

	handler.HandleStop(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockStore.AssertExpectations(t)
}
