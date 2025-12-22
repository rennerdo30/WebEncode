package ffmpeg

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFFmpegEncoder(t *testing.T) {
	t.Run("with custom paths", func(t *testing.T) {
		enc := NewFFmpegEncoder("/usr/bin/ffmpeg", "/usr/bin/ffprobe", nil)
		assert.NotNil(t, enc)
		assert.Equal(t, "/usr/bin/ffmpeg", enc.ffmpegPath)
		assert.Equal(t, "/usr/bin/ffprobe", enc.ffprobePath)
	})

	t.Run("with empty paths uses defaults", func(t *testing.T) {
		enc := NewFFmpegEncoder("", "", nil)
		assert.NotNil(t, enc)
		assert.Equal(t, "ffmpeg", enc.ffmpegPath)
		assert.Equal(t, "ffprobe", enc.ffprobePath)
	})

	t.Run("with partial paths", func(t *testing.T) {
		enc := NewFFmpegEncoder("/custom/ffmpeg", "", nil)
		assert.Equal(t, "/custom/ffmpeg", enc.ffmpegPath)
		assert.Equal(t, "ffprobe", enc.ffprobePath)
	})

	t.Run("with logger", func(t *testing.T) {
		l := logger.New("test")
		enc := NewFFmpegEncoder("", "", l)
		assert.NotNil(t, enc)
		assert.NotNil(t, enc.logger)
	})
}

func TestCalculateSegments_NoKeyframes(t *testing.T) {
	segments := CalculateSegments(nil, 120.0, 30.0)

	assert.NotEmpty(t, segments)
	assert.GreaterOrEqual(t, len(segments), 3)
	assert.LessOrEqual(t, len(segments), 5)

	assert.Equal(t, 0.0, segments[0].StartTime)

	lastSegment := segments[len(segments)-1]
	assert.Equal(t, 120.0, lastSegment.EndTime)
}

func TestCalculateSegments_WithKeyframes(t *testing.T) {
	keyframes := []float64{0, 2.5, 5.0, 10.0, 15.0, 20.0, 25.0, 30.0, 35.0, 40.0}
	segments := CalculateSegments(keyframes, 40.0, 10.0)

	assert.NotEmpty(t, segments)

	for i, seg := range segments {
		if i > 0 {
			found := false
			for _, kf := range keyframes {
				if seg.StartTime == kf {
					found = true
					break
				}
			}
			assert.True(t, found, "Segment %d start (%.2f) should be at keyframe", i, seg.StartTime)
		}
	}
}

func TestCalculateSegments_ShortVideo(t *testing.T) {
	segments := CalculateSegments(nil, 5.0, 30.0)

	assert.Equal(t, 1, len(segments))
	assert.Equal(t, 0.0, segments[0].StartTime)
	assert.Equal(t, 5.0, segments[0].EndTime)
}

func TestCalculateSegments_EdgeCases(t *testing.T) {
	t.Run("very long video", func(t *testing.T) {
		segments := CalculateSegments(nil, 3600.0, 60.0)
		assert.NotEmpty(t, segments)
		assert.Equal(t, 0.0, segments[0].StartTime)
		assert.Equal(t, 3600.0, segments[len(segments)-1].EndTime)
	})

	t.Run("tiny segment duration", func(t *testing.T) {
		segments := CalculateSegments(nil, 100.0, 1.0)
		assert.Equal(t, 100, len(segments))
	})

	t.Run("segment duration equals video duration", func(t *testing.T) {
		segments := CalculateSegments(nil, 30.0, 30.0)
		assert.Equal(t, 1, len(segments))
	})

	t.Run("with min segment duration constraint", func(t *testing.T) {
		keyframes := []float64{0, 1.0, 2.0, 3.0, 15.0, 30.0, 45.0, 60.0}
		segments := CalculateSegments(keyframes, 60.0, 10.0)
		// Segments less than minSegmentDuration (10s) should be merged
		assert.NotEmpty(t, segments)
	})

	t.Run("with max segment duration constraint", func(t *testing.T) {
		// Keyframes that would result in very long segments
		keyframes := []float64{0, 5.0, 130.0, 260.0}
		segments := CalculateSegments(keyframes, 260.0, 30.0)
		// maxSegmentDuration is 120s, so very long keyframe gaps should be handled
		assert.NotEmpty(t, segments)
	})
}

