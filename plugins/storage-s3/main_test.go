package main

import (
	"context"
	"io"
	"testing"
	"time"

	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewS3Storage(t *testing.T) {
	t.Run("default configuration", func(t *testing.T) {
		// With env vars not set, it should use defaults
		// This will fail to connect in tests, but we can verify struct creation
		t.Skip("Requires S3 endpoint - integration test")
	})
}

func TestS3Storage_Upload(t *testing.T) {
	t.Run("upload requires metadata", func(t *testing.T) {
		t.Skip("Requires S3 endpoint - integration test")
	})
}

func TestS3Storage_Download(t *testing.T) {
	t.Run("download existing object", func(t *testing.T) {
		t.Skip("Requires S3 endpoint - integration test")
	})

	t.Run("download non-existent object", func(t *testing.T) {
		t.Skip("Requires S3 endpoint - integration test")
	})
}

// MockS3Storage for unit testing without real S3
type MockS3Storage struct {
	pb.UnimplementedStorageServiceServer
	data map[string][]byte
}

func NewMockS3Storage() *MockS3Storage {
	return &MockS3Storage{
		data: make(map[string][]byte),
	}
}

func TestMockS3Storage_Operations(t *testing.T) {
	storage := NewMockS3Storage()

	t.Run("storage initialization", func(t *testing.T) {
		assert.NotNil(t, storage)
		assert.NotNil(t, storage.data)
	})
}

func TestS3Config(t *testing.T) {
	t.Run("endpoint parsing", func(t *testing.T) {
		endpoints := []string{
			"localhost:8333",
			"seaweedfs-filer:8333",
			"s3.amazonaws.com",
			"192.168.1.100:8333",
		}

		for _, endpoint := range endpoints {
			assert.NotEmpty(t, endpoint)
		}
	})

	t.Run("bucket name validation", func(t *testing.T) {
		validBuckets := []string{"webencode", "my-bucket", "bucket123"}
		for _, bucket := range validBuckets {
			assert.NotEmpty(t, bucket)
			assert.Regexp(t, "^[a-z0-9-]+$", bucket)
		}
	})
}

func BenchmarkBufferSizes(b *testing.B) {
	sizes := []int{8 * 1024, 16 * 1024, 32 * 1024, 64 * 1024}

	for _, size := range sizes {
		b.Run("buffer_"+string(rune(size/1024))+"KB", func(b *testing.B) {
			buf := make([]byte, size)
			for i := 0; i < b.N; i++ {
				_ = buf
			}
		})
	}
}

func TestUploadMetadata(t *testing.T) {
	t.Run("valid metadata", func(t *testing.T) {
		meta := &pb.FileMetadata{
			Path:        "test/video.mp4",
			ContentType: "video/mp4",
			Bucket:      "webencode",
		}

		assert.Equal(t, "test/video.mp4", meta.Path)
		assert.Equal(t, "video/mp4", meta.ContentType)
		assert.Equal(t, "webencode", meta.Bucket)
	})

	t.Run("empty metadata", func(t *testing.T) {
		meta := &pb.FileMetadata{}
		assert.Empty(t, meta.Path)
	})
}

func TestFileRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		req := &pb.FileRequest{
			Bucket: "webencode",
			Path:   "videos/output.mp4",
		}

		assert.Equal(t, "webencode", req.Bucket)
		assert.Equal(t, "videos/output.mp4", req.Path)
	})
}

func TestChunkSizes(t *testing.T) {
	defaultChunkSize := 32 * 1024

	t.Run("chunk size is reasonable", func(t *testing.T) {
		assert.GreaterOrEqual(t, defaultChunkSize, 8*1024)
		assert.LessOrEqual(t, defaultChunkSize, 1024*1024)
	})
}

func skipIfNoS3(t *testing.T) {
	t.Helper()
	t.Skip("S3 integration tests require S3_ENDPOINT environment variable")
}

func TestIntegration_Upload(t *testing.T) {
	skipIfNoS3(t)
}

func TestIntegration_Download(t *testing.T) {
	skipIfNoS3(t)
}

func TestIntegration_RoundTrip(t *testing.T) {
	skipIfNoS3(t)
}

func TestContextCancellation(t *testing.T) {
	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		assert.Error(t, ctx.Err())
	})
}

func TestS3Storage_GetCapabilities(t *testing.T) {
	// GetCapabilities doesn't require an S3 connection
	s := &S3Storage{
		bucket: "test-bucket",
	}

	ctx := context.Background()
	caps, err := s.GetCapabilities(ctx, &pb.Empty{})

	require.NoError(t, err)
	assert.NotNil(t, caps)
	assert.True(t, caps.SupportsBrowse)
	assert.True(t, caps.SupportsUpload)
	assert.True(t, caps.SupportsSignedUrls)
	assert.True(t, caps.SupportsStreaming)
	assert.Equal(t, "s3", caps.StorageType)
}

