package handlers

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rennerdo30/webencode/internal/orchestrator"
	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

// --- MockStore ---

type MockStore struct {
	mock.Mock
}

func (m *MockStore) ListWorkers(ctx context.Context) ([]store.Worker, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.Worker), args.Error(1)
}

func (m *MockStore) GetWorker(ctx context.Context, id string) (store.Worker, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(store.Worker), args.Error(1)
}

func (m *MockStore) AssignTask(ctx context.Context, arg store.AssignTaskParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}
func (m *MockStore) CancelJob(ctx context.Context, id pgtype.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockStore) CompleteTask(ctx context.Context, arg store.CompleteTaskParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}
func (m *MockStore) CountJobs(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockStore) CountJobsByStatus(ctx context.Context, status store.JobStatus) (int64, error) {
	args := m.Called(ctx, status)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockStore) CountPendingTasksForJob(ctx context.Context, jobID pgtype.UUID) (int64, error) {
	args := m.Called(ctx, jobID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockStore) CountTasksByJobAndStatus(ctx context.Context, arg store.CountTasksByJobAndStatusParams) (int64, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockStore) CreateAuditLog(ctx context.Context, arg store.CreateAuditLogParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}
func (m *MockStore) CreateEncodingProfile(ctx context.Context, arg store.CreateEncodingProfileParams) (store.EncodingProfile, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(store.EncodingProfile), args.Error(1)
}
func (m *MockStore) CreateJob(ctx context.Context, arg store.CreateJobParams) (store.Job, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(store.Job), args.Error(1)
}
func (m *MockStore) RegisterPluginConfig(ctx context.Context, arg store.RegisterPluginConfigParams) (store.PluginConfig, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(store.PluginConfig), args.Error(1)
}
func (m *MockStore) CreateRestreamJob(ctx context.Context, arg store.CreateRestreamJobParams) (store.RestreamJob, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(store.RestreamJob), args.Error(1)
}
func (m *MockStore) CreateStream(ctx context.Context, arg store.CreateStreamParams) (store.Stream, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(store.Stream), args.Error(1)
}
func (m *MockStore) CreateTask(ctx context.Context, arg store.CreateTaskParams) (store.Task, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(store.Task), args.Error(1)
}
func (m *MockStore) CreateWebhook(ctx context.Context, arg store.CreateWebhookParams) (store.Webhook, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(store.Webhook), args.Error(1)
}
func (m *MockStore) DeactivateWebhook(ctx context.Context, id pgtype.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockStore) DeleteEncodingProfile(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockStore) DeleteJob(ctx context.Context, id pgtype.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockStore) DeleteOldWorkers(ctx context.Context, lastSeen pgtype.Timestamptz) (int64, error) {
	args := m.Called(ctx, lastSeen)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockStore) DeletePluginConfig(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockStore) DeleteRestreamJob(ctx context.Context, id pgtype.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockStore) DeleteStream(ctx context.Context, id pgtype.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockStore) DeleteWebhook(ctx context.Context, id pgtype.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockStore) DeleteWorker(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockStore) DisablePlugin(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockStore) EnablePlugin(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockStore) FailTask(ctx context.Context, arg store.FailTaskParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}
func (m *MockStore) GetCompletedTaskOutputs(ctx context.Context, jobID pgtype.UUID) ([]store.GetCompletedTaskOutputsRow, error) {
	args := m.Called(ctx, jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.GetCompletedTaskOutputsRow), args.Error(1)
}
func (m *MockStore) GetEncodingProfile(ctx context.Context, id string) (store.EncodingProfile, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(store.EncodingProfile), args.Error(1)
}
func (m *MockStore) GetJob(ctx context.Context, id pgtype.UUID) (store.Job, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(store.Job), args.Error(1)
}
func (m *MockStore) GetPendingTasks(ctx context.Context, limit int32) ([]store.Task, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.Task), args.Error(1)
}
func (m *MockStore) GetPluginConfig(ctx context.Context, id string) (store.PluginConfig, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(store.PluginConfig), args.Error(1)
}
func (m *MockStore) GetRestreamJob(ctx context.Context, id pgtype.UUID) (store.RestreamJob, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(store.RestreamJob), args.Error(1)
}
func (m *MockStore) GetStream(ctx context.Context, id pgtype.UUID) (store.Stream, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(store.Stream), args.Error(1)
}
func (m *MockStore) GetStreamByKey(ctx context.Context, streamKey string) (store.Stream, error) {
	args := m.Called(ctx, streamKey)
	return args.Get(0).(store.Stream), args.Error(1)
}
func (m *MockStore) GetTask(ctx context.Context, id pgtype.UUID) (store.Task, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(store.Task), args.Error(1)
}
func (m *MockStore) GetWebhook(ctx context.Context, id pgtype.UUID) (store.Webhook, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(store.Webhook), args.Error(1)
}
func (m *MockStore) IncrementTaskAttempt(ctx context.Context, id pgtype.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockStore) IncrementWebhookFailure(ctx context.Context, id pgtype.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockStore) ListActiveWebhooksForEvent(ctx context.Context, event string) ([]store.Webhook, error) {
	args := m.Called(ctx, event)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.Webhook), args.Error(1)
}
func (m *MockStore) ListAuditLogs(ctx context.Context, arg store.ListAuditLogsParams) ([]store.AuditLog, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.AuditLog), args.Error(1)
}
func (m *MockStore) ListAuditLogsByUser(ctx context.Context, arg store.ListAuditLogsByUserParams) ([]store.AuditLog, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.AuditLog), args.Error(1)
}
func (m *MockStore) ListEncodingProfiles(ctx context.Context) ([]store.EncodingProfile, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.EncodingProfile), args.Error(1)
}
func (m *MockStore) ListHealthyWorkers(ctx context.Context) ([]store.Worker, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.Worker), args.Error(1)
}
func (m *MockStore) ListJobs(ctx context.Context, arg store.ListJobsParams) ([]store.Job, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.Job), args.Error(1)
}
func (m *MockStore) ListJobsByStatus(ctx context.Context, arg store.ListJobsByStatusParams) ([]store.Job, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.Job), args.Error(1)
}
func (m *MockStore) ListJobsByUser(ctx context.Context, arg store.ListJobsByUserParams) ([]store.Job, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.Job), args.Error(1)
}
func (m *MockStore) ListLiveStreams(ctx context.Context) ([]store.Stream, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.Stream), args.Error(1)
}
func (m *MockStore) ListPluginConfigs(ctx context.Context) ([]store.PluginConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.PluginConfig), args.Error(1)
}
func (m *MockStore) ListPluginConfigsByType(ctx context.Context, pluginType string) ([]store.PluginConfig, error) {
	args := m.Called(ctx, pluginType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.PluginConfig), args.Error(1)
}
func (m *MockStore) ListRestreamJobs(ctx context.Context, arg store.ListRestreamJobsParams) ([]store.RestreamJob, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.RestreamJob), args.Error(1)
}
func (m *MockStore) ListRestreamJobsByUser(ctx context.Context, arg store.ListRestreamJobsByUserParams) ([]store.RestreamJob, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.RestreamJob), args.Error(1)
}
func (m *MockStore) ListStreams(ctx context.Context, arg store.ListStreamsParams) ([]store.Stream, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.Stream), args.Error(1)
}
func (m *MockStore) ListStreamsByUser(ctx context.Context, arg store.ListStreamsByUserParams) ([]store.Stream, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.Stream), args.Error(1)
}
func (m *MockStore) ListTasksByJob(ctx context.Context, jobID pgtype.UUID) ([]store.Task, error) {
	args := m.Called(ctx, jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.Task), args.Error(1)
}
func (m *MockStore) ListWebhooks(ctx context.Context) ([]store.Webhook, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.Webhook), args.Error(1)
}
func (m *MockStore) ListWebhooksByUser(ctx context.Context, userID pgtype.UUID) ([]store.Webhook, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.Webhook), args.Error(1)
}
func (m *MockStore) RegisterWorker(ctx context.Context, arg store.RegisterWorkerParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}
func (m *MockStore) SetStreamArchiveJob(ctx context.Context, arg store.SetStreamArchiveJobParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}
func (m *MockStore) UpdateEncodingProfile(ctx context.Context, arg store.UpdateEncodingProfileParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}
func (m *MockStore) UpdateJobCompleted(ctx context.Context, id pgtype.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockStore) UpdateJobFailed(ctx context.Context, arg store.UpdateJobFailedParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}
func (m *MockStore) UpdateJobProgress(ctx context.Context, arg store.UpdateJobProgressParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}
func (m *MockStore) UpdateJobStarted(ctx context.Context, arg store.UpdateJobStartedParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}
func (m *MockStore) UpdateJobStatus(ctx context.Context, arg store.UpdateJobStatusParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}
func (m *MockStore) UpdatePluginConfig(ctx context.Context, arg store.UpdatePluginConfigParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}
func (m *MockStore) UpdateRestreamJobStatus(ctx context.Context, arg store.UpdateRestreamJobStatusParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}
func (m *MockStore) UpdateStreamLive(ctx context.Context, arg store.UpdateStreamLiveParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}
func (m *MockStore) UpdateStreamMetadata(ctx context.Context, arg store.UpdateStreamMetadataParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}
func (m *MockStore) UpdateStreamStats(ctx context.Context, arg store.UpdateStreamStatsParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}
func (m *MockStore) UpdateTaskStatus(ctx context.Context, arg store.UpdateTaskStatusParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}
func (m *MockStore) UpdateWebhookTriggered(ctx context.Context, id pgtype.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockStore) UpdateWorkerHeartbeat(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockStore) UpdateWorkerStatus(ctx context.Context, arg store.UpdateWorkerStatusParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}
func (m *MockStore) DeleteOldJobsByStatus(ctx context.Context, arg store.DeleteOldJobsByStatusParams) (int64, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockStore) DeleteOrphanedTasks(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockStore) DeleteOldAuditLogs(ctx context.Context, createdAt pgtype.Timestamptz) (int64, error) {
	args := m.Called(ctx, createdAt)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockStore) MarkWorkersUnhealthy(ctx context.Context) ([]store.Worker, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.Worker), args.Error(1)
}

