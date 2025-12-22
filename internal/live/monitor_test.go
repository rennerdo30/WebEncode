package live

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rennerdo30/webencode/internal/plugin_manager"
	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

// MockQuerier implements store.Querier
type MockQuerier struct {
	mock.Mock
}

// Implement minimal methods for test
func (m *MockQuerier) ListStreams(ctx context.Context, arg store.ListStreamsParams) ([]store.Stream, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]store.Stream), args.Error(1)
}

// Stub other methods... in a real test we'd generate this with mockery
// For this single test file, we can embed unimplemented to avoid implementing all methods
type StubQuerier struct {
	MockQuerier
}

// Need to implement all methods effectively for interface satisfaction,
// creating a stub wrapper is cleaner or using a full mock generator.
// For brevity, I'll just rely on the interface satisfaction check failing if I miss methods.
// I'll define ALL methods as stubs to satisfy the interface.
func (m *MockQuerier) AssignTask(ctx context.Context, arg store.AssignTaskParams) error { return nil }
func (m *MockQuerier) CancelJob(ctx context.Context, id pgtype.UUID) error              { return nil }
func (m *MockQuerier) CompleteTask(ctx context.Context, arg store.CompleteTaskParams) error {
	return nil
}
func (m *MockQuerier) CountJobs(ctx context.Context) (int64, error) { return 0, nil }
func (m *MockQuerier) CountJobsByStatus(ctx context.Context, status store.JobStatus) (int64, error) {
	return 0, nil
}
func (m *MockQuerier) CountPendingTasksForJob(ctx context.Context, jobID pgtype.UUID) (int64, error) {
	return 0, nil
}
func (m *MockQuerier) CountTasksByJobAndStatus(ctx context.Context, arg store.CountTasksByJobAndStatusParams) (int64, error) {
	return 0, nil
}
func (m *MockQuerier) CreateAuditLog(ctx context.Context, arg store.CreateAuditLogParams) error {
	return nil
}
func (m *MockQuerier) CreateEncodingProfile(ctx context.Context, arg store.CreateEncodingProfileParams) (store.EncodingProfile, error) {
	return store.EncodingProfile{}, nil
}
func (m *MockQuerier) CreateJob(ctx context.Context, arg store.CreateJobParams) (store.Job, error) {
	return store.Job{}, nil
}
func (m *MockQuerier) RegisterPluginConfig(ctx context.Context, arg store.RegisterPluginConfigParams) (store.PluginConfig, error) {
	return store.PluginConfig{}, nil
}
func (m *MockQuerier) DeleteOldWorkers(ctx context.Context, lastSeen pgtype.Timestamptz) (int64, error) {
	return 0, nil
}
func (m *MockQuerier) CreateRestreamJob(ctx context.Context, arg store.CreateRestreamJobParams) (store.RestreamJob, error) {
	return store.RestreamJob{}, nil
}
func (m *MockQuerier) CreateStream(ctx context.Context, arg store.CreateStreamParams) (store.Stream, error) {
	return store.Stream{}, nil
}
func (m *MockQuerier) CreateTask(ctx context.Context, arg store.CreateTaskParams) (store.Task, error) {
	return store.Task{}, nil
}
func (m *MockQuerier) CreateWebhook(ctx context.Context, arg store.CreateWebhookParams) (store.Webhook, error) {
	return store.Webhook{}, nil
}
func (m *MockQuerier) DeactivateWebhook(ctx context.Context, id pgtype.UUID) error { return nil }
func (m *MockQuerier) DeleteEncodingProfile(ctx context.Context, id string) error  { return nil }
func (m *MockQuerier) DeleteJob(ctx context.Context, id pgtype.UUID) error         { return nil }
func (m *MockQuerier) DeleteOldAuditLogs(ctx context.Context, createdAt pgtype.Timestamptz) (int64, error) {
	return 0, nil
}
func (m *MockQuerier) DeleteOldJobsByStatus(ctx context.Context, arg store.DeleteOldJobsByStatusParams) (int64, error) {
	return 0, nil
}
func (m *MockQuerier) DeleteOrphanedTasks(ctx context.Context) (int64, error)       { return 0, nil }
func (m *MockQuerier) DeletePluginConfig(ctx context.Context, id string) error      { return nil }
func (m *MockQuerier) DeleteRestreamJob(ctx context.Context, id pgtype.UUID) error  { return nil }
func (m *MockQuerier) DeleteStream(ctx context.Context, id pgtype.UUID) error       { return nil }
func (m *MockQuerier) DeleteWebhook(ctx context.Context, id pgtype.UUID) error      { return nil }
func (m *MockQuerier) DeleteWorker(ctx context.Context, id string) error            { return nil }
func (m *MockQuerier) DisablePlugin(ctx context.Context, id string) error           { return nil }
func (m *MockQuerier) EnablePlugin(ctx context.Context, id string) error            { return nil }
func (m *MockQuerier) FailTask(ctx context.Context, arg store.FailTaskParams) error { return nil }
func (m *MockQuerier) GetCompletedTaskOutputs(ctx context.Context, jobID pgtype.UUID) ([]store.GetCompletedTaskOutputsRow, error) {
	return nil, nil
}
func (m *MockQuerier) GetEncodingProfile(ctx context.Context, id string) (store.EncodingProfile, error) {
	return store.EncodingProfile{}, nil
}
func (m *MockQuerier) GetJob(ctx context.Context, id pgtype.UUID) (store.Job, error) {
	return store.Job{}, nil
}
func (m *MockQuerier) GetPendingTasks(ctx context.Context, limit int32) ([]store.Task, error) {
	return nil, nil
}
func (m *MockQuerier) GetPluginConfig(ctx context.Context, id string) (store.PluginConfig, error) {
	return store.PluginConfig{}, nil
}
func (m *MockQuerier) GetRestreamJob(ctx context.Context, id pgtype.UUID) (store.RestreamJob, error) {
	return store.RestreamJob{}, nil
}
func (m *MockQuerier) GetStream(ctx context.Context, id pgtype.UUID) (store.Stream, error) {
	return store.Stream{}, nil
}
func (m *MockQuerier) GetStreamByKey(ctx context.Context, streamKey string) (store.Stream, error) {
	return store.Stream{}, nil
}
func (m *MockQuerier) GetTask(ctx context.Context, id pgtype.UUID) (store.Task, error) {
	return store.Task{}, nil
}
func (m *MockQuerier) GetWebhook(ctx context.Context, id pgtype.UUID) (store.Webhook, error) {
	return store.Webhook{}, nil
}
func (m *MockQuerier) GetWorker(ctx context.Context, id string) (store.Worker, error) {
	return store.Worker{}, nil
}
func (m *MockQuerier) IncrementTaskAttempt(ctx context.Context, id pgtype.UUID) error    { return nil }
func (m *MockQuerier) IncrementWebhookFailure(ctx context.Context, id pgtype.UUID) error { return nil }
func (m *MockQuerier) ListActiveWebhooksForEvent(ctx context.Context, event string) ([]store.Webhook, error) {
	return nil, nil
}
func (m *MockQuerier) ListAuditLogs(ctx context.Context, arg store.ListAuditLogsParams) ([]store.AuditLog, error) {
	return nil, nil
}
func (m *MockQuerier) ListAuditLogsByUser(ctx context.Context, arg store.ListAuditLogsByUserParams) ([]store.AuditLog, error) {
	return nil, nil
}
func (m *MockQuerier) ListEncodingProfiles(ctx context.Context) ([]store.EncodingProfile, error) {
	return nil, nil
}
func (m *MockQuerier) ListHealthyWorkers(ctx context.Context) ([]store.Worker, error) {
	return nil, nil
}
func (m *MockQuerier) ListJobs(ctx context.Context, arg store.ListJobsParams) ([]store.Job, error) {
	return nil, nil
}
func (m *MockQuerier) ListJobsByStatus(ctx context.Context, arg store.ListJobsByStatusParams) ([]store.Job, error) {
	return nil, nil
}
func (m *MockQuerier) ListJobsByUser(ctx context.Context, arg store.ListJobsByUserParams) ([]store.Job, error) {
	return nil, nil
}
func (m *MockQuerier) ListLiveStreams(ctx context.Context) ([]store.Stream, error) { return nil, nil }
func (m *MockQuerier) ListPluginConfigs(ctx context.Context) ([]store.PluginConfig, error) {
	return nil, nil
}
func (m *MockQuerier) ListPluginConfigsByType(ctx context.Context, pluginType string) ([]store.PluginConfig, error) {
	return nil, nil
}
func (m *MockQuerier) ListRestreamJobs(ctx context.Context, arg store.ListRestreamJobsParams) ([]store.RestreamJob, error) {
	return nil, nil
}
func (m *MockQuerier) ListRestreamJobsByUser(ctx context.Context, arg store.ListRestreamJobsByUserParams) ([]store.RestreamJob, error) {
	return nil, nil
}
func (m *MockQuerier) ListStreamsByUser(ctx context.Context, arg store.ListStreamsByUserParams) ([]store.Stream, error) {
	return nil, nil
}
func (m *MockQuerier) ListTasksByJob(ctx context.Context, jobID pgtype.UUID) ([]store.Task, error) {
	return nil, nil
}
func (m *MockQuerier) ListWebhooks(ctx context.Context) ([]store.Webhook, error) { return nil, nil }
func (m *MockQuerier) ListWebhooksByUser(ctx context.Context, userID pgtype.UUID) ([]store.Webhook, error) {
	return nil, nil
}
func (m *MockQuerier) ListWorkers(ctx context.Context) ([]store.Worker, error) { return nil, nil }
func (m *MockQuerier) MarkWorkersUnhealthy(ctx context.Context) ([]store.Worker, error) {
	return nil, nil
}
func (m *MockQuerier) RegisterWorker(ctx context.Context, arg store.RegisterWorkerParams) error {
	return nil
}
func (m *MockQuerier) SetStreamArchiveJob(ctx context.Context, arg store.SetStreamArchiveJobParams) error {
	return nil
}
func (m *MockQuerier) UpdateEncodingProfile(ctx context.Context, arg store.UpdateEncodingProfileParams) error {
	return nil
}
func (m *MockQuerier) UpdateJobCompleted(ctx context.Context, id pgtype.UUID) error { return nil }
func (m *MockQuerier) UpdateJobFailed(ctx context.Context, arg store.UpdateJobFailedParams) error {
	return nil
}
func (m *MockQuerier) UpdateJobProgress(ctx context.Context, arg store.UpdateJobProgressParams) error {
	return nil
}
func (m *MockQuerier) UpdateJobStarted(ctx context.Context, arg store.UpdateJobStartedParams) error {
	return nil
}
func (m *MockQuerier) UpdateJobStatus(ctx context.Context, arg store.UpdateJobStatusParams) error {
	return nil
}
func (m *MockQuerier) UpdatePluginConfig(ctx context.Context, arg store.UpdatePluginConfigParams) error {
	return nil
}
func (m *MockQuerier) UpdateRestreamJobStatus(ctx context.Context, arg store.UpdateRestreamJobStatusParams) error {
	return nil
}
func (m *MockQuerier) UpdateStreamLive(ctx context.Context, arg store.UpdateStreamLiveParams) error {
	return nil
}
func (m *MockQuerier) UpdateStreamMetadata(ctx context.Context, arg store.UpdateStreamMetadataParams) error {
	return nil
}
func (m *MockQuerier) UpdateStreamStats(ctx context.Context, arg store.UpdateStreamStatsParams) error {
	return nil
}
func (m *MockQuerier) UpdateTaskStatus(ctx context.Context, arg store.UpdateTaskStatusParams) error {
	return nil
}
func (m *MockQuerier) UpdateWebhookTriggered(ctx context.Context, id pgtype.UUID) error { return nil }
func (m *MockQuerier) UpdateWorkerHeartbeat(ctx context.Context, id string) error       { return nil }
func (m *MockQuerier) UpdateWorkerStatus(ctx context.Context, arg store.UpdateWorkerStatusParams) error {
	return nil
}
func (m *MockQuerier) CreateErrorEvent(ctx context.Context, arg store.CreateErrorEventParams) (store.ErrorEvent, error) {
	return store.ErrorEvent{}, nil
}
func (m *MockQuerier) ListErrorEvents(ctx context.Context, arg store.ListErrorEventsParams) ([]store.ErrorEvent, error) {
	return nil, nil
}
func (m *MockQuerier) ListErrorEventsBySource(ctx context.Context, arg store.ListErrorEventsBySourceParams) ([]store.ErrorEvent, error) {
	return nil, nil
}
func (m *MockQuerier) ResolveErrorEvent(ctx context.Context, id pgtype.UUID) error {
	return nil
}
func (m *MockQuerier) DeleteOldErrorEvents(ctx context.Context, createdBefore pgtype.Timestamptz) (int64, error) {
	return 0, nil
}
func (m *MockQuerier) CreateJobLog(ctx context.Context, arg store.CreateJobLogParams) error {
	return nil
}
func (m *MockQuerier) ListJobLogs(ctx context.Context, jobID pgtype.UUID) ([]store.JobLog, error) {
	return nil, nil
}
func (m *MockQuerier) CreateNotification(ctx context.Context, arg store.CreateNotificationParams) (store.Notification, error) {
	return store.Notification{}, nil
}
func (m *MockQuerier) ListNotifications(ctx context.Context, arg store.ListNotificationsParams) ([]store.Notification, error) {
	return nil, nil
}
func (m *MockQuerier) GetUnreadNotificationCount(ctx context.Context, userID pgtype.UUID) (int64, error) {
	return 0, nil
}
func (m *MockQuerier) MarkNotificationRead(ctx context.Context, arg store.MarkNotificationReadParams) error {
	return nil
}
func (m *MockQuerier) MarkAllNotificationsRead(ctx context.Context, userID pgtype.UUID) error {
	return nil
}
func (m *MockQuerier) DeleteNotification(ctx context.Context, arg store.DeleteNotificationParams) error {
	return nil
}
func (m *MockQuerier) DeleteOldNotifications(ctx context.Context, userID pgtype.UUID) error {
	return nil
}
func (m *MockQuerier) UpdateStreamRestreamDestinations(ctx context.Context, arg store.UpdateStreamRestreamDestinationsParams) error {
	return nil
}

