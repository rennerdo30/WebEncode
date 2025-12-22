package pluginsdk

import (
	"context"

	"google.golang.org/grpc/health/grpc_health_v1"
)

// HealthCheckServer implements the standard gRPC health check service.
// This is required by go-plugin to verify the subprocess is alive.
type HealthCheckServer struct {
	grpc_health_v1.UnimplementedHealthServer
}

func (s *HealthCheckServer) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

func (s *HealthCheckServer) Watch(req *grpc_health_v1.HealthCheckRequest, server grpc_health_v1.Health_WatchServer) error {
	// Send initial status
	err := server.Send(&grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	})
	if err != nil {
		return err
	}
	// Keep connection open but don't stream updates for now
	<-server.Context().Done()
	return nil
}
