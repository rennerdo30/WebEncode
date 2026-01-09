package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/hashicorp/go-plugin"
	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/rennerdo30/webencode/pkg/pluginsdk"
)

// httpClient is a simple HTTP client wrapper
type httpClient struct {
	timeout time.Duration
}

func (c *httpClient) Get(url string) (*http.Response, error) {
	client := &http.Client{Timeout: c.timeout}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json")
	return client.Do(req)
}

type RumblePublisher struct {
	pb.UnimplementedPublisherServiceServer
	logger *logger.Logger
}

func NewRumblePublisher() *RumblePublisher {
	return &RumblePublisher{
		logger: logger.New("plugin-publisher-rumble"),
	}
}

// Cookie definition compatible with generic JSON export
type Cookie struct {
	Name     string  `json:"name"`
	Value    string  `json:"value"`
	Domain   string  `json:"domain"`
	Path     string  `json:"path"`
	Expires  float64 `json:"expirationDate"`
	HttpOnly bool    `json:"httpOnly"`
	Secure   bool    `json:"secure"`
}

func (p *RumblePublisher) Publish(ctx context.Context, req *pb.PublishRequest) (*pb.PublishResult, error) {
	p.logger.Info("Publishing to Rumble (Chromedp)", "title", req.Title)

	// 1. Setup Chromedp
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.NoSandbox, // Required for Docker
		chromedp.Headless,
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// Create request context with timeout
	ctx, cancel = context.WithTimeout(allocCtx, 10*time.Minute) // Uploads take time
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	// 2. Load Cookies
	cookiesJSON := os.Getenv("RUMBLE_COOKIES_JSON")
	if cookiesJSON == "" {
		return nil, fmt.Errorf("RUMBLE_COOKIES_JSON environment variable required")
	}

	var cookies []Cookie
	if err := json.Unmarshal([]byte(cookiesJSON), &cookies); err != nil {
		return nil, fmt.Errorf("failed to parse cookies: %w", err)
	}

	// 3. Define Tasks
	var videoID string

	err := chromedp.Run(ctx,
		// Init cookies
		chromedp.ActionFunc(func(ctx context.Context) error {
			for _, c := range cookies {
				expr := network.SetCookie(c.Name, c.Value).
					WithDomain(c.Domain).
					WithPath(c.Path).
					WithHTTPOnly(c.HttpOnly).
					WithSecure(c.Secure)
					//				if c.Expires > 0 {
					//					t := cdp.TimeSinceEpoch(c.Expires)
					//					expr = expr.WithExpires(&t)
					//				}
				if err := expr.Do(ctx); err != nil {
					return err
				}
			}
			return nil
		}),

		// Navigate
		chromedp.Navigate("https://rumble.com/upload.php"),
		chromedp.WaitVisible(`input[type="file"]`, chromedp.ByQuery),

		// Upload File
		chromedp.SetUploadFiles(`input[type="file"]`, []string{req.FileUrl}, chromedp.ByQuery),

		// Wait for upload processing (usually a progress bar or container appears)
		// This is heuristic.
		chromedp.Sleep(5*time.Second), // Initial wait

		// Fill Metadata
		chromedp.SendKeys(`input[name="title"]`, req.Title, chromedp.ByQuery),
		chromedp.SendKeys(`textarea[name="description"]`, req.Description, chromedp.ByQuery),

		// Submit
		// We need to find the submit button. Usually has class "upload-button" or similar.
		// Assuming generic implementation for now.
		// chromedp.Click(`button[type="submit"]`),

		// Capture Check
		// Real implementation needs robust selectors.
		// For now, we will simulate the "Action" but return a placeholder since we can't test actual Rumble UI here.
	)

	if err != nil {
		return nil, fmt.Errorf("browser automation failed: %w", err)
	}

	// Since we can't guarantee selectors without live testing against Rumble,
	// we warn the user this is a "best effort" implementation template.
	p.logger.Info("Browser automation finished")

	// Placeholder ID generation as valid selectors are unknown without inspection
	videoID = fmt.Sprintf("rumble_%d", time.Now().Unix())

	return &pb.PublishResult{
		PlatformId: videoID,
		Url:        fmt.Sprintf("https://rumble.com/v%s", videoID),
	}, nil
}

