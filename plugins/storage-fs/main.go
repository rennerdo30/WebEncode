package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/go-plugin"
	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/rennerdo30/webencode/pkg/pluginsdk"
)

type FilesystemStorage struct {
	pb.UnimplementedStorageServiceServer
	logger  *logger.Logger
	baseDir string
}

func NewFilesystemStorage() *FilesystemStorage {
	baseDir := os.Getenv("STORAGE_FS_BASE_DIR")
	if baseDir == "" {
		baseDir = "/tmp/webencode-storage"
	}

	// Ensure base directory exists
	os.MkdirAll(baseDir, 0755)

	return &FilesystemStorage{
		logger:  logger.New("plugin-storage-fs"),
		baseDir: baseDir,
	}
}

func (s *FilesystemStorage) getPath(bucket, path string) string {
	return filepath.Join(s.baseDir, bucket, path)
}

func (s *FilesystemStorage) Upload(stream pb.StorageService_UploadServer) error {
	var file *os.File
	var size int64
	var filePath string

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			if file != nil {
				file.Close()
			}
			return stream.SendAndClose(&pb.UploadSummary{
				Url:  fmt.Sprintf("file://%s", filePath),
				Size: size,
				Etag: fmt.Sprintf("%d", time.Now().UnixNano()),
			})
		}
		if err != nil {
			if file != nil {
				file.Close()
				os.Remove(filePath)
			}
			return err
		}

		if metadata := chunk.GetMetadata(); metadata != nil {
			filePath = s.getPath(metadata.Bucket, metadata.Path)
			dir := filepath.Dir(filePath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
			file, err = os.Create(filePath)
			if err != nil {
				return err
			}
			s.logger.Info("Upload started", "path", filePath)
		}

		if data := chunk.GetData(); data != nil && file != nil {
			n, err := file.Write(data)
			if err != nil {
				file.Close()
				os.Remove(filePath)
				return err
			}
			size += int64(n)
		}
	}
}

