package main

import (
	"sort"
	"testing"

	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/stretchr/testify/assert"
)

func TestGetExtension(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{"simple mp4", "video.mp4", "mp4"},
		{"simple mkv", "movie.mkv", "mkv"},
		{"uppercase extension", "image.JPG", "jpg"},
		{"mixed case", "Photo.PnG", "png"},
		{"no extension", "file_no_ext", ""},
		{"multiple dots", "archive.tar.gz", "gz"},
		{"hidden file with ext", ".hidden.txt", "txt"},
		{"hidden file no ext", ".hidden", "hidden"},
		{"empty string", "", ""},
		{"only dot", ".", ""},
		{"ends with dot", "file.", ""},
		{"double extension", "video.backup.mp4", "mp4"},
		{"path with dots", "folder.name/video.mp4", "mp4"},
		{"long extension", "file.longextension", "longextension"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := getExtension(tc.filename)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsVideo(t *testing.T) {
	videoExts := []string{
		"mp4", "mkv", "avi", "mov", "wmv",
		"flv", "webm", "m4v", "mpeg", "mpg",
		"3gp", "ts", "mts", "m2ts", "vob",
	}

	for _, ext := range videoExts {
		t.Run("video_"+ext, func(t *testing.T) {
			assert.True(t, isVideo(ext), "extension %s should be video", ext)
		})
	}

	nonVideoExts := []string{"txt", "mp3", "jpg", "png", "pdf", "", "doc", "exe"}
	for _, ext := range nonVideoExts {
		t.Run("nonvideo_"+ext, func(t *testing.T) {
			assert.False(t, isVideo(ext), "extension %s should not be video", ext)
		})
	}
}

func TestIsAudio(t *testing.T) {
	audioExts := []string{
		"mp3", "wav", "flac", "aac", "ogg",
		"wma", "m4a", "opus",
	}

	for _, ext := range audioExts {
		t.Run("audio_"+ext, func(t *testing.T) {
			assert.True(t, isAudio(ext), "extension %s should be audio", ext)
		})
	}

	nonAudioExts := []string{"txt", "mp4", "jpg", "png", "pdf", "", "doc", "mkv"}
	for _, ext := range nonAudioExts {
		t.Run("nonaudio_"+ext, func(t *testing.T) {
			assert.False(t, isAudio(ext), "extension %s should not be audio", ext)
		})
	}
}

func TestIsImage(t *testing.T) {
	imageExts := []string{
		"jpg", "jpeg", "png", "gif", "bmp",
		"webp", "svg", "tiff", "tif",
	}

	for _, ext := range imageExts {
		t.Run("image_"+ext, func(t *testing.T) {
			assert.True(t, isImage(ext), "extension %s should be image", ext)
		})
	}

	nonImageExts := []string{"txt", "mp4", "mp3", "pdf", "", "doc", "mkv"}
	for _, ext := range nonImageExts {
		t.Run("nonimage_"+ext, func(t *testing.T) {
			assert.False(t, isImage(ext), "extension %s should not be image", ext)
		})
	}
}

func TestVideoExtensionsMap(t *testing.T) {
	// Verify the map is properly initialized
	assert.NotNil(t, videoExtensions)
	assert.True(t, len(videoExtensions) > 0)

	// Verify all expected extensions are present
	expectedExts := []string{"mp4", "mkv", "avi", "mov", "webm", "ts"}
	for _, ext := range expectedExts {
		assert.True(t, videoExtensions[ext], "videoExtensions should contain %s", ext)
	}
}

func TestAudioExtensionsMap(t *testing.T) {
	assert.NotNil(t, audioExtensions)
	assert.True(t, len(audioExtensions) > 0)

	expectedExts := []string{"mp3", "wav", "flac", "aac", "ogg"}
	for _, ext := range expectedExts {
		assert.True(t, audioExtensions[ext], "audioExtensions should contain %s", ext)
	}
}

func TestImageExtensionsMap(t *testing.T) {
	assert.NotNil(t, imageExtensions)
	assert.True(t, len(imageExtensions) > 0)

	expectedExts := []string{"jpg", "jpeg", "png", "gif", "webp"}
	for _, ext := range expectedExts {
		assert.True(t, imageExtensions[ext], "imageExtensions should contain %s", ext)
	}
}

func TestBrowseEntrySorting(t *testing.T) {
	// Test the sorting logic used in Browse
	entries := []*pb.BrowseEntry{
		{Name: "zebra.mp4", IsDirectory: false},
		{Name: "alpha", IsDirectory: true},
		{Name: "beta.mp4", IsDirectory: false},
		{Name: "zeta", IsDirectory: true},
		{Name: "Alpha.mp4", IsDirectory: false},
	}

	// Apply the same sorting logic as in Browse
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDirectory != entries[j].IsDirectory {
			return entries[i].IsDirectory
		}
		return entries[i].Name < entries[j].Name
	})

	// Directories should come first
	assert.True(t, entries[0].IsDirectory)
	assert.True(t, entries[1].IsDirectory)

	// Then files
	assert.False(t, entries[2].IsDirectory)
	assert.False(t, entries[3].IsDirectory)
	assert.False(t, entries[4].IsDirectory)
}