// MockLivePlugin implements LiveServiceClient
type MockLivePlugin struct {
	mock.Mock
}

func (m *MockLivePlugin) StartIngest(ctx context.Context, in *pb.IngestConfig, opts ...grpc.CallOption) (*pb.IngestSession, error) {
	return nil, nil
}
func (m *MockLivePlugin) StopIngest(ctx context.Context, in *pb.SessionID, opts ...grpc.CallOption) (*pb.Empty, error) {
	return nil, nil
}
func (m *MockLivePlugin) GetTelemetry(ctx context.Context, in *pb.SessionID, opts ...grpc.CallOption) (*pb.Telemetry, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.Telemetry), args.Error(1)
}
func (m *MockLivePlugin) AddOutputTarget(ctx context.Context, in *pb.AddOutputTargetRequest, opts ...grpc.CallOption) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}
func (m *MockLivePlugin) RemoveOutputTarget(ctx context.Context, in *pb.RemoveOutputTargetRequest, opts ...grpc.CallOption) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}

// MockPublisher implements live.Publisher
type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) Publish(ctx context.Context, subject string, data []byte) error {
	args := m.Called(ctx, subject, data)
	return args.Error(0)
}

func TestPoll(t *testing.T) {
	// Setup Mocks
	mockDB := &MockQuerier{}
	mockPub := &MockPublisher{}
	pm := plugin_manager.New(logger.New("test"), "")
	l := logger.New("test")

	mockPlugin := &MockLivePlugin{}
	pm.Live["live-mediamtx"] = mockPlugin

	svc := NewMonitorService(mockDB, mockPub, pm, l)

	// Mock DB response - stream is already marked as live in DB
	// so no transition occurs (wasLive will be false but stream.IsLive is true)
	streamID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}

	// Mock stream - set IsLive: true so DB update branch is skipped
	streams := []store.Stream{
		{
			ID:        streamID,
			StreamKey: "test-key",
			IsLive:    true, // Already live in DB
		},
	}

	// Pre-populate activeStreams to prevent stream.started event
	// This simulates a stream that was already detected as live
	svc.activeStreams[streamID.String()] = true

	mockDB.On("ListStreams", mock.Anything, mock.Anything).Return(streams, nil)

	// Mock Plugin response
	mockPlugin.On("GetTelemetry", mock.Anything, &pb.SessionID{Id: "test-key"}).Return(&pb.Telemetry{
		IsLive:  true,
		Bitrate: 5000,
		Fps:     60,
		Viewers: 100,
	}, nil)

	// Expect Publish for telemetry only (no stream.started since stream was already active)
	expectedSubject := "live.telemetry." + streamID.String()

	mockPub.On("Publish", mock.Anything, expectedSubject, mock.MatchedBy(func(data []byte) bool {
		var payload map[string]interface{}
		json.Unmarshal(data, &payload)
		return payload["fps"] == 60.0 && payload["bitrate"] == 5000.0
	})).Return(nil)

	// Run poll once
	svc.poll()

	mockPub.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	mockPlugin.AssertExpectations(t)
}

