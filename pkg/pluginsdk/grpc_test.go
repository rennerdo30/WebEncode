package pluginsdk

import (
	"context"
	"net"
	"testing"

	api "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

// MockEncoder
type MockEncoder struct {
	api.UnimplementedEncoderServiceServer
}

func (m *MockEncoder) GetCapabilities(ctx context.Context, req *api.Empty) (*api.Capabilities, error) {
	return &api.Capabilities{
		VideoCodecs: []string{"h264"},
	}, nil
}

func (m *MockEncoder) Transcode(req *api.TranscodeRequest, stream api.EncoderService_TranscodeServer) error {
	stream.Send(&api.TranscodeProgress{Percent: 50})
	stream.Send(&api.TranscodeProgress{Percent: 100, Completed: true})
	return nil
}

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	api.RegisterEncoderServiceServer(s, &MockEncoder{})
	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestEncoderGRPC(t *testing.T) {
	// Client
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	assert.NoError(t, err)
	defer conn.Close()

	client := api.NewEncoderServiceClient(conn)

	// Test GetCapabilities
	caps, err := client.GetCapabilities(ctx, &api.Empty{})
	assert.NoError(t, err)
	assert.Equal(t, "h264", caps.VideoCodecs[0])

	// Test Transcode
	stream, err := client.Transcode(ctx, &api.TranscodeRequest{SourceUrl: "test"})
	assert.NoError(t, err)

	p1, err := stream.Recv()
	assert.NoError(t, err)
	assert.Equal(t, float32(50), p1.Percent)

	p2, err := stream.Recv()
	assert.NoError(t, err)
	assert.True(t, p2.Completed)
}

func TestPlugin_GRPCServer(t *testing.T) {
	// Create a Plugin with a mock encoder
	p := &Plugin{
		EncoderImpl: &MockEncoder{},
	}

	// Create a GRPC server
	s := grpc.NewServer()

	// Register via our Plugin.GRPCServer method
	err := p.GRPCServer(nil, s)
	assert.NoError(t, err)

	// In a real scenario we'd query the server info, but here we trust no error means registered.
	// We can verify by checking if ServiceInfo contains EncoderService
	info := s.GetServiceInfo()
	_, ok := info["v1.EncoderService"]
	if !ok {
		for k := range info {
			t.Logf("Registered service: %s", k)
		}
	}
	assert.True(t, ok, "EncoderService should be registered")
}

func TestPlugin_GRPCClient(t *testing.T) {
	p := &Plugin{}
	ctx := context.Background()
	// Mock client conn (we don't need it connected for this test)
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	assert.NoError(t, err)
	defer conn.Close()

	raw, err := p.GRPCClient(ctx, nil, conn)
	assert.NoError(t, err)

	res, ok := raw.(*PluginResult)
	assert.True(t, ok)
	assert.NotNil(t, res.Encoder)
	assert.NotNil(t, res.Auth)
	assert.NotNil(t, res.Storage)
	assert.NotNil(t, res.Live)
	assert.NotNil(t, res.Publisher)
}
