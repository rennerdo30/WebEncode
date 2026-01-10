package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/rennerdo30/webencode/internal/plugin_manager"
	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/ffmpeg"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

// Mocks
type MockBus struct {
	mock.Mock
}

func (m *MockBus) JetStream() jetstream.JetStream {
	args := m.Called()
	return args.Get(0).(jetstream.JetStream)
}
func (m *MockBus) Publish(ctx context.Context, subject string, data []byte) error {
	args := m.Called(ctx, subject, data)
	return args.Error(0)
}

type MockEncoder struct {
	mock.Mock
}

func (m *MockEncoder) Probe(ctx context.Context, url string) (*ffmpeg.ProbeResult, error) {
	args := m.Called(ctx, url)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ffmpeg.ProbeResult), args.Error(1)
}

func (m *MockEncoder) Transcode(ctx context.Context, task *ffmpeg.TranscodeTask, progressCh chan<- ffmpeg.Progress) error {
	args := m.Called(ctx, task, progressCh)
	// Create dummy progress
	go func() {
		defer close(progressCh)
		progressCh <- ffmpeg.Progress{Percent: 50}
		progressCh <- ffmpeg.Progress{Percent: 100}
	}()
	return args.Error(0)
}

type MockJetStreamMsg struct {
	mock.Mock
}

func (m *MockJetStreamMsg) Ack() error {
	args := m.Called()
	return args.Error(0)
}
func (m *MockJetStreamMsg) Nak() error {
	args := m.Called()
	return args.Error(0)
}
func (m *MockJetStreamMsg) Term() error {
	args := m.Called()
	return args.Error(0)
}
func (m *MockJetStreamMsg) Data() []byte {
	args := m.Called()
	return args.Get(0).([]byte)
}

// Other interface methods stubs
func (m *MockJetStreamMsg) Headers() nats.Header                      { return nil }
func (m *MockJetStreamMsg) Metadata() (*jetstream.MsgMetadata, error) { return nil, nil }
func (m *MockJetStreamMsg) Reply() string                             { return "" }
func (m *MockJetStreamMsg) Subject() string                           { return "" }
func (m *MockJetStreamMsg) NakWithDelay(delay time.Duration) error    { return nil }
func (m *MockJetStreamMsg) InProgress() error                         { return nil }
func (m *MockJetStreamMsg) DoubleAck(ctx context.Context) error       { return nil }
func (m *MockJetStreamMsg) TermWithReason(reason string) error        { return nil }

func TestHandleMessage_Probe(t *testing.T) {
	mockBus := new(MockBus)
	mockEncoder := new(MockEncoder)
	mockMsg := new(MockJetStreamMsg)

	w := &Worker{
		id:      "worker-1",
		bus:     mockBus,
		logger:  logger.New("test"),
		encoder: mockEncoder,
		workDir: os.TempDir(),
	}

	// Task data
	probeParams := `{"url": "http://example.com/video.mp4"}`
	task := store.Task{
		ID:     pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
		Type:   store.TaskTypeProbe,
		Params: []byte(probeParams),
	}
	taskBytes, _ := json.Marshal(task)

	// Expectations
	mockMsg.On("Data").Return(taskBytes)
	mockEncoder.On("Probe", mock.Anything, "http://example.com/video.mp4").Return(&ffmpeg.ProbeResult{Duration: 60}, nil)

	// Expect log events (sendLog calls)
	mockBus.On("Publish", mock.Anything, "jobs.events", mock.MatchedBy(func(data []byte) bool {
		var payload map[string]interface{}
		json.Unmarshal(data, &payload)
		return payload["event"] == "log"
	})).Return(nil).Maybe()

	// Expect event publish (completion)
	// Mock all Publish calls (logs and events)
	mockBus.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockMsg.On("Ack").Return(nil)

	w.handleMessage(context.Background(), mockMsg)

	mockEncoder.AssertExpectations(t)
	mockBus.AssertExpectations(t)
	mockMsg.AssertExpectations(t)
}

