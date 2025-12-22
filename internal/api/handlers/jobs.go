package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rennerdo30/webencode/internal/api/middleware"
	"github.com/rennerdo30/webencode/internal/orchestrator"
	"github.com/rennerdo30/webencode/internal/plugin_manager"
	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/errors"
	"github.com/rennerdo30/webencode/pkg/logger"
)

// Pagination limits
const (
	maxLimit     = 100
	defaultLimit = 10
)

// Valid source types for job creation
var validSourceTypes = map[string]bool{
	"url":      true,
	"upload":   true,
	"stream":   true,
	"restream": true,
	"":         true, // Empty defaults to "url"
}

// isValidURL checks if the given string is a valid URL with http/https or s3 scheme
func isValidURL(s string) bool {
	if s == "" {
		return false
	}
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	scheme := strings.ToLower(u.Scheme)
	return scheme == "http" || scheme == "https" || scheme == "s3" || scheme == "file"
}

// clampPagination ensures limit and offset are within valid bounds
func clampPagination(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

type JobResponse struct {
	ID           string          `json:"id"`
	SourceURL    string          `json:"source_url"`
	Profiles     []string        `json:"profiles"`
	Status       store.JobStatus `json:"status"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	ProgressPct  int             `json:"progress_pct"`
	ErrorMessage string          `json:"error_message,omitempty"`
	StartedAt    *time.Time      `json:"started_at,omitempty"`
	FinishedAt   *time.Time      `json:"finished_at,omitempty"`
	ETASeconds   int             `json:"eta_seconds,omitempty"`
}

type TaskResponse struct {
	ID            string          `json:"id"`
	JobID         string          `json:"job_id"`
	Type          store.TaskType  `json:"type"`
	Status        string          `json:"status"`
	SequenceIndex int             `json:"sequence_index"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	OutputKey     string          `json:"output_key,omitempty"`
	Result        json.RawMessage `json:"result,omitempty"`
}

func toJobResponse(j store.Job) JobResponse {
	var start, finish *time.Time
	if j.StartedAt.Valid {
		t := j.StartedAt.Time
		start = &t
	}
	if j.FinishedAt.Valid {
		t := j.FinishedAt.Time
		finish = &t
	}

	return JobResponse{
		ID:           j.ID.String(),
		SourceURL:    j.SourceUrl,
		Profiles:     j.Profiles,
		Status:       j.Status,
		CreatedAt:    j.CreatedAt.Time,
		UpdatedAt:    j.UpdatedAt.Time,
		ProgressPct:  int(j.ProgressPct.Int32),
		ErrorMessage: j.ErrorMessage.String,
		StartedAt:    start,
		FinishedAt:   finish,
		ETASeconds:   int(j.EtaSeconds.Int32),
	}
}

func toTaskResponse(t store.Task) TaskResponse {
	var result json.RawMessage
	if len(t.Result) > 0 {
		result = json.RawMessage(t.Result)
	}
	return TaskResponse{
		ID:            t.ID.String(),
		JobID:         t.JobID.String(),
		Type:          t.Type,
		Status:        t.Status,
		SequenceIndex: int(t.SequenceIndex.Int32),
		CreatedAt:     t.CreatedAt.Time,
		UpdatedAt:     t.UpdatedAt.Time,
		OutputKey:     t.OutputKey.String,
		Result:        result,
	}
}

type JobsHandler struct {
	orch   orchestrator.Service
	pm     *plugin_manager.Manager
	logger *logger.Logger
}

func NewJobsHandler(orch orchestrator.Service, pm *plugin_manager.Manager, l *logger.Logger) *JobsHandler {
	return &JobsHandler{orch: orch, pm: pm, logger: l}
}

func (h *JobsHandler) Register(r chi.Router) {
	r.Post("/v1/jobs", h.CreateJob)
	r.Get("/v1/jobs", h.ListJobs)
	r.Get("/v1/jobs/{id}", h.GetJob)
	r.Delete("/v1/jobs/{id}", h.DeleteJob)
	r.Post("/v1/jobs/{id}/cancel", h.CancelJob)
	r.Post("/v1/jobs/{id}/retry", h.RetryJob)
	r.Get("/v1/jobs/{id}/logs", h.GetJobLogs)
	r.Get("/v1/jobs/{id}/outputs", h.GetJobOutputs)
	r.Post("/v1/jobs/{id}/publish", h.PublishJob)
}

type CreateJobRequest struct {
	SourceURL  string   `json:"source_url"`
	SourceType string   `json:"source_type,omitempty"` // "url", "upload", "stream", "restream"
	Profiles   []string `json:"profiles"`
}

func (h *JobsHandler) CreateJob(w http.ResponseWriter, r *http.Request) {
	var req CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	// Validate source_url format
	if !isValidURL(req.SourceURL) {
		h.logger.Warn("Invalid source_url provided", "url", req.SourceURL)
		http.Error(w, "Invalid source_url: must be a valid URL with http, https, s3, or file scheme", http.StatusBadRequest)
		return
	}

	// Validate source_type
	if !validSourceTypes[req.SourceType] {
		h.logger.Warn("Invalid source_type provided", "type", req.SourceType)
		http.Error(w, "Invalid source_type: must be one of url, upload, stream, restream", http.StatusBadRequest)
		return
	}

	// Validate profiles array
	if len(req.Profiles) == 0 {
		h.logger.Warn("Empty profiles array provided")
		http.Error(w, "Profiles array must not be empty", http.StatusBadRequest)
		return
	}

	// Extract UserID from auth context
	userID := middleware.GetUserID(r.Context())

	job, err := h.orch.SubmitJob(r.Context(), orchestrator.JobRequest{
		UserID:     userID,
		SourceURL:  req.SourceURL,
		SourceType: req.SourceType,
		Profiles:   req.Profiles,
	})
	if err != nil {
		h.logger.Error("Failed to submit job", "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(toJobResponse(*job)); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

func (h *JobsHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	limit := defaultLimit
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil {
			offset = v
		}
	}

	// Validate and clamp pagination parameters
	limit, offset = clampPagination(limit, offset)

	jobs, err := h.orch.ListJobs(r.Context(), int32(limit), int32(offset))
	if err != nil {
		h.logger.Error("Failed to list jobs", "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	resp := make([]JobResponse, 0, len(jobs))
	for _, job := range jobs {
		resp = append(resp, toJobResponse(job))
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

func (h *JobsHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	job, err := h.orch.GetJob(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get job", "id", id, "error", err)
		errors.Response(w, r, errors.ErrNotFound)
		return
	}

	tasks, err := h.orch.GetJobTasks(r.Context(), id)
	if err != nil {
		h.logger.Warn("Failed to get tasks for job", "id", id, "error", err)
	}

	taskResp := make([]TaskResponse, 0, len(tasks))
	for _, task := range tasks {
		taskResp = append(taskResp, toTaskResponse(task))
	}

	response := map[string]interface{}{
		"job":   toJobResponse(*job),
		"tasks": taskResp,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

func (h *JobsHandler) DeleteJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	if err := h.orch.DeleteJob(r.Context(), id); err != nil {
		h.logger.Error("Failed to delete job", "id", id, "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *JobsHandler) CancelJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	if err := h.orch.CancelJob(r.Context(), id); err != nil {
		h.logger.Error("Failed to cancel job", "id", id, "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	if err := json.NewEncoder(w).Encode(map[string]string{"status": "cancelled"}); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

func (h *JobsHandler) RetryJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	job, err := h.orch.RestartJob(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to restart job", "id", id, "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(toJobResponse(*job)); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

func (h *JobsHandler) GetJobLogs(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	logs, err := h.orch.GetJobLogs(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get job logs", "id", id, "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	// Transform to response if needed
	resp := make([]JobLogResponse, 0, len(logs))
	for _, l := range logs {
		var metadata json.RawMessage
		if l.Metadata != nil {
			metadata = json.RawMessage(l.Metadata)
		}
		resp = append(resp, JobLogResponse{
			ID:        l.ID.String(),
			JobID:     l.JobID.String(),
			Level:     l.Level,
			Message:   l.Message,
			Metadata:  metadata,
			CreatedAt: l.CreatedAt.Time,
		})
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

type JobLogResponse struct {
	ID        string          `json:"id"`
	JobID     string          `json:"job_id"`
	Level     string          `json:"level"`
	Message   string          `json:"message"`
	Metadata  json.RawMessage `json:"metadata"`
	CreatedAt time.Time       `json:"created_at"`
}

// JobOutput represents an output file from a completed job
type JobOutput struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // "final", "segment", "manifest"
	URL         string `json:"url"`
	DownloadURL string `json:"download_url,omitempty"`
	Size        int64  `json:"size,omitempty"`
	Profile     string `json:"profile,omitempty"`
}

// JobOutputsResponse contains all outputs for a job
type JobOutputsResponse struct {
	JobID   string      `json:"job_id"`
	Status  string      `json:"status"`
	Outputs []JobOutput `json:"outputs"`
}

// GetJobOutputs returns the output files for a completed job
func (h *JobsHandler) GetJobOutputs(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	job, err := h.orch.GetJob(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get job", "id", id, "error", err)
		errors.Response(w, r, errors.ErrNotFound)
		return
	}

	// Get all tasks for this job
	tasks, err := h.orch.GetJobTasks(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get tasks", "id", id, "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	outputs := []JobOutput{}

	// Find the stitch task (final output) and transcode tasks
	for _, task := range tasks {
		if task.Status != "completed" || !task.OutputKey.Valid {
			continue
		}

		outputPath := task.OutputKey.String
		var outputType, profile string

		switch task.Type {
		case store.TaskTypeStitch:
			outputType = "final"
		case store.TaskTypeTranscode:
			outputType = "segment"
			// Extract profile from output key if present
			profile = "default"
		default:
			continue
		}

		// Generate download URL using storage plugin
		downloadURL := ""
		if h.pm != nil {
			for pluginID, client := range h.pm.Storage {
				storageClient, ok := client.(pb.StorageServiceClient)
				if !ok {
					continue
				}

				// Try to get a signed download URL
				resp, err := storageClient.GetURL(r.Context(), &pb.SignedUrlRequest{
					ObjectKey:     outputPath,
					ExpirySeconds: 3600, // 1 hour
					Method:        "GET",
				})
				if err == nil && resp.Url != "" {
					downloadURL = resp.Url
					h.logger.Debug("Generated download URL", "plugin", pluginID, "key", outputPath)
					break
				}
			}
		}

		outputs = append(outputs, JobOutput{
			Name:        outputPath,
			Type:        outputType,
			URL:         outputPath,
			DownloadURL: downloadURL,
			Profile:     profile,
		})
	}

	response := JobOutputsResponse{
		JobID:   id,
		Status:  string(job.Status),
		Outputs: outputs,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

// PublishJobRequest contains the request body for publishing a job
type PublishJobRequest struct {
	Platform    string `json:"platform"`     // e.g., "twitch", "youtube", "kick"
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	AccessToken string `json:"access_token"` // OAuth token for the platform
	OutputKey   string `json:"output_key,omitempty"` // Optional: specific output to publish (defaults to final)
}

// PublishJobResponse contains the result of publishing a job
type PublishJobResponse struct {
	Success     bool   `json:"success"`
	PlatformID  string `json:"platform_id,omitempty"`
	PlatformURL string `json:"platform_url,omitempty"`
	Message     string `json:"message,omitempty"`
}

// PublishJob publishes a completed job to an external platform
func (h *JobsHandler) PublishJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	var req PublishJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	if req.Platform == "" || req.AccessToken == "" {
		http.Error(w, "platform and access_token are required", http.StatusBadRequest)
		return
	}

	// Get the job and verify it's completed
	job, err := h.orch.GetJob(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get job", "id", id, "error", err)
		errors.Response(w, r, errors.ErrNotFound)
		return
	}

	if job.Status != store.JobStatusCompleted {
		http.Error(w, "Job must be completed before publishing", http.StatusBadRequest)
		return
	}

	// Get the output file to publish
	tasks, err := h.orch.GetJobTasks(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get tasks", "id", id, "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	// Find the final output (stitch task) or specified output
	var outputKey string
	for _, task := range tasks {
		if task.Status != "completed" || !task.OutputKey.Valid {
			continue
		}

		if req.OutputKey != "" {
			// Use the specified output
			if task.OutputKey.String == req.OutputKey {
				outputKey = task.OutputKey.String
				break
			}
		} else if task.Type == store.TaskTypeStitch {
			// Use the final stitched output
			outputKey = task.OutputKey.String
			break
		}
	}

	if outputKey == "" {
		http.Error(w, "No output file found for publishing", http.StatusBadRequest)
		return
	}

	// Get the download URL for the output file
	var fileURL string
	if h.pm != nil {
		for _, client := range h.pm.Storage {
			storageClient, ok := client.(pb.StorageServiceClient)
			if !ok {
				continue
			}
			resp, err := storageClient.GetURL(r.Context(), &pb.SignedUrlRequest{
				ObjectKey:     outputKey,
				ExpirySeconds: 86400, // 24 hours for upload
				Method:        "GET",
			})
			if err == nil && resp.Url != "" {
				fileURL = resp.Url
				break
			}
		}
	}

	if fileURL == "" {
		http.Error(w, "Could not generate download URL for output file", http.StatusInternalServerError)
		return
	}

	// Find the publisher plugin for the requested platform
	pluginID := "publisher-" + req.Platform
	if h.pm == nil {
		http.Error(w, "Plugin manager not available", http.StatusInternalServerError)
		return
	}

	publisherClient, ok := h.pm.Publisher[pluginID]
	if !ok {
		http.Error(w, "Publisher plugin not found: "+pluginID, http.StatusBadRequest)
		return
	}

	pubClient, ok := publisherClient.(pb.PublisherServiceClient)
	if !ok {
		http.Error(w, "Invalid publisher plugin", http.StatusInternalServerError)
		return
	}

	// Call the publisher plugin
	title := req.Title
	if title == "" {
		title = "Video from WebEncode"
	}

	publishReq := &pb.PublishRequest{
		Platform:    req.Platform,
		FileUrl:     fileURL,
		Title:       title,
		Description: req.Description,
		AccessToken: req.AccessToken,
	}

	h.logger.Info("Publishing job", "job_id", id, "platform", req.Platform, "file", outputKey)

	result, err := pubClient.Publish(r.Context(), publishReq)
	if err != nil {
		h.logger.Error("Publish failed", "job_id", id, "platform", req.Platform, "error", err)
		http.Error(w, "Publishing failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := PublishJobResponse{
		Success:     true,
		PlatformID:  result.PlatformId,
		PlatformURL: result.Url,
		Message:     "Successfully published to " + req.Platform,
	}

	h.logger.Info("Job published successfully", "job_id", id, "platform", req.Platform, "url", result.Url)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}
