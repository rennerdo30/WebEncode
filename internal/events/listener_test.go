package events

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rennerdo30/webencode/internal/orchestrator"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- MockStore ---

type MockStore struct {
	mock.Mock
}

func (m *MockStore) UpdateTaskStatus(ctx context.Context, arg store.UpdateTaskStatusParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockStore) GetTask(ctx context.Context, id pgtype.UUID) (store.Task, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(store.Task), args.Error(1)
}

func (m *MockStore) CountPendingTasksForJob(ctx context.Context, jobID pgtype.UUID) (int64, error) {
	args := m.Called(ctx, jobID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStore) UpdateJobStatus(ctx context.Context, arg store.UpdateJobStatusParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockStore) CreateErrorEvent(ctx context.Context, arg store.CreateErrorEventParams) (store.ErrorEvent, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(store.ErrorEvent), args.Error(1)
}

// Stubs for interface compliance (EventListener only uses methods above)
func (m *MockStore) CreateJob(ctx context.Context, arg store.CreateJobParams) (store.Job, error) {
	return store.Job{}, nil
}
func (m *MockStore) GetJob(ctx context.Context, id pgtype.UUID) (store.Job, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(store.Job), args.Error(1)
}
func (m *MockStore) CreateNotification(ctx context.Context, arg store.CreateNotificationParams) (store.Notification, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(store.Notification), args.Error(1)
}
func (m *MockStore) ListJobs(ctx context.Context, arg store.ListJobsParams) ([]store.Job, error) {
	return nil, nil
}
func (m *MockStore) UpdateJobProgress(ctx context.Context, arg store.UpdateJobProgressParams) error {
	return nil
}
func (m *MockStore) DeleteJob(ctx context.Context, id pgtype.UUID) error { return nil }
func (m *MockStore) CreateTask(ctx context.Context, arg store.CreateTaskParams) (store.Task, error) {
	return store.Task{}, nil
}
func (m *MockStore) ListTasksByJob(ctx context.Context, jobID pgtype.UUID) ([]store.Task, error) {
	return nil, nil
}

// ... (Adding minimal stubs to satisfy store.Querier interface if needed by the compiler,
// using a struct embedding would be better but explicit is safe for now.
// Assuming the interface is large, we might need a helper or just implement the required ones if passing strict interface)
// Since Go interfaces are implicit, we only strictly need to implement methods CALLED if we cast to interface.
// However, NewEventListener takes store.Querier interface. So we MUST implement all.
// To save space, let's embed a type that implements all with panic/nil if possible, or list them.
// Given strict compiler check, I will try to implement necessary ones and rely on the fact that existing mocks_test.go implementations suggest the interface is large.
// PROPOSAL: I will modify the test to NOT implement the full interface manually if possible, or use a generate mock.
// Since I cannot run mockgen, I will implement a minimal MockStore that satisfies `store.Querier` by embedding a struct that implements it or listing all methods.
// Inspecting `internal/api/handlers/mocks_test.go` showed a lot of methods.
// Strategy: define `type StubQuerier struct{}` with all methods returning nil/empty, embed it in MockStore.
// I don't have the full list of methods easily copy-pasteable without reading `store/querier.go`.
// Use previous `view_file` of `internal/api/handlers/mocks_test.go` as reference. It had ~260 lines.
// I will copy the stubs from there to ensure compilation.

func (m *MockStore) AssignTask(ctx context.Context, arg store.AssignTaskParams) error     { return nil }
func (m *MockStore) CancelJob(ctx context.Context, id pgtype.UUID) error                  { return nil }
func (m *MockStore) CompleteTask(ctx context.Context, arg store.CompleteTaskParams) error { return nil }
func (m *MockStore) CountJobs(ctx context.Context) (int64, error)                         { return 0, nil }
func (m *MockStore) CountJobsByStatus(ctx context.Context, status store.JobStatus) (int64, error) {
	return 0, nil
}
func (m *MockStore) CountTasksByJobAndStatus(ctx context.Context, arg store.CountTasksByJobAndStatusParams) (int64, error) {
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
func (m *MockStore) DeleteOldWorkers(ctx context.Context, lastSeen pgtype.Timestamptz) (int64, error) {
	return 0, nil
}
func (m *MockStore) DeletePluginConfig(ctx context.Context, id string) error      { return nil }
func (m *MockStore) DeleteRestreamJob(ctx context.Context, id pgtype.UUID) error  { return nil }
func (m *MockStore) DeleteStream(ctx context.Context, id pgtype.UUID) error       { return nil }
func (m *MockStore) DeleteWebhook(ctx context.Context, id pgtype.UUID) error      { return nil }
func (m *MockStore) DeleteWorker(ctx context.Context, id string) error            { return nil }
func (m *MockStore) DisablePlugin(ctx context.Context, id string) error           { return nil }
func (m *MockStore) EnablePlugin(ctx context.Context, id string) error            { return nil }
func (m *MockStore) FailTask(ctx context.Context, arg store.FailTaskParams) error { return nil }
func (m *MockStore) GetCompletedTaskOutputs(ctx context.Context, jobID pgtype.UUID) ([]store.GetCompletedTaskOutputsRow, error) {
	return nil, nil
}
func (m *MockStore) GetEncodingProfile(ctx context.Context, id string) (store.EncodingProfile, error) {
	return store.EncodingProfile{}, nil
}
func (m *MockStore) GetPendingTasks(ctx context.Context, limit int32) ([]store.Task, error) {
	return nil, nil
}
func (m *MockStore) GetPluginConfig(ctx context.Context, id string) (store.PluginConfig, error) {
	return store.PluginConfig{}, nil
}
func (m *MockStore) GetRestreamJob(ctx context.Context, id pgtype.UUID) (store.RestreamJob, error) {
	return store.RestreamJob{}, nil
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
func (m *MockStore) IncrementTaskAttempt(ctx context.Context, id pgtype.UUID) error    { return nil }
func (m *MockStore) IncrementWebhookFailure(ctx context.Context, id pgtype.UUID) error { return nil }
func (m *MockStore) ListActiveWebhooksForEvent(ctx context.Context, event string) ([]store.Webhook, error) {
	return nil, nil
}
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
func (m *MockStore) RegisterWorker(ctx context.Context, arg store.RegisterWorkerParams) error {
	return nil
}
func (m *MockStore) SetStreamArchiveJob(ctx context.Context, arg store.SetStreamArchiveJobParams) error {
	return nil
}
func (m *MockStore) UpdateEncodingProfile(ctx context.Context, arg store.UpdateEncodingProfileParams) error {
	return nil
}
func (m *MockStore) UpdateJobCompleted(ctx context.Context, id pgtype.UUID) error { return nil }
func (m *MockStore) UpdateJobFailed(ctx context.Context, arg store.UpdateJobFailedParams) error {
	return nil
}
func (m *MockStore) UpdateJobStarted(ctx context.Context, arg store.UpdateJobStartedParams) error {
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
func (m *MockStore) UpdateWorkerHeartbeat(ctx context.Context, id string) error       { return nil }
func (m *MockStore) UpdateWorkerStatus(ctx context.Context, arg store.UpdateWorkerStatusParams) error {
	return nil
}
func (m *MockStore) DeleteOldJobsByStatus(ctx context.Context, arg store.DeleteOldJobsByStatusParams) (int64, error) {
	return 0, nil
}
func (m *MockStore) DeleteOrphanedTasks(ctx context.Context) (int64, error) { return 0, nil }
func (m *MockStore) DeleteOldAuditLogs(ctx context.Context, createdAt pgtype.Timestamptz) (int64, error) {
	return 0, nil
}
func (m *MockStore) GetWorker(ctx context.Context, id string) (store.Worker, error) {
	return store.Worker{}, nil
}
func (m *MockStore) ListWorkers(ctx context.Context) ([]store.Worker, error) { return nil, nil }
func (m *MockStore) MarkWorkersUnhealthy(ctx context.Context) ([]store.Worker, error) {
	return nil, nil
}
func (m *MockStore) ListErrorEvents(ctx context.Context, arg store.ListErrorEventsParams) ([]store.ErrorEvent, error) {
	return nil, nil
}
func (m *MockStore) ListErrorEventsBySource(ctx context.Context, arg store.ListErrorEventsBySourceParams) ([]store.ErrorEvent, error) {
	return nil, nil
}
func (m *MockStore) ResolveErrorEvent(ctx context.Context, id pgtype.UUID) error { return nil }
func (m *MockStore) DeleteOldErrorEvents(ctx context.Context, createdBefore pgtype.Timestamptz) (int64, error) {
	return 0, nil
}
func (m *MockStore) CreateJobLog(ctx context.Context, arg store.CreateJobLogParams) error { return nil }
func (m *MockStore) ListJobLogs(ctx context.Context, jobID pgtype.UUID) ([]store.JobLog, error) {
	return nil, nil
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
func (m *MockStore) DeleteOldNotifications(ctx context.Context, userID pgtype.UUID) error { return nil }
func (m *MockStore) UpdateStreamRestreamDestinations(ctx context.Context, arg store.UpdateStreamRestreamDestinationsParams) error {
	return nil
}

// --- MockOrchestrator ---

type MockOrchestrator struct {
	mock.Mock
}

func (m *MockOrchestrator) SubmitJob(ctx context.Context, req orchestrator.JobRequest) (*store.Job, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*store.Job), args.Error(1)
}

func (m *MockOrchestrator) GetJob(ctx context.Context, id string) (*store.Job, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*store.Job), args.Error(1)
}

func (m *MockOrchestrator) DeleteJob(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrchestrator) CancelJob(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrchestrator) ListJobs(ctx context.Context, limit, offset int32) ([]store.Job, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.Job), args.Error(1)
}

func (m *MockOrchestrator) GetJobTasks(ctx context.Context, jobID string) ([]store.Task, error) {
	args := m.Called(ctx, jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.Task), args.Error(1)
}

func (m *MockOrchestrator) GetJobLogs(ctx context.Context, jobID string) ([]store.JobLog, error) {
	args := m.Called(ctx, jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.JobLog), args.Error(1)
}

func (m *MockOrchestrator) RestartJob(ctx context.Context, id string) (*store.Job, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*store.Job), args.Error(1)
}

func (m *MockOrchestrator) SubmitRestream(ctx context.Context, restreamID string) error {
	args := m.Called(ctx, restreamID)
	return args.Error(0)
}

func (m *MockOrchestrator) StopRestream(ctx context.Context, restreamID string) error {
	args := m.Called(ctx, restreamID)
	return args.Error(0)
}

func (m *MockOrchestrator) HandleTaskEvent(ctx context.Context, taskID string, eventType string, result json.RawMessage) error {
	args := m.Called(ctx, taskID, eventType, result)
	return args.Error(0)
}

// --- Tests ---

func TestTaskEvent_Serialization(t *testing.T) {
	event := TaskEvent{
		TaskID: "00000000-0000-0000-0000-000000000001",
		Event:  "completed",
		Result: json.RawMessage(`{"output_path": "/path/to/file"}`),
	}

	data, err := json.Marshal(event)
	assert.NoError(t, err)

	var decoded TaskEvent
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, event.TaskID, decoded.TaskID)
	assert.Equal(t, event.Event, decoded.Event)
}

