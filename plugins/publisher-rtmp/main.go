package main

import (
	"context"

	"github.com/hashicorp/go-plugin"
	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/rennerdo30/webencode/pkg/pluginsdk"
)

// GenericRTMPPublisher handles publishing to any RTMP endpoint
type GenericRTMPPublisher struct {
	pb.UnimplementedPublisherServiceServer
	logger *logger.Logger
}

func NewGenericRTMPPublisher() *GenericRTMPPublisher {
	return &GenericRTMPPublisher{
		logger: logger.New("plugin-publisher-rtmp"),
	}
}

// Publish is not used for live streaming - this plugin is for live RTMP relay only
func (p *GenericRTMPPublisher) Publish(ctx context.Context, req *pb.PublishRequest) (*pb.PublishResult, error) {
	// Generic RTMP doesn't support VOD upload
	p.logger.Info("Publish called on generic RTMP - not supported for VOD")
	return &pb.PublishResult{
		PlatformId: "rtmp",
		Url:        "N/A - Generic RTMP is for live streaming only",
	}, nil
}

func (p *GenericRTMPPublisher) Retract(ctx context.Context, req *pb.RetractRequest) (*pb.Empty, error) {
	// Nothing to retract for generic RTMP
	return &pb.Empty{}, nil
}

// GetLiveStreamEndpoint returns the user-configured RTMP URL and stream key
// For generic RTMP, the access_token field contains the full RTMP URL with key
// Format: "rtmp://server.com/app/streamkey"
func (p *GenericRTMPPublisher) GetLiveStreamEndpoint(ctx context.Context, req *pb.GetLiveStreamEndpointRequest) (*pb.GetLiveStreamEndpointResponse, error) {
	// For generic RTMP, the access_token contains the full RTMP URL
	// We parse it to extract the base URL and stream key
	rtmpURL := req.AccessToken

	// The access_token should be in format: "rtmp://server/app/key"
	// We return it as-is since MediaMTX can use the full URL
	p.logger.Info("GetLiveStreamEndpoint called", "rtmp_url", rtmpURL)

	// Split the URL to get base and key
	// Most RTMP URLs are: rtmp://server/app/streamkey
	// We'll return the full URL as the rtmp_url and empty key
	// since the key is embedded in the URL
	return &pb.GetLiveStreamEndpointResponse{
		RtmpUrl:   rtmpURL,
		StreamKey: "", // Key is embedded in URL
	}, nil
}

// GetChatMessages - Generic RTMP doesn't support chat
func (p *GenericRTMPPublisher) GetChatMessages(ctx context.Context, req *pb.GetChatMessagesRequest) (*pb.GetChatMessagesResponse, error) {
	return &pb.GetChatMessagesResponse{
		Messages: []*pb.ChatMessage{},
	}, nil
}

// SendChatMessage - Generic RTMP doesn't support chat
func (p *GenericRTMPPublisher) SendChatMessage(ctx context.Context, req *pb.SendChatMessageRequest) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: pluginsdk.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"webencode": &pluginsdk.Plugin{
				PublisherImpl: NewGenericRTMPPublisher(),
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