func TestPoll_ErrorCases(t *testing.T) {
	mockDB := &MockQuerier{}
	mockPub := &MockPublisher{}
	pm := plugin_manager.New(logger.New("test"), "")
	l := logger.New("test")
	svc := NewMonitorService(mockDB, mockPub, pm, l)

	// Case 1: DB Error
	mockDB.On("ListStreams", mock.Anything, mock.Anything).Return([]store.Stream{}, fmt.Errorf("db error")).Once()
	svc.poll()

	// Case 2: No active streams
	mockDB.On("ListStreams", mock.Anything, mock.Anything).Return([]store.Stream{}, nil).Once()
	svc.poll()

	// Case 3: Stream not live
	mockDB.On("ListStreams", mock.Anything, mock.Anything).Return([]store.Stream{{IsLive: false}}, nil).Once()
	svc.poll()

	// Case 4: No plugin found
	streamID := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}
	mockDB.On("ListStreams", mock.Anything, mock.Anything).Return([]store.Stream{{ID: streamID, IsLive: true}}, nil).Once()
	// pm.Live is empty
	svc.poll()

	// Case 5: Plugin Error
	mockPlugin := &MockLivePlugin{}
	pm.Live["live-mediamtx"] = mockPlugin
	mockDB.On("ListStreams", mock.Anything, mock.Anything).Return([]store.Stream{{ID: streamID, StreamKey: "k1", IsLive: true}}, nil).Once()
	mockPlugin.On("GetTelemetry", mock.Anything, &pb.SessionID{Id: "k1"}).Return(&pb.Telemetry{}, fmt.Errorf("plugin fail")).Once()
	svc.poll()

	// Case 6: Telemetry not live
	mockDB.On("ListStreams", mock.Anything, mock.Anything).Return([]store.Stream{{ID: streamID, StreamKey: "k2", IsLive: true}}, nil).Once()
	mockPlugin.On("GetTelemetry", mock.Anything, &pb.SessionID{Id: "k2"}).Return(&pb.Telemetry{IsLive: false}, nil).Once()
	svc.poll()

	mockDB.AssertExpectations(t)
}

