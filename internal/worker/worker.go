package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/rennerdo30/webencode/internal/plugin_manager"
	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/rennerdo30/webencode/pkg/bus"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/ffmpeg"
	"github.com/rennerdo30/webencode/pkg/hardware"
	"github.com/rennerdo30/webencode/pkg/logger"
)

// execCommandContext allows mocking exec.CommandContext
var execCommandContext = exec.CommandContext

type MessageBus interface {
	JetStream() jetstream.JetStream
	Publish(ctx context.Context, subject string, data []byte) error
}

type Encoder interface {
	Probe(ctx context.Context, url string) (*ffmpeg.ProbeResult, error)
	Transcode(ctx context.Context, task *ffmpeg.TranscodeTask, progressCh chan<- ffmpeg.Progress) error
}

type Worker struct {
	id            string
	bus           MessageBus
	caps          *hardware.Capabilities
	logger        *logger.Logger
	pluginManager *plugin_manager.Manager
	encoder       Encoder
	workDir       string
	isBusy        bool
}

func (w *Worker) Status() string {
	if w.isBusy {
		return "busy"
	}
	return "idle"
}

func New(id string, b MessageBus, caps *hardware.Capabilities, l *logger.Logger, pluginDir string) *Worker {
	// Create work directory
	workDir := filepath.Join(os.TempDir(), "webencode_worker_"+id)
	os.MkdirAll(workDir, 0755)

	return &Worker{
		id:            id,
		bus:           b,
		caps:          caps,
		logger:        l,
		pluginManager: plugin_manager.New(l, pluginDir),
		encoder:       ffmpeg.NewFFmpegEncoder("ffmpeg", "ffprobe", l), // Assumes in PATH
		workDir:       workDir,
	}
}

func (w *Worker) Start(ctx context.Context) error {
	// Initialize Plugins (Workers need Encoder/Storage plugins)
	if err := w.pluginManager.LoadAll(); err != nil {
		w.logger.Error("Failed to load plugins", "error", err)
	}

	w.logger.Info("Worker started", "id", w.id, "work_dir", w.workDir)

	// Subscribe to Work Queue via NATS
	// Using "worker" queue group for load balancing
	consumer, err := w.bus.JetStream().CreateOrUpdateConsumer(ctx, bus.StreamWork, jetstream.ConsumerConfig{
		Durable:       "worker_pool",
		AckPolicy:     jetstream.AckExplicitPolicy,
		FilterSubject: bus.SubjectJobDispatch,
	})
	if err != nil {
		return fmt.Errorf("create consumer: %w", err)
	}

	cons, err := consumer.Consume(func(msg jetstream.Msg) {
		w.handleMessage(ctx, msg)
	})
	if err != nil {
		return err
	}

	// Keep running until context cancelled
	<-ctx.Done()
	cons.Stop()
	return nil
}

func (w *Worker) handleMessage(ctx context.Context, msg jetstream.Msg) {
	// Don't ack immediately. Ack on completion.
	// We should set a timeout/NAK policy but simple Ack on success is MVP.

	// Panic Recovery
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
			stack := string(debug.Stack())
			w.logger.Error("Worker panicked", "error", err, "stack", stack)

			// Report system error
			w.reportError(err, map[string]interface{}{
				"panic": true,
				"stack": stack,
			})

			msg.Term() // Don't retry panicking tasks indefinitely
		}
	}()

	var task store.Task
	if err := json.Unmarshal(msg.Data(), &task); err != nil {
		w.logger.Error("Invalid task format", "error", err)
		msg.Term() // Don't retry malformed
		return
	}

	w.logger.Info("Processing task", "id", task.ID, "type", task.Type)
	w.isBusy = true
	defer func() { w.isBusy = false }()

	var result []byte
	var err error

	switch task.Type {
	case store.TaskTypeProbe:
		result, err = w.handleProbe(ctx, task)
	case store.TaskTypeTranscode:
		result, err = w.handleTranscode(ctx, task)
	case store.TaskTypeStitch:
		result, err = w.handleStitch(ctx, task)
	case store.TaskTypeRestream:
		result, err = w.handleRestream(ctx, task)
	case store.TaskTypeManifest:
		result, err = w.handleManifest(ctx, task)
	default:
		w.logger.Warn("Unknown task type", "type", task.Type)
		msg.Term()
		return
	}

	if err != nil {
		w.logger.Error("Task failed", "id", task.ID, "error", err)
		errorPayload, _ := json.Marshal(map[string]string{"error": err.Error()})
		w.sendEvent(ctx, task.ID.String(), "failed", errorPayload)

		// Report as system error if it's not a user error?
		// For now, let's keep task failures as task events, and panics as system errors.
		// Unless it's an unexpected error.

		msg.Nak() // Retry? Or Term if fatal? For now Nak to retry.
		return
	}

	// Send completion event
	w.sendEvent(ctx, task.ID.String(), "completed", result)
	msg.Ack()
	w.logger.Info("Task completed", "id", task.ID)
}

