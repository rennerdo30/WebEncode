package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStreamsHandler_ListStreams(t *testing.T) {
	mockDB := new(MockStore)
	handler := NewStreamsHandler(mockDB, logger.New("test"))

	mockDB.On("ListStreams", mock.Anything, mock.Anything).Return([]store.Stream{
		{
			ID:        pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
			StreamKey: "live_abc",
			Title:     pgtype.Text{String: "My Stream", Valid: true},
			IsLive:    true,
			CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			IngestUrl: pgtype.Text{String: "rtmp://ingest", Valid: true},
		},
	}, nil)

	req := httptest.NewRequest("GET", "/v1/streams", nil)
	w := httptest.NewRecorder()

	handler.ListStreams(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var streams []StreamResponse
	json.NewDecoder(w.Body).Decode(&streams)
	assert.Len(t, streams, 1)
	assert.Equal(t, "live_abc", streams[0].StreamKey)
}

func TestStreamsHandler_GetStream(t *testing.T) {
	mockDB := new(MockStore)
	handler := NewStreamsHandler(mockDB, logger.New("test"))

	r := chi.NewRouter()
	r.Get("/v1/streams/{id}", handler.GetStream)

	// Valid UUID for testing
	uuidStr := "00000000-0000-0000-0000-000000000001"
	req := httptest.NewRequest("GET", "/v1/streams/"+uuidStr, nil)
	w := httptest.NewRecorder()

	mockDB.On("GetStream", mock.Anything, mock.MatchedBy(func(id pgtype.UUID) bool {
		return id.Valid
	})).Return(store.Stream{
		ID:        pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
		StreamKey: "live_123",
	}, nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStreamsHandler_GetStream_NotFound(t *testing.T) {
	mockDB := new(MockStore)
	handler := NewStreamsHandler(mockDB, logger.New("test"))

	r := chi.NewRouter()
	r.Get("/v1/streams/{id}", handler.GetStream)

	uuidStr := "00000000-0000-0000-0000-000000000001"
	req := httptest.NewRequest("GET", "/v1/streams/"+uuidStr, nil)
	w := httptest.NewRecorder()

	mockDB.On("GetStream", mock.Anything, mock.Anything).Return(store.Stream{}, errors.New("not found"))

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestStreamsHandler_CreateStream(t *testing.T) {
	mockDB := new(MockStore)
	handler := NewStreamsHandler(mockDB, logger.New("test"))

	reqBody := CreateStreamRequest{
		Title:       "New Stream",
		Description: "Desc",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/streams", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mockDB.On("CreateStream", mock.Anything, mock.MatchedBy(func(arg store.CreateStreamParams) bool {
		return arg.Title.String == "New Stream" && strings.HasPrefix(arg.StreamKey, "live_")
	})).Return(store.Stream{
		ID:        pgtype.UUID{Bytes: [16]byte{2}, Valid: true},
		StreamKey: "live_generated",
		Title:     pgtype.Text{String: "New Stream", Valid: true},
	}, nil)

	handler.CreateStream(w, req)

	assert.Equal(t, http.StatusOK, w.Code) // creates returns 200 in current impl (should be 201 but checking impl)
	// Actually implementation: w.WriteHeader not called explicitly with 201?
	// streams.go:162 just encodes. Default is 200.
	// Oh wait, `CreateRestream` uses 201. `CreateJob` uses 201. `CreateStream` uses default 200?
	// Let's check `streams.go`.
	// Line 162: if err := json.NewEncoder(w).Encode...
	// No WriteHeader(StatusCreated). So it is 200.
}

func TestStreamsHandler_GetStreamDestinations(t *testing.T) {
	mockDB := new(MockStore)
	handler := NewStreamsHandler(mockDB, logger.New("test"))

	r := chi.NewRouter()
	r.Get("/v1/streams/{id}/destinations", handler.GetStreamDestinations)

	uuidStr := "00000000-0000-0000-0000-000000000001"
	req := httptest.NewRequest("GET", "/v1/streams/"+uuidStr+"/destinations", nil)
	w := httptest.NewRecorder()

	mockDB.On("GetStream", mock.Anything, mock.Anything).Return(store.Stream{
		RestreamDestinations: []byte(`[{"plugin_id":"twitch","enabled":true}]`),
	}, nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var dests []RestreamDestination
	json.NewDecoder(w.Body).Decode(&dests)
	assert.Len(t, dests, 1)
	assert.Equal(t, "twitch", dests[0].PluginID)
}

func TestStreamsHandler_UpdateStreamDestinations(t *testing.T) {
	mockDB := new(MockStore)
	handler := NewStreamsHandler(mockDB, logger.New("test"))

	r := chi.NewRouter()
	r.Put("/v1/streams/{id}/destinations", handler.UpdateStreamDestinations)

	uuidStr := "00000000-0000-0000-0000-000000000001"
	destBody := []RestreamDestination{
		{PluginID: "twitch", AccessToken: "token", Enabled: true},
	}
	body, _ := json.Marshal(destBody)
	req := httptest.NewRequest("PUT", "/v1/streams/"+uuidStr+"/destinations", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mockDB.On("UpdateStreamRestreamDestinations", mock.Anything, mock.MatchedBy(func(arg store.UpdateStreamRestreamDestinationsParams) bool {
		return string(arg.RestreamDestinations) != "" // simplified
	})).Return(nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStreamsHandler_Register(t *testing.T) {
	mockDB := new(MockStore)
	handler := NewStreamsHandler(mockDB, logger.New("test"))
	r := chi.NewRouter()
	handler.Register(r)
	assert.NotNil(t, r)
}