func (m *MockStore) CreateErrorEvent(ctx context.Context, arg store.CreateErrorEventParams) (store.ErrorEvent, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(store.ErrorEvent), args.Error(1)
}
func (m *MockStore) ListErrorEvents(ctx context.Context, arg store.ListErrorEventsParams) ([]store.ErrorEvent, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.ErrorEvent), args.Error(1)
}
func (m *MockStore) ListErrorEventsBySource(ctx context.Context, arg store.ListErrorEventsBySourceParams) ([]store.ErrorEvent, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.ErrorEvent), args.Error(1)
}
func (m *MockStore) ResolveErrorEvent(ctx context.Context, id pgtype.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockStore) DeleteOldErrorEvents(ctx context.Context, createdBefore pgtype.Timestamptz) (int64, error) {
	args := m.Called(ctx, createdBefore)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStore) CreateJobLog(ctx context.Context, arg store.CreateJobLogParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}
func (m *MockStore) ListJobLogs(ctx context.Context, jobID pgtype.UUID) ([]store.JobLog, error) {
	args := m.Called(ctx, jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.JobLog), args.Error(1)
}

func (m *MockStore) CreateNotification(ctx context.Context, arg store.CreateNotificationParams) (store.Notification, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(store.Notification), args.Error(1)
}
func (m *MockStore) ListNotifications(ctx context.Context, arg store.ListNotificationsParams) ([]store.Notification, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.Notification), args.Error(1)
}
func (m *MockStore) GetUnreadNotificationCount(ctx context.Context, userID pgtype.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockStore) MarkNotificationRead(ctx context.Context, arg store.MarkNotificationReadParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}
func (m *MockStore) MarkAllNotificationsRead(ctx context.Context, userID pgtype.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
func (m *MockStore) DeleteNotification(ctx context.Context, arg store.DeleteNotificationParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}
func (m *MockStore) DeleteOldNotifications(ctx context.Context, userID pgtype.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
func (m *MockStore) UpdateStreamRestreamDestinations(ctx context.Context, arg store.UpdateStreamRestreamDestinationsParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

// --- MockOrchestratorService ---
type MockOrchestratorService struct {
	mock.Mock
}

func (m *MockOrchestratorService) SubmitJob(ctx context.Context, req orchestrator.JobRequest) (*store.Job, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*store.Job), args.Error(1)
}

func (m *MockOrchestratorService) GetJob(ctx context.Context, id string) (*store.Job, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*store.Job), args.Error(1)
}

func (m *MockOrchestratorService) DeleteJob(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrchestratorService) CancelJob(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrchestratorService) ListJobs(ctx context.Context, limit, offset int32) ([]store.Job, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.Job), args.Error(1)
}

