package pluginsdk

import (
	"context"

	"github.com/hashicorp/go-plugin"
	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"google.golang.org/grpc"
)

// Plugin implements the gRPC side of the plugin system.
type Plugin struct {
	plugin.NetRPCUnsupportedPlugin
	// One of these will be non-nil on the server side depending on what the plugin implements
	AuthImpl      pb.AuthServiceServer
	StorageImpl   pb.StorageServiceServer
	EncoderImpl   pb.EncoderServiceServer
	LiveImpl      pb.LiveServiceServer
	PublisherImpl pb.PublisherServiceServer
}

func (p *Plugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &PluginResult{
		Auth:      pb.NewAuthServiceClient(c),
		Storage:   pb.NewStorageServiceClient(c),
		Encoder:   pb.NewEncoderServiceClient(c),
		Live:      pb.NewLiveServiceClient(c),
		Publisher: pb.NewPublisherServiceClient(c),
	}, nil
}

func (p *Plugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	// Note: Health check is automatically registered by go-plugin's DefaultGRPCServer
	// Do not register grpc_health_v1.RegisterHealthServer here as it causes duplicate registration

	if p.AuthImpl != nil {
		pb.RegisterAuthServiceServer(s, p.AuthImpl)
	}
	if p.StorageImpl != nil {
		pb.RegisterStorageServiceServer(s, p.StorageImpl)
	}
	if p.EncoderImpl != nil {
		pb.RegisterEncoderServiceServer(s, p.EncoderImpl)
	}
	if p.LiveImpl != nil {
		pb.RegisterLiveServiceServer(s, p.LiveImpl)
	}
	if p.PublisherImpl != nil {
		pb.RegisterPublisherServiceServer(s, p.PublisherImpl)
	}
	return nil
}

// PluginResult is a flexible container for whatever client interface we got
type PluginResult struct {
	Auth      pb.AuthServiceClient
	Storage   pb.StorageServiceClient
	Encoder   pb.EncoderServiceClient
	Live      pb.LiveServiceClient
	Publisher pb.PublisherServiceClient
}
