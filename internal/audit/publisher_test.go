package audit

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockBus struct {
	mock.Mock
}

func (m *MockBus) Publish(ctx context.Context, subject string, data []byte) error {
	args := m.Called(ctx, subject, data)
	return args.Error(0)
}

func TestPublishUserAction(t *testing.T) {
	mockBus := new(MockBus)
	l := logger.New("test")
	pub := NewPublisher(mockBus, l)

	event := Event{
		Type:   EventTypeJobCreated,
		UserID: "user1",
		Action: "create",
	}

	mockBus.On("Publish", mock.Anything, "audit.user_action.job.created", mock.MatchedBy(func(data []byte) bool {
		var e Event
		json.Unmarshal(data, &e)
		return e.UserID == "user1" && e.Type == EventTypeJobCreated && !e.Timestamp.IsZero() && e.ID != ""
	})).Return(nil)

	err := pub.PublishUserAction(context.Background(), event)
	assert.NoError(t, err)
	mockBus.AssertExpectations(t)
}

func TestPublishSystemEvent(t *testing.T) {
	mockBus := new(MockBus)
	l := logger.New("test")
	pub := NewPublisher(mockBus, l)

	event := Event{
		Type:   EventTypeWorkerJoined,
		Action: "join",
	}

	mockBus.On("Publish", mock.Anything, "audit.system.worker.joined", mock.MatchedBy(func(data []byte) bool {
		var e Event
		json.Unmarshal(data, &e)
		return e.Type == EventTypeWorkerJoined && !e.Timestamp.IsZero()
	})).Return(nil)

	err := pub.PublishSystemEvent(context.Background(), event)
	assert.NoError(t, err)
	mockBus.AssertExpectations(t)
}

func TestPublishError(t *testing.T) {
	mockBus := new(MockBus)
	l := logger.New("test")
	pub := NewPublisher(mockBus, l)

	mockBus.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("nats error"))

	err := pub.PublishUserAction(context.Background(), Event{Type: EventTypeJobCreated})
	assert.Error(t, err)
	assert.Equal(t, "nats error", err.Error())
}

func TestConvenienceMethods(t *testing.T) {
	mockBus := new(MockBus)
	l := logger.New("test")
	pub := NewPublisher(mockBus, l)

	// JobCreated
	mockBus.On("Publish", mock.Anything, "audit.user_action.job.created", mock.MatchedBy(func(data []byte) bool {
		var e Event
		json.Unmarshal(data, &e)
		return e.UserID == "u1" && e.ResourceID == "j1" && e.Details["source_url"] == "http://src"
	})).Return(nil).Once()

	pub.JobCreated(context.Background(), "u1", "j1", "http://src")

	// JobCancelled
	mockBus.On("Publish", mock.Anything, "audit.user_action.job.cancelled", mock.MatchedBy(func(data []byte) bool {
		var e Event
		json.Unmarshal(data, &e)
		return e.Action == "cancel"
	})).Return(nil).Once()

	pub.JobCancelled(context.Background(), "u1", "j1")

	// WorkerJoined
	mockBus.On("Publish", mock.Anything, "audit.system.worker.joined", mock.MatchedBy(func(data []byte) bool {
		var e Event
		json.Unmarshal(data, &e)
		return e.ResourceID == "w1" && e.Details["hostname"] == "host1"
	})).Return(nil).Once()

	pub.WorkerJoined(context.Background(), "w1", "host1")

	// JobDeleted
	mockBus.On("Publish", mock.Anything, "audit.user_action.job.deleted", mock.MatchedBy(func(data []byte) bool {
		var e Event
		json.Unmarshal(data, &e)
		return e.Action == "delete" && e.ResourceID == "j1"
	})).Return(nil).Once()
	pub.JobDeleted(context.Background(), "u1", "j1")

	// StreamCreated
	mockBus.On("Publish", mock.Anything, "audit.user_action.stream.created", mock.MatchedBy(func(data []byte) bool {
		var e Event
		json.Unmarshal(data, &e)
		return e.Action == "create" && e.ResourceID == "s1" && e.Details["title"] == "My Stream"
	})).Return(nil).Once()
	pub.StreamCreated(context.Background(), "u1", "s1", "My Stream")

	// WorkerLeft
	mockBus.On("Publish", mock.Anything, "audit.system.worker.left", mock.MatchedBy(func(data []byte) bool {
		var e Event
		json.Unmarshal(data, &e)
		return e.Action == "leave" && e.ResourceID == "w1"
	})).Return(nil).Once()
	pub.WorkerLeft(context.Background(), "w1")

	// PluginCrashed
	mockBus.On("Publish", mock.Anything, "audit.system.plugin.crashed", mock.MatchedBy(func(data []byte) bool {
		var e Event
		json.Unmarshal(data, &e)
		return e.Action == "crash" && e.ResourceID == "p1" && e.Details["error"] == "oops"
	})).Return(nil).Once()
	pub.PluginCrashed(context.Background(), "p1", "oops")

	// UserLogin
	mockBus.On("Publish", mock.Anything, "audit.user_action.user.login", mock.MatchedBy(func(data []byte) bool {
		var e Event
		json.Unmarshal(data, &e)
		return e.Action == "login" && e.UserID == "u1" && e.ClientIP == "127.0.0.1"
	})).Return(nil).Once()
	pub.UserLogin(context.Background(), "u1", "127.0.0.1", "Mozilla/5.0")

	mockBus.AssertExpectations(t)
}
