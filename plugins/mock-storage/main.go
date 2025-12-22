package main

import (
	"context"
	"io"
	"log"

	"github.com/hashicorp/go-plugin"
	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/rennerdo30/webencode/pkg/pluginsdk"
)

type MockStorage struct {
	pb.UnimplementedStorageServiceServer
}

func (s *MockStorage) Upload(stream pb.StorageService_UploadServer) error {
	log.Println("MockStorage: Upload started")
	var size int64
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.UploadSummary{
				Url:  "mock://bucket/file",
				Size: size,
				Etag: "mock-etag",
			})
		}
		if err != nil {
			return err
		}
		if data := chunk.GetData(); data != nil {
			size += int64(len(data))
		}
	}
}

func (s *MockStorage) Download(req *pb.FileRequest, stream pb.StorageService_DownloadServer) error {
	log.Println("MockStorage: Download requested", req.Path)
	return nil
}

func (s *MockStorage) Delete(ctx context.Context, req *pb.FileRequest) (*pb.Empty, error) {
	log.Println("MockStorage: Delete requested", req.Path)
	return &pb.Empty{}, nil
}

func (s *MockStorage) GetURL(ctx context.Context, req *pb.SignedUrlRequest) (*pb.SignedUrlResponse, error) {
	return &pb.SignedUrlResponse{
		Url:       "mock://" + req.ObjectKey,
		ExpiresAt: 9999999999,
	}, nil
}

func (s *MockStorage) GetUploadURL(ctx context.Context, req *pb.SignedUrlRequest) (*pb.SignedUrlResponse, error) {
	return &pb.SignedUrlResponse{
		Url:       "mock://upload/" + req.ObjectKey,
		ExpiresAt: 9999999999,
	}, nil
}

func (s *MockStorage) ListObjects(ctx context.Context, req *pb.ListObjectsRequest) (*pb.ListObjectsResponse, error) {
	return &pb.ListObjectsResponse{
		Objects:     []*pb.ObjectInfo{},
		IsTruncated: false,
	}, nil
}

func (s *MockStorage) GetObjectMetadata(ctx context.Context, req *pb.FileRequest) (*pb.ObjectMetadata, error) {
	return &pb.ObjectMetadata{
		Key:         req.Path,
		Size:        0,
		ContentType: "application/octet-stream",
	}, nil
}

func (s *MockStorage) GetCapabilities(ctx context.Context, req *pb.Empty) (*pb.StorageCapabilities, error) {
	return &pb.StorageCapabilities{
		SupportsBrowse:     true,
		SupportsUpload:     true,
		SupportsSignedUrls: true,
		SupportsStreaming:  false,
		StorageType:        "mock",
	}, nil
}

func (s *MockStorage) BrowseRoots(ctx context.Context, req *pb.Empty) (*pb.BrowseRootsResponse, error) {
	return &pb.BrowseRootsResponse{
		Roots: []*pb.BrowseEntry{
			{Name: "mock-bucket-1", Path: "mock-bucket-1", IsDirectory: true},
			{Name: "mock-bucket-2", Path: "mock-bucket-2", IsDirectory: true},
		},
	}, nil
}

func (s *MockStorage) Browse(ctx context.Context, req *pb.BrowseRequest) (*pb.BrowseResponse, error) {
	return &pb.BrowseResponse{
		CurrentPath: req.Path,
		ParentPath:  "", // simplified
		Entries: []*pb.BrowseEntry{
			{Name: "fake-video.mp4", Path: req.Path + "/fake-video.mp4", IsDirectory: false, IsVideo: true, Size: 1024 * 1024 * 50},
			{Name: "subdir", Path: req.Path + "/subdir", IsDirectory: true},
		},
	}, nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: pluginsdk.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"webencode": &pluginsdk.Plugin{
				StorageImpl: &MockStorage{},
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
