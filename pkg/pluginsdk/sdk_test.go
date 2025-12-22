package pluginsdk

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func TestHealthCheckServer_Check(t *testing.T) {
	s := &HealthCheckServer{}
	req := &grpc_health_v1.HealthCheckRequest{}

	resp, err := s.Check(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, grpc_health_v1.HealthCheckResponse_SERVING, resp.Status)
}

type mockWatchServer struct {
	grpc_health_v1.Health_WatchServer
	sent chan *grpc_health_v1.HealthCheckResponse
	ctx  context.Context
}

func (m *mockWatchServer) Send(resp *grpc_health_v1.HealthCheckResponse) error {
	m.sent <- resp
	return nil
}

func (m *mockWatchServer) Context() context.Context {
	return m.ctx
}

func TestHealthCheckServer_Watch(t *testing.T) {
	s := &HealthCheckServer{}
	req := &grpc_health_v1.HealthCheckRequest{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sent := make(chan *grpc_health_v1.HealthCheckResponse, 1)
	mockServer := &mockWatchServer{
		sent: sent,
		ctx:  ctx,
	}

	// Run Watch in goroutine
	errChan := make(chan error)
	go func() {
		errChan <- s.Watch(req, mockServer)
	}()

	// Should receive initial status
	select {
	case resp := <-sent:
		assert.Equal(t, grpc_health_v1.HealthCheckResponse_SERVING, resp.Status)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for initial status")
	}

	// Cancel context to stop Watch
	cancel()

	select {
	case err := <-errChan:
		assert.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for Watch to return")
	}
}
