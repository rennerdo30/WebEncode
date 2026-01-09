package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/metadata"
)

// Mock Streams
type MockUploadServer struct {
	mock.Mock
	pb.StorageService_UploadServer
}

func (m *MockUploadServer) Recv() (*pb.FileChunk, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.FileChunk), args.Error(1)
}

func (m *MockUploadServer) SendAndClose(summary *pb.UploadSummary) error {
	args := m.Called(summary)
	return args.Error(0)
}

type MockDownloadServer struct {
	mock.Mock
	pb.StorageService_DownloadServer
}

func (m *MockDownloadServer) Send(chunk *pb.FileChunk) error {
	args := m.Called(chunk)
	return args.Error(0)
}

func (m *MockDownloadServer) SetHeader(md metadata.MD) error  { return nil }
func (m *MockDownloadServer) SendHeader(md metadata.MD) error { return nil }
func (m *MockDownloadServer) SetTrailer(md metadata.MD)       {}
func (m *MockDownloadServer) Context() context.Context        { return context.Background() }

func TestFilesystemStorage_Upload(t *testing.T) {
	// Setup temp dir
	tmpDir, err := os.MkdirTemp("", "storage-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	os.Setenv("STORAGE_FS_BASE_DIR", tmpDir)
	storage := NewFilesystemStorage()

	mockStream := new(MockUploadServer)

	// Sequences for Recv
	// 1. Metadata
	mockStream.On("Recv").Return(&pb.FileChunk{
		Content: &pb.FileChunk_Metadata{
			Metadata: &pb.FileMetadata{
				Bucket: "test-bucket",
				Path:   "test-file.txt",
			},
		},
	}, nil).Once()

	// 2. Data
	mockStream.On("Recv").Return(&pb.FileChunk{
		Content: &pb.FileChunk_Data{
			Data: []byte("hello world"),
		},
	}, nil).Once()

	// 3. EOF
	mockStream.On("Recv").Return(nil, io.EOF).Once()

	// Expect SendAndClose
	mockStream.On("SendAndClose", mock.MatchedBy(func(s *pb.UploadSummary) bool {
		return s.Size == 11
	})).Return(nil)

	err = storage.Upload(mockStream)
	assert.NoError(t, err)

	// Verify file exists
	content, err := os.ReadFile(filepath.Join(tmpDir, "test-bucket", "test-file.txt"))
	assert.NoError(t, err)
	assert.Equal(t, "hello world", string(content))
}

func TestFilesystemStorage_Download(t *testing.T) {
	// Setup temp dir and file
	tmpDir, err := os.MkdirTemp("", "storage-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	os.MkdirAll(filepath.Join(tmpDir, "test-bucket"), 0755)
	err = os.WriteFile(filepath.Join(tmpDir, "test-bucket", "test-file.txt"), []byte("hello world"), 0644)
	assert.NoError(t, err)

	os.Setenv("STORAGE_FS_BASE_DIR", tmpDir)
	storage := NewFilesystemStorage()

	mockStream := new(MockDownloadServer)

	// Expect metadata send
	mockStream.On("Send", mock.MatchedBy(func(c *pb.FileChunk) bool {
		return c.GetMetadata() != nil && c.GetMetadata().Size == 11
	})).Return(nil).Once()

	// Expect data send
	mockStream.On("Send", mock.MatchedBy(func(c *pb.FileChunk) bool {
		return c.GetData() != nil && string(c.GetData()) == "hello world"
	})).Return(nil).Once()

	err = storage.Download(&pb.FileRequest{
		Bucket: "test-bucket",
		Path:   "test-file.txt",
	}, mockStream)

	assert.NoError(t, err)
	mockStream.AssertExpectations(t)
}

func TestFilesystemStorage_Delete(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	os.MkdirAll(filepath.Join(tmpDir, "test-bucket"), 0755)
	filePath := filepath.Join(tmpDir, "test-bucket", "delete-me.txt")
	os.WriteFile(filePath, []byte("bye"), 0644)

	os.Setenv("STORAGE_FS_BASE_DIR", tmpDir)
	storage := NewFilesystemStorage()

	_, err = storage.Delete(context.Background(), &pb.FileRequest{
		Bucket: "test-bucket",
		Path:   "delete-me.txt",
	})
	assert.NoError(t, err)

	_, err = os.Stat(filePath)
	assert.True(t, os.IsNotExist(err))
}

func TestFilesystemStorage_ListObjects(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	os.MkdirAll(filepath.Join(tmpDir, "test-bucket", "subdir"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "test-bucket", "file1.txt"), []byte("1"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "test-bucket", "subdir", "file2.txt"), []byte("2"), 0644)

	os.Setenv("STORAGE_FS_BASE_DIR", tmpDir)
	storage := NewFilesystemStorage()

	resp, err := storage.ListObjects(context.Background(), &pb.ListObjectsRequest{
		Bucket: "test-bucket",
	})
	assert.NoError(t, err)
	assert.Len(t, resp.Objects, 2)
}

func TestFilesystemStorage_GetURL(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	os.Setenv("STORAGE_FS_BASE_DIR", tmpDir)
	storage := NewFilesystemStorage()

	resp, err := storage.GetURL(context.Background(), &pb.SignedUrlRequest{
		Bucket:        "test-bucket",
		ObjectKey:     "test-file.txt",
		ExpirySeconds: 3600,
	})
	assert.NoError(t, err)
	assert.Contains(t, resp.Url, "file://")
	assert.Contains(t, resp.Url, "test-bucket")
	assert.Contains(t, resp.Url, "test-file.txt")
	assert.True(t, resp.ExpiresAt > 0)
}

func TestFilesystemStorage_GetUploadURL(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	os.Setenv("STORAGE_FS_BASE_DIR", tmpDir)
	storage := NewFilesystemStorage()

	resp, err := storage.GetUploadURL(context.Background(), &pb.SignedUrlRequest{
		Bucket:        "uploads",
		ObjectKey:     "new-file.mp4",
		ExpirySeconds: 7200,
	})
	assert.NoError(t, err)
	assert.Contains(t, resp.Url, "file://")
	assert.Contains(t, resp.Url, "uploads")
	assert.Contains(t, resp.Url, "new-file.mp4")
}

func TestFilesystemStorage_GetObjectMetadata(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	os.MkdirAll(filepath.Join(tmpDir, "test-bucket"), 0755)
	testContent := []byte("test content for metadata")
	os.WriteFile(filepath.Join(tmpDir, "test-bucket", "meta-file.txt"), testContent, 0644)

	os.Setenv("STORAGE_FS_BASE_DIR", tmpDir)
	storage := NewFilesystemStorage()

	meta, err := storage.GetObjectMetadata(context.Background(), &pb.FileRequest{
		Bucket: "test-bucket",
		Path:   "meta-file.txt",
	})
	assert.NoError(t, err)
	assert.Equal(t, "meta-file.txt", meta.Key)
	assert.Equal(t, int64(len(testContent)), meta.Size)
	assert.Equal(t, "application/octet-stream", meta.ContentType)
	assert.True(t, meta.LastModified > 0)
}

func TestFilesystemStorage_GetObjectMetadata_NotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	os.Setenv("STORAGE_FS_BASE_DIR", tmpDir)
	storage := NewFilesystemStorage()

	_, err = storage.GetObjectMetadata(context.Background(), &pb.FileRequest{
		Bucket: "test-bucket",
		Path:   "nonexistent.txt",
	})
	assert.Error(t, err)
}

func TestFilesystemStorage_GetCapabilities(t *testing.T) {
	storage := NewFilesystemStorage()

	caps, err := storage.GetCapabilities(context.Background(), &pb.Empty{})
	assert.NoError(t, err)
	assert.True(t, caps.SupportsBrowse)
	assert.True(t, caps.SupportsUpload)
	assert.False(t, caps.SupportsSignedUrls)
	assert.True(t, caps.SupportsStreaming)
	assert.Equal(t, "filesystem", caps.StorageType)
}

func TestFilesystemStorage_BrowseRoots(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	os.Setenv("STORAGE_FS_BASE_DIR", tmpDir)
	os.Setenv("STORAGE_FS_BROWSE_ROOTS", "")
	storage := NewFilesystemStorage()

	resp, err := storage.BrowseRoots(context.Background(), &pb.Empty{})
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Roots)

	// Should contain the base dir
	found := false
	for _, root := range resp.Roots {
		if root.Path == tmpDir {
			found = true
			assert.True(t, root.IsDirectory)
			break
		}
	}
	assert.True(t, found, "base dir should be in roots")
}

