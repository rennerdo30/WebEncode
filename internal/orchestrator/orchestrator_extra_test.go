package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rennerdo30/webencode/pkg/bus"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandleTaskProgress_Success(t *testing.T) {
	mockStore := new(MockStore)
	mockBus := new(MockBus)
	l := logger.New("test")
	orch := New(mockStore, mockBus, l)

	ctx := context.Background()
	jobID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	taskID := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}

	task := store.Task{
		ID:    taskID,
		JobID: jobID,
		Type:  store.TaskTypeTranscode,
	}

	mockStore.On("GetTask", ctx, taskID).Return(task, nil)

	// Expect UpdateJobProgress
	mockStore.On("UpdateJobProgress", mock.Anything, mock.MatchedBy(func(arg store.UpdateJobProgressParams) bool {
		return arg.ID == jobID && arg.ProgressPct.Int32 == 50
	})).Return(nil)

	// Expect Bus Publish
	mockBus.On("Publish", mock.Anything, bus.SubjectJobEvents, mock.MatchedBy(func(data []byte) bool {
		var evt map[string]interface{}
		json.Unmarshal(data, &evt)
		return evt["event"] == "progress" && evt["progress"].(float64) == 50
	})).Return(nil)

	err := orch.HandleTaskEvent(ctx, taskID.String(), "progress", json.RawMessage(`{"percent": 50}`))
	assert.NoError(t, err)

	mockStore.AssertExpectations(t)
	mockBus.AssertExpectations(t)
}

func TestHandleTaskEvent_TaskNotFound(t *testing.T) {
	mockStore := new(MockStore)
	mockBus := new(MockBus)
	l := logger.New("test")
	orch := New(mockStore, mockBus, l)

	taskID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	mockStore.On("GetTask", context.Background(), taskID).Return(store.Task{}, errors.New("not found"))

	err := orch.HandleTaskEvent(context.Background(), taskID.String(), "completed", nil)
	assert.Error(t, err)
}

func TestHandleTaskFailed_DBError(t *testing.T) {
	mockStore := new(MockStore)
	mockBus := new(MockBus)
	l := logger.New("test")
	orch := New(mockStore, mockBus, l)

	taskID := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}
	task := store.Task{ID: taskID}

	mockStore.On("GetTask", mock.Anything, taskID).Return(task, nil)
	mockStore.On("FailTask", mock.Anything, mock.Anything).Return(assert.AnError)

	err := orch.HandleTaskEvent(context.Background(), taskID.String(), "failed", json.RawMessage(`{}`))
	assert.Error(t, err)
}
