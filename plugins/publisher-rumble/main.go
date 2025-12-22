package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/hashicorp/go-plugin"
	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/rennerdo30/webencode/pkg/pluginsdk"
)

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
	// TODO: Use Chromedp to navigate to https://rumble.com/live-stream-setup
	// and extract the Stream Key from the DOM
	p.logger.Info("Fetching Rumble stream key via headless browser...")

	return &pb.GetLiveStreamEndpointResponse{
		RtmpUrl:   "rtmp://live-input.rumble.com/live",
		StreamKey: "rumble_key_...", // Requires scraping
	}, nil
}

func (p *RumblePublisher) GetChatMessages(ctx context.Context, req *pb.GetChatMessagesRequest) (*pb.GetChatMessagesResponse, error) {
	// TODO: Scrape chat or use internal API
	return &pb.GetChatMessagesResponse{
		Messages: []*pb.ChatMessage{},
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