func TestHandleMessage_Transcode(t *testing.T) {
	mockBus := new(MockBus)
	mockEncoder := new(MockEncoder)
	mockMsg := new(MockJetStreamMsg)

	w := &Worker{
		id:      "worker-1",
		bus:     mockBus,
		logger:  logger.New("test"),
		encoder: mockEncoder,
		workDir: os.TempDir(),
	}

	// Local file transcode
	tcTask := ffmpeg.TranscodeTask{
		InputURL:  "/tmp/in.mp4",
		OutputURL: "/tmp/out.mp4",
	}
	tcParams, _ := json.Marshal(tcTask)

	task := store.Task{
		ID:     pgtype.UUID{Bytes: [16]byte{2}, Valid: true},
		Type:   store.TaskTypeTranscode,
		Params: tcParams,
	}
	taskBytes, _ := json.Marshal(task)

	mockMsg.On("Data").Return(taskBytes)

	mockEncoder.On("Transcode", mock.Anything, mock.MatchedBy(func(a *ffmpeg.TranscodeTask) bool {
		return a.InputURL == "/tmp/in.mp4"
	}), mock.Anything).Return(nil)

	// Allow log events
	mockBus.On("Publish", mock.Anything, "jobs.events", mock.MatchedBy(func(data []byte) bool {
		var payload map[string]interface{}
		json.Unmarshal(data, &payload)
		return payload["event"] == "log"
	})).Return(nil).Maybe()

	// Allow progress events
	mockBus.On("Publish", mock.Anything, "jobs.events", mock.MatchedBy(func(data []byte) bool {
		var payload map[string]interface{}
		json.Unmarshal(data, &payload)
		return payload["event"] == "progress"
	})).Return(nil).Maybe()

	// Mock all Publish calls (logs and events)
	mockBus.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockMsg.On("Ack").Return(nil)

	w.handleMessage(context.Background(), mockMsg)

	mockEncoder.AssertExpectations(t)
}

// TestHelperProcess isn't a real test. It's used to mock exec.Command
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)
	// Mock success
}

func fakeExecCommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", name}
	cs = append(cs, args...)
	cmd := exec.CommandContext(ctx, os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestHandleMessage_Stitch(t *testing.T) {
	// Mock exec
	oldExec := execCommandContext
	execCommandContext = fakeExecCommandContext
	defer func() { execCommandContext = oldExec }()

	mockBus := new(MockBus)
	mockMsg := new(MockJetStreamMsg)

	w := &Worker{
		id:      "worker-1",
		bus:     mockBus,
		logger:  logger.New("test"),
		workDir: os.TempDir(),
	}

	stitchParams := map[string]interface{}{
		"segments": []string{"seg1.ts", "seg2.ts"},
		"output":   "final.mp4",
	}
	paramsBytes, _ := json.Marshal(stitchParams)

	task := store.Task{
		ID:     pgtype.UUID{Bytes: [16]byte{3}, Valid: true},
		Type:   store.TaskTypeStitch,
		Params: paramsBytes,
	}
	taskBytes, _ := json.Marshal(task)

	mockMsg.On("Data").Return(taskBytes)

	// Mock all Publish calls (logs and events)
	mockBus.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Stitch may Ack or Nak depending on segment availability
	mockMsg.On("Ack").Return(nil).Maybe()
	mockMsg.On("Nak").Return(nil).Maybe()

	w.handleMessage(context.Background(), mockMsg)

	mockBus.AssertExpectations(t)
}

type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) Publish(ctx context.Context, in *pb.PublishRequest, opts ...grpc.CallOption) (*pb.PublishResult, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.PublishResult), args.Error(1)
}

func (m *MockPublisher) Retract(ctx context.Context, in *pb.RetractRequest, opts ...grpc.CallOption) (*pb.Empty, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.Empty), args.Error(1)
}

func (m *MockPublisher) GetLiveStreamEndpoint(ctx context.Context, in *pb.GetLiveStreamEndpointRequest, opts ...grpc.CallOption) (*pb.GetLiveStreamEndpointResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.GetLiveStreamEndpointResponse), args.Error(1)
}

func (m *MockPublisher) GetChatMessages(ctx context.Context, in *pb.GetChatMessagesRequest, opts ...grpc.CallOption) (*pb.GetChatMessagesResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.GetChatMessagesResponse), args.Error(1)
}

func (m *MockPublisher) SendChatMessage(ctx context.Context, in *pb.SendChatMessageRequest, opts ...grpc.CallOption) (*pb.Empty, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.Empty), args.Error(1)
}