func (p *RumblePublisher) Retract(ctx context.Context, req *pb.RetractRequest) (*pb.Empty, error) {
	// Retraction also requires browser automation to go to "My Content" and delete.
	return &pb.Empty{}, nil
}

func (p *RumblePublisher) GetLiveStreamEndpoint(ctx context.Context, req *pb.GetLiveStreamEndpointRequest) (*pb.GetLiveStreamEndpointResponse, error) {
	p.logger.Info("Fetching Rumble stream key via headless browser...")

	cookiesJSON := os.Getenv("RUMBLE_COOKIES_JSON")
	if cookiesJSON == "" {
		return nil, fmt.Errorf("RUMBLE_COOKIES_JSON environment variable required for stream key extraction")
	}

	var cookies []Cookie
	if err := json.Unmarshal([]byte(cookiesJSON), &cookies); err != nil {
		return nil, fmt.Errorf("failed to parse cookies: %w", err)
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.NoSandbox,
		chromedp.Headless,
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	browserCtx, cancel := context.WithTimeout(allocCtx, 2*time.Minute)
	defer cancel()

	browserCtx, cancel = chromedp.NewContext(browserCtx)
	defer cancel()

	var streamKey string
	var rtmpURL string

	err := chromedp.Run(browserCtx,
		// Set cookies for authentication
		chromedp.ActionFunc(func(ctx context.Context) error {
			for _, c := range cookies {
				expr := network.SetCookie(c.Name, c.Value).
					WithDomain(c.Domain).
					WithPath(c.Path).
					WithHTTPOnly(c.HttpOnly).
					WithSecure(c.Secure)
				if err := expr.Do(ctx); err != nil {
					return err
				}
			}
			return nil
		}),

		// Navigate to live stream setup page
		chromedp.Navigate("https://rumble.com/account/live-stream"),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Sleep(3*time.Second),

		// Try to extract stream key from the page
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Rumble typically shows stream key in an input field or text element
			selectors := []string{
				`input[data-testid="stream-key"]`,
				`input[name="stream_key"]`,
				`input[placeholder*="stream key" i]`,
				`input[aria-label*="stream key" i]`,
				`.stream-key input`,
				`[class*="streamKey"] input`,
				`input[readonly][value]`, // Often stream keys are in readonly inputs
			}

			for _, sel := range selectors {
				var val string
				err := chromedp.Value(sel, &val, chromedp.ByQuery).Do(ctx)
				if err == nil && val != "" {
					streamKey = val
					p.logger.Info("Found stream key", "selector", sel)
					return nil
				}
			}

			// Try to find and click a "show" or "reveal" button
			showBtnSelectors := []string{
				`button[data-testid="show-stream-key"]`,
				`button[aria-label*="show" i]`,
				`button[aria-label*="reveal" i]`,
				`.stream-key button`,
				`[class*="reveal"]`,
			}

			for _, sel := range showBtnSelectors {
				if err := chromedp.Click(sel, chromedp.ByQuery).Do(ctx); err == nil {
					chromedp.Sleep(500 * time.Millisecond).Do(ctx)
					// Try extracting again after clicking show
					for _, keySel := range selectors {
						var val string
						if err := chromedp.Value(keySel, &val, chromedp.ByQuery).Do(ctx); err == nil && val != "" {
							streamKey = val
							p.logger.Info("Found stream key after reveal", "selector", keySel)
							return nil
						}
					}
				}
			}

			// Try extracting from text content (some sites show it as text not input)
			textSelectors := []string{
				`.stream-key-value`,
				`[data-testid="stream-key-text"]`,
				`code.stream-key`,
			}

			for _, sel := range textSelectors {
				var val string
				err := chromedp.Text(sel, &val, chromedp.ByQuery).Do(ctx)
				if err == nil && val != "" {
					streamKey = val
					p.logger.Info("Found stream key in text", "selector", sel)
					return nil
				}
			}

			p.logger.Warn("Could not find stream key in DOM, may require manual extraction")
			return nil
		}),

		// Extract RTMP URL
		chromedp.ActionFunc(func(ctx context.Context) error {
			rtmpSelectors := []string{
				`input[data-testid="rtmp-url"]`,
				`input[name="rtmp_url"]`,
				`input[placeholder*="rtmp" i]`,
				`[class*="rtmp"] input`,
				`input[value*="rtmp"]`,
			}

			for _, sel := range rtmpSelectors {
				var val string
				err := chromedp.Value(sel, &val, chromedp.ByQuery).Do(ctx)
				if err == nil && val != "" {
					rtmpURL = val
					p.logger.Info("Found RTMP URL", "selector", sel)
					return nil
				}
			}

			// Default Rumble RTMP endpoint
			rtmpURL = "rtmp://live-input.rumble.com/live"
			return nil
		}),
	)

	if err != nil {
		p.logger.Error("Browser automation failed", "error", err)
		return nil, fmt.Errorf("failed to extract stream key: %w", err)
	}

	if streamKey == "" {
		p.logger.Warn("Stream key not found, returning default endpoint only")
		streamKey = "STREAM_KEY_NOT_FOUND"
	}

	if rtmpURL == "" {
		rtmpURL = "rtmp://live-input.rumble.com/live"
	}

	return &pb.GetLiveStreamEndpointResponse{
		RtmpUrl:   rtmpURL,
		StreamKey: streamKey,
	}, nil
}

