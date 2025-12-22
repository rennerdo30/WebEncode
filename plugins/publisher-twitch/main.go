package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/hashicorp/go-plugin"
	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/rennerdo30/webencode/pkg/pluginsdk"
)

const (
	twitchAPIBase     = "https://api.twitch.tv/helix"
	twitchUploadBase  = "https://uploads.twitch.tv/upload"
	twitchVideoCreate = "/videos"
)

type TwitchPublisher struct {
	pb.UnimplementedPublisherServiceServer
	logger   *logger.Logger
	clientID string
	apiURL   string
}

func NewTwitchPublisher() *TwitchPublisher {
	api := os.Getenv("TWITCH_API_URL")
	if api == "" {
		api = twitchAPIBase
	}

	return &TwitchPublisher{
		logger:   logger.New("plugin-publisher-twitch"),
		clientID: os.Getenv("TWITCH_CLIENT_ID"),
		apiURL:   api,
	}
}

// TwitchVideoResponse is the response from Twitch video creation API
type TwitchVideoResponse struct {
	Data []struct {
		ID        string `json:"id"`
		StreamID  string `json:"stream_id"`
		UserID    string `json:"user_id"`
		Title     string `json:"title"`
		URL       string `json:"url"`
		UploadURL string `json:"upload_url,omitempty"`
	} `json:"data"`
}

func (p *TwitchPublisher) Publish(ctx context.Context, req *pb.PublishRequest) (*pb.PublishResult, error) {
	p.logger.Info("Publishing to Twitch", "title", req.Title, "source", req.FileUrl)

	if p.clientID == "" {
		return nil, fmt.Errorf("TWITCH_CLIENT_ID not configured")
	}

	if req.AccessToken == "" {
		return nil, fmt.Errorf("access_token required for Twitch upload")
	}

	// Step 1: Create video entry on Twitch
	videoID, uploadURL, err := p.createVideo(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create video: %w", err)
	}

	// Step 2: Download source file
	sourceResp, err := http.Get(req.FileUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to download source: %w", err)
	}
	defer sourceResp.Body.Close()

	// Step 3: Upload to Twitch
	err = p.uploadVideo(ctx, uploadURL, req.AccessToken, sourceResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to upload video: %w", err)
	}

	// Step 4: Complete upload
	err = p.completeUpload(ctx, videoID, req.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to complete upload: %w", err)
	}

	p.logger.Info("Upload successful", "video_id", videoID)
	return &pb.PublishResult{
		PlatformId: videoID,
		Url:        fmt.Sprintf("https://www.twitch.tv/videos/%s", videoID),
	}, nil
}

func (p *TwitchPublisher) createVideo(ctx context.Context, req *pb.PublishRequest) (string, string, error) {
	payload := map[string]string{
		"title":       req.Title,
		"description": req.Description,
	}
	body, _ := json.Marshal(payload)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.apiURL+twitchVideoCreate, bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}

	httpReq.Header.Set("Authorization", "Bearer "+req.AccessToken)
	httpReq.Header.Set("Client-Id", p.clientID)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("twitch API error: %s - %s", resp.Status, string(respBody))
	}

	var twitchResp TwitchVideoResponse
	if err := json.NewDecoder(resp.Body).Decode(&twitchResp); err != nil {
		return "", "", err
	}

	if len(twitchResp.Data) == 0 {
		return "", "", fmt.Errorf("no video data in response")
	}

	return twitchResp.Data[0].ID, twitchResp.Data[0].UploadURL, nil
}

func (p *TwitchPublisher) uploadVideo(ctx context.Context, uploadURL, accessToken string, body io.Reader) error {
	httpReq, err := http.NewRequestWithContext(ctx, "PUT", uploadURL, body)
	if err != nil {
		return err
	}

	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Client-Id", p.clientID)
	httpReq.Header.Set("Content-Type", "video/mp4")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed: %s - %s", resp.Status, string(respBody))
	}

	return nil
}

func (p *TwitchPublisher) completeUpload(ctx context.Context, videoID, accessToken string) error {
	url := fmt.Sprintf("%s%s/%s/complete", p.apiURL, twitchVideoCreate, videoID)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return err
	}

	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Client-Id", p.clientID)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("complete failed: %s - %s", resp.Status, string(respBody))
	}

	return nil
}

func (p *TwitchPublisher) Retract(ctx context.Context, req *pb.RetractRequest) (*pb.Empty, error) {
	p.logger.Info("Retracting from Twitch", "video_id", req.PlatformId)

	if req.AccessToken == "" {
		return nil, fmt.Errorf("access_token required")
	}

	url := fmt.Sprintf("%s%s?id=%s", p.apiURL, twitchVideoCreate, req.PlatformId)

	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+req.AccessToken)
	httpReq.Header.Set("Client-Id", p.clientID)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("delete failed: %s - %s", resp.Status, string(respBody))
	}

	return &pb.Empty{}, nil
}

