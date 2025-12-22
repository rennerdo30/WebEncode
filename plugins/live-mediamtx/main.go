package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/go-plugin"
	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/rennerdo30/webencode/pkg/pluginsdk"
)

const mediamtxConfigTmpl = `
paths:
  all:
    runOnPublish: sh -c 'wget -q -O - --post-data="path=$MTX_PATH" http://kernel:8090/v1/live/start || true'
    runOnPublishRequest: http://kernel:8090/v1/live/auth
    runOnUnpublish: sh -c 'wget -q -O - --post-data="path=$MTX_PATH" http://kernel:8090/v1/live/stop || true'
`

type MediaMTXPlugin struct {
	pb.UnimplementedLiveServiceServer
	logger *logger.Logger
	apiURL string
}

func NewMediaMTXPlugin() *MediaMTXPlugin {
	url := os.Getenv("MEDIAMTX_API_URL")
	if url == "" {
		url = "http://localhost:9997"
	}
	return &MediaMTXPlugin{
		logger: logger.New("plugin-live-mediamtx"),
		apiURL: url,
	}
}

func (p *MediaMTXPlugin) StartIngest(ctx context.Context, req *pb.IngestConfig) (*pb.IngestSession, error) {
	sessionID := uuid.NewString()
	streamKey := req.StreamKey
	if streamKey == "" {
		streamKey = sessionID
	}

	// Ingress via RTMP (standard 1935)
	// Playback via HLS (standard 8888)
	// These URLs are returned to the client/kernel
	return &pb.IngestSession{
		Id:          sessionID,
		IngestUrl:   fmt.Sprintf("rtmp://localhost:1935/live/%s", streamKey),
		PlaybackUrl: fmt.Sprintf("http://localhost:8888/live/%s/index.m3u8", streamKey),
	}, nil
}

func (p *MediaMTXPlugin) StopIngest(ctx context.Context, req *pb.SessionID) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}

type PathListResponse struct {
	Items []struct {
		Name   string `json:"name"`
		Source struct {
			Type string `json:"type"`
		} `json:"source"`
		Ready   bool          `json:"ready"`
		Readers []interface{} `json:"readers"`
	} `json:"items"`
}

func (p *MediaMTXPlugin) GetTelemetry(ctx context.Context, req *pb.SessionID) (*pb.Telemetry, error) {
	// Query MediaMTX API /v3/paths/list
	apiEndpoint := fmt.Sprintf("%s/v3/paths/list", p.apiURL)

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, "GET", apiEndpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		// Log error but return default
		p.logger.Error("Failed to query MediaMTX API", "error", err)
		return &pb.Telemetry{IsLive: false}, nil
	}
	defer resp.Body.Close()

	var paths PathListResponse
	if err := json.NewDecoder(resp.Body).Decode(&paths); err != nil {
		return &pb.Telemetry{IsLive: false}, nil
	}

	// Find the path corresponding to this session/stream
	// MediaMTX path names in the API can be "streamkey" or "app/streamkey"
	for _, item := range paths.Items {
		// Try exact match, then match without prefix
		nameMatch := (item.Name == req.Id) ||
			(strings.HasSuffix(item.Name, "/"+req.Id)) ||
			(req.Id == "all")

		if nameMatch {
			return &pb.Telemetry{
				IsLive:  item.Ready,
				Bitrate: 3500000,
				Fps:     60,
				Viewers: int64(len(item.Readers)),
			}, nil
		}
	}

	return &pb.Telemetry{IsLive: false}, nil
}

func (p *MediaMTXPlugin) AddOutputTarget(ctx context.Context, req *pb.AddOutputTargetRequest) (*pb.Empty, error) {
	p.logger.Info("Adding output target to MediaMTX", "session", req.SessionId, "target", req.TargetUrl)

	// MediaMTX v3 API: PATCH /v3/config/paths/patch/{name}
	// We need to update the path configuration to add an RTMP output
	pathName := req.SessionId

	// Build the path configuration patch
	pathConfig := map[string]interface{}{
		"rtmpPublishings": []string{req.TargetUrl},
	}

	body, err := json.Marshal(pathConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	apiURL := fmt.Sprintf("%s/v3/config/paths/patch/%s", p.apiURL, pathName)
	httpReq, err := http.NewRequestWithContext(ctx, "PATCH", apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		p.logger.Error("Failed to add output target", "error", err)
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("MediaMTX API error: %s - %s", resp.Status, string(bodyBytes))
	}

	p.logger.Info("Output target added successfully")
	return &pb.Empty{}, nil
}

func (p *MediaMTXPlugin) RemoveOutputTarget(ctx context.Context, req *pb.RemoveOutputTargetRequest) (*pb.Empty, error) {
	p.logger.Info("Removing output target from MediaMTX", "session", req.SessionId, "target", req.TargetUrl)

	// MediaMTX v3 API: PATCH /v3/config/paths/patch/{name}
	// Remove the RTMP publishing from the path configuration
	pathName := req.SessionId

	// Build the path configuration patch to clear rtmpPublishings
	pathConfig := map[string]interface{}{
		"rtmpPublishings": []string{},
	}

	body, err := json.Marshal(pathConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	apiURL := fmt.Sprintf("%s/v3/config/paths/patch/%s", p.apiURL, pathName)
	httpReq, err := http.NewRequestWithContext(ctx, "PATCH", apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		p.logger.Error("Failed to remove output target", "error", err)
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("MediaMTX API error: %s - %s", resp.Status, string(bodyBytes))
	}

	p.logger.Info("Output target removed successfully")
	return &pb.Empty{}, nil
}

func main() {
	cwd, _ := os.Getwd()
	configPath := filepath.Join(cwd, "mediamtx.yml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t := template.Must(template.New("conf").Parse(mediamtxConfigTmpl))
		f, _ := os.Create(configPath)
		t.Execute(f, nil)
		f.Close()
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: pluginsdk.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"webencode": &pluginsdk.Plugin{
				LiveImpl: NewMediaMTXPlugin(),
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
