package encoder

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/rennerdo30/webencode/pkg/logger"
)

// TestHelperProcess isn't a real test. It's used to mock exec.Command
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(2)
	}

	cmd, args := args[0], args[1:]
	switch {
	case strings.Contains(cmd, "ffprobe"):
		handleFFprobe(args)
	case strings.Contains(cmd, "ffmpeg"):
		handleFFmpeg(args)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %q\n", cmd)
		os.Exit(2)
	}
}

func handleFFprobe(args []string) {
	// Check if we are probing keyframes or metadata
	isKeyframes := false
	for _, arg := range args {
		if strings.Contains(arg, "key_frame") {
			isKeyframes = true
			break
		}
	}

	if isKeyframes {
		// Mock keyframes output: pts,key_frame
		fmt.Println("0.000000,1")
		fmt.Println("10.000000,1")
		fmt.Println("20.000000,1")
	} else {
		// Mock metadata output
		probe := struct {
			Format struct {
				Duration   string `json:"duration"`
				FormatName string `json:"format_name"`
				BitRate    string `json:"bit_rate"`
			} `json:"format"`
			Streams []struct {
				Index     int    `json:"index"`
				CodecType string `json:"codec_type"`
				CodecName string `json:"codec_name"`
				Width     int    `json:"width"`
				Height    int    `json:"height"`
			} `json:"streams"`
		}{}
		probe.Format.Duration = "30.000000"
		probe.Format.FormatName = "mov,mp4"
		probe.Format.BitRate = "1000000"
		probe.Streams = append(probe.Streams, struct {
			Index     int    `json:"index"`
			CodecType string `json:"codec_type"`
			CodecName string `json:"codec_name"`
			Width     int    `json:"width"`
			Height    int    `json:"height"`
		}{
			Index:     0,
			CodecType: "video",
			CodecName: "h264",
			Width:     1920,
			Height:    1080,
		})
		json.NewEncoder(os.Stdout).Encode(probe)
	}
}

func handleFFmpeg(args []string) {
	// Simulate progress
	// format: key=value
	steps := []string{
		"out_time_ms=0\nspeed=0x\nfps=0\nbitrate=0kbits/s\nprogress=continue",
		"out_time_ms=10000000\nspeed=1.5x\nfps=30\nbitrate=1000kbits/s\nprogress=continue",
		"out_time_ms=20000000\nspeed=1.5x\nfps=30\nbitrate=1000kbits/s\nprogress=end",
	}

	for _, step := range steps {
		fmt.Fprintln(os.Stderr, step)
		time.Sleep(10 * time.Millisecond)
	}
}

func fakeExecCommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", name}
	cs = append(cs, args...)
	cmd := exec.CommandContext(ctx, os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestCalculateSegments(t *testing.T) {
	tests := []struct {
		name            string
		keyframes       []float64
		totalDuration   float64
		targetSegment   float64
		expectedCount   int
		expectedLengths []float64 // Approximate durations
	}{
		{
			name:            "Empty keyframes",
			keyframes:       []float64{},
			totalDuration:   30.0,
			targetSegment:   10.0,
			expectedCount:   3,
			expectedLengths: []float64{10.0, 10.0, 10.0},
		},
		{
			name:            "Perfect alignment",
			keyframes:       []float64{0.0, 10.0, 20.0},
			totalDuration:   30.0,
			targetSegment:   10.0,
			expectedCount:   3,
			expectedLengths: []float64{10.0, 10.0, 10.0},
		},
		{
			name:            "Weird alignment",
			keyframes:       []float64{0.0, 5.0, 10.0, 12.0, 22.0},
			totalDuration:   30.0,
			targetSegment:   10.0,
			expectedCount:   3,
			expectedLengths: []float64{10.0, 12.0, 8.0}, // 0-10, 10-22 (merged 10-12-22), 22-30
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			segments := CalculateSegments(tt.keyframes, tt.totalDuration, tt.targetSegment)
			if len(segments) != tt.expectedCount {
				t.Errorf("expected %d segments, got %d", tt.expectedCount, len(segments))
			}
			for i, seg := range segments {
				if i < len(tt.expectedLengths) {
					if diff := seg.Duration - tt.expectedLengths[i]; diff > 0.001 || diff < -0.001 {
						t.Errorf("segment %d duration mismatch: got %f, want %f", i, seg.Duration, tt.expectedLengths[i])
					}
				}
			}
		})
	}
}

func TestProbe(t *testing.T) {
	// Swap execCommandContext
	oldExec := execCommandContext
	execCommandContext = fakeExecCommandContext
	defer func() { execCommandContext = oldExec }()

	l := logger.New("test")
	enc := NewFFmpegEncoder("ffmpeg", "ffprobe", l)

	ctx := context.Background()
	result, err := enc.Probe(ctx, "test.mp4")
	if err != nil {
		t.Fatalf("Probe failed: %v", err)
	}

	if result.Duration != 30.0 {
		t.Errorf("expected duration 30.0, got %f", result.Duration)
	}
	if result.Width != 1920 {
		t.Errorf("expected width 1920, got %d", result.Width)
	}
	if len(result.Keyframes) != 3 {
		t.Errorf("expected 3 keyframes, got %d", len(result.Keyframes))
	}
}

func TestTranscode(t *testing.T) {
	// Swap execCommandContext
	oldExec := execCommandContext
	execCommandContext = fakeExecCommandContext
	defer func() { execCommandContext = oldExec }()

	l := logger.New("test")
	enc := NewFFmpegEncoder("ffmpeg", "ffprobe", l)

	task := &TranscodeTask{
		InputURL:    "in.mp4",
		OutputURL:   "out.mp4",
		Duration:    30.0,
		VideoCodec:  "libx264",
		AudioCodec:  "aac",
		Width:       1280,
		Height:      720,
		BitrateKbps: 1000,
	}

	progressCh := make(chan Progress, 10)
	ctx := context.Background()

	err := enc.Transcode(ctx, task, progressCh)
	if err != nil {
		t.Fatalf("Transcode failed: %v", err)
	}

	// Drain channel
	var count int
	for p := range progressCh {
		count++
		if p.Percent < 0 || p.Percent > 100 {
			t.Errorf("invalid percentage: %f", p.Percent)
		}
	}
	if count == 0 {
		t.Error("expected some progress updates")
	}
}

func TestBuildTranscodeArgs(t *testing.T) {
	l := logger.New("test")
	enc := NewFFmpegEncoder("ffmpeg", "ffprobe", l)

	task := &TranscodeTask{
		InputURL:    "in.mp4",
		OutputURL:   "out.mp4",
		Duration:    60.0,
		StartTime:   10.0,
		VideoCodec:  "libx264",
		AudioCodec:  "aac",
		Width:       1920,
		Height:      1080,
		BitrateKbps: 5000,
		Preset:      "slow",
		Container:   "mp4",
	}

	args := enc.buildTranscodeArgs(task)
	joined := strings.Join(args, " ")

	expectedSubstrings := []string{
		"-ss 10.000",
		"-i in.mp4",
		"-t 60.000",
		"-c:v libx264",
		"-b:v 5000k",
		"-maxrate 5500k", // 5000 * 1.1
		"-vf scale=1920:1080",
		"-preset slow",
		"-c:a aac",
		"out.mp4",
	}

	for _, sub := range expectedSubstrings {
		if !strings.Contains(joined, sub) {
			t.Errorf("expected args to contain %q, got: %s", sub, joined)
		}
	}
}
