package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Reusing MockStore from cleanup package via duplication or improved shared mock?
// MockStore is defined in mocks_test.go

func TestWorkersHandler_ListWorkers(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewWorkersHandler(mockStore, logger.New("test"))

	req := httptest.NewRequest("GET", "/v1/workers", nil)
	w := httptest.NewRecorder()

	mockStore.On("ListWorkers", mock.Anything).Return([]store.Worker{
		{ID: "worker-1", Status: "active"},
	}, nil)

	handler.ListWorkers(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var workers []store.Worker
	json.Unmarshal(w.Body.Bytes(), &workers)
	assert.Len(t, workers, 1)
	assert.Equal(t, "worker-1", workers[0].ID)
}

func TestWorkersHandler_GetWorker(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewWorkersHandler(mockStore, logger.New("test"))

	r := chi.NewRouter()
	r.Get("/v1/workers/{id}", handler.GetWorker)

	req := httptest.NewRequest("GET", "/v1/workers/worker-1", nil)
	w := httptest.NewRecorder()

	mockStore.On("GetWorker", mock.Anything, "worker-1").Return(store.Worker{
		ID: "worker-1", Status: "active",
	}, nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var worker store.Worker
	json.Unmarshal(w.Body.Bytes(), &worker)
	assert.Equal(t, "worker-1", worker.ID)
}

func TestWorkersHandler_GetWorker_NotFound(t *testing.T) {
	mockStore := new(MockStore)
	handler := NewWorkersHandler(mockStore, logger.New("test"))

	r := chi.NewRouter()
	r.Get("/v1/workers/{id}", handler.GetWorker)

	req := httptest.NewRequest("GET", "/v1/workers/missing", nil)
	w := httptest.NewRecorder()

	mockStore.On("GetWorker", mock.Anything, "missing").Return(store.Worker{}, errors.New("not found"))

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