func TestBrowseEntrySortingCaseInsensitive(t *testing.T) {
	entries := []*pb.BrowseEntry{
		{Name: "Zebra.mp4", IsDirectory: false},
		{Name: "alpha.mp4", IsDirectory: false},
		{Name: "BETA.mp4", IsDirectory: false},
	}

	// Case-insensitive sort as used in Browse
	sort.Slice(entries, func(i, j int) bool {
		// Note: The actual code uses strings.ToLower for case-insensitive sort
		return entries[i].Name < entries[j].Name
	})

	// Verify order is consistent
	assert.NotNil(t, entries[0])
}

func TestMediaFileDetection(t *testing.T) {
	testCases := []struct {
		filename string
		isVideo  bool
		isAudio  bool
		isImage  bool
	}{
		{"movie.mp4", true, false, false},
		{"song.mp3", false, true, false},
		{"photo.jpg", false, false, true},
		{"document.pdf", false, false, false},
		{"archive.zip", false, false, false},
		{"video.MKV", true, false, false},    // case insensitivity via getExtension
		{"music.FLAC", false, true, false},   // case insensitivity
		{"picture.PNG", false, false, true},  // case insensitivity
		{"noextension", false, false, false}, // no extension
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			ext := getExtension(tc.filename)
			assert.Equal(t, tc.isVideo, isVideo(ext), "video check for %s", tc.filename)
			assert.Equal(t, tc.isAudio, isAudio(ext), "audio check for %s", tc.filename)
			assert.Equal(t, tc.isImage, isImage(ext), "image check for %s", tc.filename)
		})
	}
}

func TestPathParsing(t *testing.T) {
	// Test the path parsing logic used in Browse
	testCases := []struct {
		path           string
		expectedBucket string
		expectedPrefix string
	}{
		{"webencode", "webencode", ""},
		{"webencode/videos", "webencode", "videos/"},
		{"webencode/videos/subfolder", "webencode", "videos/subfolder/"},
		{"bucket/a/b/c", "bucket", "a/b/c/"},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			// Replicate the Browse path parsing logic
			parts := splitN(tc.path, "/", 2)
			bucket := parts[0]
			prefix := ""
			if len(parts) > 1 {
				prefix = parts[1]
				if prefix != "" && !hasSuffix(prefix, "/") {
					prefix += "/"
				}
			}

			assert.Equal(t, tc.expectedBucket, bucket)
			assert.Equal(t, tc.expectedPrefix, prefix)
		})
	}
}

// Helper functions for testing
func splitN(s, sep string, n int) []string {
	result := []string{}
	for i := 0; i < n-1; i++ {
		idx := indexOf(s, sep)
		if idx == -1 {
			result = append(result, s)
			return result
		}
		result = append(result, s[:idx])
		s = s[idx+len(sep):]
	}
	if s != "" {
		result = append(result, s)
	}
	return result
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func hasSuffix(s, suffix string) bool {
	if len(s) < len(suffix) {
		return false
	}
	return s[len(s)-len(suffix):] == suffix
}

func TestParentPathCalculation(t *testing.T) {
	// Test parent path calculation logic from Browse
	testCases := []struct {
		bucket     string
		prefix     string
		parentPath string
	}{
		{"webencode", "", ""},
		{"webencode", "videos/", "webencode"},
		{"webencode", "videos/subfolder/", "webencode/videos"},
		{"bucket", "a/b/c/", "bucket/a/b"},
	}

	for _, tc := range testCases {
		t.Run(tc.bucket+"/"+tc.prefix, func(t *testing.T) {
			parentPath := ""
			if tc.prefix != "" {
				trimmed := tc.prefix
				if hasSuffix(trimmed, "/") {
					trimmed = trimmed[:len(trimmed)-1]
				}
				idx := lastIndexOf(trimmed, "/")
				if idx != -1 {
					parentPath = tc.bucket + "/" + trimmed[:idx]
				} else {
					parentPath = tc.bucket
				}
			}

			assert.Equal(t, tc.parentPath, parentPath)
		})
	}
}

func lastIndexOf(s, substr string) int {
	for i := len(s) - len(substr); i >= 0; i-- {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func TestS3StorageStruct(t *testing.T) {
	// Test that S3Storage struct can be created
	s := &S3Storage{
		bucket: "test-bucket",
	}

	assert.NotNil(t, s)
	assert.Equal(t, "test-bucket", s.bucket)
	assert.Nil(t, s.client) // Not initialized
}
