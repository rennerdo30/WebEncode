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