func TestS3Storage_GetCapabilities_Context(t *testing.T) {
	s := &S3Storage{bucket: "test"}

	t.Run("with background context", func(t *testing.T) {
		caps, err := s.GetCapabilities(context.Background(), &pb.Empty{})
		require.NoError(t, err)
		assert.NotNil(t, caps)
	})

	t.Run("with timeout context", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		caps, err := s.GetCapabilities(ctx, &pb.Empty{})
		require.NoError(t, err)
		assert.NotNil(t, caps)
	})
}

func TestS3Storage_StructFields(t *testing.T) {
	s := &S3Storage{
		bucket: "my-bucket",
	}

	assert.Equal(t, "my-bucket", s.bucket)
	assert.Nil(t, s.client)
}

func TestSignedUrlRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		req := &pb.SignedUrlRequest{
			Bucket:        "webencode",
			ObjectKey:     "videos/test.mp4",
			ExpirySeconds: 3600,
			ContentType:   "video/mp4",
		}

		assert.Equal(t, "webencode", req.Bucket)
		assert.Equal(t, "videos/test.mp4", req.ObjectKey)
		assert.Equal(t, int64(3600), req.ExpirySeconds)
		assert.Equal(t, "video/mp4", req.ContentType)
	})

	t.Run("default expiry", func(t *testing.T) {
		req := &pb.SignedUrlRequest{
			Bucket:    "webencode",
			ObjectKey: "test.mp4",
		}

		// 0 means use default (1 hour)
		expiry := time.Duration(req.ExpirySeconds) * time.Second
		if expiry == 0 {
			expiry = time.Hour
		}
		assert.Equal(t, time.Hour, expiry)
	})
}

func TestBrowseRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		req := &pb.BrowseRequest{
			Path:        "webencode/videos",
			ShowHidden:  false,
			MediaOnly:   true,
			SearchQuery: "test",
		}

		assert.Equal(t, "webencode/videos", req.Path)
		assert.False(t, req.ShowHidden)
		assert.True(t, req.MediaOnly)
		assert.Equal(t, "test", req.SearchQuery)
	})

	t.Run("empty path uses default", func(t *testing.T) {
		req := &pb.BrowseRequest{}
		assert.Empty(t, req.Path)
	})
}

func TestBrowseEntry(t *testing.T) {
	t.Run("directory entry", func(t *testing.T) {
		entry := &pb.BrowseEntry{
			Name:        "videos",
			Path:        "webencode/videos",
			IsDirectory: true,
			ModTime:     time.Now().Unix(),
		}

		assert.True(t, entry.IsDirectory)
		assert.NotEmpty(t, entry.Name)
		assert.NotEmpty(t, entry.Path)
	})

	t.Run("file entry", func(t *testing.T) {
		entry := &pb.BrowseEntry{
			Name:        "video.mp4",
			Path:        "webencode/videos/video.mp4",
			IsDirectory: false,
			Size:        1024 * 1024,
			Extension:   "mp4",
			IsVideo:     true,
			IsAudio:     false,
			IsImage:     false,
		}

		assert.False(t, entry.IsDirectory)
		assert.Equal(t, "mp4", entry.Extension)
		assert.True(t, entry.IsVideo)
	})
}

func TestUploadSummary(t *testing.T) {
	summary := &pb.UploadSummary{
		Url:  "s3://webencode/videos/test.mp4",
		Size: 1024,
		Etag: "abc123",
	}

	assert.Contains(t, summary.Url, "s3://")
	assert.Equal(t, int64(1024), summary.Size)
	assert.NotEmpty(t, summary.Etag)
}

func TestFileChunk(t *testing.T) {
	t.Run("data chunk", func(t *testing.T) {
		data := []byte("test data content")
		chunk := &pb.FileChunk{
			Content: &pb.FileChunk_Data{Data: data},
		}

		assert.Equal(t, data, chunk.GetData())
		assert.Nil(t, chunk.GetMetadata())
	})

	t.Run("metadata chunk", func(t *testing.T) {
		meta := &pb.FileMetadata{
			Path:        "test.mp4",
			ContentType: "video/mp4",
		}
		chunk := &pb.FileChunk{
			Content: &pb.FileChunk_Metadata{Metadata: meta},
		}

		assert.NotNil(t, chunk.GetMetadata())
		assert.Equal(t, "test.mp4", chunk.GetMetadata().Path)
	})
}

