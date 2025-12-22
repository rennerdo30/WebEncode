package ffmpeg

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rennerdo30/webencode/pkg/logger"
)

// ThumbnailConfig configures thumbnail generation
type ThumbnailConfig struct {
	// Timestamps to capture (in seconds). If empty, uses default positions.
	Timestamps []float64
	// Width of thumbnails (height calculated from aspect ratio)
	Width int
	// Quality for JPEG (1-31, lower is better)
	Quality int
	// Format: "jpg" or "png"
	Format string
}

// DefaultThumbnailConfig returns sensible defaults
func DefaultThumbnailConfig() ThumbnailConfig {
	return ThumbnailConfig{
		Width:   320,
		Quality: 5,
		Format:  "jpg",
	}
}

// ThumbnailResult contains generated thumbnail information
type ThumbnailResult struct {
	Timestamp float64
	FilePath  string
}

// GenerateThumbnails creates thumbnail images from a video at specified timestamps
func (e *FFmpegEncoder) GenerateThumbnails(ctx context.Context, inputURL string, outputDir string, config ThumbnailConfig) ([]ThumbnailResult, error) {
	// Set defaults
	if config.Width == 0 {
		config.Width = 320
	}
	if config.Quality == 0 {
		config.Quality = 5
	}
	if config.Format == "" {
		config.Format = "jpg"
	}

	// If no timestamps specified, probe video and generate at intervals
	timestamps := config.Timestamps
	if len(timestamps) == 0 {
		probe, err := e.Probe(ctx, inputURL)
		if err != nil {
			return nil, fmt.Errorf("failed to probe video for thumbnails: %w", err)
		}
		timestamps = generateThumbnailTimestamps(probe.Duration)
	}

	results := make([]ThumbnailResult, 0, len(timestamps))

	for i, ts := range timestamps {
		outputPath := filepath.Join(outputDir, fmt.Sprintf("thumb_%03d.%s", i, config.Format))

		err := e.generateSingleThumbnail(ctx, inputURL, outputPath, ts, config)
		if err != nil {
			e.logger.Warn("Failed to generate thumbnail", "timestamp", ts, "error", err)
			continue
		}

		results = append(results, ThumbnailResult{
			Timestamp: ts,
			FilePath:  outputPath,
		})
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("failed to generate any thumbnails")
	}

	e.logger.Info("Generated thumbnails", "count", len(results), "output_dir", outputDir)
	return results, nil
}

// generateSingleThumbnail creates one thumbnail at a specific timestamp
func (e *FFmpegEncoder) generateSingleThumbnail(ctx context.Context, inputURL, outputPath string, timestamp float64, config ThumbnailConfig) error {
	args := []string{
		"-hide_banner",
		"-loglevel", "error",
		"-y",
		"-ss", fmt.Sprintf("%.3f", timestamp),
		"-i", inputURL,
		"-vframes", "1",
		"-vf", fmt.Sprintf("scale=%d:-1", config.Width),
	}

	// Format-specific options
	if config.Format == "jpg" {
		args = append(args, "-q:v", fmt.Sprintf("%d", config.Quality))
	}

	args = append(args, outputPath)

	cmd := exec.CommandContext(ctx, e.ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg thumbnail failed: %w, output: %s", err, string(output))
	}

	return nil
}

// GenerateThumbnailGrid creates a sprite sheet of thumbnails (useful for video scrubbing)
func (e *FFmpegEncoder) GenerateThumbnailGrid(ctx context.Context, inputURL string, outputPath string, duration float64, cols, rows int) error {
	if cols == 0 {
		cols = 5
	}
	if rows == 0 {
		rows = 5
	}

	totalThumbs := cols * rows
	interval := duration / float64(totalThumbs)

	// Generate video filter for grid
	vf := fmt.Sprintf("fps=1/%.3f,scale=160:-1,tile=%dx%d", interval, cols, rows)

	args := []string{
		"-hide_banner",
		"-loglevel", "error",
		"-y",
		"-i", inputURL,
		"-vf", vf,
		"-frames:v", "1",
		outputPath,
	}

	cmd := exec.CommandContext(ctx, e.ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg grid failed: %w, output: %s", err, string(output))
	}

	e.logger.Info("Generated thumbnail grid", "path", outputPath, "grid", fmt.Sprintf("%dx%d", cols, rows))
	return nil
}

