package ffmpeg

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseProbeOutput_ValidJSON(t *testing.T) {
	validJSON := `{
		"format": {
			"duration": "120.5",
			"format_name": "mov,mp4",
			"bit_rate": "5000000"
		},
		"streams": [
			{"index": 0, "codec_type": "video", "codec_name": "h264", "width": 1920, "height": 1080},
			{"index": 1, "codec_type": "audio", "codec_name": "aac", "width": 0, "height": 0}
		]
	}`

	result, err := parseProbeOutput([]byte(validJSON))

	require.NoError(t, err)
	assert.Equal(t, 120.5, result.Duration)
	assert.Equal(t, "mov,mp4", result.Format)
	assert.Equal(t, int64(5000000), result.Bitrate)
	assert.Equal(t, 1920, result.Width)
	assert.Equal(t, 1080, result.Height)
	assert.Len(t, result.Streams, 2)
	assert.Equal(t, "video", result.Streams[0].CodecType)
	assert.Equal(t, "audio", result.Streams[1].CodecType)
}

func TestParseProbeOutput_InvalidJSON(t *testing.T) {
	invalidJSON := `{invalid json}`

	result, err := parseProbeOutput([]byte(invalidJSON))

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to parse probe output")
}

func TestParseProbeOutput_EmptyJSON(t *testing.T) {
	emptyJSON := `{}`

	result, err := parseProbeOutput([]byte(emptyJSON))

	require.NoError(t, err)
	assert.Equal(t, 0.0, result.Duration)
	assert.Empty(t, result.Format)
	assert.Empty(t, result.Streams)
}

func TestParseProbeOutput_NoVideoStream(t *testing.T) {
	audioOnlyJSON := `{
		"format": {"duration": "180.0", "format_name": "mp3"},
		"streams": [
			{"index": 0, "codec_type": "audio", "codec_name": "mp3"}
		]
	}`

	result, err := parseProbeOutput([]byte(audioOnlyJSON))

	require.NoError(t, err)
	assert.Equal(t, 0, result.Width)
	assert.Equal(t, 0, result.Height)
	assert.Len(t, result.Streams, 1)
}

func TestParseProbeOutput_MultipleVideoStreams(t *testing.T) {
	multiVideoJSON := `{
		"format": {"duration": "60.0"},
		"streams": [
			{"index": 0, "codec_type": "video", "codec_name": "h264", "width": 1920, "height": 1080},
			{"index": 1, "codec_type": "video", "codec_name": "h264", "width": 1280, "height": 720}
		]
	}`

	result, err := parseProbeOutput([]byte(multiVideoJSON))

	require.NoError(t, err)
	// Should use first video stream dimensions
	assert.Equal(t, 1920, result.Width)
	assert.Equal(t, 1080, result.Height)
}

func TestParseProbeOutput_InvalidDuration(t *testing.T) {
	invalidDurationJSON := `{
		"format": {"duration": "not-a-number"},
		"streams": []
	}`

	result, err := parseProbeOutput([]byte(invalidDurationJSON))

	require.NoError(t, err)
	assert.Equal(t, 0.0, result.Duration) // Failed to parse, defaults to 0
}

func TestPresetProfiles(t *testing.T) {
	assert.NotNil(t, PresetProfiles)
	assert.NotEmpty(t, PresetProfiles)

	expectedProfiles := []string{
		"1080p_h264",
		"720p_h264",
		"480p_h264",
		"4k_hevc",
		"1080p_vp9",
	}

	for _, name := range expectedProfiles {
		t.Run(name, func(t *testing.T) {
			profile, ok := PresetProfiles[name]
			assert.True(t, ok, "profile %s should exist", name)
			assert.NotEmpty(t, profile.VideoCodec)
			assert.NotEmpty(t, profile.AudioCodec)
			assert.Greater(t, profile.Width, 0)
			assert.Greater(t, profile.Height, 0)
			assert.Greater(t, profile.BitrateKbps, 0)
		})
	}
}

func TestGetProfile_Exists(t *testing.T) {
	profile, ok := GetProfile("1080p_h264")

	assert.True(t, ok)
	assert.NotNil(t, profile)
	assert.Equal(t, "libx264", profile.VideoCodec)
	assert.Equal(t, 1920, profile.Width)
	assert.Equal(t, 1080, profile.Height)
}

func TestGetProfile_NotExists(t *testing.T) {
	profile, ok := GetProfile("nonexistent_profile")

	assert.False(t, ok)
	assert.Nil(t, profile)
}

func TestGetProfile_ReturnsCopy(t *testing.T) {
	profile1, _ := GetProfile("1080p_h264")
	profile2, _ := GetProfile("1080p_h264")

	// Modify profile1
	profile1.BitrateKbps = 99999

	// profile2 should be unchanged
	assert.NotEqual(t, profile1.BitrateKbps, profile2.BitrateKbps)
	assert.Equal(t, 5000, profile2.BitrateKbps)
}

func TestGetAvailableProfiles(t *testing.T) {
	profiles := GetAvailableProfiles()

	assert.NotEmpty(t, profiles)
	assert.GreaterOrEqual(t, len(profiles), 5)

	// Check that all expected profiles are in the list
	expectedProfiles := []string{"1080p_h264", "720p_h264", "480p_h264", "4k_hevc", "1080p_vp9"}
	for _, expected := range expectedProfiles {
		found := false
		for _, profile := range profiles {
			if profile == expected {
				found = true
				break
			}
		}
		assert.True(t, found, "expected profile %s not found", expected)
	}
}