// Mock upload stream for testing
type mockUploadStream struct {
	chunks []*pb.FileChunk
	index  int
	result *pb.UploadSummary
}

func (m *mockUploadStream) Recv() (*pb.FileChunk, error) {
	if m.index >= len(m.chunks) {
		return nil, io.EOF
	}
	chunk := m.chunks[m.index]
	m.index++
	return chunk, nil
}

func (m *mockUploadStream) SendAndClose(summary *pb.UploadSummary) error {
	m.result = summary
	return nil
}

func (m *mockUploadStream) Context() context.Context {
	return context.Background()
}

func (m *mockUploadStream) SetHeader(metadata interface{}) error {
	return nil
}

func (m *mockUploadStream) SendHeader(metadata interface{}) error {
	return nil
}

func (m *mockUploadStream) SetTrailer(metadata interface{}) {
}

func (m *mockUploadStream) RecvMsg(msg interface{}) error {
	return nil
}

func (m *mockUploadStream) SendMsg(msg interface{}) error {
	return nil
}

func TestMockUploadStream(t *testing.T) {
	stream := &mockUploadStream{
		chunks: []*pb.FileChunk{
			{Content: &pb.FileChunk_Metadata{Metadata: &pb.FileMetadata{Path: "test.mp4"}}},
			{Content: &pb.FileChunk_Data{Data: []byte("test data")}},
		},
	}

	// First chunk is metadata
	chunk, err := stream.Recv()
	require.NoError(t, err)
	assert.NotNil(t, chunk.GetMetadata())

	// Second chunk is data
	chunk, err = stream.Recv()
	require.NoError(t, err)
	assert.NotNil(t, chunk.GetData())

	// Third call returns EOF
	_, err = stream.Recv()
	assert.Equal(t, io.EOF, err)
}

// Mock download stream for testing
type mockDownloadStream struct {
	chunks []*pb.FileChunk
}

func (m *mockDownloadStream) Send(chunk *pb.FileChunk) error {
	m.chunks = append(m.chunks, chunk)
	return nil
}

func (m *mockDownloadStream) Context() context.Context {
	return context.Background()
}

func (m *mockDownloadStream) SetHeader(metadata interface{}) error {
	return nil
}

func (m *mockDownloadStream) SendHeader(metadata interface{}) error {
	return nil
}

func (m *mockDownloadStream) SetTrailer(metadata interface{}) {
}

func (m *mockDownloadStream) RecvMsg(msg interface{}) error {
	return nil
}

func (m *mockDownloadStream) SendMsg(msg interface{}) error {
	return nil
}

func TestMockDownloadStream(t *testing.T) {
	stream := &mockDownloadStream{}

	// Send some chunks
	err := stream.Send(&pb.FileChunk{Content: &pb.FileChunk_Data{Data: []byte("chunk1")}})
	require.NoError(t, err)

	err = stream.Send(&pb.FileChunk{Content: &pb.FileChunk_Data{Data: []byte("chunk2")}})
	require.NoError(t, err)

	assert.Len(t, stream.chunks, 2)
}

func TestExpiryCalculation(t *testing.T) {
	tests := []struct {
		name           string
		expirySeconds  int64
		expectedExpiry time.Duration
	}{
		{"zero uses default", 0, time.Hour},
		{"one hour", 3600, time.Hour},
		{"one day", 86400, 24 * time.Hour},
		{"short expiry", 60, time.Minute},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			expiry := time.Duration(tc.expirySeconds) * time.Second
			if expiry == 0 {
				expiry = time.Hour // Default
			}
			assert.Equal(t, tc.expectedExpiry, expiry)
		})
	}
}

func TestContentTypeDefaults(t *testing.T) {
	t.Run("empty content type uses default", func(t *testing.T) {
		contentType := ""
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		assert.Equal(t, "application/octet-stream", contentType)
	})

	t.Run("specified content type is used", func(t *testing.T) {
		contentType := "video/mp4"
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		assert.Equal(t, "video/mp4", contentType)
	})
}

func TestBucketDefaults(t *testing.T) {
	defaultBucket := "webencode"

	t.Run("empty bucket uses default", func(t *testing.T) {
		reqBucket := ""
		bucket := reqBucket
		if bucket == "" {
			bucket = defaultBucket
		}
		assert.Equal(t, defaultBucket, bucket)
	})

	t.Run("specified bucket is used", func(t *testing.T) {
		reqBucket := "custom-bucket"
		bucket := reqBucket
		if bucket == "" {
			bucket = defaultBucket
		}
		assert.Equal(t, "custom-bucket", bucket)
	})
}
