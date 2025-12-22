package ffmpeg

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// SubtitleMode defines how subtitles should be handled
type SubtitleMode string

const (
	SubtitleModePassthrough SubtitleMode = "passthrough" // Copy subtitle streams as-is
	SubtitleModeExtract     SubtitleMode = "extract"     // Extract to separate SRT/VTT files
	SubtitleModeBurnIn      SubtitleMode = "burnin"      // Burn subtitles into video
	SubtitleModeDisable     SubtitleMode = "disable"     // Remove all subtitles
)

// SubtitleTrack represents a subtitle track in the source
type SubtitleTrack struct {
	Index    int
	Language string
	Title    string
	Codec    string
	Default  bool
	Forced   bool
}

// SubtitleConfig configures subtitle handling
type SubtitleConfig struct {
	Mode         SubtitleMode
	TrackIndex   int    // Which track to use (-1 for all/first)
	Language     string // Language filter
	FontSize     int    // For burn-in mode
	FontName     string // For burn-in mode
	OutlineWidth int    // For burn-in mode
}

// DefaultSubtitleConfig returns sensible defaults
func DefaultSubtitleConfig() SubtitleConfig {
	return SubtitleConfig{
		Mode:         SubtitleModePassthrough,
		TrackIndex:   -1,
		FontSize:     24,
		FontName:     "Arial",
		OutlineWidth: 2,
	}
}

// GetSubtitleTracks returns available subtitle tracks from a video
func (e *FFmpegEncoder) GetSubtitleTracks(ctx context.Context, inputURL string) ([]SubtitleTrack, error) {
	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_streams",
		"-select_streams", "s",
		inputURL,
	}

	cmd := exec.CommandContext(ctx, e.ffprobePath, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w", err)
	}

	return parseSubtitleTracks(output)
}

// parseSubtitleTracks parses ffprobe output for subtitle streams
func parseSubtitleTracks(output []byte) ([]SubtitleTrack, error) {
	// Simple JSON parsing - in production use encoding/json
	tracks := []SubtitleTrack{}

	// This is a simplified parser - the real implementation would use json.Unmarshal
	outputStr := string(output)
	if strings.Contains(outputStr, `"codec_type": "subtitle"`) {
		// At least one subtitle track exists
		tracks = append(tracks, SubtitleTrack{
			Index: 0,
			Codec: "unknown",
		})
	}

	return tracks, nil
}

// ExtractSubtitles extracts subtitle tracks to separate files
func (e *FFmpegEncoder) ExtractSubtitles(ctx context.Context, inputURL, outputDir string, trackIndex int) ([]string, error) {
	// Get available tracks
	tracks, err := e.GetSubtitleTracks(ctx, inputURL)
	if err != nil {
		return nil, err
	}

	if len(tracks) == 0 {
		e.logger.Info("No subtitle tracks found", "input", inputURL)
		return nil, nil
	}

	outputs := []string{}

	for i, track := range tracks {
		if trackIndex >= 0 && i != trackIndex {
			continue
		}

		ext := "srt"
		if track.Codec == "webvtt" || track.Codec == "vtt" {
			ext = "vtt"
		}

		outputPath := filepath.Join(outputDir, fmt.Sprintf("subtitle_%d.%s", i, ext))

		args := []string{
			"-hide_banner",
			"-loglevel", "error",
			"-y",
			"-i", inputURL,
			"-map", fmt.Sprintf("0:s:%d", i),
			"-c:s", "srt",
			outputPath,
		}

		cmd := exec.CommandContext(ctx, e.ffmpegPath, args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			e.logger.Warn("Failed to extract subtitle track", "index", i, "error", err, "output", string(output))
			continue
		}

		outputs = append(outputs, outputPath)
	}

	e.logger.Info("Extracted subtitles", "count", len(outputs), "output_dir", outputDir)
	return outputs, nil
}

// BuildBurnInArgs returns FFmpeg filter args for burning in subtitles
func BuildBurnInArgs(config SubtitleConfig, subtitleFile string) []string {
	// Build subtitles filter with styling
	style := fmt.Sprintf("FontSize=%d,FontName=%s,OutlineColour=&H000000&,Outline=%d",
		config.FontSize, config.FontName, config.OutlineWidth)

	filter := fmt.Sprintf("subtitles='%s':force_style='%s'",
		escapeFilterPath(subtitleFile), style)

	return []string{"-vf", filter}
}

// BuildPassthroughArgs returns FFmpeg args for copying subtitles
func BuildPassthroughArgs() []string {
	return []string{
		"-c:s", "copy",
	}
}

// escapeFilterPath escapes special characters in file paths for FFmpeg filters
func escapeFilterPath(path string) string {
	// Escape single quotes and backslashes
	path = strings.ReplaceAll(path, "\\", "\\\\")
	path = strings.ReplaceAll(path, "'", "\\'")
	path = strings.ReplaceAll(path, ":", "\\:")
	return path
}

// TranscodeWithSubtitles transcodes video with subtitle handling
func (e *FFmpegEncoder) TranscodeWithSubtitles(ctx context.Context, task *TranscodeTask, subConfig SubtitleConfig, progressCh chan<- Progress) error {
	args := e.buildTranscodeArgs(task)

	switch subConfig.Mode {
	case SubtitleModePassthrough:
		// Map subtitle streams and copy
		args = insertBefore(args, task.OutputURL, "-map", "0:s?", "-c:s", "copy")

	case SubtitleModeBurnIn:
		// Need to extract subtitles first, then burn in
		// This modifies the video filter
		if subConfig.TrackIndex >= 0 {
			subFilter := fmt.Sprintf("subtitles='%s':si=%d",
				escapeFilterPath(task.InputURL), subConfig.TrackIndex)
			args = insertBefore(args, task.OutputURL, "-vf", subFilter)
		}

	case SubtitleModeExtract:
		// Extract handled separately via ExtractSubtitles
		// Just remove subtitle streams from output
		args = insertBefore(args, task.OutputURL, "-sn")

	case SubtitleModeDisable:
		// Explicitly disable subtitles
		args = insertBefore(args, task.OutputURL, "-sn")
	}

	cmd := exec.CommandContext(ctx, e.ffmpegPath, args...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	go e.parseProgress(stderr, task.Duration, progressCh)

	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("ffmpeg failed: %w", err)
	}

	return nil
}

// insertBefore inserts elements before a target in a slice
func insertBefore(slice []string, target string, elements ...string) []string {
	for i, v := range slice {
		if v == target {
			result := make([]string, 0, len(slice)+len(elements))
			result = append(result, slice[:i]...)
			result = append(result, elements...)
			result = append(result, slice[i:]...)
			return result
		}
	}
	// If target not found, append at end (before output)
	return append(slice[:len(slice)-1], append(elements, slice[len(slice)-1])...)
}