func TestFilesystemStorage_Browse(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test structure
	os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "video.mp4"), []byte("video"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "audio.mp3"), []byte("audio"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "document.txt"), []byte("doc"), 0644)
	os.WriteFile(filepath.Join(tmpDir, ".hidden"), []byte("hidden"), 0644)

	os.Setenv("STORAGE_FS_BASE_DIR", tmpDir)
	storage := NewFilesystemStorage()

	t.Run("list all files", func(t *testing.T) {
		resp, err := storage.Browse(context.Background(), &pb.BrowseRequest{
			Path: tmpDir,
		})
		assert.NoError(t, err)
		assert.Equal(t, tmpDir, resp.CurrentPath)
		// Should have subdir, video.mp4, audio.mp3, document.txt (no hidden)
		assert.Len(t, resp.Entries, 4)
	})

	t.Run("show hidden files", func(t *testing.T) {
		resp, err := storage.Browse(context.Background(), &pb.BrowseRequest{
			Path:       tmpDir,
			ShowHidden: true,
		})
		assert.NoError(t, err)
		assert.Len(t, resp.Entries, 5) // includes .hidden
	})

	t.Run("media only filter", func(t *testing.T) {
		resp, err := storage.Browse(context.Background(), &pb.BrowseRequest{
			Path:      tmpDir,
			MediaOnly: true,
		})
		assert.NoError(t, err)
		// Should have subdir (directories always included), video.mp4, audio.mp3
		assert.Len(t, resp.Entries, 3)
	})

	t.Run("search filter", func(t *testing.T) {
		resp, err := storage.Browse(context.Background(), &pb.BrowseRequest{
			Path:        tmpDir,
			SearchQuery: "video",
		})
		assert.NoError(t, err)
		assert.Len(t, resp.Entries, 1)
		assert.Equal(t, "video.mp4", resp.Entries[0].Name)
	})

	t.Run("nonexistent path", func(t *testing.T) {
		_, err := storage.Browse(context.Background(), &pb.BrowseRequest{
			Path: "/nonexistent/path/12345",
		})
		assert.Error(t, err)
	})

	t.Run("file path instead of directory", func(t *testing.T) {
		_, err := storage.Browse(context.Background(), &pb.BrowseRequest{
			Path: filepath.Join(tmpDir, "video.mp4"),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a directory")
	})
}

func TestFilesystemStorage_ListObjects_WithMaxKeys(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	os.MkdirAll(filepath.Join(tmpDir, "test-bucket"), 0755)
	for i := 0; i < 10; i++ {
		os.WriteFile(filepath.Join(tmpDir, "test-bucket", "file"+string(rune('a'+i))+".txt"), []byte("x"), 0644)
	}

	os.Setenv("STORAGE_FS_BASE_DIR", tmpDir)
	storage := NewFilesystemStorage()

	resp, err := storage.ListObjects(context.Background(), &pb.ListObjectsRequest{
		Bucket:  "test-bucket",
		MaxKeys: 5,
	})
	assert.NoError(t, err)
	assert.Len(t, resp.Objects, 5)
	assert.True(t, resp.IsTruncated)
}

func TestFilesystemStorage_Delete_NonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	os.Setenv("STORAGE_FS_BASE_DIR", tmpDir)
	storage := NewFilesystemStorage()

	// Deleting non-existent file should not error
	_, err = storage.Delete(context.Background(), &pb.FileRequest{
		Bucket: "test-bucket",
		Path:   "nonexistent.txt",
	})
	assert.NoError(t, err)
}
