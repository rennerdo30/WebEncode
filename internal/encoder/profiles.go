package encoder

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// parseProbeOutput parses ffprobe JSON output
func parseProbeOutput(data []byte) (*ProbeResult, error) {
	var probe struct {
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
	}

	if err := json.Unmarshal(data, &probe); err != nil {
		return nil, fmt.Errorf("failed to parse probe output: %w", err)
	}

	result := &ProbeResult{
		Format:  probe.Format.FormatName,
		Streams: make([]StreamInfo, 0, len(probe.Streams)),
	}

	// Parse duration
	if duration, err := strconv.ParseFloat(probe.Format.Duration, 64); err == nil {
		result.Duration = duration
	}

	// Parse bitrate
	if bitrate, err := strconv.ParseInt(probe.Format.BitRate, 10, 64); err == nil {
		result.Bitrate = bitrate
	}

	// Parse streams
	for _, s := range probe.Streams {
		result.Streams = append(result.Streams, StreamInfo{
			Index:     s.Index,
			CodecType: s.CodecType,
			CodecName: s.CodecName,
		})

		// Get dimensions from first video stream
		if s.CodecType == "video" && result.Width == 0 {
			result.Width = s.Width
			result.Height = s.Height
		}
	}

	return result, nil
}

// Profile presets for common encoding scenarios
var PresetProfiles = map[string]*TranscodeTask{
	"1080p_h264": {
		VideoCodec:  "libx264",
		AudioCodec:  "aac",
		Width:       1920,
		Height:      1080,
		BitrateKbps: 5000,
		Preset:      "fast",
		Container:   "mp4",
	},
	"720p_h264": {
		VideoCodec:  "libx264",
		AudioCodec:  "aac",
		Width:       1280,
		Height:      720,
		BitrateKbps: 2500,
		Preset:      "fast",
		Container:   "mp4",
	},
	"480p_h264": {
		VideoCodec:  "libx264",
		AudioCodec:  "aac",
		Width:       854,
		Height:      480,
		BitrateKbps: 1000,
		Preset:      "fast",
		Container:   "mp4",
	},
	"4k_hevc": {
		VideoCodec:  "libx265",
		AudioCodec:  "aac",
		Width:       3840,
		Height:      2160,
		BitrateKbps: 15000,
		Preset:      "medium",
		Container:   "mp4",
	},
	"1080p_vp9": {
		VideoCodec:  "libvpx-vp9",
		AudioCodec:  "libopus",
		Width:       1920,
		Height:      1080,
		BitrateKbps: 4000,
		Preset:      "good",
		Container:   "webm",
	},
}

// GetProfile returns a preset profile by name
func GetProfile(name string) (*TranscodeTask, bool) {
	profile, ok := PresetProfiles[name]
	if !ok {
		return nil, false
	}
	// Return a copy
	copy := *profile
	return &copy, true
}

// GetAvailableProfiles returns all available profile names
func GetAvailableProfiles() []string {
	profiles := make([]string, 0, len(PresetProfiles))
	for name := range PresetProfiles {
		profiles = append(profiles, name)
	}
	return profiles
}

// BuildConcatDemuxer creates ffmpeg concat demuxer file content
func BuildConcatDemuxer(segmentPaths []string) string {
	var sb strings.Builder
	for _, path := range segmentPaths {
		// Escape single quotes in path
		escaped := strings.ReplaceAll(path, "'", "'\\''")
		sb.WriteString(fmt.Sprintf("file '%s'\n", escaped))
	}
	return sb.String()
}

// StitchCommand returns FFmpeg args for stitching segments
func StitchCommand(concatFilePath, outputPath string) []string {
	return []string{
		"-hide_banner",
		"-loglevel", "warning",
		"-y",
		"-f", "concat",
		"-safe", "0",
		"-i", concatFilePath,
		"-c", "copy",
		"-movflags", "+faststart",
		outputPath,
	}
}
