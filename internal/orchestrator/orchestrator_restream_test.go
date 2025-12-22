package orchestrator

import (
	"context"
	"testing"

	"github.com/rennerdo30/webencode/pkg/bus"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSubmitRestream(t *testing.T) {
	mockStore := new(MockStore)
	mockBus := new(MockBus)
	orch := New(mockStore, mockBus, logger.New("test"))
	ctx := context.Background()

	restreamID := validUUID1
	restreamJob := store.RestreamJob{
		ID:                 toUUID(restreamID),
		OutputDestinations: []byte(`["rtmp://dest1", "rtmp://dest2"]`),
	}

	// Expect GetRestreamJob
	mockStore.On("GetRestreamJob", ctx, toUUID(restreamID)).Return(restreamJob, nil)

	// Expect CreateTask
	mockStore.On("CreateTask", ctx, mock.MatchedBy(func(arg store.CreateTaskParams) bool {
		return arg.JobID == toUUID(restreamID) &&
			arg.Type == store.TaskTypeRestream &&
			string(arg.Params) == string(restreamJob.OutputDestinations)
	})).Return(store.Task{ID: toUUID(validUUID2)}, nil)

	// Expect dispatch
	mockBus.On("Publish", ctx, bus.SubjectJobDispatch, mock.Anything).Return(nil)

	err := orch.SubmitRestream(ctx, restreamID)
	assert.NoError(t, err)

	mockStore.AssertExpectations(t)
	mockBus.AssertExpectations(t)
}

func TestStopRestream(t *testing.T) {
	mockStore := new(MockStore)
	mockBus := new(MockBus)
	orch := New(mockStore, mockBus, logger.New("test"))
	ctx := context.Background()

	restreamID := validUUID1

	// Mock ListTasksByJob
	tasks := []store.Task{
		{ID: toUUID(validUUID2), Status: "assigned"},
		{ID: toUUID(validUUID3), Status: "completed"},
	}
	mockStore.On("ListTasksByJob", ctx, toUUID(restreamID)).Return(tasks, nil)

	// Expect UpdateTaskStatus for assigned task only
	mockStore.On("UpdateTaskStatus", ctx, mock.MatchedBy(func(arg store.UpdateTaskStatusParams) bool {
		return arg.ID == toUUID(validUUID2) && arg.Status == "cancelled"
	})).Return(nil)

	err := orch.StopRestream(ctx, restreamID)
	assert.NoError(t, err)

	mockStore.AssertExpectations(t)
}

func TestSubmitRestream_Errors(t *testing.T) {
	mockStore := new(MockStore)
	mockBus := new(MockBus)
	orch := New(mockStore, mockBus, logger.New("test"))
	ctx := context.Background()

	// DB Error on GetRestreamJob
	mockStore.On("GetRestreamJob", ctx, mock.Anything).Return(store.RestreamJob{}, assert.AnError).Once()
	err := orch.SubmitRestream(ctx, validUUID1)
	assert.Error(t, err)

	// DB Error on CreateTask
	mockStore.On("GetRestreamJob", ctx, mock.Anything).Return(store.RestreamJob{ID: toUUID(validUUID1)}, nil)
	mockStore.On("CreateTask", ctx, mock.Anything).Return(store.Task{}, assert.AnError).Once()
	err = orch.SubmitRestream(ctx, validUUID1)
	assert.Error(t, err)

	// Invalid UUID
	err = orch.SubmitRestream(ctx, "invalid-uuid")
	assert.Error(t, err)
}

func TestStopRestream_Errors(t *testing.T) {
	mockStore := new(MockStore)
	mockBus := new(MockBus)
	orch := New(mockStore, mockBus, logger.New("test"))
	ctx := context.Background()

	// DB Error on ListTasksByJob
	mockStore.On("ListTasksByJob", ctx, mock.Anything).Return([]store.Task{}, assert.AnError).Once()
	err := orch.StopRestream(ctx, validUUID1)
	assert.Error(t, err)
}
