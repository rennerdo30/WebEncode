package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rennerdo30/webencode/internal/encoder"
	"github.com/rennerdo30/webencode/pkg/bus"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/logger"
)

type EventBus interface {
	Publish(ctx context.Context, subject string, data []byte) error
}

// Service defines the orchestrator service interface
type Service interface {
	SubmitJob(ctx context.Context, req JobRequest) (*store.Job, error)
	GetJob(ctx context.Context, id string) (*store.Job, error)
	DeleteJob(ctx context.Context, id string) error
	CancelJob(ctx context.Context, id string) error
	ListJobs(ctx context.Context, limit, offset int32) ([]store.Job, error)
	GetJobTasks(ctx context.Context, jobID string) ([]store.Task, error)
	GetJobLogs(ctx context.Context, jobID string) ([]store.JobLog, error)
	RestartJob(ctx context.Context, id string) (*store.Job, error)

	SubmitRestream(ctx context.Context, restreamID string) error
	StopRestream(ctx context.Context, restreamID string) error

	HandleTaskEvent(ctx context.Context, taskID string, eventType string, result json.RawMessage) error
}

type Orchestrator struct {
	db     store.Querier
	bus    EventBus
	logger *logger.Logger
}

func New(db store.Querier, b EventBus, l *logger.Logger) *Orchestrator {
	return &Orchestrator{db: db, bus: b, logger: l}
}

// JobRequest contains parameters for creating a new job
type JobRequest struct {
	UserID     string
	SourceURL  string
	SourceType string // "url", "upload", "stream", "restream" - defaults to "url"
	Profiles   []string
}