func TestCreateDefaultSegments(t *testing.T) {
	t.Run("standard case", func(t *testing.T) {
		segments := createDefaultSegments(100.0, 20.0)
		assert.Equal(t, 5, len(segments))

		for i, seg := range segments {
			assert.Equal(t, i, seg.Index)
			assert.InDelta(t, 20.0, seg.Duration, 0.1)
		}
	})

	t.Run("last segment shorter", func(t *testing.T) {
		segments := createDefaultSegments(90.0, 40.0)
		assert.Equal(t, 3, len(segments))
		assert.Equal(t, 10.0, segments[2].Duration) // Last segment is 10s
	})

	t.Run("single segment", func(t *testing.T) {
		segments := createDefaultSegments(10.0, 100.0)
		assert.Equal(t, 1, len(segments))
	})
}

func TestTranscodeTask_Fields(t *testing.T) {
	task := TranscodeTask{
		TaskID:      "test-123",
		InputURL:    "/input/video.mp4",
		OutputURL:   "/output/video.mp4",
		VideoCodec:  "libx264",
		AudioCodec:  "aac",
		Width:       1920,
		Height:      1080,
		BitrateKbps: 5000,
		Preset:      "medium",
		Container:   "mp4",
		StartTime:   10.0,
		Duration:    60.0,
	}

	assert.Equal(t, "test-123", task.TaskID)
	assert.Equal(t, "libx264", task.VideoCodec)
	assert.Equal(t, 1920, task.Width)
	assert.Equal(t, 5000, task.BitrateKbps)
	assert.Equal(t, 10.0, task.StartTime)
	assert.Equal(t, 60.0, task.Duration)
}

func TestProgress_Fields(t *testing.T) {
	progress := Progress{
		Percent:     50.0,
		Speed:       1.5,
		FPS:         30.0,
		CurrentTime: 60.0,
		Bitrate:     5000000,
	}

	assert.Equal(t, 50.0, progress.Percent)
	assert.Equal(t, 1.5, progress.Speed)
	assert.Equal(t, 30.0, progress.FPS)
	assert.Equal(t, 60.0, progress.CurrentTime)
	assert.Equal(t, int64(5000000), progress.Bitrate)
}

func TestProbeResult_Fields(t *testing.T) {
	result := ProbeResult{
		Duration: 120.5,
		Width:    1920,
		Height:   1080,
		Format:   "mov,mp4,m4a,3gp,3g2,mj2",
		Bitrate:  5000000,
		Streams: []StreamInfo{
			{Index: 0, CodecType: "video", CodecName: "h264"},
			{Index: 1, CodecType: "audio", CodecName: "aac"},
		},
		Keyframes: []float64{0, 2.0, 4.0},
	}

	assert.Equal(t, 120.5, result.Duration)
	assert.Equal(t, 1920, result.Width)
	assert.Equal(t, 2, len(result.Streams))
	assert.Equal(t, "video", result.Streams[0].CodecType)
	assert.Equal(t, 3, len(result.Keyframes))
}

func TestStreamInfo(t *testing.T) {
	streams := []StreamInfo{
		{Index: 0, CodecType: "video", CodecName: "h264"},
		{Index: 1, CodecType: "audio", CodecName: "aac"},
		{Index: 2, CodecType: "subtitle", CodecName: "srt"},
	}

	assert.Equal(t, "h264", streams[0].CodecName)
	assert.Equal(t, "audio", streams[1].CodecType)
	assert.Equal(t, "subtitle", streams[2].CodecType)
}

func TestSegment_Duration(t *testing.T) {
	segment := Segment{
		Index:     0,
		StartTime: 10.0,
		EndTime:   20.0,
		Duration:  10.0,
	}

	assert.Equal(t, 10.0, segment.Duration)
	assert.Equal(t, segment.EndTime-segment.StartTime, segment.Duration)
}

func TestBuildTranscodeArgs_BasicCase(t *testing.T) {
	enc := NewFFmpegEncoder("/usr/bin/ffmpeg", "/usr/bin/ffprobe", nil)

	task := &TranscodeTask{
		InputURL:    "/input/video.mp4",
		OutputURL:   "/output/video.mp4",
		VideoCodec:  "libx264",
		AudioCodec:  "aac",
		Width:       1280,
		Height:      720,
		BitrateKbps: 2500,
		Preset:      "medium",
		Container:   "mp4",
	}

	args := enc.buildTranscodeArgs(task)
	argsStr := strings.Join(args, " ")

	assert.Contains(t, argsStr, "-i /input/video.mp4")
	assert.Contains(t, argsStr, "-c:v libx264")
	assert.Contains(t, argsStr, "-c:a aac")
	assert.Contains(t, argsStr, "-b:v 2500k")
	assert.Contains(t, argsStr, "-preset medium")
	assert.Contains(t, argsStr, "/output/video.mp4")
	assert.Contains(t, argsStr, "-movflags +faststart")
}

