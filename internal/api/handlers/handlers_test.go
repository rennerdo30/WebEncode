package handlers

import (
	"encoding/json"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestCreateJobRequest_Serialization(t *testing.T) {
	req := CreateJobRequest{
		SourceURL: "s3://bucket/video.mp4",
		Profiles:  []string{"1080p", "720p"},
	}

	data, err := json.Marshal(req)
	assert.NoError(t, err)

	var decoded CreateJobRequest
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, req.SourceURL, decoded.SourceURL)
	assert.Equal(t, req.Profiles, decoded.Profiles)
}

func TestHealthResponse_Serialization(t *testing.T) {
	resp := HealthResponse{
		Status:  "healthy",
		Version: "1.0.0",
		Services: map[string]string{
			"database": "healthy",
			"nats":     "healthy",
		},
	}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)

	var decoded HealthResponse
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, "healthy", decoded.Status)
	assert.Equal(t, "1.0.0", decoded.Version)
	assert.Equal(t, 2, len(decoded.Services))
}

func TestSystemStats_Serialization(t *testing.T) {
	stats := SystemStats{}
	stats.Jobs.Total = 100
	stats.Jobs.Completed = 90
	stats.Jobs.Failed = 5
	stats.Workers.Total = 3
	stats.Workers.Healthy = 2
	stats.Streams.Total = 5
	stats.Streams.Live = 1

	data, err := json.Marshal(stats)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"total":100`)
}

// Router integration test helper
func newTestRouter() chi.Router {
	r := chi.NewRouter()
	return r
}

func TestRouterSetup(t *testing.T) {
	r := newTestRouter()
	assert.NotNil(t, r)
}

func TestNewJobsHandler(t *testing.T) {
	// Just test that it can be created
	handler := NewJobsHandler(nil, nil, nil)
	assert.NotNil(t, handler)
}

func TestNewSystemHandler(t *testing.T) {
	handler := NewSystemHandler(nil, nil)
	assert.NotNil(t, handler)
}

func TestNewWorkersHandler(t *testing.T) {
	handler := NewWorkersHandler(nil, nil)
	assert.NotNil(t, handler)
}