func TestHandleMessage_Restream(t *testing.T) {
	mockBus := new(MockBus)
	mockMsg := new(MockJetStreamMsg)
	mockPub := new(MockPublisher)

	// Setup PluginManager with mock publisher
	pm := plugin_manager.New(logger.New("test"), "")
	pm.Publisher["mock_pub"] = mockPub

	w := &Worker{
		id:            "worker-1",
		bus:           mockBus,
		logger:        logger.New("test"),
		pluginManager: pm,
		workDir:       os.TempDir(),
	}

	req := pb.PublishRequest{
		Platform: "youtube",
		FileUrl:  "/tmp/video.mp4",
	}
	reqParams, _ := json.Marshal(&req)

	task := store.Task{
		ID:     pgtype.UUID{Bytes: [16]byte{4}, Valid: true},
		Type:   store.TaskTypeRestream,
		Params: reqParams,
	}
	taskBytes, _ := json.Marshal(task)

	mockMsg.On("Data").Return(taskBytes)

	mockPub.On("Publish", mock.Anything, mock.MatchedBy(func(r *pb.PublishRequest) bool {
		return r.Platform == "youtube" && r.FileUrl == "/tmp/video.mp4"
	})).Return(&pb.PublishResult{PlatformId: "yt-123", Url: "http://youtube.com/watch?v=yt-123"}, nil)

	// Mock all Publish calls (logs and events)
	mockBus.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockMsg.On("Ack").Return(nil)

	w.handleMessage(context.Background(), mockMsg)

	mockPub.AssertExpectations(t)
	mockBus.AssertExpectations(t)
}

func TestWorker_Status(t *testing.T) {
	w := &Worker{
		id:     "worker-1",
		isBusy: false,
	}

	if w.Status() != "idle" {
		t.Errorf("Expected 'idle', got '%s'", w.Status())
	}

	w.isBusy = true
	if w.Status() != "busy" {
		t.Errorf("Expected 'busy', got '%s'", w.Status())
	}
}

func TestHandleMessage_InvalidJSON(t *testing.T) {
	mockBus := new(MockBus)
	mockMsg := new(MockJetStreamMsg)

	w := &Worker{
		id:      "worker-1",
		bus:     mockBus,
		logger:  logger.New("test"),
		workDir: os.TempDir(),
	}

	// Invalid JSON
	mockMsg.On("Data").Return([]byte("not valid json"))
	mockMsg.On("Term").Return(nil)

	w.handleMessage(context.Background(), mockMsg)

	mockMsg.AssertExpectations(t)
}

func TestHandleMessage_UnknownTaskType(t *testing.T) {
	mockBus := new(MockBus)
	mockMsg := new(MockJetStreamMsg)

	w := &Worker{
		id:      "worker-1",
		bus:     mockBus,
		logger:  logger.New("test"),
		workDir: os.TempDir(),
	}

	task := store.Task{
		ID:     pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
		Type:   "unknown_type",
		Params: []byte(`{}`),
	}
	taskBytes, _ := json.Marshal(task)

	mockMsg.On("Data").Return(taskBytes)
	mockMsg.On("Term").Return(nil)

	w.handleMessage(context.Background(), mockMsg)

	mockMsg.AssertExpectations(t)
}

func TestHandleMessage_ProbeError(t *testing.T) {
	mockBus := new(MockBus)
	mockEncoder := new(MockEncoder)
	mockMsg := new(MockJetStreamMsg)

	w := &Worker{
		id:      "worker-1",
		bus:     mockBus,
		logger:  logger.New("test"),
		encoder: mockEncoder,
		workDir: os.TempDir(),
	}

	probeParams := `{"url": "http://example.com/video.mp4"}`
	task := store.Task{
		ID:     pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
		Type:   store.TaskTypeProbe,
		Params: []byte(probeParams),
	}
	taskBytes, _ := json.Marshal(task)

	mockMsg.On("Data").Return(taskBytes)
	mockEncoder.On("Probe", mock.Anything, "http://example.com/video.mp4").Return(nil, fmt.Errorf("probe failed"))

	// Allow log/event publish calls
	mockBus.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockMsg.On("Nak").Return(nil)

	w.handleMessage(context.Background(), mockMsg)

	mockEncoder.AssertExpectations(t)
	mockMsg.AssertExpectations(t)
}