func TestBuildTranscodeArgs_WithTimeRange(t *testing.T) {
	enc := NewFFmpegEncoder("/usr/bin/ffmpeg", "/usr/bin/ffprobe", nil)

	task := &TranscodeTask{
		InputURL:   "/input/video.mp4",
		OutputURL:  "/output/segment.mp4",
		StartTime:  10.0,
		Duration:   30.0,
		VideoCodec: "libx264",
	}

	args := enc.buildTranscodeArgs(task)
	argsStr := strings.Join(args, " ")

	assert.Contains(t, argsStr, "-ss 10")
	assert.Contains(t, argsStr, "-t 30")
}

func TestBuildTranscodeArgs_CopyCodec(t *testing.T) {
	enc := NewFFmpegEncoder("/usr/bin/ffmpeg", "/usr/bin/ffprobe", nil)

	task := &TranscodeTask{
		InputURL:   "/input/video.mp4",
		OutputURL:  "/output/copy.mp4",
		VideoCodec: "copy",
		AudioCodec: "copy",
	}

	args := enc.buildTranscodeArgs(task)
	argsStr := strings.Join(args, " ")

	assert.Contains(t, argsStr, "-c:v copy")
	assert.Contains(t, argsStr, "-c:a copy")
}

func TestBuildTranscodeArgs_DefaultCodecs(t *testing.T) {
	enc := NewFFmpegEncoder("", "", nil)

	task := &TranscodeTask{
		InputURL:  "/input/video.mp4",
		OutputURL: "/output/video.mp4",
	}

	args := enc.buildTranscodeArgs(task)
	argsStr := strings.Join(args, " ")

	// Should use default codecs
	assert.Contains(t, argsStr, "-c:v libx264")
	assert.Contains(t, argsStr, "-c:a aac")
	assert.Contains(t, argsStr, "-preset fast")
}

func TestBuildTranscodeArgs_H265WithPreset(t *testing.T) {
	enc := NewFFmpegEncoder("", "", nil)

	task := &TranscodeTask{
		InputURL:   "/input/video.mp4",
		OutputURL:  "/output/video.mp4",
		VideoCodec: "libx265",
		Preset:     "slow",
	}

	args := enc.buildTranscodeArgs(task)
	argsStr := strings.Join(args, " ")

	assert.Contains(t, argsStr, "-c:v libx265")
	assert.Contains(t, argsStr, "-preset slow")
}

func TestBuildTranscodeArgs_WebM(t *testing.T) {
	enc := NewFFmpegEncoder("", "", nil)

	task := &TranscodeTask{
		InputURL:   "/input/video.mp4",
		OutputURL:  "/output/video.webm",
		VideoCodec: "libvpx-vp9",
		AudioCodec: "libopus",
		Container:  "webm",
	}

	args := enc.buildTranscodeArgs(task)
	argsStr := strings.Join(args, " ")

	assert.Contains(t, argsStr, "-c:v libvpx-vp9")
	assert.Contains(t, argsStr, "-c:a libopus")
	assert.Contains(t, argsStr, "-f webm")
	// WebM should NOT have movflags
	assert.NotContains(t, argsStr, "-movflags")
}

func TestBuildTranscodeArgs_NoBitrate(t *testing.T) {
	enc := NewFFmpegEncoder("", "", nil)

	task := &TranscodeTask{
		InputURL:  "/input/video.mp4",
		OutputURL: "/output/video.mp4",
	}

	args := enc.buildTranscodeArgs(task)
	argsStr := strings.Join(args, " ")

	// Should NOT contain bitrate settings
	assert.NotContains(t, argsStr, "-b:v")
	assert.NotContains(t, argsStr, "-maxrate")
	assert.NotContains(t, argsStr, "-bufsize")
}

func TestBuildTranscodeArgs_NoResolution(t *testing.T) {
	enc := NewFFmpegEncoder("", "", nil)

	task := &TranscodeTask{
		InputURL:  "/input/video.mp4",
		OutputURL: "/output/video.mp4",
	}

	args := enc.buildTranscodeArgs(task)
	argsStr := strings.Join(args, " ")

	// Should NOT contain scale filter
	assert.NotContains(t, argsStr, "-vf")
}