func TestNewEventListener(t *testing.T) {
	listener := NewEventListener(nil, nil, nil, nil)
	assert.NotNil(t, listener)
}

func TestEventListener_HandleEvent_Completed(t *testing.T) {
	mockStore := new(MockStore)
	mockOrch := new(MockOrchestrator)
	listener := NewEventListener(nil, mockStore, mockOrch, logger.New("test"))

	taskIDStr := "00000000-0000-0000-0000-000000000001"

	event := TaskEvent{
		TaskID: taskIDStr,
		Event:  "completed",
		Result: json.RawMessage(`{}`),
	}
	data, _ := json.Marshal(event)

	// Expect HandleTaskEvent to be called on orchestrator
	mockOrch.On("HandleTaskEvent", mock.Anything, taskIDStr, "completed", mock.Anything).Return(nil)

	// Call private method directly since we are in same package
	listener.handleEvent(context.Background(), data)

	mockOrch.AssertExpectations(t)
}

func TestEventListener_HandleEvent_Failed(t *testing.T) {
	mockStore := new(MockStore)
	mockOrch := new(MockOrchestrator)
	listener := NewEventListener(nil, mockStore, mockOrch, logger.New("test"))

	taskIDStr := "00000000-0000-0000-0000-000000000001"
	result := json.RawMessage(`{"error":"ffmpeg exited with code 1"}`)

	event := TaskEvent{
		TaskID: taskIDStr,
		Event:  "failed",
		Result: result,
	}
	data, _ := json.Marshal(event)

	// Expect HandleTaskEvent to be called on orchestrator - use mock.Anything for the result
	mockOrch.On("HandleTaskEvent", mock.Anything, taskIDStr, "failed", mock.Anything).Return(nil)

	listener.handleEvent(context.Background(), data)

	mockOrch.AssertExpectations(t)
}