func TestStartStop(t *testing.T) {
	mockDB := &MockQuerier{}
	mockPub := &MockPublisher{}
	pm := plugin_manager.New(logger.New("test"), "")
	l := logger.New("test")
	svc := NewMonitorService(mockDB, mockPub, pm, l)

	// Mock DB to return empty streams so no processing happens if poll runs
	mockDB.On("ListStreams", mock.Anything, mock.Anything).Return([]store.Stream{}, nil).Maybe()

	// Test Start
	assert.False(t, svc.running, "Service should not be running initially")

	svc.Start()
	assert.True(t, svc.running, "Service should be running after Start")

	// Call Start again should be no-op
	svc.Start()
	assert.True(t, svc.running, "Service should still be running after duplicate Start")

	// Stop immediately to prevent any polling
	svc.Stop()
	assert.False(t, svc.running, "Service should not be running after Stop")

	// Call Stop again should be no-op
	svc.Stop()
	assert.False(t, svc.running, "Service should still not be running after duplicate Stop")
}

func TestStreamTransition(t *testing.T) {
	// Test that stream start/end transitions trigger the correct events
	mockDB := &MockQuerier{}
	mockPub := &MockPublisher{}
	pm := plugin_manager.New(logger.New("test"), "")
	l := logger.New("test")

	mockPlugin := &MockLivePlugin{}
	pm.Live["live-mediamtx"] = mockPlugin

	svc := NewMonitorService(mockDB, mockPub, pm, l)

	streamID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	userID := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}

	streams := []store.Stream{
		{
			ID:        streamID,
			UserID:    userID,
			StreamKey: "test-key",
			IsLive:    false, // Not live in DB yet
		},
	}

	mockDB.On("ListStreams", mock.Anything, mock.Anything).Return(streams, nil)
	mockDB.On("UpdateStreamLive", mock.Anything, mock.Anything).Return(nil).Maybe()
	mockDB.On("CreateNotification", mock.Anything, mock.Anything).Return(store.Notification{}, nil).Maybe()

	// Plugin says stream is now live - this triggers a transition
	mockPlugin.On("GetTelemetry", mock.Anything, &pb.SessionID{Id: "test-key"}).Return(&pb.Telemetry{
		IsLive:  true,
		Bitrate: 5000,
		Fps:     60,
		Viewers: 100,
	}, nil)

	// Expect both telemetry and stream.started events
	mockPub.On("Publish", mock.Anything, mock.MatchedBy(func(subject string) bool {
		return subject == "live.telemetry."+streamID.String() || subject == "events.stream.started"
	}), mock.Anything).Return(nil)

	// Run poll once
	svc.poll()

	// Give goroutine time to run handleStreamStart
	time.Sleep(100 * time.Millisecond)

	// Verify stream is now tracked as live
	assert.True(t, svc.activeStreams[streamID.String()], "Stream should be tracked as live")
}