func (p *TwitchPublisher) GetLiveStreamEndpoint(ctx context.Context, req *pb.GetLiveStreamEndpointRequest) (*pb.GetLiveStreamEndpointResponse, error) {
	// GET https://api.twitch.tv/helix/streams/key
	// API Docs: GET https://api.twitch.tv/helix/streams/key
	// Requires User Access Token with channel:read:stream_key

	httpReq, err := http.NewRequestWithContext(ctx, "GET", p.apiURL+"/streams/key", nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+req.AccessToken)
	httpReq.Header.Set("Client-Id", p.clientID)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("twitch api error: %s - %s", resp.Status, string(body))
	}

	var data struct {
		Data []struct {
			StreamKey string `json:"stream_key"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	if len(data.Data) == 0 {
		return nil, fmt.Errorf("no stream key found")
	}

	return &pb.GetLiveStreamEndpointResponse{
		RtmpUrl:   "rtmp://live.twitch.tv/app/", // Default ingest, could be dynamic
		StreamKey: data.Data[0].StreamKey,
	}, nil
}

func (p *TwitchPublisher) GetChatMessages(ctx context.Context, req *pb.GetChatMessagesRequest) (*pb.GetChatMessagesResponse, error) {
	// Note: Twitch requires EventSub WebSocket or IRC for realtime chat
	// For polling, we can use the moderator chat API
	if req.AccessToken == "" {
		return nil, fmt.Errorf("access_token required")
	}

	// Get user's channel ID first
	userReq, err := http.NewRequestWithContext(ctx, "GET", p.apiURL+"/users", nil)
	if err != nil {
		return nil, err
	}
	userReq.Header.Set("Authorization", "Bearer "+req.AccessToken)
	userReq.Header.Set("Client-Id", p.clientID)

	userResp, err := http.DefaultClient.Do(userReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer userResp.Body.Close()

	var userData struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(userResp.Body).Decode(&userData); err != nil {
		return nil, err
	}

	if len(userData.Data) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	broadcasterID := userData.Data[0].ID
	_ = broadcasterID // Will be used when IRC/EventSub is implemented

	// Note: Twitch doesn't have a REST API to fetch historical chat
	// This would require IRC connection or EventSub WebSocket
	// For now, return empty with a note that this needs IRC/EventSub
	p.logger.Warn("Twitch chat requires IRC or EventSub WebSocket for real-time messages")

	return &pb.GetChatMessagesResponse{
		Messages:    []*pb.ChatMessage{},
		NextSinceId: "",
	}, nil
}

func (p *TwitchPublisher) SendChatMessage(ctx context.Context, req *pb.SendChatMessageRequest) (*pb.Empty, error) {
	if req.AccessToken == "" {
		return nil, fmt.Errorf("access_token required")
	}

	// Get user's channel ID
	userReq, err := http.NewRequestWithContext(ctx, "GET", p.apiURL+"/users", nil)
	if err != nil {
		return nil, err
	}
	userReq.Header.Set("Authorization", "Bearer "+req.AccessToken)
	userReq.Header.Set("Client-Id", p.clientID)

	userResp, err := http.DefaultClient.Do(userReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer userResp.Body.Close()

	var userData struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(userResp.Body).Decode(&userData); err != nil {
		return nil, err
	}

	if len(userData.Data) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	broadcasterID := userData.Data[0].ID

	// Send chat message using moderator endpoint
	payload := map[string]string{
		"broadcaster_id": broadcasterID,
		"sender_id":      broadcasterID,
		"message":        req.Message,
	}
	body, _ := json.Marshal(payload)

	chatReq, err := http.NewRequestWithContext(ctx, "POST", p.apiURL+"/chat/messages", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	chatReq.Header.Set("Authorization", "Bearer "+req.AccessToken)
	chatReq.Header.Set("Client-Id", p.clientID)
	chatReq.Header.Set("Content-Type", "application/json")

	chatResp, err := http.DefaultClient.Do(chatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}
	defer chatResp.Body.Close()

	if chatResp.StatusCode != http.StatusOK && chatResp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(chatResp.Body)
		return nil, fmt.Errorf("send message failed: %s - %s", chatResp.Status, string(respBody))
	}

	return &pb.Empty{}, nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: pluginsdk.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"webencode": &pluginsdk.Plugin{
				PublisherImpl: NewTwitchPublisher(),
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