func (w *Worker) sendLog(ctx context.Context, taskID string, level string, message string) {
	payload := map[string]interface{}{
		"task_id": taskID,
		"event":   "log",
		"result": map[string]string{
			"level":   level,
			"message": message,
		},
	}
	data, _ := json.Marshal(payload)
	w.bus.Publish(ctx, bus.SubjectJobEvents, data)
}

func (w *Worker) sendEvent(ctx context.Context, taskID string, eventType string, result []byte) {
	// Payload for job events
	payload := map[string]interface{}{
		"task_id": taskID,
		"event":   eventType,
		"result":  json.RawMessage(result),
	}
	data, _ := json.Marshal(payload)

	w.bus.Publish(ctx, bus.SubjectJobEvents, data)
}

func (w *Worker) reportError(err error, contextData map[string]interface{}) {
	payload := map[string]interface{}{
		"source":       fmt.Sprintf("worker:%s", w.id),
		"severity":     "critical",
		"message":      err.Error(),
		"context_data": contextData,
	}

	if stack, ok := contextData["stack"].(string); ok {
		payload["stack_trace"] = stack
	}

	data, _ := json.Marshal(payload)
	w.bus.Publish(context.Background(), bus.SubjectErrorEvents, data)
}

func (w *Worker) handleProbe(ctx context.Context, task store.Task) ([]byte, error) {
	w.sendLog(ctx, task.ID.String(), "info", "Starting media probe")

	var params map[string]string
	if err := json.Unmarshal(task.Params, &params); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	url := params["url"]

	// Handle S3 URLs - download file first since ffprobe can't access s3:// directly
	probeURL := url
	var localFile string
	if strings.HasPrefix(url, "s3://") {
		bucket, key, err := parseS3URL(url)
		if err != nil {
			return nil, fmt.Errorf("invalid s3 url: %w", err)
		}
		localFile = filepath.Join(w.workDir, "probe_"+filepath.Base(key))
		w.sendLog(ctx, task.ID.String(), "info", fmt.Sprintf("Downloading from S3: %s", url))
		if err := w.downloadFile(ctx, bucket, key, localFile); err != nil {
			return nil, fmt.Errorf("s3 download failed: %w", err)
		}
		probeURL = localFile
		defer os.Remove(localFile) // Cleanup after probe
	}

	w.sendLog(ctx, task.ID.String(), "info", fmt.Sprintf("Probing: %s", probeURL))

	res, err := w.encoder.Probe(ctx, probeURL)
	if err != nil {
		w.sendLog(ctx, task.ID.String(), "error", fmt.Sprintf("Probe failed: %s", err))
		return nil, err
	}

	w.sendLog(ctx, task.ID.String(), "info", fmt.Sprintf("Probe successful. Duration: %.2fs", res.Duration))
	return json.Marshal(res)
}

