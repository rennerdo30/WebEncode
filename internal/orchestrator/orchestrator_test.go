package orchestrator

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rennerdo30/webencode/internal/encoder"
	"github.com/rennerdo30/webencode/pkg/bus"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	validUUID1 = "11111111-1111-1111-1111-111111111111"
	validUUID2 = "22222222-2222-2222-2222-222222222222"
	validUUID3 = "33333333-3333-3333-3333-333333333333"
)

func toUUID(s string) pgtype.UUID {
	var u pgtype.UUID
	_ = u.Scan(s)
	return u
}

type MockStore struct {
	mock.Mock
}

func (m *MockStore) CreateJob(ctx context.Context, arg store.CreateJobParams) (store.Job, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(store.Job), args.Error(1)
}

func (m *MockStore) CreateTask(ctx context.Context, arg store.CreateTaskParams) (store.Task, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(store.Task), args.Error(1)
}

func (m *MockStore) GetJob(ctx context.Context, id pgtype.UUID) (store.Job, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(store.Job), args.Error(1)
}

func (m *MockStore) ListJobs(ctx context.Context, arg store.ListJobsParams) ([]store.Job, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]store.Job), args.Error(1)
}

func (m *MockStore) GetTask(ctx context.Context, id pgtype.UUID) (store.Task, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(store.Task), args.Error(1)
}

func (m *MockStore) UpdateJobStatus(ctx context.Context, arg store.UpdateJobStatusParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockStore) UpdateJobStarted(ctx context.Context, arg store.UpdateJobStartedParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockStore) UpdateJobFailed(ctx context.Context, arg store.UpdateJobFailedParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockStore) UpdateJobCompleted(ctx context.Context, id pgtype.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStore) CompleteTask(ctx context.Context, arg store.CompleteTaskParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockStore) FailTask(ctx context.Context, arg store.FailTaskParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockStore) DeleteJob(ctx context.Context, id pgtype.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStore) CancelJob(ctx context.Context, id pgtype.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStore) ListTasksByJob(ctx context.Context, jobID pgtype.UUID) ([]store.Task, error) {
	args := m.Called(ctx, jobID)
	return args.Get(0).([]store.Task), args.Error(1)
}

