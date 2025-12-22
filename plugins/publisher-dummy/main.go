package main

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-plugin"
	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/rennerdo30/webencode/pkg/pluginsdk"
)

type DummyPublisher struct {
	pb.UnimplementedPublisherServiceServer
	logger *logger.Logger
}

func NewDummyPublisher() *DummyPublisher {
	return &DummyPublisher{
		logger: logger.New("plugin-publisher-dummy"),
	}
}

func (p *DummyPublisher) Publish(ctx context.Context, req *pb.PublishRequest) (*pb.PublishResult, error) {
	p.logger.Info("Publish request", "platform", req.Platform, "title", req.Title)

	// Simulate upload/push time
	time.Sleep(2 * time.Second)

	// Return fake URL
	url := fmt.Sprintf("https://%s.com/video/%s", req.Platform, req.JobId)

	p.logger.Info("Publish complete", "url", url)

	return &pb.PublishResult{
		PlatformId: "dummy-id-" + req.JobId,
		Url:        url,
	}, nil
}

func (p *DummyPublisher) Retract(ctx context.Context, req *pb.RetractRequest) (*pb.Empty, error) {
	p.logger.Info("Retract request", "platform", req.Platform, "id", req.PlatformId)
	return &pb.Empty{}, nil
}

func (p *DummyPublisher) GetLiveStreamEndpoint(ctx context.Context, req *pb.GetLiveStreamEndpointRequest) (*pb.GetLiveStreamEndpointResponse, error) {
	return &pb.GetLiveStreamEndpointResponse{
		RtmpUrl:   "rtmp://localhost/dummy",
		StreamKey: "dummy_key",
	}, nil
}

func (p *DummyPublisher) GetChatMessages(ctx context.Context, req *pb.GetChatMessagesRequest) (*pb.GetChatMessagesResponse, error) {
	return &pb.GetChatMessagesResponse{
		Messages: []*pb.ChatMessage{
			{
				Id:         "msg1",
				Platform:   "dummy",
				AuthorName: "User1",
				Content:    "Hello World",
				Timestamp:  time.Now().Unix(),
			},
		},
	}, nil
}

func (p *DummyPublisher) SendChatMessage(ctx context.Context, req *pb.SendChatMessageRequest) (*pb.Empty, error) {
	p.logger.Info("Sending chat message", "content", req.Message)
	return &pb.Empty{}, nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: pluginsdk.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"webencode": &pluginsdk.Plugin{
				PublisherImpl: NewDummyPublisher(),
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