// SubmitJob creates the initial state for a job and dispatches the probe task
func (o *Orchestrator) SubmitJob(ctx context.Context, req JobRequest) (*store.Job, error) {
	// Parse UserID
	var userID pgtype.UUID
	if err := userID.Scan(req.UserID); err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	// Determine source type with default to "url"
	sourceType := store.JobSourceTypeUrl
	switch req.SourceType {
	case "upload":
		sourceType = store.JobSourceTypeUpload
	case "stream":
		sourceType = store.JobSourceTypeStream
	case "restream":
		sourceType = store.JobSourceTypeRestream
	}

	// 1. Create Job in DB
	job, err := o.db.CreateJob(ctx, store.CreateJobParams{
		SourceUrl: req.SourceURL,
		Profiles:  req.Profiles,
		Metadata:  []byte("{}"),
		UserID:    userID,
		SourceType: store.NullJobSourceType{
			JobSourceType: sourceType,
			Valid:         true,
		},
		ProfileID:    pgtype.Text{String: "default", Valid: true},
		OutputConfig: []byte("{}"),
	})
	if err != nil {
		return nil, fmt.Errorf("create job: %w", err)
	}

	// 2. Create Probe Task
	probeParams, _ := json.Marshal(map[string]string{
		"url": req.SourceURL,
	})

	task, err := o.db.CreateTask(ctx, store.CreateTaskParams{
		JobID:         job.ID,
		Type:          store.TaskTypeProbe,
		Params:        probeParams,
		SequenceIndex: pgtype.Int4{Int32: 0, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create probe task: %w", err)
	}

	// 3. Dispatch
	if err := o.dispatchTask(ctx, task); err != nil {
		return nil, err
	}

	// 4. Audit Log
	_ = o.db.CreateAuditLog(ctx, store.CreateAuditLogParams{
		UserID:       userID,
		Action:       "job.create",
		ResourceType: "job",
		ResourceID:   pgtype.Text{String: job.ID.String(), Valid: true},
		Details:      []byte(fmt.Sprintf(`{"source_url": "%s"}`, req.SourceURL)),
	})

	o.logger.Info("Job submitted", "job_id", job.ID, "task_id", task.ID)
	return &job, nil
}

func (o *Orchestrator) GetJob(ctx context.Context, id string) (*store.Job, error) {
	var uid pgtype.UUID
	if err := uid.Scan(id); err != nil {
		return nil, fmt.Errorf("invalid uuid: %w", err)
	}
	job, err := o.db.GetJob(ctx, uid)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (o *Orchestrator) ListJobs(ctx context.Context, limit, offset int32) ([]store.Job, error) {
	if limit <= 0 {
		limit = 10
	}
	return o.db.ListJobs(ctx, store.ListJobsParams{
		Limit:  limit,
		Offset: offset,
	})
}

// GetJobTasks returns all tasks for a job
func (o *Orchestrator) GetJobTasks(ctx context.Context, jobID string) ([]store.Task, error) {
	var uid pgtype.UUID
	if err := uid.Scan(jobID); err != nil {
		return nil, fmt.Errorf("invalid uuid: %w", err)
	}
	return o.db.ListTasksByJob(ctx, uid)
}

func (o *Orchestrator) GetJobLogs(ctx context.Context, jobID string) ([]store.JobLog, error) {
	var uid pgtype.UUID
	if err := uid.Scan(jobID); err != nil {
		return nil, fmt.Errorf("invalid uuid: %w", err)
	}
	return o.db.ListJobLogs(ctx, uid)
}

// DeleteJob deletes a job and all its tasks
func (o *Orchestrator) DeleteJob(ctx context.Context, id string) error {
	var uid pgtype.UUID
	if err := uid.Scan(id); err != nil {
		return fmt.Errorf("invalid uuid: %w", err)
	}

	// Tasks are deleted by foreign key cascade
	if err := o.db.DeleteJob(ctx, uid); err != nil {
		return fmt.Errorf("delete job: %w", err)
	}

	o.logger.Info("Job deleted", "job_id", id)
	return nil
}

// CancelJob cancels a running job
func (o *Orchestrator) CancelJob(ctx context.Context, id string) error {
	var uid pgtype.UUID
	if err := uid.Scan(id); err != nil {
		return fmt.Errorf("invalid uuid: %w", err)
	}

	if err := o.db.CancelJob(ctx, uid); err != nil {
		return fmt.Errorf("cancel job: %w", err)
	}

	// Fetch job to get UserID for audit log
	job, _ := o.db.GetJob(ctx, uid)

	// Audit Log
	_ = o.db.CreateAuditLog(ctx, store.CreateAuditLogParams{
		UserID:       job.UserID,
		Action:       "job.cancel",
		ResourceType: "job",
		ResourceID:   pgtype.Text{String: id, Valid: true},
	})

	// Publish cancellation event
	eventData, _ := json.Marshal(map[string]string{
		"job_id": id,
		"status": "cancelled",
	})
	o.bus.Publish(ctx, bus.SubjectJobEvents, eventData)

	o.logger.Info("Job cancelled", "job_id", id)
	return nil
}

// RestartJob creates a new job with the same parameters as the given job
func (o *Orchestrator) RestartJob(ctx context.Context, id string) (*store.Job, error) {
	job, err := o.GetJob(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get job for restart: %w", err)
	}

	// Map internal source type string for request
	sourceType := "url"
	if job.SourceType.Valid {
		switch job.SourceType.JobSourceType {
		case store.JobSourceTypeUpload:
			sourceType = "upload"
		case store.JobSourceTypeStream:
			sourceType = "stream"
		case store.JobSourceTypeRestream:
			sourceType = "restream"
		}
	}

	req := JobRequest{
		UserID:     job.UserID.String(), // Note: UserID mismatch if UUID handling differs, but String() should work
		SourceURL:  job.SourceUrl,
		SourceType: sourceType,
		Profiles:   job.Profiles,
	}

	// Create audit log for retry
	var jobIDStr pgtype.Text
	jobIDStr.String = id
	jobIDStr.Valid = true

	_ = o.db.CreateAuditLog(ctx, store.CreateAuditLogParams{
		UserID:       job.UserID,
		Action:       "job.restart",
		ResourceType: "job",
		ResourceID:   jobIDStr,
		Details:      []byte(fmt.Sprintf(`{"original_job_id": "%s"}`, id)),
	})

	return o.SubmitJob(ctx, req)
}

func (o *Orchestrator) dispatchTask(ctx context.Context, task store.Task) error {
	payload, _ := json.Marshal(task)
	subject := bus.SubjectJobDispatch
	if err := o.bus.Publish(ctx, subject, payload); err != nil {
		return fmt.Errorf("dispatch bus error: %w", err)
	}
	return nil
}

// HandleTaskEvent processes events from workers
func (o *Orchestrator) HandleTaskEvent(ctx context.Context, taskID string, eventType string, result json.RawMessage) error {
	var tid pgtype.UUID
	if err := tid.Scan(taskID); err != nil {
		return fmt.Errorf("invalid task id: %w", err)
	}

	task, err := o.db.GetTask(ctx, tid)
	if err != nil {
		return fmt.Errorf("get task: %w", err)
	}

	o.logger.Info("Handling task event", "task_id", taskID, "type", task.Type, "event", eventType)

	switch eventType {
	case "progress":
		return o.handleTaskProgress(ctx, task, result)
	case "completed":
		return o.handleTaskCompleted(ctx, task, result)
	case "failed":
		return o.handleTaskFailed(ctx, task, result)
	case "log":
		return o.handleTaskLog(ctx, task, result)
	}
	return nil
}

func (o *Orchestrator) handleTaskCompleted(ctx context.Context, task store.Task, result json.RawMessage) error {
	// Extract output_path from result if available (for transcode tasks)
	var taskResult struct {
		OutputPath string `json:"output_path"`
	}
	_ = json.Unmarshal(result, &taskResult)

	outputKey := pgtype.Text{}
	if taskResult.OutputPath != "" {
		outputKey = pgtype.Text{String: taskResult.OutputPath, Valid: true}
	}

	// Mark task completed with output_key
	if err := o.db.CompleteTask(ctx, store.CompleteTaskParams{
		ID:        task.ID,
		Result:    result,
		OutputKey: outputKey,
	}); err != nil {
		return fmt.Errorf("mark task complete: %w", err)
	}

	switch task.Type {
	case store.TaskTypeProbe:
		return o.handleProbeComplete(ctx, task, result)
	case store.TaskTypeTranscode:
		return o.handleTranscodeComplete(ctx, task)
	case store.TaskTypeStitch:
		return o.handleStitchComplete(ctx, task)
	}
	return nil
}

func (o *Orchestrator) handleTaskProgress(ctx context.Context, task store.Task, result json.RawMessage) error {
	var progress struct {
		Percent float64 `json:"percent"`
	}
	if err := json.Unmarshal(result, &progress); err != nil {
		return err
	}

	// Update job progress
	if err := o.db.UpdateJobProgress(ctx, store.UpdateJobProgressParams{
		ID:          task.JobID,
		ProgressPct: pgtype.Int4{Int32: int32(progress.Percent), Valid: true},
	}); err != nil {
		return err
	}

	// Publish progress event for SSE/UI
	eventData, _ := json.Marshal(map[string]interface{}{
		"job_id":   task.JobID.String(),
		"event":    "progress",
		"progress": int(progress.Percent),
	})
	o.bus.Publish(ctx, bus.SubjectJobEvents, eventData)

	return nil
}

func (o *Orchestrator) handleTaskFailed(ctx context.Context, task store.Task, result json.RawMessage) error {
	// Mark task failed
	if err := o.db.FailTask(ctx, store.FailTaskParams{
		ID:     task.ID,
		Result: result,
	}); err != nil {
		return fmt.Errorf("mark task failed: %w", err)
	}

	// Fail job
	return o.db.UpdateJobFailed(ctx, store.UpdateJobFailedParams{
		ID:           task.JobID,
		ErrorMessage: pgtype.Text{String: fmt.Sprintf("Task %s failed: %s", task.ID, string(result)), Valid: true},
	})
}

func (o *Orchestrator) handleTaskLog(ctx context.Context, task store.Task, result json.RawMessage) error {
	var logData struct {
		Level   string `json:"level"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(result, &logData); err != nil {
		return err
	}

	return o.db.CreateJobLog(ctx, store.CreateJobLogParams{
		JobID:    task.JobID,
		Level:    logData.Level,
		Message:  logData.Message,
		Metadata: []byte("{}"),
	})
}

func (o *Orchestrator) handleProbeComplete(ctx context.Context, task store.Task, result json.RawMessage) error {
	// Parse probe result
	var probeData encoder.ProbeResult
	if err := json.Unmarshal(result, &probeData); err != nil {
		return fmt.Errorf("parse probe result: %w", err)
	}

	job, err := o.db.GetJob(ctx, task.JobID)
	if err != nil {
		return err
	}

	// Get healthy worker count to determine segmentation strategy
	workers, err := o.db.ListHealthyWorkers(ctx)
	workerCount := 1
	if err == nil && len(workers) > 0 {
		workerCount = len(workers)
	}

	// Smart chunking: only segment if multiple workers are available
	// With single worker, segmenting adds overhead without parallelism benefit
	var segments []encoder.Segment
	if workerCount <= 1 {
		// Single worker: process entire file as one segment
		o.logger.Info("Single worker detected, skipping segmentation", "duration", probeData.Duration)
		segments = []encoder.Segment{{
			Index:     0,
			StartTime: 0,
			EndTime:   probeData.Duration,
			Duration:  probeData.Duration,
		}}
	} else {
		// Multiple workers: segment for parallel processing
		// Target segment duration scales with worker count for optimal parallelism
		targetDuration := 30.0 // Default 30s segments
		if workerCount >= 4 {
			targetDuration = 15.0 // Smaller segments for more workers
		} else if workerCount >= 8 {
			targetDuration = 10.0
		}
		segments = encoder.CalculateSegments(probeData.Keyframes, probeData.Duration, targetDuration)
		o.logger.Info("Multiple workers detected, using parallel segmentation",
			"workers", workerCount, "segments", len(segments), "target_duration", targetDuration)
	}

	// Create transcode tasks for each profile & segment
	// For simplicity, we just use the first profile or a default
	profileName := "1080p_h264"
	if len(job.Profiles) > 0 {
		profileName = job.Profiles[0]
	}

	profile, ok := encoder.GetProfile(profileName)
	if !ok {
		profile, _ = encoder.GetProfile("720p_h264")
	}

	o.logger.Info("Creating transcode tasks", "segments", len(segments), "profile", profileName)

	for _, seg := range segments {
		// Construct output filename
		outputName := fmt.Sprintf("%x_%s_%03d.mp4", job.ID.Bytes, profileName, seg.Index)

		transcodeParams := encoder.TranscodeTask{
			InputURL:    job.SourceUrl,
			OutputURL:   outputName, // In reality, this would be a bucket URL
			StartTime:   seg.StartTime,
			Duration:    seg.Duration,
			VideoCodec:  profile.VideoCodec,
			AudioCodec:  profile.AudioCodec,
			Width:       profile.Width,
			Height:      profile.Height,
			BitrateKbps: profile.BitrateKbps,
			Preset:      profile.Preset,
			Container:   profile.Container,
		}

		paramsBytes, _ := json.Marshal(transcodeParams)

		t, err := o.db.CreateTask(ctx, store.CreateTaskParams{
			JobID:         job.ID,
			Type:          store.TaskTypeTranscode,
			Params:        paramsBytes,
			SequenceIndex: pgtype.Int4{Int32: int32(seg.Index), Valid: true},
			StartTimeSec:  pgtype.Float8{Float64: seg.StartTime, Valid: true},
			EndTimeSec:    pgtype.Float8{Float64: seg.EndTime, Valid: true},
		})
		if err != nil {
			return err
		}

		if err := o.dispatchTask(ctx, t); err != nil {
			o.logger.Error("Failed to dispatch task", "error", err)
		}
	}

	// Update job status
	return o.db.UpdateJobStarted(ctx, store.UpdateJobStartedParams{
		ID:                 job.ID,
		AssignedToWorkerID: pgtype.Text{String: "orchestrator", Valid: true},
	})
}

func (o *Orchestrator) handleTranscodeComplete(ctx context.Context, task store.Task) error {
	// Check if pending tasks remain for this job
	count, err := o.db.CountTasksByJobAndStatus(ctx, store.CountTasksByJobAndStatusParams{
		JobID:  task.JobID,
		Status: "pending",
	})
	if err != nil {
		return err
	}

	// Also check "assigned" status (running)
	running, err := o.db.CountTasksByJobAndStatus(ctx, store.CountTasksByJobAndStatusParams{
		JobID:  task.JobID,
		Status: "assigned",
	})
	if err != nil {
		return err
	}

	if count == 0 && running == 0 {
		// All transcode tasks done -> Create Stitch Task
		o.logger.Info("All transcode tasks done, creating stitch task", "job_id", task.JobID)
		return o.createStitchTask(ctx, task.JobID)
	}

	return nil
}

func (o *Orchestrator) createStitchTask(ctx context.Context, jobID pgtype.UUID) error {
	// Get all completed tasks to gather output filenames
	outputs, err := o.db.GetCompletedTaskOutputs(ctx, jobID)
	if err != nil {
		return err
	}

	var segmentFiles []string
	for _, out := range outputs {
		if out.OutputKey.Valid {
			segmentFiles = append(segmentFiles, out.OutputKey.String)
		}
	}

	// Final output name
	finalOutput := fmt.Sprintf("%x_final.mp4", jobID.Bytes)

	stitchParams := map[string]interface{}{
		"segments": segmentFiles,
		"output":   finalOutput,
	}
	paramsBytes, _ := json.Marshal(stitchParams)

	task, err := o.db.CreateTask(ctx, store.CreateTaskParams{
		JobID:         jobID,
		Type:          store.TaskTypeStitch,
		Params:        paramsBytes,
		SequenceIndex: pgtype.Int4{Int32: -1, Valid: true}, // -1 indicates a non-segment task (stitch/finalize)
	})
	if err != nil {
		return err
	}

	// Update job status to stitching
	if err := o.db.UpdateJobStatus(ctx, store.UpdateJobStatusParams{
		ID:     jobID,
		Status: store.JobStatusStitching,
	}); err != nil {
		return err
	}

	return o.dispatchTask(ctx, task)
}

func (o *Orchestrator) handleStitchComplete(ctx context.Context, task store.Task) error {
	o.logger.Info("Stitch completed, job finished", "job_id", task.JobID)
	return o.db.UpdateJobCompleted(ctx, task.JobID)
}

func (o *Orchestrator) SubmitRestream(ctx context.Context, restreamID string) error {
	var uid pgtype.UUID
	if err := uid.Scan(restreamID); err != nil {
		return err
	}

	restream, err := o.db.GetRestreamJob(ctx, uid)
	if err != nil {
		return err
	}

	// Create a Restream Task
	task, err := o.db.CreateTask(ctx, store.CreateTaskParams{
		JobID:         restream.ID,
		Type:          store.TaskTypeRestream,
		Params:        restream.OutputDestinations, // OutputDestinations contains the necessary info
		SequenceIndex: pgtype.Int4{Int32: 0, Valid: true},
	})
	if err != nil {
		return err
	}

	return o.dispatchTask(ctx, task)
}

func (o *Orchestrator) StopRestream(ctx context.Context, restreamID string) error {
	// For restreaming, stopping usually means cancelling the task
	// This would involve finding the active task for this job and cancelling it
	var uid pgtype.UUID
	if err := uid.Scan(restreamID); err != nil {
		return err
	}

	tasks, err := o.db.ListTasksByJob(ctx, uid)
	if err != nil {
		return err
	}

	for _, t := range tasks {
		if t.Status == "assigned" || t.Status == "pending" {
			// Fail or cancel the task
			o.db.UpdateTaskStatus(ctx, store.UpdateTaskStatusParams{
				ID:     t.ID,
				Status: "cancelled",
			})
		}
	}

	return nil
}