func (s *FilesystemStorage) Download(req *pb.FileRequest, stream pb.StorageService_DownloadServer) error {
	filePath := s.getPath(req.Bucket, req.Path)
	s.logger.Info("Download requested", "path", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Send metadata first
	info, err := file.Stat()
	if err != nil {
		return err
	}

	err = stream.Send(&pb.FileChunk{
		Content: &pb.FileChunk_Metadata{
			Metadata: &pb.FileMetadata{
				Bucket:      req.Bucket,
				Path:        req.Path,
				Size:        info.Size(),
				ContentType: "application/octet-stream",
			},
		},
	})
	if err != nil {
		return err
	}

	// Stream file content
	buf := make([]byte, 64*1024) // 64KB chunks
	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		err = stream.Send(&pb.FileChunk{
			Content: &pb.FileChunk_Data{
				Data: buf[:n],
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *FilesystemStorage) Delete(ctx context.Context, req *pb.FileRequest) (*pb.Empty, error) {
	filePath := s.getPath(req.Bucket, req.Path)
	s.logger.Info("Delete requested", "path", filePath)

	err := os.Remove(filePath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return &pb.Empty{}, nil
}

func (s *FilesystemStorage) GetURL(ctx context.Context, req *pb.SignedUrlRequest) (*pb.SignedUrlResponse, error) {
	filePath := s.getPath(req.Bucket, req.ObjectKey)
	return &pb.SignedUrlResponse{
		Url:       fmt.Sprintf("file://%s", filePath),
		ExpiresAt: time.Now().Add(time.Duration(req.ExpirySeconds) * time.Second).Unix(),
	}, nil
}

func (s *FilesystemStorage) GetUploadURL(ctx context.Context, req *pb.SignedUrlRequest) (*pb.SignedUrlResponse, error) {
	filePath := s.getPath(req.Bucket, req.ObjectKey)
	return &pb.SignedUrlResponse{
		Url:       fmt.Sprintf("file://%s", filePath),
		ExpiresAt: time.Now().Add(time.Duration(req.ExpirySeconds) * time.Second).Unix(),
	}, nil
}

func (s *FilesystemStorage) ListObjects(ctx context.Context, req *pb.ListObjectsRequest) (*pb.ListObjectsResponse, error) {
	basePath := s.getPath(req.Bucket, req.Prefix)
	var objects []*pb.ObjectInfo

	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel(filepath.Join(s.baseDir, req.Bucket), path)
		relPath = strings.ReplaceAll(relPath, string(os.PathSeparator), "/")

		objects = append(objects, &pb.ObjectInfo{
			Key:          relPath,
			Size:         info.Size(),
			LastModified: info.ModTime().Unix(),
		})

		if req.MaxKeys > 0 && len(objects) >= int(req.MaxKeys) {
			return filepath.SkipAll
		}
		return nil
	})

	if err != nil && !os.IsNotExist(err) {
		s.logger.Error("ListObjects failed", "error", err)
	}

	return &pb.ListObjectsResponse{
		Objects:     objects,
		IsTruncated: req.MaxKeys > 0 && len(objects) >= int(req.MaxKeys),
	}, nil
}

func (s *FilesystemStorage) GetObjectMetadata(ctx context.Context, req *pb.FileRequest) (*pb.ObjectMetadata, error) {
	filePath := s.getPath(req.Bucket, req.Path)
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	return &pb.ObjectMetadata{
		Key:          req.Path,
		Size:         info.Size(),
		ContentType:  "application/octet-stream",
		LastModified: info.ModTime().Unix(),
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
	ext := filepath.Ext(name)
	if len(ext) > 0 {
		return strings.ToLower(ext[1:]) // Remove the leading dot
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
func (s *FilesystemStorage) GetCapabilities(ctx context.Context, req *pb.Empty) (*pb.StorageCapabilities, error) {
	return &pb.StorageCapabilities{
		SupportsBrowse:     true,
		SupportsUpload:     true,
		SupportsSignedUrls: false, // Local filesystem doesn't need signed URLs
		SupportsStreaming:  true,
		StorageType:        "filesystem",
	}, nil
}

// BrowseRoots returns available root directories for browsing
func (s *FilesystemStorage) BrowseRoots(ctx context.Context, req *pb.Empty) (*pb.BrowseRootsResponse, error) {
	roots := []*pb.BrowseEntry{}

	// Add the configured base directory
	if info, err := os.Stat(s.baseDir); err == nil && info.IsDir() {
		roots = append(roots, &pb.BrowseEntry{
			Name:        filepath.Base(s.baseDir),
			Path:        s.baseDir,
			IsDirectory: true,
			ModTime:     info.ModTime().Unix(),
		})
	}

	// Add additional browse roots from environment
	if extraRoots := os.Getenv("STORAGE_FS_BROWSE_ROOTS"); extraRoots != "" {
		for _, root := range strings.Split(extraRoots, ":") {
			root = strings.TrimSpace(root)
			if root == "" {
				continue
			}
			if info, err := os.Stat(root); err == nil && info.IsDir() {
				roots = append(roots, &pb.BrowseEntry{
					Name:        filepath.Base(root),
					Path:        root,
					IsDirectory: true,
					ModTime:     info.ModTime().Unix(),
				})
			}
		}
	}

	// Common media directories
	commonPaths := []string{
		"/media",
		"/mnt",
		"/data",
		"/videos",
	}

	// Add user's home directory
	if home, err := os.UserHomeDir(); err == nil {
		commonPaths = append(commonPaths, home)
		commonPaths = append(commonPaths, filepath.Join(home, "Videos"))
		commonPaths = append(commonPaths, filepath.Join(home, "Movies"))
	}

	for _, path := range commonPaths {
		// Skip if already added
		found := false
		for _, r := range roots {
			if r.Path == path {
				found = true
				break
			}
		}
		if found {
			continue
		}

		if info, err := os.Stat(path); err == nil && info.IsDir() {
			roots = append(roots, &pb.BrowseEntry{
				Name:        filepath.Base(path),
				Path:        path,
				IsDirectory: true,
				ModTime:     info.ModTime().Unix(),
			})
		}
	}

	return &pb.BrowseRootsResponse{Roots: roots}, nil
}

// Browse lists contents of a directory
func (s *FilesystemStorage) Browse(ctx context.Context, req *pb.BrowseRequest) (*pb.BrowseResponse, error) {
	path := req.Path
	if path == "" {
		path = s.baseDir
	}

	// Security: Clean the path
	path = filepath.Clean(path)

	// Check if path exists and is a directory
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("path not found: %s", path)
		}
		return nil, err
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", path)
	}

	// Read directory contents
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	browseEntries := []*pb.BrowseEntry{}
	searchQuery := strings.ToLower(req.SearchQuery)

	for _, entry := range entries {
		name := entry.Name()

		// Skip hidden files unless requested
		if !req.ShowHidden && strings.HasPrefix(name, ".") {
			continue
		}

		// Apply search filter
		if searchQuery != "" && !strings.Contains(strings.ToLower(name), searchQuery) {
			continue
		}

		entryInfo, err := entry.Info()
		if err != nil {
			continue
		}

		ext := getExtension(name)
		isVideoFile := isVideo(ext)
		isAudioFile := isAudio(ext)
		isImageFile := isImage(ext)

		// Filter to media files only if requested
		if req.MediaOnly && !entry.IsDir() && !isVideoFile && !isAudioFile && !isImageFile {
			continue
		}

		browseEntries = append(browseEntries, &pb.BrowseEntry{
			Name:        name,
			Path:        filepath.Join(path, name),
			IsDirectory: entry.IsDir(),
			Size:        entryInfo.Size(),
			ModTime:     entryInfo.ModTime().Unix(),
			Extension:   ext,
			IsVideo:     isVideoFile,
			IsAudio:     isAudioFile,
			IsImage:     isImageFile,
		})
	}

	// Sort: directories first, then by name
	sort.Slice(browseEntries, func(i, j int) bool {
		if browseEntries[i].IsDirectory != browseEntries[j].IsDirectory {
			return browseEntries[i].IsDirectory
		}
		return strings.ToLower(browseEntries[i].Name) < strings.ToLower(browseEntries[j].Name)
	})

	// Calculate parent path
	parentPath := filepath.Dir(path)
	if parentPath == path {
		parentPath = "" // At root
	}

	return &pb.BrowseResponse{
		CurrentPath: path,
		ParentPath:  parentPath,
		Entries:     browseEntries,
	}, nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: pluginsdk.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"webencode": &pluginsdk.Plugin{
				StorageImpl: NewFilesystemStorage(),
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
