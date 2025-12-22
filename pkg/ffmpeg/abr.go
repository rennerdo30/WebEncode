package ffmpeg

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// ABRVariant represents a quality variant in the ABR ladder
type ABRVariant struct {
	Name         string
	Width        int
	Height       int
	VideoBitrate int // kbps
	AudioBitrate int // kbps
	MaxBitrate   int // kbps (typically 110% of VideoBitrate)
	BufferSize   int // kbps (typically 200% of VideoBitrate)
}

// DefaultABRLadder returns a standard ABR ladder for HLS
func DefaultABRLadder() []ABRVariant {
	return []ABRVariant{
		{Name: "1080p", Width: 1920, Height: 1080, VideoBitrate: 5000, AudioBitrate: 192, MaxBitrate: 5500, BufferSize: 10000},
		{Name: "720p", Width: 1280, Height: 720, VideoBitrate: 2500, AudioBitrate: 128, MaxBitrate: 2750, BufferSize: 5000},
		{Name: "480p", Width: 854, Height: 480, VideoBitrate: 1000, AudioBitrate: 96, MaxBitrate: 1100, BufferSize: 2000},
		{Name: "360p", Width: 640, Height: 360, VideoBitrate: 600, AudioBitrate: 64, MaxBitrate: 660, BufferSize: 1200},
	}
}

// CompactABRLadder returns a smaller ladder for bandwidth-constrained scenarios
func CompactABRLadder() []ABRVariant {
	return []ABRVariant{
		{Name: "720p", Width: 1280, Height: 720, VideoBitrate: 2500, AudioBitrate: 128, MaxBitrate: 2750, BufferSize: 5000},
		{Name: "480p", Width: 854, Height: 480, VideoBitrate: 1000, AudioBitrate: 96, MaxBitrate: 1100, BufferSize: 2000},
	}
}

// HLSConfig configures HLS output
type HLSConfig struct {
	SegmentDuration int    // seconds (default 6)
	PlaylistType    string // "vod" or "event"
	MasterPlaylist  string // name of master playlist (default "master.m3u8")
}

// DefaultHLSConfig returns sensible HLS defaults
func DefaultHLSConfig() HLSConfig {
	return HLSConfig{
		SegmentDuration: 6,
		PlaylistType:    "vod",
		MasterPlaylist:  "master.m3u8",
	}
}

// HLSResult contains the output of HLS generation
type HLSResult struct {
	MasterPlaylist string
	Variants       []HLSVariantResult
}

// HLSVariantResult contains info about a generated variant
type HLSVariantResult struct {
	Name         string
	Playlist     string
	Width        int
	Height       int
	Bandwidth    int
	SegmentCount int
}

