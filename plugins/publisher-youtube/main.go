package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/go-plugin"
	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/rennerdo30/webencode/pkg/pluginsdk"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type YouTubePublisher struct {
	pb.UnimplementedPublisherServiceServer
	logger *logger.Logger
	config *Config
}

type Config struct {
	ClientID     string
	ClientSecret string
	EndpointURL  string // Optional override
}

func NewYouTubePublisher() *YouTubePublisher {
	cfg := &Config{
		ClientID:     os.Getenv("YOUTUBE_CLIENT_ID"),
		ClientSecret: os.Getenv("YOUTUBE_CLIENT_SECRET"),
		EndpointURL:  os.Getenv("YOUTUBE_API_ENDPOINT"), // For testing
	}

	return &YouTubePublisher{
		logger: logger.New("plugin-publisher-youtube"),
		config: cfg,
	}
}

func (p *YouTubePublisher) getYouTubeService(ctx context.Context, accessToken string) (*youtube.Service, error) {
	config := &oauth2.Config{
		ClientID:     p.config.ClientID,
		ClientSecret: p.config.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{youtube.YoutubeUploadScope},
	}

	token := &oauth2.Token{
		AccessToken: accessToken,
	}

	// Check if it's a refresh token
	if strings.HasPrefix(accessToken, "refresh:") {
		token = &oauth2.Token{
			RefreshToken: strings.TrimPrefix(accessToken, "refresh:"),
		}
	}

	client := config.Client(ctx, token)

	opts := []option.ClientOption{option.WithHTTPClient(client)}
	if p.config.EndpointURL != "" {
		opts = append(opts, option.WithEndpoint(p.config.EndpointURL))
	}

	return youtube.NewService(ctx, opts...)
}

func (p *YouTubePublisher) Publish(ctx context.Context, req *pb.PublishRequest) (*pb.PublishResult, error) {
	p.logger.Info("Publishing to YouTube", "title", req.Title, "source", req.FileUrl)

	// Validate config
	if p.config.ClientID == "" || p.config.ClientSecret == "" {
		p.logger.Error("YouTube credentials not configured")
		return nil, fmt.Errorf("YouTube credentials not configured")
	}

	if req.AccessToken == "" {
		return nil, fmt.Errorf("access_token required for YouTube upload")
	}

	// Get YouTube service
	service, err := p.getYouTubeService(ctx, req.AccessToken)
	if err != nil {
		p.logger.Error("Failed to create YouTube service", "error", err)
		return nil, fmt.Errorf("failed to create YouTube service: %w", err)
	}

	// Download the source video
	resp, err := http.Get(req.FileUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to download source: %w", err)
	}
	defer resp.Body.Close()

	// Create upload
	upload := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       req.Title,
			Description: req.Description,
			CategoryId:  "22", // People & Blogs (default)
		},
		Status: &youtube.VideoStatus{
			PrivacyStatus: "private", // Default to private
		},
	}

	call := service.Videos.Insert([]string{"snippet", "status"}, upload)
	call.Media(resp.Body)

	video, err := call.Do()
	if err != nil {
		p.logger.Error("Failed to upload to YouTube", "error", err)
		return nil, fmt.Errorf("upload failed: %w", err)
	}

	p.logger.Info("Upload successful", "video_id", video.Id)
	return &pb.PublishResult{
		PlatformId: video.Id,
		Url:        fmt.Sprintf("https://www.youtube.com/watch?v=%s", video.Id),
	}, nil
}

func (p *YouTubePublisher) Retract(ctx context.Context, req *pb.RetractRequest) (*pb.Empty, error) {
	p.logger.Info("Retracting from YouTube", "video_id", req.PlatformId)

	if req.AccessToken == "" {
		return nil, fmt.Errorf("access_token required")
	}

	service, err := p.getYouTubeService(ctx, req.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube service: %w", err)
	}

	err = service.Videos.Delete(req.PlatformId).Do()
	if err != nil {
		p.logger.Error("Failed to delete video", "error", err)
		return nil, fmt.Errorf("delete failed: %w", err)
	}

	return &pb.Empty{}, nil
}

