package ffmpeg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubtitleMode_Constants(t *testing.T) {
	assert.Equal(t, SubtitleMode("passthrough"), SubtitleModePassthrough)
	assert.Equal(t, SubtitleMode("extract"), SubtitleModeExtract)
	assert.Equal(t, SubtitleMode("burnin"), SubtitleModeBurnIn)
	assert.Equal(t, SubtitleMode("disable"), SubtitleModeDisable)
}

func TestSubtitleTrack_Fields(t *testing.T) {
	track := SubtitleTrack{
		Index:    0,
		Language: "eng",
		Title:    "English",
		Codec:    "srt",
		Default:  true,
		Forced:   false,
	}

	assert.Equal(t, 0, track.Index)
	assert.Equal(t, "eng", track.Language)
	assert.Equal(t, "English", track.Title)
	assert.Equal(t, "srt", track.Codec)
	assert.True(t, track.Default)
	assert.False(t, track.Forced)
}

func TestSubtitleConfig_Fields(t *testing.T) {
	config := SubtitleConfig{
		Mode:         SubtitleModeBurnIn,
		TrackIndex:   0,
		Language:     "eng",
		FontSize:     32,
		FontName:     "DejaVu Sans",
		OutlineWidth: 3,
	}

	assert.Equal(t, SubtitleModeBurnIn, config.Mode)
	assert.Equal(t, 0, config.TrackIndex)
	assert.Equal(t, "eng", config.Language)
	assert.Equal(t, 32, config.FontSize)
	assert.Equal(t, "DejaVu Sans", config.FontName)
	assert.Equal(t, 3, config.OutlineWidth)
}

func TestDefaultSubtitleConfig(t *testing.T) {
	config := DefaultSubtitleConfig()

	assert.Equal(t, SubtitleModePassthrough, config.Mode)
	assert.Equal(t, -1, config.TrackIndex)
	assert.Equal(t, 24, config.FontSize)
	assert.Equal(t, "Arial", config.FontName)
	assert.Equal(t, 2, config.OutlineWidth)
}

func TestParseSubtitleTracks_WithSubtitles(t *testing.T) {
	// Output that indicates a subtitle stream exists
	output := []byte(`{"streams": [{"codec_type": "subtitle"}]}`)

	tracks, err := parseSubtitleTracks(output)

	assert.NoError(t, err)
	assert.Len(t, tracks, 1)
	assert.Equal(t, 0, tracks[0].Index)
}

func TestParseSubtitleTracks_NoSubtitles(t *testing.T) {
	output := []byte(`{"streams": [{"codec_type": "video"}]}`)

	tracks, err := parseSubtitleTracks(output)

	assert.NoError(t, err)
	assert.Empty(t, tracks)
}

func TestParseSubtitleTracks_EmptyOutput(t *testing.T) {
	output := []byte(`{}`)

	tracks, err := parseSubtitleTracks(output)

	assert.NoError(t, err)
	assert.Empty(t, tracks)
}

func TestBuildBurnInArgs(t *testing.T) {
	config := SubtitleConfig{
		FontSize:     28,
		FontName:     "Arial",
		OutlineWidth: 2,
	}

	args := BuildBurnInArgs(config, "/path/to/subs.srt")

	assert.Len(t, args, 2)
	assert.Equal(t, "-vf", args[0])
	assert.Contains(t, args[1], "subtitles=")
	assert.Contains(t, args[1], "FontSize=28")
	assert.Contains(t, args[1], "FontName=Arial")
	assert.Contains(t, args[1], "Outline=2")
}

func TestBuildBurnInArgs_PathEscaping(t *testing.T) {
	config := DefaultSubtitleConfig()

	// Path with special characters
	args := BuildBurnInArgs(config, "/path/with:colon/and'quote.srt")

	assert.Contains(t, args[1], "\\:")  // Colon should be escaped
	assert.Contains(t, args[1], "\\'") // Quote should be escaped
}

func TestBuildPassthroughArgs(t *testing.T) {
	args := BuildPassthroughArgs()

	assert.Len(t, args, 2)
	assert.Equal(t, "-c:s", args[0])
	assert.Equal(t, "copy", args[1])
}

func TestEscapeFilterPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/simple/path.srt", "/simple/path.srt"},
		{"/path/with'quote.srt", "/path/with\\'quote.srt"},
		{"/path/with:colon.srt", "/path/with\\:colon.srt"},
		{"/path/with\\backslash.srt", "/path/with\\\\backslash.srt"},
		{"C:\\Windows\\path.srt", "C\\:\\\\Windows\\\\path.srt"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := escapeFilterPath(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestInsertBefore_TargetFound(t *testing.T) {
	slice := []string{"a", "b", "c", "d"}

	result := insertBefore(slice, "c", "x", "y")

	assert.Equal(t, []string{"a", "b", "x", "y", "c", "d"}, result)
}

func TestInsertBefore_TargetNotFound(t *testing.T) {
	slice := []string{"a", "b", "c"}

	result := insertBefore(slice, "z", "x", "y")

	// Should append before last element
	assert.Equal(t, []string{"a", "b", "x", "y", "c"}, result)
}

func TestInsertBefore_SingleElement(t *testing.T) {
	slice := []string{"output.mp4"}

	result := insertBefore(slice, "output.mp4", "-sn")

	assert.Equal(t, []string{"-sn", "output.mp4"}, result)
}

// Note: insertBefore panics on empty slice - this is expected behavior
// as it's always called with non-empty args slices from buildTranscodeArgs

func TestSubtitleTrack_Codec_VTT(t *testing.T) {
	track := SubtitleTrack{
		Index: 0,
		Codec: "webvtt",
	}

	// Check that VTT codec is recognized
	assert.Equal(t, "webvtt", track.Codec)
}

func TestSubtitleTrack_Codec_SRT(t *testing.T) {
	track := SubtitleTrack{
		Index: 0,
		Codec: "srt",
	}

	assert.Equal(t, "srt", track.Codec)
}

func TestSubtitleConfig_AllModes(t *testing.T) {
	modes := []SubtitleMode{
		SubtitleModePassthrough,
		SubtitleModeExtract,
		SubtitleModeBurnIn,
		SubtitleModeDisable,
	}

	for _, mode := range modes {
		t.Run(string(mode), func(t *testing.T) {
			config := SubtitleConfig{Mode: mode}
			assert.Equal(t, mode, config.Mode)
		})
	}
}

func TestBuildBurnInArgs_CustomStyle(t *testing.T) {
	config := SubtitleConfig{
		FontSize:     48,
		FontName:     "Comic Sans MS",
		OutlineWidth: 5,
	}

	args := BuildBurnInArgs(config, "/subs.srt")

	assert.Contains(t, args[1], "FontSize=48")
	assert.Contains(t, args[1], "FontName=Comic Sans MS")
	assert.Contains(t, args[1], "Outline=5")
}

func TestSubtitleTrack_ForcedAndDefault(t *testing.T) {
	forcedTrack := SubtitleTrack{
		Index:   0,
		Forced:  true,
		Default: false,
	}

	defaultTrack := SubtitleTrack{
		Index:   1,
		Forced:  false,
		Default: true,
	}

	assert.True(t, forcedTrack.Forced)
	assert.False(t, forcedTrack.Default)
	assert.False(t, defaultTrack.Forced)
	assert.True(t, defaultTrack.Default)
}