func TestBuildTranscodeArgs_BitrateSettings(t *testing.T) {
	enc := NewFFmpegEncoder("", "", nil)

	task := &TranscodeTask{
		InputURL:    "/input/video.mp4",
		OutputURL:   "/output/video.mp4",
		BitrateKbps: 5000,
	}

	args := enc.buildTranscodeArgs(task)
	argsStr := strings.Join(args, " ")

	assert.Contains(t, argsStr, "-b:v 5000k")
	assert.Contains(t, argsStr, "-maxrate 5500k") // 110% of bitrate
	assert.Contains(t, argsStr, "-bufsize 10000k") // 200% of bitrate
}

func TestParseProgress(t *testing.T) {
	enc := NewFFmpegEncoder("", "", nil)

	progressOutput := `frame=100
fps=30.00
stream_0_0_q=-1.0
bitrate=2500.0kbits/s
total_size=5000000
out_time_us=60000000
out_time_ms=60000000
out_time=00:01:00.000000
dup_frames=0
drop_frames=0
speed=2.0x
progress=continue
frame=200
fps=30.00
speed=1.5x
out_time_ms=120000000
progress=end`

	reader := bytes.NewBufferString(progressOutput)
	progressCh := make(chan Progress, 10)

	go enc.parseProgress(reader, 120.0, progressCh)

	var lastProgress Progress
	for p := range progressCh {
		lastProgress = p
	}

	// Check final progress values
	assert.True(t, lastProgress.Percent > 0)
}

func TestParseProgress_EdgeCases(t *testing.T) {
	enc := NewFFmpegEncoder("", "", nil)

	t.Run("empty input", func(t *testing.T) {
		reader := bytes.NewBufferString("")
		progressCh := make(chan Progress, 10)

		go enc.parseProgress(reader, 60.0, progressCh)

		count := 0
		for range progressCh {
			count++
		}
		assert.Equal(t, 0, count)
	})

	t.Run("malformed input", func(t *testing.T) {
		reader := bytes.NewBufferString("not=valid\nformat=without\nproper=data")
		progressCh := make(chan Progress, 10)

		go enc.parseProgress(reader, 60.0, progressCh)

		for range progressCh {
			// Should handle gracefully without crashing
		}
	})

	t.Run("percent cap at 100", func(t *testing.T) {
		// out_time exceeds duration
		progressOutput := `out_time_ms=120000000
progress=continue`
		reader := bytes.NewBufferString(progressOutput)
		progressCh := make(chan Progress, 10)

		go enc.parseProgress(reader, 60.0, progressCh) // duration is 60s but out_time is 120s

		for p := range progressCh {
			assert.LessOrEqual(t, p.Percent, 100.0)
		}
	})
}