func (m *MockOrchestratorService) GetJobTasks(ctx context.Context, jobID string) ([]store.Task, error) {
	args := m.Called(ctx, jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.Task), args.Error(1)
}

func (m *MockOrchestratorService) GetJobLogs(ctx context.Context, jobID string) ([]store.JobLog, error) {
	args := m.Called(ctx, jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.JobLog), args.Error(1)
}

func (m *MockOrchestratorService) SubmitRestream(ctx context.Context, restreamID string) error {
	args := m.Called(ctx, restreamID)
	return args.Error(0)
}

func (m *MockOrchestratorService) StopRestream(ctx context.Context, restreamID string) error {
	args := m.Called(ctx, restreamID)
	return args.Error(0)
}

func (m *MockOrchestratorService) HandleTaskEvent(ctx context.Context, taskID string, eventType string, result json.RawMessage) error {
	args := m.Called(ctx, taskID, eventType, result)
	return args.Error(0)
}

func (m *MockOrchestratorService) RestartJob(ctx context.Context, id string) (*store.Job, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*store.Job), args.Error(1)
}

// --- MockStorageClient ---

type MockStorageClient struct {
	mock.Mock
}

func (m *MockStorageClient) GetCapabilities(ctx context.Context, in *pb.Empty, opts ...grpc.CallOption) (*pb.StorageCapabilities, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.StorageCapabilities), args.Error(1)
}