func (w *Worker) handleTranscode(ctx context.Context, task store.Task) ([]byte, error) {
	var tTask ffmpeg.TranscodeTask
	if err := json.Unmarshal(task.Params, &tTask); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	// 1. Handle Input (Download if S3)
	originalInput := tTask.InputURL
	localInput := originalInput
	if strings.HasPrefix(originalInput, "s3://") {
		bucket, key, err := parseS3URL(originalInput)
		if err != nil {
			return nil, fmt.Errorf("invalid input url: %w", err)
		}
		localInput = filepath.Join(w.workDir, "input_"+filepath.Base(key))
		w.logger.Info("Downloading input", "url", originalInput, "local", localInput)
		if err := w.downloadFile(ctx, bucket, key, localInput); err != nil {
			return nil, fmt.Errorf("download failed: %w", err)
		}
		// Defer cleanup? Maybe
	}
	tTask.InputURL = localInput

	// 2. Handle Output (Determine local path)
	originalOutput := tTask.OutputURL
	localOutput := filepath.Join(w.workDir, filepath.Base(originalOutput))
	if strings.HasPrefix(originalOutput, "s3://") {
		localOutput = filepath.Join(w.workDir, "output_"+filepath.Base(originalOutput))
	}
	tTask.OutputURL = localOutput

	progressCh := make(chan ffmpeg.Progress)

	// Monitor progress
	go func() {
		for p := range progressCh {
			// Log every 10%
			if int(p.Percent)%10 == 0 {
				w.logger.Debug("Encoding progress", "task_id", task.ID, "percent", p.Percent)
				progressData := fmt.Sprintf(`{"percent": %f}`, p.Percent)
				w.sendEvent(ctx, task.ID.String(), "progress", []byte(progressData))
			}
		}
	}()

	w.logger.Info("Starting transcode", "input", localInput, "output", localOutput)
	if err := w.encoder.Transcode(ctx, &tTask, progressCh); err != nil {
		return nil, fmt.Errorf("transcode failed: %w", err)
	}

	// 3. Handle Output Upload
	finalURL := localOutput // Default to local path return
	if strings.HasPrefix(originalOutput, "s3://") {
		bucket, key, err := parseS3URL(originalOutput)
		if err != nil {
			return nil, fmt.Errorf("invalid output url: %w", err)
		}
		w.logger.Info("Uploading output", "local", localOutput, "url", originalOutput)
		url, err := w.uploadFile(ctx, localOutput, bucket, key)
		if err != nil {
			return nil, fmt.Errorf("upload failed: %w", err)
		}
		finalURL = url
	}

	w.sendLog(ctx, task.ID.String(), "info", fmt.Sprintf("Transcode finished. Output: %s", finalURL))

	return json.Marshal(map[string]string{
		"output_path": finalURL,
	})
}