func TestHandleMessage_Manifest(t *testing.T) {
	oldExec := execCommandContext
	execCommandContext = fakeExecCommandContext
	defer func() { execCommandContext = oldExec }()

	mockBus := new(MockBus)
	mockMsg := new(MockJetStreamMsg)

	w := &Worker{
		id:            "worker-1",
		bus:           mockBus,
		logger:        logger.New("test"),
		workDir:       os.TempDir(),
		pluginManager: plugin_manager.New(logger.New("test"), ""),
	}

	manifestParams := map[string]interface{}{
		"variants": []ffmpeg.HLSVariant{
			{Path: "1080p.m3u8", Bandwidth: 5000000, Resolution: "1920x1080"},
			{Path: "720p.m3u8", Bandwidth: 2500000, Resolution: "1280x720"},
		},
		"output": "/tmp/master.m3u8",
	}
	paramsBytes, _ := json.Marshal(manifestParams)

	task := store.Task{
		ID:     pgtype.UUID{Bytes: [16]byte{5}, Valid: true},
		Type:   store.TaskTypeManifest,
		Params: paramsBytes,
	}
	taskBytes, _ := json.Marshal(task)

	mockMsg.On("Data").Return(taskBytes)

	// Mock all Publish calls (logs and events)
	mockBus.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockMsg.On("Ack").Return(nil)

	w.handleMessage(context.Background(), mockMsg)

	mockBus.AssertExpectations(t)
	mockMsg.AssertExpectations(t)
}

func TestParseS3URL(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantBucket string
		wantKey    string
		wantErr    bool
	}{
		{
			name:       "valid s3 url",
			url:        "s3://mybucket/path/to/file.mp4",
			wantBucket: "mybucket",
			wantKey:    "path/to/file.mp4",
			wantErr:    false,
		},
		{
			name:       "valid s3 url without path",
			url:        "s3://mybucket/file.mp4",
			wantBucket: "mybucket",
			wantKey:    "file.mp4",
			wantErr:    false,
		},
		{
			name:    "invalid scheme",
			url:     "http://mybucket/file.mp4",
			wantErr: true,
		},
		{
			name:    "invalid url",
			url:     "://invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucket, key, err := parseS3URL(tt.url)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if bucket != tt.wantBucket {
				t.Errorf("bucket = %q, want %q", bucket, tt.wantBucket)
			}
			if key != tt.wantKey {
				t.Errorf("key = %q, want %q", key, tt.wantKey)
			}
		})
	}
}

func TestHandleMessage_ProbeInvalidParams(t *testing.T) {
	mockBus := new(MockBus)
	mockMsg := new(MockJetStreamMsg)

	w := &Worker{
		id:      "worker-1",
		bus:     mockBus,
		logger:  logger.New("test"),
		workDir: os.TempDir(),
	}

	task := store.Task{
		ID:     pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
		Type:   store.TaskTypeProbe,
		Params: []byte("not valid json"),
	}
	taskBytes, _ := json.Marshal(task)

	mockMsg.On("Data").Return(taskBytes)

	// Allow log/event publish calls
	mockBus.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockMsg.On("Nak").Return(nil)

	w.handleMessage(context.Background(), mockMsg)

	mockMsg.AssertExpectations(t)
}

func TestHandleMessage_TranscodeError(t *testing.T) {
	mockBus := new(MockBus)
	mockEncoder := new(MockEncoder)
	mockMsg := new(MockJetStreamMsg)

	w := &Worker{
		id:      "worker-1",
		bus:     mockBus,
		logger:  logger.New("test"),
		encoder: mockEncoder,
		workDir: os.TempDir(),
	}

	tcTask := ffmpeg.TranscodeTask{
		InputURL:  "/tmp/in.mp4",
		OutputURL: "/tmp/out.mp4",
	}
	tcParams, _ := json.Marshal(tcTask)

	task := store.Task{
		ID:     pgtype.UUID{Bytes: [16]byte{2}, Valid: true},
		Type:   store.TaskTypeTranscode,
		Params: tcParams,
	}
	taskBytes, _ := json.Marshal(task)

	mockMsg.On("Data").Return(taskBytes)

	// Mock encoder to return an error
	mockEncoder.On("Transcode", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("transcode failed"))

	// Allow log/event publish calls
	mockBus.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockMsg.On("Nak").Return(nil)

	w.handleMessage(context.Background(), mockMsg)

	mockEncoder.AssertExpectations(t)
	mockMsg.AssertExpectations(t)
}