func (m *MockStorageClient) BrowseRoots(ctx context.Context, in *pb.Empty, opts ...grpc.CallOption) (*pb.BrowseRootsResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.BrowseRootsResponse), args.Error(1)
}

func (m *MockStorageClient) Browse(ctx context.Context, in *pb.BrowseRequest, opts ...grpc.CallOption) (*pb.BrowseResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.BrowseResponse), args.Error(1)
}

func (m *MockStorageClient) Download(ctx context.Context, in *pb.FileRequest, opts ...grpc.CallOption) (pb.StorageService_DownloadClient, error) {
	args := m.Called(ctx, in)
	return nil, args.Error(1)
}

func (m *MockStorageClient) GetUploadURL(ctx context.Context, in *pb.SignedUrlRequest, opts ...grpc.CallOption) (*pb.SignedUrlResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.SignedUrlResponse), args.Error(1)
}

func (m *MockStorageClient) Delete(ctx context.Context, in *pb.FileRequest, opts ...grpc.CallOption) (*pb.Empty, error) {
	args := m.Called(ctx, in)
	return &pb.Empty{}, args.Error(1)
}

func (m *MockStorageClient) Upload(ctx context.Context, opts ...grpc.CallOption) (pb.StorageService_UploadClient, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(pb.StorageService_UploadClient), args.Error(1)
}

func (m *MockStorageClient) GetURL(ctx context.Context, in *pb.SignedUrlRequest, opts ...grpc.CallOption) (*pb.SignedUrlResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.SignedUrlResponse), args.Error(1)
}

func (m *MockStorageClient) ListObjects(ctx context.Context, in *pb.ListObjectsRequest, opts ...grpc.CallOption) (*pb.ListObjectsResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.ListObjectsResponse), args.Error(1)
}

func (m *MockStorageClient) GetObjectMetadata(ctx context.Context, in *pb.FileRequest, opts ...grpc.CallOption) (*pb.ObjectMetadata, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.ObjectMetadata), args.Error(1)
}

// --- MockUploadStream ---

type MockUploadStream struct {
	mock.Mock
	grpc.ClientStream
}

func (m *MockUploadStream) Send(chunk *pb.FileChunk) error {
	args := m.Called(chunk)
	return args.Error(0)
}

func (m *MockUploadStream) CloseAndRecv() (*pb.UploadSummary, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.UploadSummary), args.Error(1)
}
