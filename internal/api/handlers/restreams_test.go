package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRestreamsHandler_ListRestreams(t *testing.T) {
	mockDB := new(MockStore)
	mockSvc := new(MockOrchestratorService)
	handler := NewRestreamsHandler(mockDB, mockSvc, logger.New("test"))

	mockDB.On("ListRestreamJobs", mock.Anything, mock.Anything).Return([]store.RestreamJob{
		{
			ID:                 pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
			Title:              pgtype.Text{String: "Restream 1", Valid: true},
			Status:             pgtype.Text{String: "idle", Valid: true},
			CreatedAt:          pgtype.Timestamptz{Time: time.Now(), Valid: true},
			UpdatedAt:          pgtype.Timestamptz{Time: time.Now(), Valid: true},
			OutputDestinations: []byte(`[]`),
		},
	}, nil)

	req := httptest.NewRequest("GET", "/v1/restreams", nil)
	w := httptest.NewRecorder()

	handler.ListRestreams(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res []RestreamResponse
	json.NewDecoder(w.Body).Decode(&res)
	assert.Len(t, res, 1)
	assert.Equal(t, "Restream 1", res[0].Title)
}

func TestRestreamsHandler_GetRestream(t *testing.T) {
	mockDB := new(MockStore)
	mockSvc := new(MockOrchestratorService)
	handler := NewRestreamsHandler(mockDB, mockSvc, logger.New("test"))

	r := chi.NewRouter()
	r.Get("/v1/restreams/{id}", handler.GetRestream)

	uuidStr := "00000000-0000-0000-0000-000000000001"
	req := httptest.NewRequest("GET", "/v1/restreams/"+uuidStr, nil)
	w := httptest.NewRecorder()

	mockDB.On("GetRestreamJob", mock.Anything, mock.Anything).Return(store.RestreamJob{
		ID:                 pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
		Title:              pgtype.Text{String: "Restream 1", Valid: true},
		OutputDestinations: []byte(`[]`),
	}, nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRestreamsHandler_CreateRestream(t *testing.T) {
	mockDB := new(MockStore)
	mockSvc := new(MockOrchestratorService)
	handler := NewRestreamsHandler(mockDB, mockSvc, logger.New("test"))

	reqBody := CreateRestreamRequest{
		Title:       "New Restream",
		InputType:   "rtmp",
		InputURL:    "rtmp://src",
		LoopEnabled: true,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/restreams", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mockDB.On("CreateRestreamJob", mock.Anything, mock.MatchedBy(func(arg store.CreateRestreamJobParams) bool {
		return arg.Title.String == "New Restream" && arg.LoopEnabled.Bool
	})).Return(store.RestreamJob{
		ID:                 pgtype.UUID{Bytes: [16]byte{2}, Valid: true},
		Title:              pgtype.Text{String: "New Restream", Valid: true},
		OutputDestinations: []byte(`null`),
	}, nil)

	handler.CreateRestream(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestRestreamsHandler_StartRestream(t *testing.T) {
	mockDB := new(MockStore)
	mockSvc := new(MockOrchestratorService)
	handler := NewRestreamsHandler(mockDB, mockSvc, logger.New("test"))

	r := chi.NewRouter()
	r.Post("/v1/restreams/{id}/start", handler.StartRestream)

	uuidStr := "00000000-0000-0000-0000-000000000001"
	req := httptest.NewRequest("POST", "/v1/restreams/"+uuidStr+"/start", nil)
	w := httptest.NewRecorder()

	// Update Status
	mockDB.On("UpdateRestreamJobStatus", mock.Anything, mock.MatchedBy(func(arg store.UpdateRestreamJobStatusParams) bool {
		return arg.Status.String == "streaming"
	})).Return(nil)

	// Submit Dispatch
	mockSvc.On("SubmitRestream", mock.Anything, uuidStr).Return(nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res map[string]string
	json.NewDecoder(w.Body).Decode(&res)
	assert.Equal(t, "streaming", res["status"])
}

func TestRestreamsHandler_StopRestream(t *testing.T) {
	mockDB := new(MockStore)
	mockSvc := new(MockOrchestratorService)
	handler := NewRestreamsHandler(mockDB, mockSvc, logger.New("test"))

	r := chi.NewRouter()
	r.Post("/v1/restreams/{id}/stop", handler.StopRestream)

	uuidStr := "00000000-0000-0000-0000-000000000001"
	req := httptest.NewRequest("POST", "/v1/restreams/"+uuidStr+"/stop", nil)
	w := httptest.NewRecorder()

	// Update Status
	mockDB.On("UpdateRestreamJobStatus", mock.Anything, mock.MatchedBy(func(arg store.UpdateRestreamJobStatusParams) bool {
		return arg.Status.String == "stopped"
	})).Return(nil)

	// Stop Dispatch
	mockSvc.On("StopRestream", mock.Anything, uuidStr).Return(nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRestreamsHandler_DeleteRestream(t *testing.T) {
	mockDB := new(MockStore)
	mockSvc := new(MockOrchestratorService)
	handler := NewRestreamsHandler(mockDB, mockSvc, logger.New("test"))

	r := chi.NewRouter()
	r.Delete("/v1/restreams/{id}", handler.DeleteRestream)

	uuidStr := "00000000-0000-0000-0000-000000000001"
	req := httptest.NewRequest("DELETE", "/v1/restreams/"+uuidStr, nil)
	w := httptest.NewRecorder()

	mockDB.On("DeleteRestreamJob", mock.Anything, mock.Anything).Return(nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestRestreamsHandler_Register(t *testing.T) {
	mockDB := new(MockStore)
	mockSvc := new(MockOrchestratorService)
	handler := NewRestreamsHandler(mockDB, mockSvc, logger.New("test"))
	r := chi.NewRouter()
	handler.Register(r)
	assert.NotNil(t, r)
}