func (p *YouTubePublisher) GetLiveStreamEndpoint(ctx context.Context, req *pb.GetLiveStreamEndpointRequest) (*pb.GetLiveStreamEndpointResponse, error) {
	if req.AccessToken == "" {
		return nil, fmt.Errorf("access_token required")
	}

	service, err := p.getYouTubeService(ctx, req.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube service: %w", err)
	}

	// List live broadcasts to find active or upcoming broadcast
	broadcastsCall := service.LiveBroadcasts.List([]string{"id", "snippet", "contentDetails"})
	broadcastsCall.BroadcastStatus("active")
	broadcastsCall.Mine(true)

	broadcasts, err := broadcastsCall.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list broadcasts: %w", err)
	}

	var streamID string
	if len(broadcasts.Items) > 0 {
		// Use existing broadcast's stream
		streamID = broadcasts.Items[0].ContentDetails.BoundStreamId
	} else {
		// Create a new live broadcast
		broadcast := &youtube.LiveBroadcast{
			Snippet: &youtube.LiveBroadcastSnippet{
				Title:              "Live Stream",
				ScheduledStartTime: time.Now().Format(time.RFC3339),
			},
			Status: &youtube.LiveBroadcastStatus{
				PrivacyStatus: "unlisted",
			},
		}

		createCall := service.LiveBroadcasts.Insert([]string{"snippet", "status", "contentDetails"}, broadcast)
		created, err := createCall.Do()
		if err != nil {
			return nil, fmt.Errorf("failed to create broadcast: %w", err)
		}

		// Create a live stream
		stream := &youtube.LiveStream{
			Snippet: &youtube.LiveStreamSnippet{
				Title: "Primary Stream",
			},
			Cdn: &youtube.CdnSettings{
				IngestionType: "rtmp",
				FrameRate:     "variable",
				Resolution:    "variable",
			},
		}

		streamCall := service.LiveStreams.Insert([]string{"snippet", "cdn"}, stream)
		streamCreated, err := streamCall.Do()
		if err != nil {
			return nil, fmt.Errorf("failed to create stream: %w", err)
		}

		streamID = streamCreated.Id

		// Bind stream to broadcast
		bindCall := service.LiveBroadcasts.Bind(created.Id, []string{"id", "contentDetails"})
		bindCall.StreamId(streamID)
		_, err = bindCall.Do()
		if err != nil {
			return nil, fmt.Errorf("failed to bind stream: %w", err)
		}
	}

	// Get stream details with ingestion info
	streamCall := service.LiveStreams.List([]string{"cdn"})
	streamCall.Id(streamID)

	streams, err := streamCall.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get stream details: %w", err)
	}

	if len(streams.Items) == 0 {
		return nil, fmt.Errorf("stream not found")
	}

	streamInfo := streams.Items[0]

	return &pb.GetLiveStreamEndpointResponse{
		RtmpUrl:   streamInfo.Cdn.IngestionInfo.IngestionAddress,
		StreamKey: streamInfo.Cdn.IngestionInfo.StreamName,
	}, nil
}

func (p *YouTubePublisher) GetChatMessages(ctx context.Context, req *pb.GetChatMessagesRequest) (*pb.GetChatMessagesResponse, error) {
	if req.AccessToken == "" {
		return nil, fmt.Errorf("access_token required")
	}

	service, err := p.getYouTubeService(ctx, req.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube service: %w", err)
	}

	// Get the live chat ID from active broadcast
	broadcastsCall := service.LiveBroadcasts.List([]string{"snippet", "contentDetails"})
	broadcastsCall.BroadcastStatus("active")
	broadcastsCall.Mine(true)

	broadcasts, err := broadcastsCall.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list broadcasts: %w", err)
	}

	if len(broadcasts.Items) == 0 {
		// No active broadcast, return empty
		return &pb.GetChatMessagesResponse{Messages: []*pb.ChatMessage{}}, nil
	}

	liveChatID := broadcasts.Items[0].Snippet.LiveChatId
	if liveChatID == "" {
		return &pb.GetChatMessagesResponse{Messages: []*pb.ChatMessage{}}, nil
	}

	// Fetch chat messages
	chatCall := service.LiveChatMessages.List(liveChatID, []string{"snippet", "authorDetails"})
	chatCall.MaxResults(100)

	// Use since_id for pagination (as pageToken)
	if req.SinceId != "" {
		chatCall.PageToken(req.SinceId)
	}

	chatResp, err := chatCall.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch chat messages: %w", err)
	}

	messages := make([]*pb.ChatMessage, 0, len(chatResp.Items))
	for _, item := range chatResp.Items {
		// Parse timestamp
		publishedAt, _ := time.Parse(time.RFC3339, item.Snippet.PublishedAt)

		messages = append(messages, &pb.ChatMessage{
			Id:         item.Id,
			Platform:   "youtube",
			AuthorName: item.AuthorDetails.DisplayName,
			Content:    item.Snippet.DisplayMessage,
			Timestamp:  publishedAt.Unix(),
		})
	}

	return &pb.GetChatMessagesResponse{
		Messages:    messages,
		NextSinceId: chatResp.NextPageToken,
	}, nil
}

func (p *YouTubePublisher) SendChatMessage(ctx context.Context, req *pb.SendChatMessageRequest) (*pb.Empty, error) {
	if req.AccessToken == "" {
		return nil, fmt.Errorf("access_token required")
	}

	service, err := p.getYouTubeService(ctx, req.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube service: %w", err)
	}

	// Get the live chat ID from active broadcast
	broadcastsCall := service.LiveBroadcasts.List([]string{"snippet"})
	broadcastsCall.BroadcastStatus("active")
	broadcastsCall.Mine(true)

	broadcasts, err := broadcastsCall.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list broadcasts: %w", err)
	}

	if len(broadcasts.Items) == 0 {
		return nil, fmt.Errorf("no active broadcast found")
	}

	liveChatID := broadcasts.Items[0].Snippet.LiveChatId
	if liveChatID == "" {
		return nil, fmt.Errorf("live chat not enabled for this broadcast")
	}

	// Send the message
	message := &youtube.LiveChatMessage{
		Snippet: &youtube.LiveChatMessageSnippet{
			LiveChatId: liveChatID,
			Type:       "textMessageEvent",
			TextMessageDetails: &youtube.LiveChatTextMessageDetails{
				MessageText: req.Message,
			},
		},
	}

	_, err = service.LiveChatMessages.Insert([]string{"snippet"}, message).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return &pb.Empty{}, nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: pluginsdk.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"webencode": &pluginsdk.Plugin{
				PublisherImpl: NewYouTubePublisher(),
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
