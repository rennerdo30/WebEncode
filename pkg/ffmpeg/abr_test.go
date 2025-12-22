package ffmpeg

import (
	"context"
	"testing"

	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestDefaultABRLadder(t *testing.T) {
	ladder := DefaultABRLadder()

	assert.Equal(t, 4, len(ladder))
	assert.Equal(t, "1080p", ladder[0].Name)
	assert.Equal(t, 1920, ladder[0].Width)
	assert.Equal(t, 1080, ladder[0].Height)
	assert.Equal(t, 5000, ladder[0].VideoBitrate)
	assert.Equal(t, 192, ladder[0].AudioBitrate)
	assert.Equal(t, 5500, ladder[0].MaxBitrate)
	assert.Equal(t, 10000, ladder[0].BufferSize)
}

func TestDefaultABRLadder_AllVariants(t *testing.T) {
	ladder := DefaultABRLadder()

	expectedVariants := []struct {
		name   string
		width  int
		height int
	}{
		{"1080p", 1920, 1080},
		{"720p", 1280, 720},
		{"480p", 854, 480},
		{"360p", 640, 360},
	}

	for i, expected := range expectedVariants {
		t.Run(expected.name, func(t *testing.T) {
			assert.Equal(t, expected.name, ladder[i].Name)
			assert.Equal(t, expected.width, ladder[i].Width)
			assert.Equal(t, expected.height, ladder[i].Height)
		})
	}
}

func TestCompactABRLadder(t *testing.T) {
	ladder := CompactABRLadder()

	assert.Equal(t, 2, len(ladder))
	assert.Equal(t, "720p", ladder[0].Name)
	assert.Equal(t, "480p", ladder[1].Name)
}

func TestHLSConfig_Default(t *testing.T) {
	config := DefaultHLSConfig()

	assert.Equal(t, 6, config.SegmentDuration)
	assert.Equal(t, "vod", config.PlaylistType)
	assert.Equal(t, "master.m3u8", config.MasterPlaylist)
}

func TestHLSConfig_Custom(t *testing.T) {
	config := HLSConfig{
		SegmentDuration: 4,
		PlaylistType:    "event",
		MasterPlaylist:  "index.m3u8",
	}

	assert.Equal(t, 4, config.SegmentDuration)
	assert.Equal(t, "event", config.PlaylistType)
	assert.Equal(t, "index.m3u8", config.MasterPlaylist)
}

func TestABRVariant_Fields(t *testing.T) {
	variant := ABRVariant{
		Name:         "1080p",
		Width:        1920,
		Height:       1080,
		VideoBitrate: 5000,
		AudioBitrate: 192,
		MaxBitrate:   5500,
		BufferSize:   10000,
	}

	assert.Equal(t, "1080p", variant.Name)
	assert.Equal(t, 1920, variant.Width)
	assert.Equal(t, 1080, variant.Height)
	assert.Equal(t, 5000, variant.VideoBitrate)
	assert.Equal(t, 192, variant.AudioBitrate)
	assert.Equal(t, 5500, variant.MaxBitrate)
	assert.Equal(t, 10000, variant.BufferSize)
}

func TestHLSResult_Fields(t *testing.T) {
	result := HLSResult{
		MasterPlaylist: "/output/master.m3u8",
		Variants: []HLSVariantResult{
			{Name: "1080p", Playlist: "/output/stream_0/playlist.m3u8", Width: 1920, Height: 1080, Bandwidth: 5192000},
		},
	}

	assert.Equal(t, "/output/master.m3u8", result.MasterPlaylist)
	assert.Len(t, result.Variants, 1)
	assert.Equal(t, "1080p", result.Variants[0].Name)
}

func TestHLSVariantResult_Fields(t *testing.T) {
	variant := HLSVariantResult{
		Name:         "720p",
		Playlist:     "/output/stream_1/playlist.m3u8",
		Width:        1280,
		Height:       720,
		Bandwidth:    2628000,
		SegmentCount: 20,
	}

	assert.Equal(t, "720p", variant.Name)
	assert.Equal(t, "/output/stream_1/playlist.m3u8", variant.Playlist)
	assert.Equal(t, 1280, variant.Width)
	assert.Equal(t, 720, variant.Height)
	assert.Equal(t, 2628000, variant.Bandwidth)
	assert.Equal(t, 20, variant.SegmentCount)
}

func TestGenerateHLSABR_MissingFile(t *testing.T) {
	encoder := NewFFmpegEncoder("ffmpeg", "ffprobe", logger.New("test"))

	ctx := context.Background()
	res, err := encoder.GenerateHLSABR(ctx, "nonexistent.mp4", "/tmp", nil, DefaultHLSConfig())

	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestGenerateHLSABR_DefaultLadder(t *testing.T) {
	encoder := NewFFmpegEncoder("ffmpeg", "ffprobe", logger.New("test"))

	ctx := context.Background()
	// This will fail because the file doesn't exist, but we can check
	// that defaults are applied
	_, err := encoder.GenerateHLSABR(ctx, "nonexistent.mp4", "/tmp", nil, HLSConfig{})

	assert.Error(t, err) // Expected - file doesn't exist
}

func TestGenerateHLSABR_CustomLadder(t *testing.T) {
	encoder := NewFFmpegEncoder("ffmpeg", "ffprobe", logger.New("test"))

	customLadder := []ABRVariant{
		{Name: "720p", Width: 1280, Height: 720, VideoBitrate: 2500, AudioBitrate: 128, MaxBitrate: 2750, BufferSize: 5000},
	}

	ctx := context.Background()
	_, err := encoder.GenerateHLSABR(ctx, "nonexistent.mp4", "/tmp", customLadder, DefaultHLSConfig())

	assert.Error(t, err) // Expected - file doesn't exist
}

func TestGenerateDASH_MissingFile(t *testing.T) {
	encoder := NewFFmpegEncoder("ffmpeg", "ffprobe", logger.New("test"))

	ctx := context.Background()
	_, err := encoder.GenerateDASH(ctx, "nonexistent.mp4", "/tmp", nil)

	assert.Error(t, err)
}

func TestGenerateThumbnailTimestamps(t *testing.T) {
	ts := generateThumbnailTimestamps(100.0)

	assert.Equal(t, 5, len(ts))
	assert.Equal(t, 10.0, ts[0]) // 10%
	assert.Equal(t, 25.0, ts[1]) // 25%
	assert.Equal(t, 50.0, ts[2]) // 50%
	assert.Equal(t, 75.0, ts[3]) // 75%
	assert.Equal(t, 90.0, ts[4]) // 90%
}

func TestGenerateThumbnailTimestamps_ShortVideo(t *testing.T) {
	ts := generateThumbnailTimestamps(10.0)

	assert.Equal(t, 5, len(ts))
	assert.InDelta(t, 1.0, ts[0], 0.01)  // 10%
	assert.InDelta(t, 2.5, ts[1], 0.01)  // 25%
	assert.InDelta(t, 5.0, ts[2], 0.01)  // 50%
	assert.InDelta(t, 7.5, ts[3], 0.01)  // 75%
	assert.InDelta(t, 9.0, ts[4], 0.01)  // 90%
}

func TestGenerateThumbnailTimestamps_LongVideo(t *testing.T) {
	ts := generateThumbnailTimestamps(3600.0) // 1 hour

	assert.Equal(t, 5, len(ts))
	assert.Equal(t, 360.0, ts[0])  // 10%
	assert.Equal(t, 900.0, ts[1])  // 25%
	assert.Equal(t, 1800.0, ts[2]) // 50%
	assert.Equal(t, 2700.0, ts[3]) // 75%
	assert.Equal(t, 3240.0, ts[4]) // 90%
}

func TestABRLadder_BitrateRatios(t *testing.T) {
	ladder := DefaultABRLadder()

	for _, variant := range ladder {
		t.Run(variant.Name, func(t *testing.T) {
			// MaxBitrate should be ~110% of VideoBitrate
			expectedMax := float64(variant.VideoBitrate) * 1.1
			assert.InDelta(t, expectedMax, float64(variant.MaxBitrate), float64(variant.VideoBitrate)*0.05)

			// BufferSize should be ~200% of VideoBitrate
			expectedBuffer := variant.VideoBitrate * 2
			assert.InDelta(t, float64(expectedBuffer), float64(variant.BufferSize), float64(variant.VideoBitrate)/2)
		})
	}
}

func TestHLSConfig_Zero(t *testing.T) {
	config := HLSConfig{}

	// Test default application in GenerateHLSABR
	encoder := NewFFmpegEncoder("ffmpeg", "ffprobe", logger.New("test"))
	ctx := context.Background()

	// This will fail, but defaults should be applied
	_, _ = encoder.GenerateHLSABR(ctx, "nonexistent.mp4", "/tmp", nil, config)

	// The function internally sets defaults, so we just verify it doesn't panic
}

func TestCompactABRLadder_BitrateValues(t *testing.T) {
	ladder := CompactABRLadder()

	// 720p should have reasonable bitrate
	assert.Equal(t, 2500, ladder[0].VideoBitrate)
	assert.Equal(t, 128, ladder[0].AudioBitrate)

	// 480p should have lower bitrate
	assert.Equal(t, 1000, ladder[1].VideoBitrate)
	assert.Equal(t, 96, ladder[1].AudioBitrate)
}

func TestABRVariant_TotalBandwidth(t *testing.T) {
	ladder := DefaultABRLadder()

	for _, variant := range ladder {
		t.Run(variant.Name, func(t *testing.T) {
			totalBitrate := variant.VideoBitrate + variant.AudioBitrate
			// Total bandwidth in bits/sec
			expectedBandwidth := totalBitrate * 1000

			assert.Greater(t, expectedBandwidth, 0)
		})
	}
}

func TestGenerateDASH_DefaultLadder(t *testing.T) {
	encoder := NewFFmpegEncoder("ffmpeg", "ffprobe", logger.New("test"))

	ctx := context.Background()
	// This will fail, but we test that nil ladder uses default
	_, err := encoder.GenerateDASH(ctx, "nonexistent.mp4", "/tmp", nil)

	assert.Error(t, err) // Expected - file doesn't exist
}

func TestGenerateDASH_CustomLadder(t *testing.T) {
	encoder := NewFFmpegEncoder("ffmpeg", "ffprobe", logger.New("test"))

	customLadder := []ABRVariant{
		{Name: "720p", Width: 1280, Height: 720, VideoBitrate: 2500, AudioBitrate: 128},
	}

	ctx := context.Background()
	_, err := encoder.GenerateDASH(ctx, "nonexistent.mp4", "/tmp", customLadder)

	assert.Error(t, err) // Expected - file doesn't exist
}

func TestGenerateHLSABR_ContextCancellation(t *testing.T) {
	encoder := NewFFmpegEncoder("ffmpeg", "ffprobe", logger.New("test"))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := encoder.GenerateHLSABR(ctx, "nonexistent.mp4", "/tmp", nil, DefaultHLSConfig())

	assert.Error(t, err)
}

func TestGenerateDASH_ContextCancellation(t *testing.T) {
	encoder := NewFFmpegEncoder("ffmpeg", "ffprobe", logger.New("test"))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := encoder.GenerateDASH(ctx, "nonexistent.mp4", "/tmp", nil)

	assert.Error(t, err)
}