func (w *Worker) handleStitch(ctx context.Context, task store.Task) ([]byte, error) {
	w.sendLog(ctx, task.ID.String(), "info", "Starting stitch operation")

	var params struct {
		Segments []string `json:"segments"`
		Output   string   `json:"output"`
	}
	if err := json.Unmarshal(task.Params, &params); err != nil {
		w.sendLog(ctx, task.ID.String(), "error", fmt.Sprintf("Invalid stitch params: %s", err))
		return nil, err
	}

	w.sendLog(ctx, task.ID.String(), "info", fmt.Sprintf("Stitching %d segments into %s", len(params.Segments), params.Output))

	// Re-map paths if needed (assuming shared dir)
	var segmentPaths []string
	var missingSegments []string
	for _, seg := range params.Segments {
		// Check if path is already absolute, otherwise join with workDir
		segPath := seg
		if !filepath.IsAbs(seg) {
			segPath = filepath.Join(w.workDir, seg)
		}
		segmentPaths = append(segmentPaths, segPath)

		// Check if segment exists
		if _, err := os.Stat(segPath); os.IsNotExist(err) {
			missingSegments = append(missingSegments, segPath)
		}
	}

	// Log missing segments
	if len(missingSegments) > 0 {
		errMsg := fmt.Sprintf("Missing %d segment files: %v", len(missingSegments), missingSegments)
		w.sendLog(ctx, task.ID.String(), "error", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	outputPath := filepath.Join(w.workDir, params.Output)

	// Create concat file
	concatContent := ffmpeg.BuildConcatDemuxer(segmentPaths)
	concatFile := filepath.Join(w.workDir, fmt.Sprintf("concat_%x.txt", task.ID.Bytes))
	w.sendLog(ctx, task.ID.String(), "info", fmt.Sprintf("Creating concat file: %s", concatFile))

	if err := os.WriteFile(concatFile, []byte(concatContent), 0644); err != nil {
		w.sendLog(ctx, task.ID.String(), "error", fmt.Sprintf("Failed to write concat file: %s", err))
		return nil, err
	}

	// Log the concat file content for debugging
	w.sendLog(ctx, task.ID.String(), "debug", fmt.Sprintf("Concat file content:\n%s", concatContent))

	// Stitch using FFmpeg
	args := ffmpeg.StitchCommand(concatFile, outputPath)
	w.sendLog(ctx, task.ID.String(), "info", fmt.Sprintf("Executing: ffmpeg %s", strings.Join(args, " ")))

	cmd := w.makeCmd(ctx, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Log the full FFmpeg output for debugging
		ffmpegOutput := string(out)
		w.logger.Error("Stitch failed", "output", ffmpegOutput, "error", err)
		w.sendLog(ctx, task.ID.String(), "error", fmt.Sprintf("FFmpeg stitch failed with exit code: %v", err))
		w.sendLog(ctx, task.ID.String(), "error", fmt.Sprintf("FFmpeg output:\n%s", ffmpegOutput))
		return nil, fmt.Errorf("stitch failed: %w (output: %s)", err, ffmpegOutput)
	}

	w.sendLog(ctx, task.ID.String(), "info", fmt.Sprintf("Stitch completed successfully. Output: %s", outputPath))

	// Return the absolute path as the "URL" for local/shared volume access
	// The frontend proxy will handle reading this path from the shared /tmp volume.
	return json.Marshal(map[string]string{
		"output_path": outputPath,
	})
}

func (w *Worker) handleRestream(ctx context.Context, task store.Task) ([]byte, error) {
	var req pb.PublishRequest
	if err := json.Unmarshal(task.Params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	// Handle S3 Input
	if strings.HasPrefix(req.FileUrl, "s3://") {
		bucket, key, err := parseS3URL(req.FileUrl)
		if err != nil {
			return nil, fmt.Errorf("invalid input url: %w", err)
		}
		localInput := filepath.Join(w.workDir, "restream_input_"+filepath.Base(key))
		w.logger.Info("Downloading restream input", "url", req.FileUrl, "local", localInput)

		if err := w.downloadFile(ctx, bucket, key, localInput); err != nil {
			return nil, fmt.Errorf("download failed: %w", err)
		}

		// Ensure cleanup of downloaded file
		defer func() {
			os.Remove(localInput)
		}()

		req.FileUrl = localInput
	}

	// Find Publisher Plugin
	var client pb.PublisherServiceClient
	for _, c := range w.pluginManager.Publisher {
		if sc, ok := c.(pb.PublisherServiceClient); ok {
			client = sc
			break
		}
	}
	if client == nil {
		return nil, fmt.Errorf("no publisher plugin available")
	}

	w.logger.Info("Publishing", "platform", req.Platform, "url", req.FileUrl)
	res, err := client.Publish(ctx, &req)
	if err != nil {
		return nil, err
	}

	return json.Marshal(res)
}

func (w *Worker) handleManifest(ctx context.Context, task store.Task) ([]byte, error) {
	var params struct {
		Variants []ffmpeg.HLSVariant `json:"variants"`
		Output   string              `json:"output"`
	}
	if err := json.Unmarshal(task.Params, &params); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	content := ffmpeg.BuildMasterPlaylist(params.Variants)

	// Determine Output Path
	localOutput := filepath.Join(w.workDir, "master.m3u8")
	if err := os.WriteFile(localOutput, []byte(content), 0644); err != nil {
		return nil, err
	}

	// Upload if needed
	finalURL := localOutput
	if strings.HasPrefix(params.Output, "s3://") {
		bucket, key, err := parseS3URL(params.Output)
		if err != nil {
			return nil, err
		}
		url, err := w.uploadFile(ctx, localOutput, bucket, key)
		if err != nil {
			return nil, err
		}
		finalURL = url
	}

	return json.Marshal(map[string]string{"output_path": finalURL})
}

// makeCmd helper for testing
func (w *Worker) makeCmd(ctx context.Context, args ...string) *exec.Cmd {
	return execCommandContext(ctx, "ffmpeg", args...)
}
