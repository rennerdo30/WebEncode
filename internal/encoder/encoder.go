package encoder

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"

	"github.com/rennerdo30/webencode/pkg/logger"
)

// execCommandContext allows mocking exec.CommandContext in tests
var execCommandContext = exec.CommandContext

// FFmpegEncoder wraps FFmpeg command execution
type FFmpegEncoder struct {
	ffmpegPath  string
	ffprobePath string
	logger      *logger.Logger
}

// NewFFmpegEncoder creates a new FFmpeg encoder
func NewFFmpegEncoder(ffmpegPath, ffprobePath string, l *logger.Logger) *FFmpegEncoder {
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg"
	}
	if ffprobePath == "" {
		ffprobePath = "ffprobe"
	}
	return &FFmpegEncoder{
		ffmpegPath:  ffmpegPath,
		ffprobePath: ffprobePath,
		logger:      l,
	}
}

// ProbeResult contains media file metadata
type ProbeResult struct {
	Duration  float64
	Width     int
	Height    int
	Format    string
	Bitrate   int64
	Streams   []StreamInfo
	Keyframes []float64
}

// StreamInfo contains stream metadata
type StreamInfo struct {
	Index     int
	CodecType string // video, audio, subtitle
	CodecName string
}

// TranscodeTask represents a transcoding task
type TranscodeTask struct {
	TaskID      string
	InputURL    string
	OutputURL   string
	StartTime   float64
	Duration    float64
	VideoCodec  string
	AudioCodec  string
	Width       int
	Height      int
	BitrateKbps int
	Preset      string
	Container   string
}

// Progress represents encoding progress
type Progress struct {
	Percent     float64
	Speed       float64
	FPS         float64
	CurrentTime float64
	Bitrate     int64
}

// Probe analyzes a media file and returns its metadata
func (e *FFmpegEncoder) Probe(ctx context.Context, url string) (*ProbeResult, error) {
	// Get basic metadata
	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		url,
	}

	cmd := execCommandContext(ctx, e.ffprobePath, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w", err)
	}

	result, err := parseProbeOutput(output)
	if err != nil {
		return nil, err
	}

	// Get keyframes for segmentation
	keyframes, err := e.getKeyframes(ctx, url)
	if err != nil {
		e.logger.Warn("Failed to get keyframes, using default segmentation", "error", err)
	} else {
		result.Keyframes = keyframes
	}

	return result, nil
}

// getKeyframes extracts keyframe timestamps for segment alignment
func (e *FFmpegEncoder) getKeyframes(ctx context.Context, url string) ([]float64, error) {
	args := []string{
		"-v", "quiet",
		"-select_streams", "v:0",
		"-show_entries", "frame=pts_time,key_frame",
		"-of", "csv=p=0",
		url,
	}

	cmd := execCommandContext(ctx, e.ffprobePath, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var keyframes []float64
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		parts := strings.Split(strings.TrimSpace(line), ",")
		if len(parts) == 2 && parts[1] == "1" {
			if pts, err := strconv.ParseFloat(parts[0], 64); err == nil {
				keyframes = append(keyframes, pts)
			}
		}
	}

	return keyframes, nil
}

// Segment represents a video segment for parallel processing
type Segment struct {
	Index     int
	StartTime float64
	EndTime   float64
	Duration  float64
}

// CalculateSegments creates keyframe-aligned segments for parallel transcoding
func CalculateSegments(keyframes []float64, totalDuration float64, targetSegmentDuration float64) []Segment {
	if len(keyframes) == 0 {
		// Fallback: create segments without keyframe alignment
		return createDefaultSegments(totalDuration, targetSegmentDuration)
	}

	const (
		minSegmentDuration = 10.0  // seconds
		maxSegmentDuration = 120.0 // seconds
	)

	segments := []Segment{}
	currentStart := 0.0
	segmentIndex := 0

	for i, kf := range keyframes {
		timeSinceStart := kf - currentStart

		// If we've accumulated enough time, or this is the last keyframe
		if timeSinceStart >= targetSegmentDuration || i == len(keyframes)-1 {
			// Ensure minimum segment duration (unless it's the last)
			if timeSinceStart < minSegmentDuration && i != len(keyframes)-1 {
				continue
			}

			// Cap at maximum segment duration
			endTime := kf
			if timeSinceStart > maxSegmentDuration {
				// Find a keyframe closer to the target
				for j := i; j >= 0; j-- {
					if keyframes[j]-currentStart >= targetSegmentDuration &&
						keyframes[j]-currentStart <= maxSegmentDuration {
						endTime = keyframes[j]
						break
					}
				}
			}

			segments = append(segments, Segment{
				Index:     segmentIndex,
				StartTime: currentStart,
				EndTime:   endTime,
				Duration:  endTime - currentStart,
			})
			currentStart = endTime
			segmentIndex++
		}
	}

	// Handle remaining duration
	if currentStart < totalDuration-1 {
		segments = append(segments, Segment{
			Index:     segmentIndex,
			StartTime: currentStart,
			EndTime:   totalDuration,
			Duration:  totalDuration - currentStart,
		})
	}

	return segments
}

