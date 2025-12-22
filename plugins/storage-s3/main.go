package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/go-plugin"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/rennerdo30/webencode/pkg/pluginsdk"
)

type S3Storage struct {
	pb.UnimplementedStorageServiceServer
	client *minio.Client
	bucket string // In real app, might be dynamic or configured via env
}

func NewS3Storage() *S3Storage {
	endpoint := os.Getenv("S3_ENDPOINT")
	if endpoint == "" {
		endpoint = "seaweedfs-filer:8333"
	}
	accessKeyID := os.Getenv("S3_ACCESS_KEY")
	secretAccessKey := os.Getenv("S3_SECRET_KEY")
	bucket := os.Getenv("S3_BUCKET")
	if bucket == "" {
		bucket = "webencode"
	}

	opts := &minio.Options{
		Secure: false,
	}
	if accessKeyID != "" && secretAccessKey != "" {
		opts.Creds = credentials.NewStaticV4(accessKeyID, secretAccessKey, "")
	}

	client, err := minio.New(endpoint, opts)
	if err != nil {
		log.Fatalln(err)
	}

	// Ensure bucket exists
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucket)
	if err == nil && !exists {
		client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
	}

	return &S3Storage{client: client, bucket: bucket}
}

func (s *S3Storage) Upload(stream pb.StorageService_UploadServer) error {
	// Create a temporary file to buffer the upload
	tempFile, err := os.CreateTemp("", "webencode-upload-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}()

	var metadata *pb.FileMetadata
	var size int64

	// Read stream to temp file
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if meta := chunk.GetMetadata(); meta != nil {
			metadata = meta
		}

		if data := chunk.GetData(); data != nil {
			n, err := tempFile.Write(data)
			if err != nil {
				return fmt.Errorf("failed to write to temp file: %w", err)
			}
			size += int64(n)
		}
	}

	// Seek to beginning
	if _, err := tempFile.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek temp file: %w", err)
	}

	// Determine bucket and path
	bucket := s.bucket
	path := fmt.Sprintf("upload_%d", time.Now().UnixNano())
	contentType := "application/octet-stream"

	if metadata != nil {
		if metadata.Bucket != "" {
			bucket = metadata.Bucket
		}
		if metadata.Path != "" {
			path = metadata.Path
		}
		if metadata.ContentType != "" {
			contentType = metadata.ContentType
		}
	}

	// Upload to S3
	info, err := s.client.PutObject(context.Background(), bucket, path, tempFile, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to put object to s3: %w", err)
	}

	return stream.SendAndClose(&pb.UploadSummary{
		Url:  fmt.Sprintf("s3://%s/%s", bucket, path),
		Size: info.Size,
		Etag: info.ETag,
	})
}