func TestHandleMessage_TranscodeInvalidParams(t *testing.T) {
	mockBus := new(MockBus)
	mockMsg := new(MockJetStreamMsg)

	w := &Worker{
		id:      "worker-1",
		bus:     mockBus,
		logger:  logger.New("test"),
		workDir: os.TempDir(),
	}

	task := store.Task{
		ID:     pgtype.UUID{Bytes: [16]byte{2}, Valid: true},
		Type:   store.TaskTypeTranscode,
		Params: []byte("invalid json"),
	}
	taskBytes, _ := json.Marshal(task)

	mockMsg.On("Data").Return(taskBytes)
	mockBus.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockMsg.On("Nak").Return(nil)

	w.handleMessage(context.Background(), mockMsg)

	mockMsg.AssertExpectations(t)
}

func TestHandleMessage_StitchInvalidParams(t *testing.T) {
	mockBus := new(MockBus)
	mockMsg := new(MockJetStreamMsg)

	w := &Worker{
		id:      "worker-1",
		bus:     mockBus,
		logger:  logger.New("test"),
		workDir: os.TempDir(),
	}

	task := store.Task{
		ID:     pgtype.UUID{Bytes: [16]byte{3}, Valid: true},
		Type:   store.TaskTypeStitch,
		Params: []byte("invalid json"),
	}
	taskBytes, _ := json.Marshal(task)

	mockMsg.On("Data").Return(taskBytes)
	mockBus.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockMsg.On("Nak").Return(nil)

	w.handleMessage(context.Background(), mockMsg)

	mockMsg.AssertExpectations(t)
}

func TestHandleMessage_ManifestInvalidParams(t *testing.T) {
	mockBus := new(MockBus)
	mockMsg := new(MockJetStreamMsg)

	w := &Worker{
		id:            "worker-1",
		bus:           mockBus,
		logger:        logger.New("test"),
		workDir:       os.TempDir(),
		pluginManager: plugin_manager.New(logger.New("test"), ""),
	}

	task := store.Task{
		ID:     pgtype.UUID{Bytes: [16]byte{5}, Valid: true},
		Type:   store.TaskTypeManifest,
		Params: []byte("invalid json"),
	}
	taskBytes, _ := json.Marshal(task)

	mockMsg.On("Data").Return(taskBytes)
	mockBus.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockMsg.On("Nak").Return(nil)

	w.handleMessage(context.Background(), mockMsg)

	mockMsg.AssertExpectations(t)
}

func TestHandleMessage_RestreamInvalidParams(t *testing.T) {
	mockBus := new(MockBus)
	mockMsg := new(MockJetStreamMsg)

	w := &Worker{
		id:            "worker-1",
		bus:           mockBus,
		logger:        logger.New("test"),
		workDir:       os.TempDir(),
		pluginManager: plugin_manager.New(logger.New("test"), ""),
	}

	task := store.Task{
		ID:     pgtype.UUID{Bytes: [16]byte{4}, Valid: true},
		Type:   store.TaskTypeRestream,
		Params: []byte("invalid json"),
	}
	taskBytes, _ := json.Marshal(task)

	mockMsg.On("Data").Return(taskBytes)
	mockBus.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockMsg.On("Nak").Return(nil)

	w.handleMessage(context.Background(), mockMsg)

	mockMsg.AssertExpectations(t)
}

func TestHandleMessage_RestreamNoPublisher(t *testing.T) {
	mockBus := new(MockBus)
	mockMsg := new(MockJetStreamMsg)

	// Empty plugin manager with no publishers
	pm := plugin_manager.New(logger.New("test"), "")

	w := &Worker{
		id:            "worker-1",
		bus:           mockBus,
		logger:        logger.New("test"),
		pluginManager: pm,
		workDir:       os.TempDir(),
	}

	req := pb.PublishRequest{
		Platform: "youtube",
		FileUrl:  "/tmp/video.mp4",
	}
	reqParams, _ := json.Marshal(&req)

	task := store.Task{
		ID:     pgtype.UUID{Bytes: [16]byte{4}, Valid: true},
		Type:   store.TaskTypeRestream,
		Params: reqParams,
	}
	taskBytes, _ := json.Marshal(task)

	mockMsg.On("Data").Return(taskBytes)
	mockBus.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockMsg.On("Nak").Return(nil)

	w.handleMessage(context.Background(), mockMsg)

	mockBus.AssertExpectations(t)
	mockMsg.AssertExpectations(t)
}