func TestEventListener_HandleEvent_Progress(t *testing.T) {
	mockStore := new(MockStore)
	mockOrch := new(MockOrchestrator)
	listener := NewEventListener(nil, mockStore, mockOrch, logger.New("test"))

	taskIDStr := "00000000-0000-0000-0000-000000000001"
	result := json.RawMessage(`{"progress_pct":50}`)

	event := TaskEvent{
		TaskID: taskIDStr,
		Event:  "progress",
		Result: result,
	}
	data, _ := json.Marshal(event)

	mockOrch.On("HandleTaskEvent", mock.Anything, taskIDStr, "progress", mock.Anything).Return(nil)

	listener.handleEvent(context.Background(), data)

	mockOrch.AssertExpectations(t)
}

func TestEventListener_HandleEvent_OrchestratorError(t *testing.T) {
	mockStore := new(MockStore)
	mockOrch := new(MockOrchestrator)
	listener := NewEventListener(nil, mockStore, mockOrch, logger.New("test"))

	taskIDStr := "00000000-0000-0000-0000-000000000001"

	event := TaskEvent{
		TaskID: taskIDStr,
		Event:  "completed",
		Result: json.RawMessage(`{}`),
	}
	data, _ := json.Marshal(event)

	// Orchestrator returns an error
	mockOrch.On("HandleTaskEvent", mock.Anything, taskIDStr, "completed", mock.Anything).Return(assert.AnError)

	// Should not panic, just log the error
	listener.handleEvent(context.Background(), data)

	mockOrch.AssertExpectations(t)
}