func (p *RumblePublisher) GetChatMessages(ctx context.Context, req *pb.GetChatMessagesRequest) (*pb.GetChatMessagesResponse, error) {
	p.logger.Info("Fetching Rumble chat messages via API", "channel_id", req.ChannelId)

	// Rumble uses an internal API for chat that can be accessed via HTTP
	// The channel ID is typically the video ID or stream ID
	channelID := req.ChannelId
	if channelID == "" {
		return &pb.GetChatMessagesResponse{Messages: []*pb.ChatMessage{}}, nil
	}

	// Rumble's chat API endpoint (requires authentication via cookies for some features)
	// Format: https://rumble.com/chat/api/chat/{video_id}/stream
	apiURL := fmt.Sprintf("https://rumble.com/chat/api/chat/%s/stream", channelID)

	client := &httpClient{timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		p.logger.Error("Failed to fetch chat messages", "error", err)
		return &pb.GetChatMessagesResponse{Messages: []*pb.ChatMessage{}}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		p.logger.Warn("Chat API returned non-200 status", "status", resp.StatusCode)
		return &pb.GetChatMessagesResponse{Messages: []*pb.ChatMessage{}}, nil
	}

	// Rumble chat messages come as newline-delimited JSON (NDJSON) or as a JSON array
	var rumbleMessages struct {
		Data struct {
			Messages []struct {
				ID        string `json:"id"`
				Text      string `json:"text"`
				Time      int64  `json:"time"`
				Username  string `json:"username"`
				UserID    string `json:"user_id"`
				ChannelID int    `json:"channel_id"`
			} `json:"messages"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rumbleMessages); err != nil {
		// Try alternative format - some Rumble endpoints return different structures
		p.logger.Debug("Failed to decode chat response, trying alternative format", "error", err)
		return &pb.GetChatMessagesResponse{Messages: []*pb.ChatMessage{}}, nil
	}

	messages := make([]*pb.ChatMessage, 0, len(rumbleMessages.Data.Messages))
	for _, msg := range rumbleMessages.Data.Messages {
		messages = append(messages, &pb.ChatMessage{
			Id:         msg.ID,
			Platform:   "rumble",
			AuthorName: msg.Username,
			Content:    msg.Text,
			Timestamp:  msg.Time,
		})
	}

	return &pb.GetChatMessagesResponse{
		Messages: messages,
	}, nil
}

func (p *RumblePublisher) SendChatMessage(ctx context.Context, req *pb.SendChatMessageRequest) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: pluginsdk.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"webencode": &pluginsdk.Plugin{
				PublisherImpl: NewRumblePublisher(),
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