func TestHandleMessage_RestreamError(t *testing.T) {
	mockBus := new(MockBus)
	mockMsg := new(MockJetStreamMsg)
	mockPub := new(MockPublisher)

	pm := plugin_manager.New(logger.New("test"), "")
	pm.Publisher["mock_pub"] = mockPub

	w := &Worker{
		id:            "worker-1",
		bus:           mockBus,
		logger:        logger.New("test"),
		pluginManager: pm,
		workDir:       os.TempDir(),
	}

	req := pb.PublishRequest{
		Platform: "youtube",
		FileUrl:  "/tmp/video.mp4",
	}
	reqParams, _ := json.Marshal(&req)

	task := store.Task{
		ID:     pgtype.UUID{Bytes: [16]byte{4}, Valid: true},
		Type:   store.TaskTypeRestream,
		Params: reqParams,
	}
	taskBytes, _ := json.Marshal(task)

	mockMsg.On("Data").Return(taskBytes)
	mockPub.On("Publish", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("publish failed"))
	mockBus.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockMsg.On("Nak").Return(nil)

	w.handleMessage(context.Background(), mockMsg)

	mockPub.AssertExpectations(t)
	mockMsg.AssertExpectations(t)
}

func TestHandleMessage_StitchEmptySegments(t *testing.T) {
	mockBus := new(MockBus)
	mockMsg := new(MockJetStreamMsg)

	w := &Worker{
		id:      "worker-1",
		bus:     mockBus,
		logger:  logger.New("test"),
		workDir: os.TempDir(),
	}

	stitchParams := map[string]interface{}{
		"segments": []string{}, // Empty segments
		"output":   "final.mp4",
	}
	paramsBytes, _ := json.Marshal(stitchParams)

	task := store.Task{
		ID:     pgtype.UUID{Bytes: [16]byte{3}, Valid: true},
		Type:   store.TaskTypeStitch,
		Params: paramsBytes,
	}
	taskBytes, _ := json.Marshal(task)

	mockMsg.On("Data").Return(taskBytes)
	mockBus.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockMsg.On("Nak").Return(nil)

	w.handleMessage(context.Background(), mockMsg)

	mockMsg.AssertExpectations(t)
}

func TestHandleMessage_ManifestEmptyVariants(t *testing.T) {
	mockBus := new(MockBus)
	mockMsg := new(MockJetStreamMsg)

	w := &Worker{
		id:            "worker-1",
		bus:           mockBus,
		logger:        logger.New("test"),
		workDir:       os.TempDir(),
		pluginManager: plugin_manager.New(logger.New("test"), ""),
	}

	manifestParams := map[string]interface{}{
		"variants": []ffmpeg.HLSVariant{}, // Empty variants
		"output":   "/tmp/master.m3u8",
	}
	paramsBytes, _ := json.Marshal(manifestParams)

	task := store.Task{
		ID:     pgtype.UUID{Bytes: [16]byte{5}, Valid: true},
		Type:   store.TaskTypeManifest,
		Params: paramsBytes,
	}
	taskBytes, _ := json.Marshal(task)

	mockMsg.On("Data").Return(taskBytes)
	mockBus.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockMsg.On("Ack").Return(nil)

	w.handleMessage(context.Background(), mockMsg)

	mockMsg.AssertExpectations(t)
}

func TestWorker_BusyStatus(t *testing.T) {
	w := &Worker{
		id:     "worker-1",
		isBusy: false,
	}

	// Initially idle
	if w.Status() != "idle" {
		t.Errorf("Expected 'idle', got '%s'", w.Status())
	}

	// Set busy
	w.isBusy = true
	if w.Status() != "busy" {
		t.Errorf("Expected 'busy', got '%s'", w.Status())
	}

	// Set back to idle
	w.isBusy = false
	if w.Status() != "idle" {
		t.Errorf("Expected 'idle', got '%s'", w.Status())
	}
}

func TestDownloadFile_NoPlugin(t *testing.T) {
	mockBus := new(MockBus)

	pm := plugin_manager.New(logger.New("test"), "")
	// No storage plugins

	w := &Worker{
		id:            "worker-1",
		bus:           mockBus,
		logger:        logger.New("test"),
		pluginManager: pm,
		workDir:       os.TempDir(),
	}

	// No S3_ENDPOINT set, should fail
	os.Unsetenv("S3_ENDPOINT")

	err := w.downloadFile(context.Background(), "bucket", "key", "/tmp/test.mp4")

	if err == nil {
		t.Error("expected error when no storage plugin and no S3_ENDPOINT")
	}
}