func TestProbe_MissingFile(t *testing.T) {
	enc := NewFFmpegEncoder("ffprobe", "ffprobe", logger.New("test"))
	ctx := context.Background()

	result, err := enc.Probe(ctx, "/nonexistent/file.mp4")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestTranscode_MissingFile(t *testing.T) {
	enc := NewFFmpegEncoder("ffmpeg", "ffprobe", logger.New("test"))
	ctx := context.Background()

	task := &TranscodeTask{
		InputURL:  "/nonexistent/file.mp4",
		OutputURL: "/tmp/output.mp4",
	}

	progressCh := make(chan Progress, 10)
	err := enc.Transcode(ctx, task, progressCh)

	assert.Error(t, err)
}

func TestTranscode_ContextCancellation(t *testing.T) {
	enc := NewFFmpegEncoder("ffmpeg", "ffprobe", logger.New("test"))
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	task := &TranscodeTask{
		InputURL:  "/input/video.mp4",
		OutputURL: "/tmp/output.mp4",
	}

	progressCh := make(chan Progress, 10)
	err := enc.Transcode(ctx, task, progressCh)

	assert.Error(t, err)
}

func TestFFmpegEncoder_StructFields(t *testing.T) {
	l := logger.New("test")
	enc := &FFmpegEncoder{
		ffmpegPath:  "/custom/ffmpeg",
		ffprobePath: "/custom/ffprobe",
		logger:      l,
	}

	assert.Equal(t, "/custom/ffmpeg", enc.ffmpegPath)
	assert.Equal(t, "/custom/ffprobe", enc.ffprobePath)
	assert.NotNil(t, enc.logger)
}

func TestBuildTranscodeArgs_AllOptions(t *testing.T) {
	enc := NewFFmpegEncoder("", "", nil)

	task := &TranscodeTask{
		TaskID:      "test-task",
		InputURL:    "/input/video.mp4",
		OutputURL:   "/output/video.mp4",
		StartTime:   10.0,
		Duration:    60.0,
		VideoCodec:  "libx264",
		AudioCodec:  "aac",
		Width:       1920,
		Height:      1080,
		BitrateKbps: 8000,
		Preset:      "slow",
		Container:   "mp4",
	}

	args := enc.buildTranscodeArgs(task)
	argsStr := strings.Join(args, " ")

	// All options should be present
	assert.Contains(t, argsStr, "-ss 10")
	assert.Contains(t, argsStr, "-i /input/video.mp4")
	assert.Contains(t, argsStr, "-t 60")
	assert.Contains(t, argsStr, "-c:v libx264")
	assert.Contains(t, argsStr, "-b:v 8000k")
	assert.Contains(t, argsStr, "-vf")
	assert.Contains(t, argsStr, "scale=1920:1080")
	assert.Contains(t, argsStr, "-preset slow")
	assert.Contains(t, argsStr, "-c:a aac")
	assert.Contains(t, argsStr, "-pix_fmt yuv420p")
	assert.Contains(t, argsStr, "-movflags +faststart")
	assert.Contains(t, argsStr, "/output/video.mp4")
}

func TestGetKeyframes_MissingFile(t *testing.T) {
	enc := NewFFmpegEncoder("ffprobe", "ffprobe", logger.New("test"))
	ctx := context.Background()

	keyframes, err := enc.getKeyframes(ctx, "/nonexistent/file.mp4")

	assert.Error(t, err)
	assert.Nil(t, keyframes)
}

func TestCalculateSegments_RemainingDuration(t *testing.T) {
	// Test case where there's remaining duration after last keyframe
	keyframes := []float64{0, 30.0, 60.0}
	segments := CalculateSegments(keyframes, 100.0, 30.0)

	// Should include remaining duration
	lastSeg := segments[len(segments)-1]
	assert.Equal(t, 100.0, lastSeg.EndTime)
}

func BenchmarkCalculateSegments(b *testing.B) {
	keyframes := make([]float64, 1000)
	for i := range keyframes {
		keyframes[i] = float64(i) * 2.0
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateSegments(keyframes, 2000.0, 30.0)
	}
}

func BenchmarkBuildTranscodeArgs(b *testing.B) {
	enc := NewFFmpegEncoder("", "", nil)
	task := &TranscodeTask{
		InputURL:    "/input/video.mp4",
		OutputURL:   "/output/video.mp4",
		VideoCodec:  "libx264",
		AudioCodec:  "aac",
		Width:       1920,
		Height:      1080,
		BitrateKbps: 5000,
		Preset:      "fast",
		Container:   "mp4",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		enc.buildTranscodeArgs(task)
	}
}

func TestProgressChannel_NonBlocking(t *testing.T) {
	enc := NewFFmpegEncoder("", "", nil)

	// Create a full channel
	progressCh := make(chan Progress, 1)
	progressCh <- Progress{Percent: 50}

	// parseProgress should not block when channel is full
	progressOutput := `out_time_ms=60000000
progress=continue
out_time_ms=120000000
progress=continue`

	reader := bytes.NewBufferString(progressOutput)

	// This should complete without deadlock
	done := make(chan bool)
	go func() {
		enc.parseProgress(reader, 120.0, progressCh)
		done <- true
	}()

	// Wait for completion
	<-done
}

func TestParseProgress_SpeedParsing(t *testing.T) {
	enc := NewFFmpegEncoder("", "", nil)

	progressOutput := `speed=2.5x
progress=continue`

	reader := bytes.NewBufferString(progressOutput)
	progressCh := make(chan Progress, 10)

	go enc.parseProgress(reader, 60.0, progressCh)

	var lastProgress Progress
	for p := range progressCh {
		lastProgress = p
	}

	assert.Equal(t, 2.5, lastProgress.Speed)
}

func TestParseProgress_BitrateParsing(t *testing.T) {
	enc := NewFFmpegEncoder("", "", nil)

	progressOutput := `bitrate=5000.0kbits/s
progress=continue`

	reader := bytes.NewBufferString(progressOutput)
	progressCh := make(chan Progress, 10)

	go enc.parseProgress(reader, 60.0, progressCh)

	var lastProgress Progress
	for p := range progressCh {
		lastProgress = p
	}

	assert.Equal(t, int64(5000000), lastProgress.Bitrate)
}

func TestParseProgress_FPSParsing(t *testing.T) {
	enc := NewFFmpegEncoder("", "", nil)

	progressOutput := `fps=29.97
progress=continue`

	reader := bytes.NewBufferString(progressOutput)
	progressCh := make(chan Progress, 10)

	go enc.parseProgress(reader, 60.0, progressCh)

	var lastProgress Progress
	for p := range progressCh {
		lastProgress = p
	}

	require.InDelta(t, 29.97, lastProgress.FPS, 0.01)
}