func (m *MockStore) CountTasksByJobAndStatus(ctx context.Context, arg store.CountTasksByJobAndStatusParams) (int64, error) {
	args := m.Called(ctx, arg)
	// Handle occasional int/int64 mismatch in Testify
	if v, ok := args.Get(0).(int); ok {
		return int64(v), args.Error(1)
	}
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStore) CountPendingTasksForJob(ctx context.Context, jobID pgtype.UUID) (int64, error) {
	args := m.Called(ctx, jobID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStore) GetCompletedTaskOutputs(ctx context.Context, jobID pgtype.UUID) ([]store.GetCompletedTaskOutputsRow, error) {
	args := m.Called(ctx, jobID)
	return args.Get(0).([]store.GetCompletedTaskOutputsRow), args.Error(1)
}

// Stubs
func (m *MockStore) UpdateJobProgress(ctx context.Context, arg store.UpdateJobProgressParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}
func (m *MockStore) GetPendingTasks(ctx context.Context, limit int32) ([]store.Task, error) {
	return nil, nil
}
func (m *MockStore) AssignTask(ctx context.Context, arg store.AssignTaskParams) error { return nil }
func (m *MockStore) RegisterWorker(ctx context.Context, arg store.RegisterWorkerParams) error {
	return nil
}
func (m *MockStore) UpdateWorkerHeartbeat(ctx context.Context, id string) error { return nil }
func (m *MockStore) CountJobs(ctx context.Context) (int64, error)               { return 0, nil }
func (m *MockStore) CountJobsByStatus(ctx context.Context, status store.JobStatus) (int64, error) {
	return 0, nil
}
func (m *MockStore) CreateAuditLog(ctx context.Context, arg store.CreateAuditLogParams) error {
	return nil
}
func (m *MockStore) CreateEncodingProfile(ctx context.Context, arg store.CreateEncodingProfileParams) (store.EncodingProfile, error) {
	return store.EncodingProfile{}, nil
}
func (m *MockStore) RegisterPluginConfig(ctx context.Context, arg store.RegisterPluginConfigParams) (store.PluginConfig, error) {
	return store.PluginConfig{}, nil
}
func (m *MockStore) CreateRestreamJob(ctx context.Context, arg store.CreateRestreamJobParams) (store.RestreamJob, error) {
	return store.RestreamJob{}, nil
}
func (m *MockStore) CreateStream(ctx context.Context, arg store.CreateStreamParams) (store.Stream, error) {
	return store.Stream{}, nil
}
func (m *MockStore) CreateWebhook(ctx context.Context, arg store.CreateWebhookParams) (store.Webhook, error) {
	return store.Webhook{}, nil
}
func (m *MockStore) DeactivateWebhook(ctx context.Context, id pgtype.UUID) error { return nil }
func (m *MockStore) DeleteEncodingProfile(ctx context.Context, id string) error  { return nil }
func (m *MockStore) DeletePluginConfig(ctx context.Context, id string) error     { return nil }
func (m *MockStore) DeleteRestreamJob(ctx context.Context, id pgtype.UUID) error { return nil }
func (m *MockStore) DeleteStream(ctx context.Context, id pgtype.UUID) error      { return nil }
func (m *MockStore) DeleteWebhook(ctx context.Context, id pgtype.UUID) error     { return nil }
func (m *MockStore) DeleteWorker(ctx context.Context, id string) error           { return nil }
func (m *MockStore) GetEncodingProfile(ctx context.Context, id string) (store.EncodingProfile, error) {
	return store.EncodingProfile{}, nil
}
func (m *MockStore) GetPluginConfig(ctx context.Context, id string) (store.PluginConfig, error) {
	return store.PluginConfig{}, nil
}
func (m *MockStore) GetRestreamJob(ctx context.Context, id pgtype.UUID) (store.RestreamJob, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(store.RestreamJob), args.Error(1)
}
func (m *MockStore) UpdateTaskStatus(ctx context.Context, arg store.UpdateTaskStatusParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}
func (m *MockStore) GetStream(ctx context.Context, id pgtype.UUID) (store.Stream, error) {
	return store.Stream{}, nil
}
func (m *MockStore) GetStreamByKey(ctx context.Context, streamKey string) (store.Stream, error) {
	return store.Stream{}, nil
}
func (m *MockStore) GetWebhook(ctx context.Context, id pgtype.UUID) (store.Webhook, error) {
	return store.Webhook{}, nil
}
func (m *MockStore) GetWorker(ctx context.Context, id string) (store.Worker, error) {
	return store.Worker{}, nil
}
func (m *MockStore) IncrementTaskAttempt(ctx context.Context, id pgtype.UUID) error { return nil }
func (m *MockStore) ListAuditLogs(ctx context.Context, arg store.ListAuditLogsParams) ([]store.AuditLog, error) {
	return nil, nil
}
func (m *MockStore) ListAuditLogsByUser(ctx context.Context, arg store.ListAuditLogsByUserParams) ([]store.AuditLog, error) {
	return nil, nil
}
func (m *MockStore) ListEncodingProfiles(ctx context.Context) ([]store.EncodingProfile, error) {
	return nil, nil
}
func (m *MockStore) ListHealthyWorkers(ctx context.Context) ([]store.Worker, error) { return nil, nil }
func (m *MockStore) ListJobsByStatus(ctx context.Context, arg store.ListJobsByStatusParams) ([]store.Job, error) {
	return nil, nil
}
func (m *MockStore) ListJobsByUser(ctx context.Context, arg store.ListJobsByUserParams) ([]store.Job, error) {
	return nil, nil
}
func (m *MockStore) ListLiveStreams(ctx context.Context) ([]store.Stream, error) { return nil, nil }
func (m *MockStore) ListPluginConfigs(ctx context.Context) ([]store.PluginConfig, error) {
	return nil, nil
}
func (m *MockStore) EnablePlugin(ctx context.Context, id string) error  { return nil }
func (m *MockStore) DisablePlugin(ctx context.Context, id string) error { return nil }
func (m *MockStore) ListPluginConfigsByType(ctx context.Context, pluginType string) ([]store.PluginConfig, error) {
	return nil, nil
}
func (m *MockStore) ListRestreamJobs(ctx context.Context, arg store.ListRestreamJobsParams) ([]store.RestreamJob, error) {
	return nil, nil
}
func (m *MockStore) ListRestreamJobsByUser(ctx context.Context, arg store.ListRestreamJobsByUserParams) ([]store.RestreamJob, error) {
	return nil, nil
}
func (m *MockStore) ListStreams(ctx context.Context, arg store.ListStreamsParams) ([]store.Stream, error) {
	return nil, nil
}
func (m *MockStore) ListStreamsByUser(ctx context.Context, arg store.ListStreamsByUserParams) ([]store.Stream, error) {
	return nil, nil
}
func (m *MockStore) ListWebhooks(ctx context.Context) ([]store.Webhook, error) { return nil, nil }
func (m *MockStore) ListWebhooksByUser(ctx context.Context, userID pgtype.UUID) ([]store.Webhook, error) {
	return nil, nil
}
func (m *MockStore) ListWorkers(ctx context.Context) ([]store.Worker, error) { return nil, nil }
func (m *MockStore) MarkWorkersUnhealthy(ctx context.Context) ([]store.Worker, error) {
	return nil, nil
}
func (m *MockStore) SetStreamArchiveJob(ctx context.Context, arg store.SetStreamArchiveJobParams) error {
	return nil
}
func (m *MockStore) UpdateEncodingProfile(ctx context.Context, arg store.UpdateEncodingProfileParams) error {
	return nil
}
func (m *MockStore) UpdatePluginConfig(ctx context.Context, arg store.UpdatePluginConfigParams) error {
	return nil
}
func (m *MockStore) UpdateRestreamJobStatus(ctx context.Context, arg store.UpdateRestreamJobStatusParams) error {
	return nil
}
func (m *MockStore) UpdateStreamLive(ctx context.Context, arg store.UpdateStreamLiveParams) error {
	return nil
}
func (m *MockStore) UpdateStreamMetadata(ctx context.Context, arg store.UpdateStreamMetadataParams) error {
	return nil
}
func (m *MockStore) UpdateStreamStats(ctx context.Context, arg store.UpdateStreamStatsParams) error {
	return nil
}
func (m *MockStore) UpdateWebhookTriggered(ctx context.Context, id pgtype.UUID) error { return nil }
func (m *MockStore) UpdateWorkerStatus(ctx context.Context, arg store.UpdateWorkerStatusParams) error {
	return nil
}
func (m *MockStore) IncrementWebhookFailure(ctx context.Context, id pgtype.UUID) error { return nil }
func (m *MockStore) ListActiveWebhooksForEvent(ctx context.Context, event string) ([]store.Webhook, error) {
	return nil, nil
}
func (m *MockStore) DeleteOldWorkers(ctx context.Context, lastSeen pgtype.Timestamptz) (int64, error) {
	return 0, nil
}

func (m *MockStore) DeleteOldJobsByStatus(ctx context.Context, arg store.DeleteOldJobsByStatusParams) (int64, error) {
	return 0, nil
}
func (m *MockStore) DeleteOrphanedTasks(ctx context.Context) (int64, error) { return 0, nil }
func (m *MockStore) DeleteOldAuditLogs(ctx context.Context, createdAt pgtype.Timestamptz) (int64, error) {
	return 0, nil
}
func (m *MockStore) CreateErrorEvent(ctx context.Context, arg store.CreateErrorEventParams) (store.ErrorEvent, error) {
	return store.ErrorEvent{}, nil
}
func (m *MockStore) ListErrorEvents(ctx context.Context, arg store.ListErrorEventsParams) ([]store.ErrorEvent, error) {
	return nil, nil
}
func (m *MockStore) ListErrorEventsBySource(ctx context.Context, arg store.ListErrorEventsBySourceParams) ([]store.ErrorEvent, error) {
	return nil, nil
}
func (m *MockStore) ResolveErrorEvent(ctx context.Context, id pgtype.UUID) error {
	return nil
}
func (m *MockStore) DeleteOldErrorEvents(ctx context.Context, createdBefore pgtype.Timestamptz) (int64, error) {
	return 0, nil
}
func (m *MockStore) CreateJobLog(ctx context.Context, arg store.CreateJobLogParams) error {
	return nil
}
func (m *MockStore) ListJobLogs(ctx context.Context, jobID pgtype.UUID) ([]store.JobLog, error) {
	return nil, nil
}
func (m *MockStore) CreateNotification(ctx context.Context, arg store.CreateNotificationParams) (store.Notification, error) {
	return store.Notification{}, nil
}
func (m *MockStore) ListNotifications(ctx context.Context, arg store.ListNotificationsParams) ([]store.Notification, error) {
	return nil, nil
}
func (m *MockStore) GetUnreadNotificationCount(ctx context.Context, userID pgtype.UUID) (int64, error) {
	return 0, nil
}
func (m *MockStore) MarkNotificationRead(ctx context.Context, arg store.MarkNotificationReadParams) error {
	return nil
}
func (m *MockStore) MarkAllNotificationsRead(ctx context.Context, userID pgtype.UUID) error {
	return nil
}
func (m *MockStore) DeleteNotification(ctx context.Context, arg store.DeleteNotificationParams) error {
	return nil
}
func (m *MockStore) DeleteOldNotifications(ctx context.Context, userID pgtype.UUID) error {
	return nil
}
func (m *MockStore) UpdateStreamRestreamDestinations(ctx context.Context, arg store.UpdateStreamRestreamDestinationsParams) error {
	return nil
}

// MockBus
type MockBus struct {
	mock.Mock
}

func (m *MockBus) Publish(ctx context.Context, subject string, data []byte) error {
	args := m.Called(ctx, subject, data)
	return args.Error(0)
}

func TestSubmitJob(t *testing.T) {
	mockStore := new(MockStore)
	mockBus := new(MockBus)
	orch := New(mockStore, mockBus, logger.New("test"))
	ctx := context.Background()

	userID := validUUID1
	sourceURL := "http://example.com/video.mp4"
	profiles := []string{"1080p"}

	expectedJob := store.Job{ID: toUUID(validUUID2), SourceUrl: sourceURL, Profiles: profiles}

	mockStore.On("CreateJob", ctx, mock.MatchedBy(func(arg store.CreateJobParams) bool {
		return arg.SourceUrl == sourceURL
	})).Return(expectedJob, nil)

	mockStore.On("CreateTask", ctx, mock.MatchedBy(func(arg store.CreateTaskParams) bool {
		return arg.Type == store.TaskTypeProbe
	})).Return(store.Task{ID: toUUID(validUUID3)}, nil)

	mockBus.On("Publish", ctx, bus.SubjectJobDispatch, mock.Anything).Return(nil)

	job, err := orch.SubmitJob(ctx, JobRequest{UserID: userID, SourceURL: sourceURL, Profiles: profiles})
	assert.NoError(t, err)
	assert.Equal(t, expectedJob.ID, job.ID)
}

func TestHandleTaskEvent_ProbeComplete(t *testing.T) {
	mockStore := new(MockStore)
	mockBus := new(MockBus)
	orch := New(mockStore, mockBus, logger.New("test"))
	ctx := context.Background()

	taskID := validUUID3
	jobID := validUUID2

	task := store.Task{
		ID:    toUUID(taskID),
		JobID: toUUID(jobID),
		Type:  store.TaskTypeProbe,
	}

	job := store.Job{
		ID:        toUUID(jobID),
		SourceUrl: "http://example.com/video.mp4",
		Profiles:  []string{"1080p_h264"},
	}

	probeResult := encoder.ProbeResult{
		Duration:  30.0,
		Keyframes: []float64{0, 10, 20},
		Width:     1920,
		Height:    1080,
	}
	resultBytes, _ := json.Marshal(probeResult)

	mockStore.On("GetTask", ctx, toUUID(taskID)).Return(task, nil)
	mockStore.On("CompleteTask", ctx, mock.Anything).Return(nil)

	mockStore.On("GetJob", ctx, toUUID(jobID)).Return(job, nil)

	mockStore.On("CreateTask", ctx, mock.MatchedBy(func(arg store.CreateTaskParams) bool {
		return arg.Type == store.TaskTypeTranscode
	})).Return(store.Task{ID: toUUID(validUUID1)}, nil).Times(3)

	mockBus.On("Publish", ctx, bus.SubjectJobDispatch, mock.Anything).Return(nil).Times(3)
	mockStore.On("UpdateJobStarted", ctx, mock.Anything).Return(nil)

	err := orch.HandleTaskEvent(ctx, taskID, "completed", resultBytes)
	assert.NoError(t, err)
}

func TestHandleTaskEvent_TranscodeComplete(t *testing.T) {
	mockStore := new(MockStore)
	mockBus := new(MockBus)
	orch := New(mockStore, mockBus, logger.New("test"))
	ctx := context.Background()

	taskID := validUUID3
	jobID := validUUID2

	task := store.Task{
		ID:    toUUID(taskID),
		JobID: toUUID(jobID),
		Type:  store.TaskTypeTranscode,
	}

	mockStore.On("GetTask", ctx, toUUID(taskID)).Return(task, nil)
	mockStore.On("CompleteTask", ctx, mock.Anything).Return(nil)

	// Case 1: More tasks pending
	mockStore.On("CountTasksByJobAndStatus", ctx, mock.MatchedBy(func(arg store.CountTasksByJobAndStatusParams) bool {
		return arg.Status == "pending"
	})).Return(int64(1), nil).Once()

	mockStore.On("CountTasksByJobAndStatus", ctx, mock.MatchedBy(func(arg store.CountTasksByJobAndStatusParams) bool {
		return arg.Status == "assigned"
	})).Return(int64(0), nil).Once()

	err := orch.HandleTaskEvent(ctx, taskID, "completed", []byte("{}"))
	assert.NoError(t, err)
}

func TestHandleTaskEvent_TranscodeComplete_AllDone(t *testing.T) {
	mockStore := new(MockStore)
	mockBus := new(MockBus)
	orch := New(mockStore, mockBus, logger.New("test"))
	ctx := context.Background()

	taskID := validUUID3
	jobID := validUUID2

	task := store.Task{
		ID:    toUUID(taskID),
		JobID: toUUID(jobID),
		Type:  store.TaskTypeTranscode,
	}

	mockStore.On("GetTask", ctx, toUUID(taskID)).Return(task, nil)
	mockStore.On("CompleteTask", ctx, mock.Anything).Return(nil)

	// Case 2: All done
	mockStore.On("CountTasksByJobAndStatus", ctx, mock.MatchedBy(func(arg store.CountTasksByJobAndStatusParams) bool {
		return arg.Status == "pending"
	})).Return(int64(0), nil).Once()

	mockStore.On("CountTasksByJobAndStatus", ctx, mock.MatchedBy(func(arg store.CountTasksByJobAndStatusParams) bool {
		return arg.Status == "assigned"
	})).Return(int64(0), nil).Once()

	mockStore.On("GetCompletedTaskOutputs", ctx, toUUID(jobID)).Return([]store.GetCompletedTaskOutputsRow{
		{OutputKey: pgtype.Text{String: "out1.mp4", Valid: true}},
	}, nil)

	mockStore.On("CreateTask", ctx, mock.MatchedBy(func(arg store.CreateTaskParams) bool {
		return arg.Type == store.TaskTypeStitch
	})).Return(store.Task{ID: toUUID(validUUID1)}, nil)

	mockStore.On("UpdateJobStatus", ctx, mock.Anything).Return(nil)
	mockBus.On("Publish", ctx, bus.SubjectJobDispatch, mock.Anything).Return(nil)

	err := orch.HandleTaskEvent(ctx, taskID, "completed", []byte("{}"))
	assert.NoError(t, err)
}

func TestCRUD(t *testing.T) {
	mockStore := new(MockStore)
	mockBus := new(MockBus)
	orch := New(mockStore, mockBus, logger.New("test"))
	ctx := context.Background()

	id := validUUID1

	// GetJob
	mockStore.On("GetJob", ctx, toUUID(id)).Return(store.Job{ID: toUUID(id)}, nil)
	job, err := orch.GetJob(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, toUUID(id), job.ID)

	// DeleteJob
	mockStore.On("DeleteJob", ctx, toUUID(id)).Return(nil)
	err = orch.DeleteJob(ctx, id)
	assert.NoError(t, err)

	// CancelJob
	mockStore.On("CancelJob", ctx, toUUID(id)).Return(nil)
	mockBus.On("Publish", ctx, bus.SubjectJobEvents, mock.Anything).Return(nil)
	err = orch.CancelJob(ctx, id)
	assert.NoError(t, err)

	// ListJobs
	mockStore.On("ListJobs", ctx, mock.Anything).Return([]store.Job{{ID: toUUID(id)}}, nil)
	jobs, err := orch.ListJobs(ctx, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, jobs, 1)

	// GetJobTasks
	mockStore.On("ListTasksByJob", ctx, toUUID(id)).Return([]store.Task{{ID: toUUID(validUUID2)}}, nil)
	tasks, err := orch.GetJobTasks(ctx, id)
	assert.NoError(t, err)
	assert.Len(t, tasks, 1)
}

func TestHandleTaskFailed(t *testing.T) {
	mockStore := new(MockStore)
	mockBus := new(MockBus)
	orch := New(mockStore, mockBus, logger.New("test"))
	ctx := context.Background()

	taskID := validUUID3
	jobID := validUUID2

	task := store.Task{
		ID:    toUUID(taskID),
		JobID: toUUID(jobID),
		Type:  store.TaskTypeProbe,
	}

	mockStore.On("GetTask", ctx, toUUID(taskID)).Return(task, nil)
	mockStore.On("FailTask", ctx, mock.Anything).Return(nil)
	mockStore.On("UpdateJobFailed", ctx, mock.Anything).Return(nil)

	err := orch.HandleTaskEvent(ctx, taskID, "failed", []byte(`"error"`))
	assert.NoError(t, err)
}

func TestHandleTaskEvent_StitchComplete(t *testing.T) {
	mockStore := new(MockStore)
	mockBus := new(MockBus)
	orch := New(mockStore, mockBus, logger.New("test"))
	ctx := context.Background()

	taskID := validUUID3
	jobID := validUUID2

	task := store.Task{
		ID:    toUUID(taskID),
		JobID: toUUID(jobID),
		Type:  store.TaskTypeStitch,
	}

	mockStore.On("GetTask", ctx, toUUID(taskID)).Return(task, nil)
	mockStore.On("CompleteTask", ctx, mock.Anything).Return(nil)
	mockStore.On("UpdateJobCompleted", ctx, toUUID(jobID)).Return(nil)

	err := orch.HandleTaskEvent(ctx, taskID, "completed", []byte("{}"))
	assert.NoError(t, err)
}
