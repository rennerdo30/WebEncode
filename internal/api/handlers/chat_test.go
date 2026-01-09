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

func TestChatHandler_GetChatMessages(t *testing.T) {
	mockStore := new(MockStore)
	l := logger.New("test")

	// No plugin manager for this test - we test the handler logic
	handler := NewChatHandler(mockStore, nil, l)

	t.Run("invalid stream ID", func(t *testing.T) {
		r := chi.NewRouter()
		r.Get("/v1/streams/{id}/chat", handler.GetChatMessages)

		req := httptest.NewRequest("GET", "/v1/streams/invalid-uuid/chat", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("stream not found", func(t *testing.T) {
		r := chi.NewRouter()
		r.Get("/v1/streams/{id}/chat", handler.GetChatMessages)

		var uid pgtype.UUID
		uid.Scan("00000000-0000-0000-0000-000000000001")

		mockStore.On("GetStream", mock.Anything, uid).Return(store.Stream{}, assert.AnError).Once()

		req := httptest.NewRequest("GET", "/v1/streams/00000000-0000-0000-0000-000000000001/chat", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("returns empty messages when no destinations", func(t *testing.T) {
		r := chi.NewRouter()
		r.Get("/v1/streams/{id}/chat", handler.GetChatMessages)

		var uid pgtype.UUID
		uid.Scan("00000000-0000-0000-0000-000000000002")

		mockStore.On("GetStream", mock.Anything, uid).Return(store.Stream{
			ID:                   uid,
			StreamKey:            "test-key",
			RestreamDestinations: []byte("[]"),
		}, nil).Once()

		req := httptest.NewRequest("GET", "/v1/streams/00000000-0000-0000-0000-000000000002/chat", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var messages []ChatMessage
		err := json.NewDecoder(w.Body).Decode(&messages)
		assert.NoError(t, err)
		assert.Empty(t, messages)
	})
}

func TestChatHandler_SendChatMessage(t *testing.T) {
	mockStore := new(MockStore)
	l := logger.New("test")

	handler := NewChatHandler(mockStore, nil, l)

	t.Run("invalid stream ID", func(t *testing.T) {
		r := chi.NewRouter()
		r.Post("/v1/streams/{id}/chat", handler.SendChatMessage)

		body := bytes.NewBufferString(`{"message":"hello"}`)
		req := httptest.NewRequest("POST", "/v1/streams/invalid-uuid/chat", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("empty message rejected", func(t *testing.T) {
		r := chi.NewRouter()
		r.Post("/v1/streams/{id}/chat", handler.SendChatMessage)

		body := bytes.NewBufferString(`{"message":""}`)
		req := httptest.NewRequest("POST", "/v1/streams/00000000-0000-0000-0000-000000000001/chat", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("stream not found", func(t *testing.T) {
		r := chi.NewRouter()
		r.Post("/v1/streams/{id}/chat", handler.SendChatMessage)

		var uid pgtype.UUID
		uid.Scan("00000000-0000-0000-0000-000000000003")

		mockStore.On("GetStream", mock.Anything, uid).Return(store.Stream{}, assert.AnError).Once()

		body := bytes.NewBufferString(`{"message":"hello"}`)
		req := httptest.NewRequest("POST", "/v1/streams/00000000-0000-0000-0000-000000000003/chat", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("sends to no destinations returns success false", func(t *testing.T) {
		r := chi.NewRouter()
		r.Post("/v1/streams/{id}/chat", handler.SendChatMessage)

		var uid pgtype.UUID
		uid.Scan("00000000-0000-0000-0000-000000000004")

		mockStore.On("GetStream", mock.Anything, uid).Return(store.Stream{
			ID:                   uid,
			StreamKey:            "test-key",
			RestreamDestinations: []byte("[]"),
		}, nil).Once()

		body := bytes.NewBufferString(`{"message":"hello world"}`)
		req := httptest.NewRequest("POST", "/v1/streams/00000000-0000-0000-0000-000000000004/chat", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, false, resp["success"])
	})
}