func createDefaultSegments(totalDuration, targetDuration float64) []Segment {
	segments := []Segment{}
	for start := 0.0; start < totalDuration; start += targetDuration {
		end := start + targetDuration
		if end > totalDuration {
			end = totalDuration
		}
		segments = append(segments, Segment{
			Index:     len(segments),
			StartTime: start,
			EndTime:   end,
			Duration:  end - start,
		})
	}
	return segments
}

// Transcode executes transcoding with progress reporting
func (e *FFmpegEncoder) Transcode(ctx context.Context, task *TranscodeTask, progressCh chan<- Progress) error {
	args := e.buildTranscodeArgs(task)

	cmd := execCommandContext(ctx, e.ffmpegPath, args...)

	// Get stderr for progress
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	// Parse progress in goroutine
	go e.parseProgress(stderr, task.Duration, progressCh)

	// Wait for completion
	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("ffmpeg failed: %w", err)
	}

	return nil
}

// buildTranscodeArgs constructs FFmpeg command arguments
func (e *FFmpegEncoder) buildTranscodeArgs(task *TranscodeTask) []string {
	args := []string{
		"-hide_banner",
		"-loglevel", "warning",
		"-progress", "pipe:2",
		"-y",
	}

	// Input seeking (before input for faster seeking)
	if task.StartTime > 0 {
		args = append(args, "-ss", fmt.Sprintf("%.3f", task.StartTime))
	}

	// Input
	args = append(args, "-i", task.InputURL)

	// Duration
	if task.Duration > 0 {
		args = append(args, "-t", fmt.Sprintf("%.3f", task.Duration))
	}

	// Video codec
	videoCodec := task.VideoCodec
	if videoCodec == "" {
		videoCodec = "libx264"
	}
	args = append(args, "-c:v", videoCodec)

	// Video settings
	if task.BitrateKbps > 0 {
		bitrate := fmt.Sprintf("%dk", task.BitrateKbps)
		maxrate := fmt.Sprintf("%dk", int(float64(task.BitrateKbps)*1.1))
		bufsize := fmt.Sprintf("%dk", task.BitrateKbps*2)
		args = append(args, "-b:v", bitrate, "-maxrate", maxrate, "-bufsize", bufsize)
	}

	// Resolution
	if task.Width > 0 && task.Height > 0 {
		vf := fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2",
			task.Width, task.Height, task.Width, task.Height)
		args = append(args, "-vf", vf)
	}

	// Preset
	preset := task.Preset
	if preset == "" {
		preset = "fast"
	}
	// Only add preset for x264/x265
	if strings.Contains(videoCodec, "264") || strings.Contains(videoCodec, "265") {
		args = append(args, "-preset", preset)
	}

	// Audio codec
	audioCodec := task.AudioCodec
	if audioCodec == "" {
		audioCodec = "aac"
	}
	args = append(args, "-c:a", audioCodec, "-b:a", "192k", "-ar", "48000", "-ac", "2")

	// Pixel format for compatibility
	args = append(args, "-pix_fmt", "yuv420p")

	// Container-specific flags
	container := task.Container
	if container == "" || container == "mp4" {
		args = append(args, "-movflags", "+faststart")
	}

	// Output format
	if container != "" {
		args = append(args, "-f", container)
	}

	// Output
	args = append(args, task.OutputURL)

	return args
}

// parseProgress reads FFmpeg progress output
func (e *FFmpegEncoder) parseProgress(r io.Reader, totalDuration float64, progressCh chan<- Progress) {
	defer close(progressCh)

	scanner := bufio.NewScanner(r)
	var current Progress

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "out_time_ms":
			if ms, err := strconv.ParseInt(value, 10, 64); err == nil {
				current.CurrentTime = float64(ms) / 1_000_000.0
				if totalDuration > 0 {
					current.Percent = (current.CurrentTime / totalDuration) * 100.0
					if current.Percent > 100 {
						current.Percent = 100
					}
				}
			}
		case "speed":
			value = strings.TrimSuffix(value, "x")
			if speed, err := strconv.ParseFloat(value, 64); err == nil {
				current.Speed = speed
			}
		case "fps":
			if fps, err := strconv.ParseFloat(value, 64); err == nil {
				current.FPS = fps
			}
		case "bitrate":
			value = strings.TrimSuffix(value, "kbits/s")
			if bitrate, err := strconv.ParseFloat(value, 64); err == nil {
				current.Bitrate = int64(bitrate * 1000)
			}
		case "progress":
			// Send progress update on each "progress" line
			select {
			case progressCh <- current:
			default:
				// Don't block if channel is full
			}
		}
	}
}
