package worker

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	pb "github.com/rennerdo30/webencode/pkg/api/v1"
)

func (w *Worker) downloadFile(ctx context.Context, bucket, key, localPath string) error {
	// Try storage plugin first
	var client pb.StorageServiceClient
	for _, c := range w.pluginManager.Storage {
		if sc, ok := c.(pb.StorageServiceClient); ok {
			client = sc
			break
		}
	}

	if client != nil {
		return w.downloadViaPlugin(ctx, client, bucket, key, localPath)
	}

	// Fallback to direct S3 access using env vars
	return w.downloadViaS3(ctx, bucket, key, localPath)
}

func (w *Worker) downloadViaPlugin(ctx context.Context, client pb.StorageServiceClient, bucket, key, localPath string) error {
	w.logger.Debug("Downloading file via plugin", "bucket", bucket, "key", key)

	stream, err := client.Download(ctx, &pb.FileRequest{Bucket: bucket, Path: key})
	if err != nil {
		return fmt.Errorf("download start failed: %w", err)
	}

	f, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer f.Close()

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if data := chunk.GetData(); data != nil {
			if _, err := f.Write(data); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *Worker) downloadViaS3(ctx context.Context, bucket, key, localPath string) error {
	endpoint := os.Getenv("S3_ENDPOINT")
	accessKey := os.Getenv("S3_ACCESS_KEY")
	secretKey := os.Getenv("S3_SECRET_KEY")

	if endpoint == "" {
		return fmt.Errorf("S3_ENDPOINT not configured")
	}

	w.logger.Debug("Downloading file via S3", "endpoint", endpoint, "bucket", bucket, "key", key)

	// Create minio client
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		return fmt.Errorf("failed to create S3 client: %w", err)
	}

	// Download object
	object, err := minioClient.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to get object: %w", err)
	}
	defer object.Close()

	// Create local file
	f, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Copy data
	_, err = io.Copy(f, object)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	w.logger.Info("Downloaded file from S3", "bucket", bucket, "key", key, "local", localPath)
	return nil
}

func (w *Worker) uploadFile(ctx context.Context, localPath, bucket, key string) (string, error) {
	// 1. Try Storage Plugin
	var client pb.StorageServiceClient
	for _, c := range w.pluginManager.Storage {
		if sc, ok := c.(pb.StorageServiceClient); ok {
			client = sc
			break
		}
	}

	if client != nil {
		url, err := w.uploadViaPlugin(ctx, client, localPath, bucket, key)
		if err == nil && !strings.HasPrefix(url, "mock://") {
			return url, nil
		}
		// If plugin failed or returned a mock URL, we can fallback to S3 if configured.
		// However, usually plugins are authoritative.
		// But in this specific case, we want to force real upload if the plugin was just a mock.
		if err != nil {
			w.logger.Warn("Plugin upload failed, falling back to S3", "error", err)
		}
	}

	// 2. Fallback to direct S3 access
	return w.uploadViaS3(ctx, localPath, bucket, key)
}

func (w *Worker) uploadViaPlugin(ctx context.Context, client pb.StorageServiceClient, localPath, bucket, key string) (string, error) {
	w.logger.Debug("Uploading file via plugin", "path", localPath, "bucket", bucket)

	f, err := os.Open(localPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	stream, err := client.Upload(ctx)
	if err != nil {
		return "", err
	}

	if err := stream.Send(&pb.FileChunk{
		Content: &pb.FileChunk_Metadata{
			Metadata: &pb.FileMetadata{
				Bucket: bucket,
				Path:   key,
			},
		},
	}); err != nil {
		return "", err
	}

	buf := make([]byte, 32*1024)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			if err := stream.Send(&pb.FileChunk{
				Content: &pb.FileChunk_Data{
					Data: buf[:n],
				},
			}); err != nil {
				return "", err
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
	}

	summary, err := stream.CloseAndRecv()
	if err != nil {
		return "", err
	}

	return summary.Url, nil
}

func (w *Worker) uploadViaS3(ctx context.Context, localPath, bucket, key string) (string, error) {
	endpoint := os.Getenv("S3_ENDPOINT")
	accessKey := os.Getenv("S3_ACCESS_KEY")
	secretKey := os.Getenv("S3_SECRET_KEY")

	if endpoint == "" {
		return "", fmt.Errorf("S3_ENDPOINT not configured")
	}

	w.logger.Debug("Uploading file via S3", "endpoint", endpoint, "bucket", bucket, "key", key)

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create S3 client: %w", err)
	}

	f, err := os.Open(localPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	fileInfo, err := f.Stat()
	if err != nil {
		return "", err
	}

	// Upload object
	info, err := minioClient.PutObject(ctx, bucket, key, f, fileInfo.Size(), minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload object: %w", err)
	}

	// Construct URL
	// Assuming http for internal s3 endpoint
	url := fmt.Sprintf("http://%s/%s/%s", endpoint, bucket, key)
	w.logger.Info("Uploaded file to S3", "url", url, "size", info.Size)
	return url, nil
}

func parseS3URL(rawURL string) (bucket, key string, err error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", "", err
	}
	if u.Scheme != "s3" {
		return "", "", fmt.Errorf("invalid scheme: %s", u.Scheme)
	}
	bucket = u.Host
	key = strings.TrimPrefix(u.Path, "/")
	return bucket, key, nil
}