func TestBuildConcatDemuxer(t *testing.T) {
	t.Run("simple paths", func(t *testing.T) {
		paths := []string{"segment1.ts", "segment2.ts", "segment3.ts"}
		result := BuildConcatDemuxer(paths)

		assert.Contains(t, result, "file 'segment1.ts'")
		assert.Contains(t, result, "file 'segment2.ts'")
		assert.Contains(t, result, "file 'segment3.ts'")
	})

	t.Run("paths with spaces", func(t *testing.T) {
		paths := []string{"/path/with spaces/segment.ts"}
		result := BuildConcatDemuxer(paths)

		assert.Contains(t, result, "file '/path/with spaces/segment.ts'")
	})

	t.Run("paths with single quotes", func(t *testing.T) {
		paths := []string{"file'with'quotes.ts"}
		result := BuildConcatDemuxer(paths)

		// Single quotes should be escaped
		assert.Contains(t, result, "'\\''")
	})

	t.Run("empty paths", func(t *testing.T) {
		paths := []string{}
		result := BuildConcatDemuxer(paths)

		assert.Empty(t, result)
	})
}

func TestStitchCommand(t *testing.T) {
	args := StitchCommand("/tmp/concat.txt", "/output/final.mp4")

	assert.Contains(t, args, "-hide_banner")
	assert.Contains(t, args, "-f")
	assert.Contains(t, args, "concat")
	assert.Contains(t, args, "-i")
	assert.Contains(t, args, "/tmp/concat.txt")
	assert.Contains(t, args, "-c")
	assert.Contains(t, args, "copy")
	assert.Contains(t, args, "-movflags")
	assert.Contains(t, args, "+faststart")
	assert.Contains(t, args, "/output/final.mp4")

	// Output should be the last element
	assert.Equal(t, "/output/final.mp4", args[len(args)-1])
}

func TestHLSVariant_Fields(t *testing.T) {
	variant := HLSVariant{
		Path:       "stream_0/playlist.m3u8",
		Bandwidth:  5000000,
		Resolution: "1920x1080",
	}

	assert.Equal(t, "stream_0/playlist.m3u8", variant.Path)
	assert.Equal(t, 5000000, variant.Bandwidth)
	assert.Equal(t, "1920x1080", variant.Resolution)
}

func TestBuildMasterPlaylist(t *testing.T) {
	variants := []HLSVariant{
		{Path: "stream_0/playlist.m3u8", Bandwidth: 5000000, Resolution: "1920x1080"},
		{Path: "stream_1/playlist.m3u8", Bandwidth: 2500000, Resolution: "1280x720"},
		{Path: "stream_2/playlist.m3u8", Bandwidth: 1000000, Resolution: "854x480"},
	}

	result := BuildMasterPlaylist(variants)

	// Check header
	assert.Contains(t, result, "#EXTM3U")
	assert.Contains(t, result, "#EXT-X-VERSION:3")

	// Check variant entries
	assert.Contains(t, result, "#EXT-X-STREAM-INF:BANDWIDTH=5000000,RESOLUTION=1920x1080")
	assert.Contains(t, result, "stream_0/playlist.m3u8")
	assert.Contains(t, result, "#EXT-X-STREAM-INF:BANDWIDTH=2500000,RESOLUTION=1280x720")
	assert.Contains(t, result, "stream_1/playlist.m3u8")
}

func TestBuildMasterPlaylist_Empty(t *testing.T) {
	variants := []HLSVariant{}
	result := BuildMasterPlaylist(variants)

	assert.Contains(t, result, "#EXTM3U")
	assert.Contains(t, result, "#EXT-X-VERSION:3")
	// Should not contain any STREAM-INF entries
	assert.NotContains(t, result, "#EXT-X-STREAM-INF")
}

func TestProfile_1080p_H264(t *testing.T) {
	profile := PresetProfiles["1080p_h264"]

	assert.Equal(t, "libx264", profile.VideoCodec)
	assert.Equal(t, "aac", profile.AudioCodec)
	assert.Equal(t, 1920, profile.Width)
	assert.Equal(t, 1080, profile.Height)
	assert.Equal(t, 5000, profile.BitrateKbps)
	assert.Equal(t, "fast", profile.Preset)
	assert.Equal(t, "mp4", profile.Container)
}

func TestProfile_720p_H264(t *testing.T) {
	profile := PresetProfiles["720p_h264"]

	assert.Equal(t, 1280, profile.Width)
	assert.Equal(t, 720, profile.Height)
	assert.Equal(t, 2500, profile.BitrateKbps)
}

func TestProfile_4K_HEVC(t *testing.T) {
	profile := PresetProfiles["4k_hevc"]

	assert.Equal(t, "libx265", profile.VideoCodec)
	assert.Equal(t, 3840, profile.Width)
	assert.Equal(t, 2160, profile.Height)
	assert.Equal(t, 15000, profile.BitrateKbps)
	assert.Equal(t, "medium", profile.Preset)
}

func TestProfile_1080p_VP9(t *testing.T) {
	profile := PresetProfiles["1080p_vp9"]

	assert.Equal(t, "libvpx-vp9", profile.VideoCodec)
	assert.Equal(t, "libopus", profile.AudioCodec)
	assert.Equal(t, "webm", profile.Container)
}

func TestBuildConcatDemuxer_LineFormat(t *testing.T) {
	paths := []string{"a.ts", "b.ts"}
	result := BuildConcatDemuxer(paths)

	lines := strings.Split(strings.TrimSpace(result), "\n")
	assert.Len(t, lines, 2)

	for _, line := range lines {
		assert.True(t, strings.HasPrefix(line, "file '"))
		assert.True(t, strings.HasSuffix(line, "'"))
	}
}