// GenerateAnimatedPreview creates a short animated GIF or WebP preview
func (e *FFmpegEncoder) GenerateAnimatedPreview(ctx context.Context, inputURL string, outputPath string, duration float64, previewDuration float64) error {
	if previewDuration == 0 {
		previewDuration = 3.0 // 3 second preview
	}

	// Start from 10% into the video
	startTime := duration * 0.1

	// Determine output format from extension
	ext := strings.ToLower(filepath.Ext(outputPath))
	isGif := ext == ".gif"

	var args []string

	if isGif {
		// GIF requires palette generation for quality
		args = []string{
			"-hide_banner",
			"-loglevel", "error",
			"-y",
			"-ss", fmt.Sprintf("%.3f", startTime),
			"-t", fmt.Sprintf("%.3f", previewDuration),
			"-i", inputURL,
			"-vf", "fps=10,scale=320:-1:flags=lanczos,split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse",
			"-loop", "0",
			outputPath,
		}
	} else {
		// WebP animated
		args = []string{
			"-hide_banner",
			"-loglevel", "error",
			"-y",
			"-ss", fmt.Sprintf("%.3f", startTime),
			"-t", fmt.Sprintf("%.3f", previewDuration),
			"-i", inputURL,
			"-vf", "fps=10,scale=320:-1",
			"-c:v", "libwebp",
			"-lossless", "0",
			"-compression_level", "6",
			"-q:v", "50",
			"-loop", "0",
			outputPath,
		}
	}

	cmd := exec.CommandContext(ctx, e.ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg preview failed: %w, output: %s", err, string(output))
	}

	e.logger.Info("Generated animated preview", "path", outputPath, "duration", previewDuration)
	return nil
}

// generateThumbnailTimestamps creates default thumbnail positions
func generateThumbnailTimestamps(duration float64) []float64 {
	// Generate 5 thumbnails at 10%, 25%, 50%, 75%, 90% of the video
	positions := []float64{0.1, 0.25, 0.5, 0.75, 0.9}
	timestamps := make([]float64, len(positions))

	for i, pos := range positions {
		timestamps[i] = duration * pos
	}

	return timestamps
}

// ThumbnailGenerator is a convenience wrapper for thumbnail generation
type ThumbnailGenerator struct {
	encoder *FFmpegEncoder
	logger  *logger.Logger
}

// NewThumbnailGenerator creates a new thumbnail generator
func NewThumbnailGenerator(ffmpegPath string, l *logger.Logger) *ThumbnailGenerator {
	return &ThumbnailGenerator{
		encoder: NewFFmpegEncoder(ffmpegPath, "", l),
		logger:  l,
	}
}

// Generate creates all thumbnail assets for a video
func (g *ThumbnailGenerator) Generate(ctx context.Context, inputURL, outputDir string, duration float64) error {
	// 1. Generate individual thumbnails
	_, err := g.encoder.GenerateThumbnails(ctx, inputURL, outputDir, DefaultThumbnailConfig())
	if err != nil {
		return fmt.Errorf("failed to generate thumbnails: %w", err)
	}

	// 2. Generate sprite sheet for scrubbing
	gridPath := filepath.Join(outputDir, "sprite.jpg")
	err = g.encoder.GenerateThumbnailGrid(ctx, inputURL, gridPath, duration, 10, 10)
	if err != nil {
		g.logger.Warn("Failed to generate sprite sheet", "error", err)
		// Don't fail the whole operation
	}

	// 3. Generate animated preview
	previewPath := filepath.Join(outputDir, "preview.webp")
	err = g.encoder.GenerateAnimatedPreview(ctx, inputURL, previewPath, duration, 3.0)
	if err != nil {
		g.logger.Warn("Failed to generate animated preview", "error", err)
		// Don't fail the whole operation
	}

	return nil
}
