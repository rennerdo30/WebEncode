package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rennerdo30/webencode/internal/orchestrator"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestJobsHandler_CreateJob(t *testing.T) {
	mockSvc := new(MockOrchestratorService)
	handler := NewJobsHandler(mockSvc, nil, logger.New("test"))

	// Setup Request
	reqBody := CreateJobRequest{
		SourceURL: "http://example.com/video.mp4",
		Profiles:  []string{"720p"},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/jobs", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Mock Expectation
	mockSvc.On("SubmitJob", mock.Anything, mock.MatchedBy(func(r orchestrator.JobRequest) bool {
		return r.SourceURL == "http://example.com/video.mp4"
	})).Return(&store.Job{
		ID:     pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
		Status: store.JobStatusQueued,
	}, nil)

	handler.CreateJob(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestJobsHandler_CreateJob_InvalidJSON(t *testing.T) {
	mockSvc := new(MockOrchestratorService)
	handler := NewJobsHandler(mockSvc, nil, logger.New("test"))

	req := httptest.NewRequest("POST", "/v1/jobs", bytes.NewReader([]byte("invalid-json")))
	w := httptest.NewRecorder()

	handler.CreateJob(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestJobsHandler_CreateJob_SubmitError(t *testing.T) {
	mockSvc := new(MockOrchestratorService)
	handler := NewJobsHandler(mockSvc, nil, logger.New("test"))

	reqBody := CreateJobRequest{
		SourceURL: "http://example.com/video.mp4",
		Profiles:  []string{"720p"}, // Must include profiles to pass validation
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/jobs", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mockSvc.On("SubmitJob", mock.Anything, mock.Anything).Return(nil, errors.New("submit failed"))

	handler.CreateJob(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestJobsHandler_CreateJob_InvalidURL(t *testing.T) {
	mockSvc := new(MockOrchestratorService)
	handler := NewJobsHandler(mockSvc, nil, logger.New("test"))

	reqBody := CreateJobRequest{
		SourceURL: "not-a-valid-url",
		Profiles:  []string{"720p"},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/jobs", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.CreateJob(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestJobsHandler_CreateJob_EmptyProfiles(t *testing.T) {
	mockSvc := new(MockOrchestratorService)
	handler := NewJobsHandler(mockSvc, nil, logger.New("test"))

	reqBody := CreateJobRequest{
		SourceURL: "http://example.com/video.mp4",
		Profiles:  []string{}, // Empty profiles
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/jobs", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.CreateJob(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestJobsHandler_CreateJob_InvalidSourceType(t *testing.T) {
	mockSvc := new(MockOrchestratorService)
	handler := NewJobsHandler(mockSvc, nil, logger.New("test"))

	reqBody := CreateJobRequest{
		SourceURL:  "http://example.com/video.mp4",
		SourceType: "invalid-type",
		Profiles:   []string{"720p"},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/jobs", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.CreateJob(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestJobsHandler_GetJob(t *testing.T) {
	mockSvc := new(MockOrchestratorService)
	handler := NewJobsHandler(mockSvc, nil, logger.New("test"))

	r := chi.NewRouter()
	r.Get("/v1/jobs/{id}", handler.GetJob)

	req := httptest.NewRequest("GET", "/v1/jobs/11111111-1111-1111-1111-111111111111", nil)
	w := httptest.NewRecorder()

	mockSvc.On("GetJob", mock.Anything, "11111111-1111-1111-1111-111111111111").Return(&store.Job{
		ID: pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
	}, nil)

	mockSvc.On("GetJobTasks", mock.Anything, "11111111-1111-1111-1111-111111111111").Return([]store.Task{}, nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "job")
	assert.Contains(t, response, "tasks")
}

func TestJobsHandler_GetJob_NotFound(t *testing.T) {
	mockSvc := new(MockOrchestratorService)
	handler := NewJobsHandler(mockSvc, nil, logger.New("test"))

	r := chi.NewRouter()
	r.Get("/v1/jobs/{id}", handler.GetJob)

	req := httptest.NewRequest("GET", "/v1/jobs/missing", nil)
	w := httptest.NewRecorder()

	mockSvc.On("GetJob", mock.Anything, "missing").Return(nil, errors.New("not found"))

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestJobsHandler_GetJob_NoID(t *testing.T) {
	mockSvc := new(MockOrchestratorService)
	handler := NewJobsHandler(mockSvc, nil, logger.New("test"))

	// Directly calling GetJob without URL param in context implies empty ID usually if not set via Chi
	req := httptest.NewRequest("GET", "/v1/jobs/", nil)
	w := httptest.NewRecorder()

	handler.GetJob(w, req)
	// handler checks: id := chi.URLParam(r, "id"), if empty -> invalid params
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestJobsHandler_ListJobs(t *testing.T) {
	mockSvc := new(MockOrchestratorService)
	handler := NewJobsHandler(mockSvc, nil, logger.New("test"))

	req := httptest.NewRequest("GET", "/v1/jobs?limit=5&offset=0", nil)
	w := httptest.NewRecorder()

	mockSvc.On("ListJobs", mock.Anything, int32(5), int32(0)).Return([]store.Job{
		{Status: store.JobStatusCompleted},
	}, nil)

	handler.ListJobs(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestJobsHandler_ListJobs_Error(t *testing.T) {
	mockSvc := new(MockOrchestratorService)
	handler := NewJobsHandler(mockSvc, nil, logger.New("test"))

	req := httptest.NewRequest("GET", "/v1/jobs", nil)
	w := httptest.NewRecorder()

	// Default limit 10, offset 0
	mockSvc.On("ListJobs", mock.Anything, int32(10), int32(0)).Return(nil, errors.New("db error"))

	handler.ListJobs(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestJobsHandler_ListJobs_PaginationBounds(t *testing.T) {
	mockSvc := new(MockOrchestratorService)
	handler := NewJobsHandler(mockSvc, nil, logger.New("test"))

	// Test that limit is clamped to max 100 and negative offset becomes 0
	req := httptest.NewRequest("GET", "/v1/jobs?limit=500&offset=-10", nil)
	w := httptest.NewRecorder()

	// Should be clamped to limit=100, offset=0
	mockSvc.On("ListJobs", mock.Anything, int32(100), int32(0)).Return([]store.Job{}, nil)

	handler.ListJobs(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestJobsHandler_DeleteJob(t *testing.T) {
	mockSvc := new(MockOrchestratorService)
	handler := NewJobsHandler(mockSvc, nil, logger.New("test"))

	r := chi.NewRouter()
	r.Delete("/v1/jobs/{id}", handler.DeleteJob)

	req := httptest.NewRequest("DELETE", "/v1/jobs/123", nil)
	w := httptest.NewRecorder()

	mockSvc.On("DeleteJob", mock.Anything, "123").Return(nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestJobsHandler_DeleteJob_Error(t *testing.T) {
	mockSvc := new(MockOrchestratorService)
	handler := NewJobsHandler(mockSvc, nil, logger.New("test"))

	r := chi.NewRouter()
	r.Delete("/v1/jobs/{id}", handler.DeleteJob)

	req := httptest.NewRequest("DELETE", "/v1/jobs/123", nil)
	w := httptest.NewRecorder()

	mockSvc.On("DeleteJob", mock.Anything, "123").Return(errors.New("failed"))

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestJobsHandler_CancelJob(t *testing.T) {
	mockSvc := new(MockOrchestratorService)
	handler := NewJobsHandler(mockSvc, nil, logger.New("test"))

	r := chi.NewRouter()
	r.Post("/v1/jobs/{id}/cancel", handler.CancelJob)

	req := httptest.NewRequest("POST", "/v1/jobs/123/cancel", nil)
	w := httptest.NewRecorder()

	mockSvc.On("CancelJob", mock.Anything, "123").Return(nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestJobsHandler_CancelJob_Error(t *testing.T) {
	mockSvc := new(MockOrchestratorService)
	handler := NewJobsHandler(mockSvc, nil, logger.New("test"))

	r := chi.NewRouter()
	r.Post("/v1/jobs/{id}/cancel", handler.CancelJob)

	req := httptest.NewRequest("POST", "/v1/jobs/123/cancel", nil)
	w := httptest.NewRecorder()

	mockSvc.On("CancelJob", mock.Anything, "123").Return(errors.New("failed"))

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestJobsHandler_GetJobLogs(t *testing.T) {
	mockSvc := new(MockOrchestratorService)
	handler := NewJobsHandler(mockSvc, nil, logger.New("test"))

	r := chi.NewRouter()
	r.Get("/v1/jobs/{id}/logs", handler.GetJobLogs)

	req := httptest.NewRequest("GET", "/v1/jobs/123/logs", nil)
	w := httptest.NewRecorder()

	mockSvc.On("GetJobLogs", mock.Anything, "123").Return([]store.JobLog{
		{Message: "Log 1"},
	}, nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestJobsHandler_GetJobLogs_Error(t *testing.T) {
	mockSvc := new(MockOrchestratorService)
	handler := NewJobsHandler(mockSvc, nil, logger.New("test"))

	r := chi.NewRouter()
	r.Get("/v1/jobs/{id}/logs", handler.GetJobLogs)

	req := httptest.NewRequest("GET", "/v1/jobs/123/logs", nil)
	w := httptest.NewRecorder()

	mockSvc.On("GetJobLogs", mock.Anything, "123").Return(nil, errors.New("failed"))

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
