package webhooks

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/rennerdo30/webencode/pkg/bus"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/stretchr/testify/mock"
)

// MockStore implements store.Querier for testing
type MockStore struct {
	mock.Mock
}

// Key Webhook methods mocked
func (m *MockStore) ListActiveWebhooksForEvent(ctx context.Context, event string) ([]store.Webhook, error) {
	args := m.Called(ctx, event)
	return args.Get(0).([]store.Webhook), args.Error(1)
}

func (m *MockStore) UpdateWebhookTriggered(ctx context.Context, id pgtype.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStore) IncrementWebhookFailure(ctx context.Context, id pgtype.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStore) GetWebhook(ctx context.Context, id pgtype.UUID) (store.Webhook, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(store.Webhook), args.Error(1)
}

func (m *MockStore) DeactivateWebhook(ctx context.Context, id pgtype.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Stubs for other methods
func (m *MockStore) MarkWorkersUnhealthy(ctx context.Context) ([]store.Worker, error) {
	return nil, nil
}
func (m *MockStore) DeleteOldJobsByStatus(ctx context.Context, arg store.DeleteOldJobsByStatusParams) (int64, error) {
	return 0, nil
}
func (m *MockStore) DeleteOrphanedTasks(ctx context.Context) (int64, error) { return 0, nil }
func (m *MockStore) DeleteOldAuditLogs(ctx context.Context, createdAt pgtype.Timestamptz) (int64, error) {
	return 0, nil
}
func (m *MockStore) AssignTask(ctx context.Context, arg store.AssignTaskParams) error     { return nil }
func (m *MockStore) CancelJob(ctx context.Context, id pgtype.UUID) error                  { return nil }
func (m *MockStore) CompleteTask(ctx context.Context, arg store.CompleteTaskParams) error { return nil }
func (m *MockStore) CountJobs(ctx context.Context) (int64, error)                         { return 0, nil }
func (m *MockStore) CountJobsByStatus(ctx context.Context, status store.JobStatus) (int64, error) {
	return 0, nil
}
func (m *MockStore) CountPendingTasksForJob(ctx context.Context, jobID pgtype.UUID) (int64, error) {
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
func (m *MockStore) CreateJob(ctx context.Context, arg store.CreateJobParams) (store.Job, error) {
	return store.Job{}, nil
}
func (m *MockStore) RegisterPluginConfig(ctx context.Context, arg store.RegisterPluginConfigParams) (store.PluginConfig, error) {
	return store.PluginConfig{}, nil
}
func (m *MockStore) DeleteOldWorkers(ctx context.Context, lastSeen pgtype.Timestamptz) (int64, error) {
	return 0, nil
}
func (m *MockStore) CreateRestreamJob(ctx context.Context, arg store.CreateRestreamJobParams) (store.RestreamJob, error) {
	return store.RestreamJob{}, nil
}
func (m *MockStore) CreateStream(ctx context.Context, arg store.CreateStreamParams) (store.Stream, error) {
	return store.Stream{}, nil
}
func (m *MockStore) CreateTask(ctx context.Context, arg store.CreateTaskParams) (store.Task, error) {
	return store.Task{}, nil
}
func (m *MockStore) CreateWebhook(ctx context.Context, arg store.CreateWebhookParams) (store.Webhook, error) {
	return store.Webhook{}, nil
}
func (m *MockStore) DeleteEncodingProfile(ctx context.Context, id string) error   { return nil }
func (m *MockStore) DeleteJob(ctx context.Context, id pgtype.UUID) error          { return nil }
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
func (m *MockStore) GetJob(ctx context.Context, id pgtype.UUID) (store.Job, error) {
	return store.Job{}, nil
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
func (m *MockStore) GetTask(ctx context.Context, id pgtype.UUID) (store.Task, error) {
	return store.Task{}, nil
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
func (m *MockStore) ListJobs(ctx context.Context, arg store.ListJobsParams) ([]store.Job, error) {
	return nil, nil
}
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
func (m *MockStore) ListTasksByJob(ctx context.Context, jobID pgtype.UUID) ([]store.Task, error) {
	return nil, nil
}
func (m *MockStore) ListWebhooks(ctx context.Context) ([]store.Webhook, error) { return nil, nil }
func (m *MockStore) ListWebhooksByUser(ctx context.Context, userID pgtype.UUID) ([]store.Webhook, error) {
	return nil, nil
}
func (m *MockStore) ListWorkers(ctx context.Context) ([]store.Worker, error) { return nil, nil }
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
func (m *MockStore) UpdateJobProgress(ctx context.Context, arg store.UpdateJobProgressParams) error {
	return nil
}
func (m *MockStore) UpdateJobStarted(ctx context.Context, arg store.UpdateJobStartedParams) error {
	return nil
}
func (m *MockStore) UpdateJobStatus(ctx context.Context, arg store.UpdateJobStatusParams) error {
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
func (m *MockStore) UpdateTaskStatus(ctx context.Context, arg store.UpdateTaskStatusParams) error {
	return nil
}
func (m *MockStore) UpdateWorkerHeartbeat(ctx context.Context, id string) error { return nil }
func (m *MockStore) UpdateWorkerStatus(ctx context.Context, arg store.UpdateWorkerStatusParams) error {
	return nil
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
func (m *MockStore) ResolveErrorEvent(ctx context.Context, id pgtype.UUID) error { return nil }
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

// MockMsg implements jetstream.Msg
type MockMsg struct {
	mock.Mock
}

func (m *MockMsg) Metadata() (*jetstream.MsgMetadata, error) { return nil, nil }
func (m *MockMsg) Data() []byte                              { return m.Called().Get(0).([]byte) }
func (m *MockMsg) Headers() nats.Header                      { return nil }
func (m *MockMsg) Subject() string                           { return m.Called().String(0) }
func (m *MockMsg) Reply() string                             { return "" }
func (m *MockMsg) Ack() error {
	args := m.Called()
	return args.Error(0)
}
func (m *MockMsg) DoubleAck(ctx context.Context) error    { return nil }
func (m *MockMsg) Nak() error                             { return nil }
func (m *MockMsg) NakWithDelay(delay time.Duration) error { return nil }
func (m *MockMsg) InProgress() error                      { return nil }
func (m *MockMsg) Term() error                            { return nil }
func (m *MockMsg) TermWithReason(reason string) error     { return nil }

// MockRoundTripper for HTTP
type MockRoundTripper struct {
	mock.Mock
}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestNew(t *testing.T) {
	l := logger.New("test")
	m := New(&bus.Bus{}, nil, l)
	if m == nil {
		t.Fatal("expected manager to be non-nil")
	}
}

func TestComputeHMAC(t *testing.T) {
	secret := "my-secret-key"
	body := []byte(`{"data":"test"}`)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	result := ComputeHMAC(body, secret)
	if result != expected {
		t.Errorf("expected HMAC %s, got %s", expected, result)
	}
}

func TestHandleEvent_Success(t *testing.T) {
	mockMsg := new(MockMsg)
	mockStore := new(MockStore)
	mockRT := new(MockRoundTripper)

	l := logger.New("test")
	m := New(nil, mockStore, l)
	m.client = &http.Client{Transport: mockRT}

	ctx := context.Background()
	eventData := []byte(`{"id":"job-1","status":"completed"}`)

	// Setup MockMsg
	mockMsg.On("Subject").Return("events.job.completed")
	mockMsg.On("Data").Return(eventData)
	mockMsg.On("Ack").Return(nil)

	// Setup MockStore
	webhookID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	webhooks := []store.Webhook{
		{ID: webhookID, Url: "http://example.com/hook", Secret: "secret", Events: []string{"job.completed"}},
	}
	mockStore.On("ListActiveWebhooksForEvent", ctx, "job.completed").Return(webhooks, nil)
	mockStore.On("UpdateWebhookTriggered", ctx, webhookID).Return(nil)

	// Setup MockHTTP
	mockRT.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == "http://example.com/hook" &&
			req.Header.Get("X-WebEncode-Event") == "true" &&
			strings.HasPrefix(req.Header.Get("X-WebEncode-Signature"), "sha256=")
	})).Return(&http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString("OK")),
	}, nil)

	// Execute (calling exported handleEvent via private access)
	m.handleEvent(ctx, mockMsg)

	// Wait for goroutine
	time.Sleep(100 * time.Millisecond)

	mockStore.AssertExpectations(t)
	mockRT.AssertExpectations(t)
	mockMsg.AssertExpectations(t)
}