func (s *S3Storage) Download(req *pb.FileRequest, stream pb.StorageService_DownloadServer) error {
	obj, err := s.client.GetObject(context.Background(), req.Bucket, req.Path, minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	defer obj.Close()

	buf := make([]byte, 32*1024)
	for {
		n, err := obj.Read(buf)
		if n > 0 {
			if err := stream.Send(&pb.FileChunk{Content: &pb.FileChunk_Data{Data: buf[:n]}}); err != nil {
				return err
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *S3Storage) GetURL(ctx context.Context, req *pb.SignedUrlRequest) (*pb.SignedUrlResponse, error) {
	// Generate presigned GET URL
	expiry := time.Duration(req.ExpirySeconds) * time.Second
	if expiry == 0 {
		expiry = time.Hour
	}

	bucket := req.Bucket
	if bucket == "" {
		bucket = s.bucket
	}

	url, err := s.client.PresignedGetObject(ctx, bucket, req.ObjectKey, expiry, nil)
	if err != nil {
		return nil, err
	}

	return &pb.SignedUrlResponse{
		Url:       url.String(),
		ExpiresAt: time.Now().Add(expiry).Unix(),
	}, nil
}

func (s *S3Storage) GetUploadURL(ctx context.Context, req *pb.SignedUrlRequest) (*pb.SignedUrlResponse, error) {
	// Generate presigned PUT URL
	expiry := time.Duration(req.ExpirySeconds) * time.Second
	if expiry == 0 {
		expiry = time.Hour
	}

	bucket := req.Bucket
	if bucket == "" {
		bucket = s.bucket
	}

	// Set content type if provided, otherwise default to binary
	contentType := "application/octet-stream"
	if req.ContentType != "" {
		contentType = req.ContentType
	}

	// Presign the PUT request
	// Note: PresignedPutObject strictly enforces the object name and expiry
	url, err := s.client.PresignedPutObject(ctx, bucket, req.ObjectKey, expiry)
	if err != nil {
		return nil, err
	}

	// Manually add Content-Type to the query params or headers if needed by the client?
	// MinIO/S3 signed URLs for PUT usually don't bake in headers unless you use a different method.
	// PresignedPutObject generates a URL where you can just PUT the data.
	// However, if we want to enforce Content-Type, we should ensure the client sends it.

	headers := map[string]string{
		"Content-Type": contentType,
	}

	return &pb.SignedUrlResponse{
		Url:       url.String(),
		Headers:   headers,
		ExpiresAt: time.Now().Add(expiry).Unix(),
	}, nil
}

// Video file extensions
var videoExtensions = map[string]bool{
	"mp4": true, "mkv": true, "avi": true, "mov": true, "wmv": true,
	"flv": true, "webm": true, "m4v": true, "mpeg": true, "mpg": true,
	"3gp": true, "ts": true, "mts": true, "m2ts": true, "vob": true,
}

// Audio file extensions
var audioExtensions = map[string]bool{
	"mp3": true, "wav": true, "flac": true, "aac": true, "ogg": true,
	"wma": true, "m4a": true, "opus": true,
}

// Image file extensions
var imageExtensions = map[string]bool{
	"jpg": true, "jpeg": true, "png": true, "gif": true, "bmp": true,
	"webp": true, "svg": true, "tiff": true, "tif": true,
}

func getExtension(name string) string {
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '.' {
			return strings.ToLower(name[i+1:])
		}
	}
	return ""
}

func isVideo(ext string) bool {
	return videoExtensions[ext]
}

func isAudio(ext string) bool {
	return audioExtensions[ext]
}

func isImage(ext string) bool {
	return imageExtensions[ext]
}

// GetCapabilities returns the storage plugin capabilities
func (s *S3Storage) GetCapabilities(ctx context.Context, req *pb.Empty) (*pb.StorageCapabilities, error) {
	return &pb.StorageCapabilities{
		SupportsBrowse:     true,
		SupportsUpload:     true,
		SupportsSignedUrls: true,
		SupportsStreaming:  true,
		StorageType:        "s3",
	}, nil
}

// BrowseRoots returns available buckets as root directories
func (s *S3Storage) BrowseRoots(ctx context.Context, req *pb.Empty) (*pb.BrowseRootsResponse, error) {
	roots := []*pb.BrowseEntry{}

	// List all accessible buckets
	buckets, err := s.client.ListBuckets(ctx)
	if err != nil {
		// If listing fails, return the configured default bucket
		roots = append(roots, &pb.BrowseEntry{
			Name:        s.bucket,
			Path:        s.bucket,
			IsDirectory: true,
		})
		return &pb.BrowseRootsResponse{Roots: roots}, nil
	}

	for _, bucket := range buckets {
		roots = append(roots, &pb.BrowseEntry{
			Name:        bucket.Name,
			Path:        bucket.Name,
			IsDirectory: true,
			ModTime:     bucket.CreationDate.Unix(),
		})
	}

	return &pb.BrowseRootsResponse{Roots: roots}, nil
}

// Browse lists contents of a bucket/prefix
func (s *S3Storage) Browse(ctx context.Context, req *pb.BrowseRequest) (*pb.BrowseResponse, error) {
	path := req.Path
	if path == "" {
		path = s.bucket
	}

	// Parse bucket and prefix from path
	// Path format: "bucket" or "bucket/prefix/to/directory"
	parts := strings.SplitN(path, "/", 2)
	bucket := parts[0]
	prefix := ""
	if len(parts) > 1 {
		prefix = parts[1]
		if !strings.HasSuffix(prefix, "/") && prefix != "" {
			prefix += "/"
		}
	}

	entries := []*pb.BrowseEntry{}
	searchQuery := strings.ToLower(req.SearchQuery)

	// List objects with delimiter to get "virtual directories"
	opts := minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: false, // Use delimiter to get virtual directories
	}

	objectCh := s.client.ListObjects(ctx, bucket, opts)

	dirs := make(map[string]bool) // Track directories we've added

	for object := range objectCh {
		if object.Err != nil {
			continue
		}

		// Get the relative name (after the prefix)
		name := strings.TrimPrefix(object.Key, prefix)
		if name == "" {
			continue
		}

		// Check if this is a "directory" (ends with / or has more path components)
		isDir := false
		if strings.HasSuffix(name, "/") {
			name = strings.TrimSuffix(name, "/")
			isDir = true
		} else if idx := strings.Index(name, "/"); idx != -1 {
			// This is a file in a subdirectory - represent the directory
			dirName := name[:idx]
			if !dirs[dirName] {
				dirs[dirName] = true
				fullPath := bucket
				if prefix != "" {
					fullPath += "/" + prefix + dirName
				} else {
					fullPath += "/" + dirName
				}

				// Skip hidden if not requested
				if !req.ShowHidden && strings.HasPrefix(dirName, ".") {
					continue
				}

				// Apply search filter
				if searchQuery != "" && !strings.Contains(strings.ToLower(dirName), searchQuery) {
					continue
				}

				entries = append(entries, &pb.BrowseEntry{
					Name:        dirName,
					Path:        fullPath,
					IsDirectory: true,
				})
			}
			continue
		}

		// Skip hidden files unless requested
		if !req.ShowHidden && strings.HasPrefix(name, ".") {
			continue
		}

		// Apply search filter
		if searchQuery != "" && !strings.Contains(strings.ToLower(name), searchQuery) {
			continue
		}

		ext := getExtension(name)
		isVideoFile := isVideo(ext)
		isAudioFile := isAudio(ext)
		isImageFile := isImage(ext)

		// Filter to media files only if requested
		if req.MediaOnly && !isDir && !isVideoFile && !isAudioFile && !isImageFile {
			continue
		}

		fullPath := bucket + "/" + object.Key
		if isDir {
			fullPath = strings.TrimSuffix(fullPath, "/")
		}

		entries = append(entries, &pb.BrowseEntry{
			Name:        name,
			Path:        fullPath,
			IsDirectory: isDir,
			Size:        object.Size,
			ModTime:     object.LastModified.Unix(),
			Extension:   ext,
			IsVideo:     isVideoFile,
			IsAudio:     isAudioFile,
			IsImage:     isImageFile,
		})
	}

	// Sort: directories first, then by name
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDirectory != entries[j].IsDirectory {
			return entries[i].IsDirectory
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})

	// Calculate parent path
	parentPath := ""
	if prefix != "" {
		// Go up one directory level
		trimmed := strings.TrimSuffix(prefix, "/")
		if idx := strings.LastIndex(trimmed, "/"); idx != -1 {
			parentPath = bucket + "/" + trimmed[:idx]
		} else {
			parentPath = bucket
		}
	}

	return &pb.BrowseResponse{
		CurrentPath: path,
		ParentPath:  parentPath,
		Entries:     entries,
	}, nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: pluginsdk.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"webencode": &pluginsdk.Plugin{
				StorageImpl: NewS3Storage(),
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
