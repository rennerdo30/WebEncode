package workers

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test HeartbeatPayload serialization
func TestHeartbeatPayload_Serialization(t *testing.T) {
	payload := HeartbeatPayload{
		ID:        "test-worker",
		Timestamp: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		Capacity: map[string]interface{}{
			"cpu_count":  8,
			"has_nvidia": true,
		},
		Status: "idle",
	}

	data, err := json.Marshal(payload)
	assert.NoError(t, err)

	var decoded HeartbeatPayload
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, payload.ID, decoded.ID)
	assert.Equal(t, payload.Status, decoded.Status)
	assert.Equal(t, float64(8), decoded.Capacity["cpu_count"]) // JSON numbers decode as float64
	assert.Equal(t, true, decoded.Capacity["has_nvidia"])
}

// Test NewHeartbeatListener
func TestNewHeartbeatListener(t *testing.T) {
	listener := NewHeartbeatListener(nil, nil, nil)
	assert.NotNil(t, listener)
}

// Test HeartbeatPayload with various status values
func TestHeartbeatPayload_StatusValues(t *testing.T) {
	validStatuses := []string{"idle", "busy", "offline", "error"}

	for _, status := range validStatuses {
		payload := HeartbeatPayload{
			ID:        "worker-1",
			Timestamp: time.Now(),
			Capacity:  map[string]interface{}{},
			Status:    status,
		}

		data, err := json.Marshal(payload)
		assert.NoError(t, err)

		var decoded HeartbeatPayload
		err = json.Unmarshal(data, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, status, decoded.Status)
	}
}

// Test HeartbeatPayload with complex capacity data
func TestHeartbeatPayload_ComplexCapacity(t *testing.T) {
	payload := HeartbeatPayload{
		ID:        "gpu-worker",
		Timestamp: time.Now(),
		Capacity: map[string]interface{}{
			"cpu_count":      16,
			"has_nvidia":     true,
			"has_vaapi":      false,
			"has_amd":        false,
			"has_intel_qsv":  false,
			"gpu_type":       "nvidia",
			"ffmpeg_version": "6.0",
		},
		Status: "idle",
	}

	data, err := json.Marshal(payload)
	assert.NoError(t, err)

	var decoded HeartbeatPayload
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, float64(16), decoded.Capacity["cpu_count"])
	assert.Equal(t, "nvidia", decoded.Capacity["gpu_type"])
	assert.Equal(t, "6.0", decoded.Capacity["ffmpeg_version"])
}

func TestHandleHeartbeat(t *testing.T) {
	mockStore := new(MockStore)
	l := logger.New("test")
	listener := NewHeartbeatListener(nil, mockStore, l)

	payload := HeartbeatPayload{
		ID:        "worker-test",
		Timestamp: time.Now(),
		Status:    "active",
		Capacity:  map[string]interface{}{"cpu": 4},
	}
	data, _ := json.Marshal(payload)

	mockStore.On("RegisterWorker", mock.Anything, mock.MatchedBy(func(arg store.RegisterWorkerParams) bool {
		return arg.ID == "worker-test" && arg.Status == "active"
	})).Return(nil)

	listener.HandleHeartbeat(context.Background(), data)

	mockStore.AssertExpectations(t)
}

func TestHandleHeartbeat_Errors(t *testing.T) {
	mockStore := new(MockStore)
	l := logger.New("test")
	listener := NewHeartbeatListener(nil, mockStore, l)

	// Invalid JSON
	listener.HandleHeartbeat(context.Background(), []byte("invalid json"))
	// Should log error but not panic

	// DB Error
	payload := HeartbeatPayload{ID: "worker-error", Status: "error"}
	data, _ := json.Marshal(payload)

	mockStore.On("RegisterWorker", mock.Anything, mock.Anything).Return(assert.AnError)

	listener.HandleHeartbeat(context.Background(), data)
	// Should log error

	mockStore.AssertExpectations(t)
}