// GenerateHLSABR creates a multi-bitrate HLS package from a source video
func (e *FFmpegEncoder) GenerateHLSABR(ctx context.Context, inputURL, outputDir string, ladder []ABRVariant, hlsConfig HLSConfig) (*HLSResult, error) {
	if len(ladder) == 0 {
		ladder = DefaultABRLadder()
	}
	if hlsConfig.SegmentDuration == 0 {
		hlsConfig.SegmentDuration = 6
	}
	if hlsConfig.MasterPlaylist == "" {
		hlsConfig.MasterPlaylist = "master.m3u8"
	}

	args := []string{
		"-hide_banner",
		"-loglevel", "warning",
		"-y",
		"-i", inputURL,
	}

	// Build filter complex for splitting video to multiple qualities
	var filterParts []string
	var mapArgs []string
	var streamMaps []string

	for i, variant := range ladder {
		// Video filter for this variant
		filterParts = append(filterParts,
			fmt.Sprintf("[0:v]scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2[v%d]",
				variant.Width, variant.Height, variant.Width, variant.Height, i))

		// Map the filtered video
		mapArgs = append(mapArgs, "-map", fmt.Sprintf("[v%d]", i))

		// Video encoding settings for this variant
		mapArgs = append(mapArgs,
			fmt.Sprintf("-c:v:%d", i), "libx264",
			fmt.Sprintf("-b:v:%d", i), fmt.Sprintf("%dk", variant.VideoBitrate),
			fmt.Sprintf("-maxrate:v:%d", i), fmt.Sprintf("%dk", variant.MaxBitrate),
			fmt.Sprintf("-bufsize:v:%d", i), fmt.Sprintf("%dk", variant.BufferSize),
		)

		// Map audio (same source for all variants)
		mapArgs = append(mapArgs, "-map", "0:a")
		mapArgs = append(mapArgs,
			fmt.Sprintf("-c:a:%d", i), "aac",
			fmt.Sprintf("-b:a:%d", i), fmt.Sprintf("%dk", variant.AudioBitrate),
			fmt.Sprintf("-ar:%d", i), "48000",
		)

		// Stream map for HLS variant
		streamMaps = append(streamMaps, fmt.Sprintf("v:%d,a:%d", i, i))
	}

	// Add filter complex
	args = append(args, "-filter_complex", strings.Join(filterParts, ";"))

	// Add all map and encoding args
	args = append(args, mapArgs...)

	// Common encoding settings
	args = append(args,
		"-preset", "fast",
		"-pix_fmt", "yuv420p",
		"-g", fmt.Sprintf("%d", hlsConfig.SegmentDuration*30), // GOP = segment duration * fps
		"-keyint_min", fmt.Sprintf("%d", hlsConfig.SegmentDuration*30),
		"-sc_threshold", "0",
	)

	// HLS settings
	segmentPattern := filepath.Join(outputDir, "stream_%v", "segment_%03d.ts")
	playlistPattern := filepath.Join(outputDir, "stream_%v", "playlist.m3u8")
	masterPath := filepath.Join(outputDir, hlsConfig.MasterPlaylist)

	args = append(args,
		"-f", "hls",
		"-hls_time", fmt.Sprintf("%d", hlsConfig.SegmentDuration),
		"-hls_list_size", "0",
		"-hls_playlist_type", hlsConfig.PlaylistType,
		"-hls_segment_filename", segmentPattern,
		"-master_pl_name", hlsConfig.MasterPlaylist,
		"-var_stream_map", strings.Join(streamMaps, " "),
		playlistPattern,
	)

	e.logger.Info("Generating HLS ABR", "variants", len(ladder), "output", outputDir)

	cmd := exec.CommandContext(ctx, e.ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("HLS generation failed: %w, output: %s", err, string(output))
	}

	// Build result
	result := &HLSResult{
		MasterPlaylist: masterPath,
		Variants:       make([]HLSVariantResult, len(ladder)),
	}

	for i, variant := range ladder {
		result.Variants[i] = HLSVariantResult{
			Name:      variant.Name,
			Playlist:  filepath.Join(outputDir, fmt.Sprintf("stream_%d", i), "playlist.m3u8"),
			Width:     variant.Width,
			Height:    variant.Height,
			Bandwidth: (variant.VideoBitrate + variant.AudioBitrate) * 1000,
		}
	}

	e.logger.Info("HLS ABR generation complete", "master", masterPath, "variants", len(ladder))
	return result, nil
}

// GenerateDASH creates a DASH package with multiple qualities
func (e *FFmpegEncoder) GenerateDASH(ctx context.Context, inputURL, outputDir string, ladder []ABRVariant) (string, error) {
	if len(ladder) == 0 {
		ladder = DefaultABRLadder()
	}

	args := []string{
		"-hide_banner",
		"-loglevel", "warning",
		"-y",
		"-i", inputURL,
	}

	// Build filter complex and encoding args (similar to HLS)
	var filterParts []string
	var mapArgs []string

	for i, variant := range ladder {
		filterParts = append(filterParts,
			fmt.Sprintf("[0:v]scale=%d:%d[v%d]", variant.Width, variant.Height, i))

		mapArgs = append(mapArgs, "-map", fmt.Sprintf("[v%d]", i))
		mapArgs = append(mapArgs,
			fmt.Sprintf("-c:v:%d", i), "libx264",
			fmt.Sprintf("-b:v:%d", i), fmt.Sprintf("%dk", variant.VideoBitrate),
		)

		mapArgs = append(mapArgs, "-map", "0:a")
		mapArgs = append(mapArgs,
			fmt.Sprintf("-c:a:%d", i), "aac",
			fmt.Sprintf("-b:a:%d", i), fmt.Sprintf("%dk", variant.AudioBitrate),
		)
	}

	args = append(args, "-filter_complex", strings.Join(filterParts, ";"))
	args = append(args, mapArgs...)

	// DASH-specific settings
	manifestPath := filepath.Join(outputDir, "manifest.mpd")
	args = append(args,
		"-f", "dash",
		"-seg_duration", "4",
		"-adaptation_sets", "id=0,streams=v id=1,streams=a",
		manifestPath,
	)

	e.logger.Info("Generating DASH", "variants", len(ladder), "output", outputDir)

	cmd := exec.CommandContext(ctx, e.ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("DASH generation failed: %w, output: %s", err, string(output))
	}

	e.logger.Info("DASH generation complete", "manifest", manifestPath)
	return manifestPath, nil
}
