package main

import (
	"context"
	"testing"

	pb "github.com/rennerdo30/webencode/pkg/api/v1"
)

func TestNewGenericRTMPPublisher(t *testing.T) {
	p := NewGenericRTMPPublisher()
	if p == nil {
		t.Fatal("expected non-nil publisher")
	}
	if p.logger == nil {
		t.Error("expected logger to be initialized")
	}
}

func TestPublish_NotSupported(t *testing.T) {
	p := NewGenericRTMPPublisher()
	ctx := context.Background()

	result, err := p.Publish(ctx, &pb.PublishRequest{
		Platform: "rtmp",
		Title:    "Test",
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result.PlatformId != "rtmp" {
		t.Errorf("expected platform_id 'rtmp', got %s", result.PlatformId)
	}
}

func TestRetract(t *testing.T) {
	p := NewGenericRTMPPublisher()
	ctx := context.Background()

	_, err := p.Retract(ctx, &pb.RetractRequest{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGetLiveStreamEndpoint(t *testing.T) {
	p := NewGenericRTMPPublisher()
	ctx := context.Background()

	testCases := []struct {
		name        string
		accessToken string
		wantURL     string
	}{
		{
			name:        "full rtmp url",
			accessToken: "rtmp://live.example.com/app/streamkey123",
			wantURL:     "rtmp://live.example.com/app/streamkey123",
		},
		{
			name:        "custom server",
			accessToken: "rtmp://192.168.1.100:1935/live/mystream",
			wantURL:     "rtmp://192.168.1.100:1935/live/mystream",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := p.GetLiveStreamEndpoint(ctx, &pb.GetLiveStreamEndpointRequest{
				AccessToken: tc.accessToken,
			})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if resp.RtmpUrl != tc.wantURL {
				t.Errorf("expected rtmp_url %s, got %s", tc.wantURL, resp.RtmpUrl)
			}
		})
	}
}

func TestGetChatMessages_Empty(t *testing.T) {
	p := NewGenericRTMPPublisher()
	ctx := context.Background()

	resp, err := p.GetChatMessages(ctx, &pb.GetChatMessagesRequest{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(resp.Messages) != 0 {
		t.Errorf("expected empty messages, got %d", len(resp.Messages))
	}
}

func TestSendChatMessage(t *testing.T) {
	p := NewGenericRTMPPublisher()
	ctx := context.Background()

	_, err := p.SendChatMessage(ctx, &pb.SendChatMessageRequest{
		Message: "test",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