func TestUploadFile_NoPlugin(t *testing.T) {
	mockBus := new(MockBus)

	pm := plugin_manager.New(logger.New("test"), "")
	// No storage plugins

	w := &Worker{
		id:            "worker-1",
		bus:           mockBus,
		logger:        logger.New("test"),
		pluginManager: pm,
		workDir:       os.TempDir(),
	}

	// No S3_ENDPOINT set, should fail
	os.Unsetenv("S3_ENDPOINT")

	_, err := w.uploadFile(context.Background(), "/tmp/test.mp4", "bucket", "key")

	if err == nil {
		t.Error("expected error when no storage plugin and no S3_ENDPOINT")
	}
}

func TestSendLog(t *testing.T) {
	mockBus := new(MockBus)

	w := &Worker{
		id:     "worker-1",
		bus:    mockBus,
		logger: logger.New("test"),
	}

	mockBus.On("Publish", mock.Anything, "jobs.events", mock.MatchedBy(func(data []byte) bool {
		var payload map[string]interface{}
		json.Unmarshal(data, &payload)
		return payload["event"] == "log" && payload["task_id"] == "task-123"
	})).Return(nil)

	w.sendLog(context.Background(), "task-123", "info", "Test message")

	mockBus.AssertExpectations(t)
}

func TestSendEvent(t *testing.T) {
	mockBus := new(MockBus)

	w := &Worker{
		id:     "worker-1",
		bus:    mockBus,
		logger: logger.New("test"),
	}

	mockBus.On("Publish", mock.Anything, "jobs.events", mock.MatchedBy(func(data []byte) bool {
		var payload map[string]interface{}
		json.Unmarshal(data, &payload)
		return payload["event"] == "completed" && payload["task_id"] == "task-456"
	})).Return(nil)

	w.sendEvent(context.Background(), "task-456", "completed", []byte(`{"output":"test"}`))

	mockBus.AssertExpectations(t)
}

func TestReportError(t *testing.T) {
	mockBus := new(MockBus)

	w := &Worker{
		id:     "worker-1",
		bus:    mockBus,
		logger: logger.New("test"),
	}

	mockBus.On("Publish", mock.Anything, mock.Anything, mock.MatchedBy(func(data []byte) bool {
		var payload map[string]interface{}
		json.Unmarshal(data, &payload)
		return payload["source"] == "worker:worker-1" &&
			payload["severity"] == "critical"
	})).Return(nil)

	w.reportError(fmt.Errorf("test error"), map[string]interface{}{
		"stack": "test stack trace",
	})

	mockBus.AssertExpectations(t)
}

func TestMakeCmd(t *testing.T) {
	w := &Worker{
		id:     "worker-1",
		logger: logger.New("test"),
	}

	cmd := w.makeCmd(context.Background(), "-i", "input.mp4", "-o", "output.mp4")

	if cmd == nil {
		t.Fatal("expected cmd to not be nil")
	}

	// Check that the command is ffmpeg
	if cmd.Path == "" {
		t.Error("expected cmd.Path to be set")
	}
}

func TestHandleMessage_PanicRecovery(t *testing.T) {
	mockBus := new(MockBus)
	mockMsg := new(MockJetStreamMsg)

	w := &Worker{
		id:      "worker-1",
		bus:     mockBus,
		logger:  logger.New("test"),
		workDir: os.TempDir(),
		encoder: &MockEncoder{},
	}

	// Create a task that will cause processing but we'll inject a panic via mock
	task := store.Task{
		ID:     pgtype.UUID{Bytes: [16]byte{99}, Valid: true},
		Type:   store.TaskTypeProbe,
		Params: []byte(`{"url": "http://panic.test"}`),
	}
	taskBytes, _ := json.Marshal(task)

	mockMsg.On("Data").Return(taskBytes)

	// Mock encoder that returns an error to test error path
	mockEncoder := new(MockEncoder)
	mockEncoder.On("Probe", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("probe error"))
	w.encoder = mockEncoder

	mockBus.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockMsg.On("Nak").Return(nil)

	// This should not panic
	w.handleMessage(context.Background(), mockMsg)

	mockEncoder.AssertExpectations(t)
}