func TestEventListener_HandleEvent_InvalidPayload(t *testing.T) {
	mockStore := new(MockStore)
	mockOrch := new(MockOrchestrator)
	listener := NewEventListener(nil, mockStore, mockOrch, logger.New("test"))

	listener.handleEvent(context.Background(), []byte("invalid-json"))

	mockStore.AssertNotCalled(t, "UpdateTaskStatus")
}

func TestEventListener_HandleErrorEvent(t *testing.T) {
	mockStore := new(MockStore)
	mockOrch := new(MockOrchestrator)
	listener := NewEventListener(nil, mockStore, mockOrch, logger.New("test"))

	event := map[string]interface{}{
		"source":       "worker-1",
		"severity":     "critical",
		"message":      "Out of memory",
		"stack_trace":  "...",
		"context_data": map[string]string{"job_id": "123"},
	}
	data, _ := json.Marshal(event)

	mockStore.On("CreateErrorEvent", mock.Anything, mock.MatchedBy(func(arg store.CreateErrorEventParams) bool {
		return arg.SourceComponent == "worker-1" &&
			arg.Column2 == store.ErrorSeverityCritical &&
			arg.Message == "Out of memory"
	})).Return(store.ErrorEvent{}, nil)

	listener.handleErrorEvent(context.Background(), data)

	mockStore.AssertExpectations(t)
}

