package main

import (
	"context"
	"testing"
	"time"

	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/rennerdo30/webencode/pkg/ffmpeg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/metadata"
)

type MockEncoderBackend struct {
	mock.Mock
}

func (m *MockEncoderBackend) Transcode(ctx context.Context, task *ffmpeg.TranscodeTask, progressCh chan<- ffmpeg.Progress) error {
	args := m.Called(ctx, task, progressCh)
	// Simulate progress
	go func() {
		defer close(progressCh)
		progressCh <- ffmpeg.Progress{Percent: 50}
	}()
	return args.Error(0)
}

type MockTranscodeServer struct {
	mock.Mock
	pb.EncoderService_TranscodeServer
}

func (m *MockTranscodeServer) Send(p *pb.TranscodeProgress) error {
	args := m.Called(p)
	return args.Error(0)
}

func (m *MockTranscodeServer) Context() context.Context {
	return context.Background()
}

func (m *MockTranscodeServer) SetHeader(md metadata.MD) error  { return nil }
func (m *MockTranscodeServer) SendHeader(md metadata.MD) error { return nil }
func (m *MockTranscodeServer) SetTrailer(md metadata.MD)       {}

func TestGetCapabilities(t *testing.T) {
	e := &FFmpegEncoder{encoder: new(MockEncoderBackend)}
	caps, err := e.GetCapabilities(context.Background(), &pb.Empty{})
	assert.NoError(t, err)
	assert.Contains(t, caps.VideoCodecs, "h264")
}

func TestTranscode(t *testing.T) {
	mockBackend := new(MockEncoderBackend)
	e := &FFmpegEncoder{encoder: mockBackend}

	req := &pb.TranscodeRequest{
		SourceUrl: "/tmp/in.mp4",
		OutputUrl: "/tmp/out.mp4",
		Profile: &pb.TranscodeProfile{
			VideoCodec: "h264",
			Width:      1920,
			Height:     1080,
		},
	}

	mockBackend.On("Transcode", mock.Anything, mock.MatchedBy(func(task *ffmpeg.TranscodeTask) bool {
		return task.InputURL == "/tmp/in.mp4" && task.Width == 1920
	}), mock.Anything).Return(nil)

	mockStream := new(MockTranscodeServer)

	// Expect 50% progress
	mockStream.On("Send", mock.MatchedBy(func(p *pb.TranscodeProgress) bool {
		return p.Percent == 50
	})).Return(nil).Once()

	// Expect 100% completion
	mockStream.On("Send", mock.MatchedBy(func(p *pb.TranscodeProgress) bool {
		return p.Completed == true
	})).Return(nil).Once()

	err := e.Transcode(req, mockStream)
	assert.NoError(t, err)

	// Wait briefly for goroutine
	time.Sleep(10 * time.Millisecond)
	mockBackend.AssertExpectations(t)
	mockStream.AssertExpectations(t)
}