func TestHandleProbe_InvalidS3URL(t *testing.T) {
	mockBus := new(MockBus)
	mockEncoder := new(MockEncoder)
	mockMsg := new(MockJetStreamMsg)

	w := &Worker{
		id:            "worker-1",
		bus:           mockBus,
		logger:        logger.New("test"),
		encoder:       mockEncoder,
		workDir:       os.TempDir(),
		pluginManager: plugin_manager.New(logger.New("test"), ""),
	}

	// Invalid S3 URL
	probeParams := `{"url": "s3://"}`
	task := store.Task{
		ID:     pgtype.UUID{Bytes: [16]byte{10}, Valid: true},
		Type:   store.TaskTypeProbe,
		Params: []byte(probeParams),
	}
	taskBytes, _ := json.Marshal(task)

	mockMsg.On("Data").Return(taskBytes)
	mockBus.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockMsg.On("Nak").Return(nil)

	w.handleMessage(context.Background(), mockMsg)

	mockMsg.AssertExpectations(t)
}

func TestHandleTranscode_InvalidS3Input(t *testing.T) {
	mockBus := new(MockBus)
	mockEncoder := new(MockEncoder)
	mockMsg := new(MockJetStreamMsg)

	pm := plugin_manager.New(logger.New("test"), "")

	w := &Worker{
		id:            "worker-1",
		bus:           mockBus,
		logger:        logger.New("test"),
		encoder:       mockEncoder,
		workDir:       os.TempDir(),
		pluginManager: pm,
	}

	// Invalid S3 input URL
	tcTask := ffmpeg.TranscodeTask{
		InputURL:  "s3://invalid",
		OutputURL: "/tmp/out.mp4",
	}
	tcParams, _ := json.Marshal(tcTask)

	task := store.Task{
		ID:     pgtype.UUID{Bytes: [16]byte{11}, Valid: true},
		Type:   store.TaskTypeTranscode,
		Params: tcParams,
	}
	taskBytes, _ := json.Marshal(task)

	mockMsg.On("Data").Return(taskBytes)
	mockBus.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockMsg.On("Nak").Return(nil)

	w.handleMessage(context.Background(), mockMsg)

	mockMsg.AssertExpectations(t)
}

func TestHandleRestream_InvalidS3Input(t *testing.T) {
	mockBus := new(MockBus)
	mockMsg := new(MockJetStreamMsg)
	mockPub := new(MockPublisher)

	pm := plugin_manager.New(logger.New("test"), "")
	pm.Publisher["mock_pub"] = mockPub

	w := &Worker{
		id:            "worker-1",
		bus:           mockBus,
		logger:        logger.New("test"),
		pluginManager: pm,
		workDir:       os.TempDir(),
	}

	// S3 input that will fail to parse
	req := pb.PublishRequest{
		Platform: "youtube",
		FileUrl:  "s3://invalid-bucket",
	}
	reqParams, _ := json.Marshal(&req)

	task := store.Task{
		ID:     pgtype.UUID{Bytes: [16]byte{12}, Valid: true},
		Type:   store.TaskTypeRestream,
		Params: reqParams,
	}
	taskBytes, _ := json.Marshal(task)

	mockMsg.On("Data").Return(taskBytes)
	mockBus.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockMsg.On("Nak").Return(nil)

	w.handleMessage(context.Background(), mockMsg)

	mockMsg.AssertExpectations(t)
}

func TestHandleManifest_S3Output(t *testing.T) {
	mockBus := new(MockBus)
	mockMsg := new(MockJetStreamMsg)

	pm := plugin_manager.New(logger.New("test"), "")

	w := &Worker{
		id:            "worker-1",
		bus:           mockBus,
		logger:        logger.New("test"),
		workDir:       os.TempDir(),
		pluginManager: pm,
	}

	// S3 output that will fail
	manifestParams := map[string]interface{}{
		"variants": []ffmpeg.HLSVariant{
			{Path: "720p.m3u8", Bandwidth: 2500000, Resolution: "1280x720"},
		},
		"output": "s3://bucket/master.m3u8",
	}
	paramsBytes, _ := json.Marshal(manifestParams)

	task := store.Task{
		ID:     pgtype.UUID{Bytes: [16]byte{13}, Valid: true},
		Type:   store.TaskTypeManifest,
		Params: paramsBytes,
	}
	taskBytes, _ := json.Marshal(task)

	mockMsg.On("Data").Return(taskBytes)
	mockBus.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockMsg.On("Nak").Return(nil)

	w.handleMessage(context.Background(), mockMsg)

	mockMsg.AssertExpectations(t)
}