func TestEventListener_HandleErrorEvent_InvalidJSON(t *testing.T) {
	mockStore := new(MockStore)
	mockOrch := new(MockOrchestrator)
	listener := NewEventListener(nil, mockStore, mockOrch, logger.New("test"))

	// Should not panic, just log the error
	listener.handleErrorEvent(context.Background(), []byte("invalid-json"))

	mockStore.AssertNotCalled(t, "CreateErrorEvent")
}

func TestEventListener_HandleErrorEvent_WarningSeverity(t *testing.T) {
	mockStore := new(MockStore)
	mockOrch := new(MockOrchestrator)
	listener := NewEventListener(nil, mockStore, mockOrch, logger.New("test"))

	event := map[string]interface{}{
		"source":   "api",
		"severity": "warning",
		"message":  "Rate limit approaching",
	}
	data, _ := json.Marshal(event)

	mockStore.On("CreateErrorEvent", mock.Anything, mock.MatchedBy(func(arg store.CreateErrorEventParams) bool {
		return arg.Column2 == store.ErrorSeverityWarning
	})).Return(store.ErrorEvent{}, nil)

	listener.handleErrorEvent(context.Background(), data)

	mockStore.AssertExpectations(t)
}

func TestEventListener_HandleErrorEvent_FatalSeverity(t *testing.T) {
	mockStore := new(MockStore)
	mockOrch := new(MockOrchestrator)
	listener := NewEventListener(nil, mockStore, mockOrch, logger.New("test"))

	event := map[string]interface{}{
		"source":   "worker-1",
		"severity": "fatal",
		"message":  "System crash",
	}
	data, _ := json.Marshal(event)

	mockStore.On("CreateErrorEvent", mock.Anything, mock.MatchedBy(func(arg store.CreateErrorEventParams) bool {
		return arg.Column2 == store.ErrorSeverityFatal
	})).Return(store.ErrorEvent{}, nil)

	listener.handleErrorEvent(context.Background(), data)

	mockStore.AssertExpectations(t)
}

func TestEventListener_HandleErrorEvent_DefaultSeverity(t *testing.T) {
	mockStore := new(MockStore)
	mockOrch := new(MockOrchestrator)
	listener := NewEventListener(nil, mockStore, mockOrch, logger.New("test"))

	event := map[string]interface{}{
		"source":  "scheduler",
		"message": "Task failed",
	}
	data, _ := json.Marshal(event)

	mockStore.On("CreateErrorEvent", mock.Anything, mock.MatchedBy(func(arg store.CreateErrorEventParams) bool {
		return arg.Column2 == store.ErrorSeverityError
	})).Return(store.ErrorEvent{}, nil)

	listener.handleErrorEvent(context.Background(), data)

	mockStore.AssertExpectations(t)
}

func TestEventListener_HandleErrorEvent_DBError(t *testing.T) {
	mockStore := new(MockStore)
	mockOrch := new(MockOrchestrator)
	listener := NewEventListener(nil, mockStore, mockOrch, logger.New("test"))

	event := map[string]interface{}{
		"source":  "test",
		"message": "Test error",
	}
	data, _ := json.Marshal(event)

	mockStore.On("CreateErrorEvent", mock.Anything, mock.Anything).Return(store.ErrorEvent{}, assert.AnError)

	// Should not panic, just log the error
	listener.handleErrorEvent(context.Background(), data)

	mockStore.AssertExpectations(t)
}

func TestEventListener_HandleErrorEvent_EmptyStackTrace(t *testing.T) {
	mockStore := new(MockStore)
	mockOrch := new(MockOrchestrator)
	listener := NewEventListener(nil, mockStore, mockOrch, logger.New("test"))

	event := map[string]interface{}{
		"source":      "worker-1",
		"severity":    "error",
		"message":     "Something went wrong",
		"stack_trace": "",
	}
	data, _ := json.Marshal(event)

	mockStore.On("CreateErrorEvent", mock.Anything, mock.MatchedBy(func(arg store.CreateErrorEventParams) bool {
		return !arg.StackTrace.Valid
	})).Return(store.ErrorEvent{}, nil)

	listener.handleErrorEvent(context.Background(), data)

	mockStore.AssertExpectations(t)
}
