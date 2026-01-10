package appcontext

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/stretchr/testify/assert"
)

// MockQuerier is a mock implementation of store.Querier for testing
type MockQuerier struct{}

// Verify MockQuerier implements store.Querier at compile time
var _ store.Querier = (*MockQuerier)(nil)

// Implement all required interface methods as stubs
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
func (m *MockQuerier) CreateErrorEvent(ctx context.Context, arg store.CreateErrorEventParams) (store.ErrorEvent, error) {
	return store.ErrorEvent{}, nil
}
func (m *MockQuerier) CreateJob(ctx context.Context, arg store.CreateJobParams) (store.Job, error) {
	return store.Job{}, nil
}
func (m *MockQuerier) CreateJobLog(ctx context.Context, arg store.CreateJobLogParams) error {
	return nil
}
func (m *MockQuerier) CreateNotification(ctx context.Context, arg store.CreateNotificationParams) (store.Notification, error) {
	return store.Notification{}, nil
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
func (m *MockQuerier) DeleteNotification(ctx context.Context, arg store.DeleteNotificationParams) error {
	return nil
}
func (m *MockQuerier) DeleteOldAuditLogs(ctx context.Context, createdAt pgtype.Timestamptz) (int64, error) {
	return 0, nil
}
func (m *MockQuerier) DeleteOldErrorEvents(ctx context.Context, createdAt pgtype.Timestamptz) (int64, error) {
	return 0, nil
}
func (m *MockQuerier) DeleteOldJobsByStatus(ctx context.Context, arg store.DeleteOldJobsByStatusParams) (int64, error) {
	return 0, nil
}
func (m *MockQuerier) DeleteOldNotifications(ctx context.Context, userID pgtype.UUID) error {
	return nil
}
func (m *MockQuerier) DeleteOldWorkers(ctx context.Context, lastSeen pgtype.Timestamptz) (int64, error) {
	return 0, nil
}
func (m *MockQuerier) DeleteOrphanedTasks(ctx context.Context) (int64, error)     { return 0, nil }
func (m *MockQuerier) DeletePluginConfig(ctx context.Context, id string) error    { return nil }
func (m *MockQuerier) DeleteRestreamJob(ctx context.Context, id pgtype.UUID) error { return nil }
func (m *MockQuerier) DeleteStream(ctx context.Context, id pgtype.UUID) error     { return nil }
func (m *MockQuerier) DeleteWebhook(ctx context.Context, id pgtype.UUID) error    { return nil }
func (m *MockQuerier) DeleteWorker(ctx context.Context, id string) error          { return nil }
func (m *MockQuerier) DisablePlugin(ctx context.Context, id string) error         { return nil }
func (m *MockQuerier) EnablePlugin(ctx context.Context, id string) error          { return nil }
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
func (m *MockQuerier) GetUnreadNotificationCount(ctx context.Context, userID pgtype.UUID) (int64, error) {
	return 0, nil
}
func (m *MockQuerier) GetWebhook(ctx context.Context, id pgtype.UUID) (store.Webhook, error) {
	return store.Webhook{}, nil
}
func (m *MockQuerier) GetWorker(ctx context.Context, id string) (store.Worker, error) {
	return store.Worker{}, nil
}
func (m *MockQuerier) IncrementTaskAttempt(ctx context.Context, id pgtype.UUID) error { return nil }
func (m *MockQuerier) IncrementWebhookFailure(ctx context.Context, id pgtype.UUID) error {
	return nil
}
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
func (m *MockQuerier) ListErrorEvents(ctx context.Context, arg store.ListErrorEventsParams) ([]store.ErrorEvent, error) {
	return nil, nil
}
func (m *MockQuerier) ListErrorEventsBySource(ctx context.Context, arg store.ListErrorEventsBySourceParams) ([]store.ErrorEvent, error) {
	return nil, nil
}
func (m *MockQuerier) ListHealthyWorkers(ctx context.Context) ([]store.Worker, error) {
	return nil, nil
}
func (m *MockQuerier) ListJobLogs(ctx context.Context, jobID pgtype.UUID) ([]store.JobLog, error) {
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
func (m *MockQuerier) ListNotifications(ctx context.Context, arg store.ListNotificationsParams) ([]store.Notification, error) {
	return nil, nil
}
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
func (m *MockQuerier) ListStreams(ctx context.Context, arg store.ListStreamsParams) ([]store.Stream, error) {
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
func (m *MockQuerier) MarkAllNotificationsRead(ctx context.Context, userID pgtype.UUID) error {
	return nil
}
func (m *MockQuerier) MarkNotificationRead(ctx context.Context, arg store.MarkNotificationReadParams) error {
	return nil
}
func (m *MockQuerier) MarkWorkersUnhealthy(ctx context.Context) ([]store.Worker, error) {
	return nil, nil
}
func (m *MockQuerier) RegisterPluginConfig(ctx context.Context, arg store.RegisterPluginConfigParams) (store.PluginConfig, error) {
	return store.PluginConfig{}, nil
}
func (m *MockQuerier) RegisterWorker(ctx context.Context, arg store.RegisterWorkerParams) error {
	return nil
}
func (m *MockQuerier) ResolveErrorEvent(ctx context.Context, id pgtype.UUID) error { return nil }
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
func (m *MockQuerier) UpdateStreamRestreamDestinations(ctx context.Context, arg store.UpdateStreamRestreamDestinationsParams) error {
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

func TestGetLogger_WithLogger(t *testing.T) {
	l := logger.New("test")
	ctx := context.WithValue(context.Background(), LoggerKey, l)

	result := GetLogger(ctx)

	assert.NotNil(t, result)
	assert.Equal(t, l, result)
}

func TestGetLogger_WithoutLogger(t *testing.T) {
	ctx := context.Background()

	result := GetLogger(ctx)

	// Should return fallback logger
	assert.NotNil(t, result)
}

func TestGetLogger_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), LoggerKey, "not a logger")

	result := GetLogger(ctx)

	// Should return fallback logger
	assert.NotNil(t, result)
}

func TestGetQuerier_WithQuerier(t *testing.T) {
	q := &MockQuerier{}
	ctx := context.WithValue(context.Background(), QuerierKey, q)

	result := GetQuerier(ctx)

	assert.NotNil(t, result)
	assert.Equal(t, q, result)
}

func TestGetQuerier_WithoutQuerier(t *testing.T) {
	ctx := context.Background()

	result := GetQuerier(ctx)

	assert.Nil(t, result)
}

func TestGetQuerier_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), QuerierKey, "not a querier")

	result := GetQuerier(ctx)

	assert.Nil(t, result)
}

func TestWithLogger(t *testing.T) {
	l := logger.New("test")
	ctx := context.Background()

	result := WithLogger(ctx, l)

	assert.NotNil(t, result)
	storedLogger := result.Value(LoggerKey)
	assert.Equal(t, l, storedLogger)
}

func TestWithQuerier(t *testing.T) {
	q := &MockQuerier{}
	ctx := context.Background()

	result := WithQuerier(ctx, q)

	assert.NotNil(t, result)
	storedQuerier := result.Value(QuerierKey)
	assert.Equal(t, q, storedQuerier)
}

func TestWithLogger_ChainedContext(t *testing.T) {
	l1 := logger.New("first")
	l2 := logger.New("second")
	ctx := context.Background()

	ctx = WithLogger(ctx, l1)
	ctx = WithLogger(ctx, l2)

	result := GetLogger(ctx)
	assert.Equal(t, l2, result)
}

func TestWithQuerier_ChainedContext(t *testing.T) {
	q1 := &MockQuerier{}
	q2 := &MockQuerier{}
	ctx := context.Background()

	ctx = WithQuerier(ctx, q1)
	ctx = WithQuerier(ctx, q2)

	result := GetQuerier(ctx)
	assert.Equal(t, q2, result)
}

func TestContextKey_Constants(t *testing.T) {
	assert.Equal(t, contextKey("logger"), LoggerKey)
	assert.Equal(t, contextKey("querier"), QuerierKey)
}

func TestCombinedContext(t *testing.T) {
	l := logger.New("test")
	q := &MockQuerier{}
	ctx := context.Background()

	ctx = WithLogger(ctx, l)
	ctx = WithQuerier(ctx, q)

	resultLogger := GetLogger(ctx)
	resultQuerier := GetQuerier(ctx)

	assert.Equal(t, l, resultLogger)
	assert.Equal(t, q, resultQuerier)
}
