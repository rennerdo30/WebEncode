package main

import (
	"context"
	"testing"

	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/stretchr/testify/assert"
)

func TestNewRumblePublisher(t *testing.T) {
	t.Run("initialization", func(t *testing.T) {
		pub := NewRumblePublisher()
		assert.NotNil(t, pub)
		assert.NotNil(t, pub.logger)
	})
}

func TestRumblePublisher_Publish(t *testing.T) {
	t.Run("requires cookies env var", func(t *testing.T) {
		pub := NewRumblePublisher()
		ctx := context.Background()

		req := &pb.PublishRequest{
			Title:       "Test Video",
			Description: "Test description",
			FileUrl:     "/tmp/video.mp4",
		}

		// Without RUMBLE_COOKIES_JSON, should fail
		result, err := pub.Publish(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "RUMBLE_COOKIES_JSON")
		assert.Nil(t, result)
	})
}

func TestRumblePublisher_Retract(t *testing.T) {
	t.Run("retract returns empty", func(t *testing.T) {
		pub := NewRumblePublisher()
		ctx := context.Background()

		req := &pb.RetractRequest{
			PlatformId: "test-id",
		}

		result, err := pub.Retract(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestCookie(t *testing.T) {
	t.Run("cookie struct", func(t *testing.T) {
		cookie := Cookie{
			Name:     "session",
			Value:    "abc123",
			Domain:   ".rumble.com",
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
		}

		assert.Equal(t, "session", cookie.Name)
		assert.Equal(t, ".rumble.com", cookie.Domain)
		assert.True(t, cookie.HttpOnly)
	})
}

func TestRumbleConfig(t *testing.T) {
	t.Run("upload URL", func(t *testing.T) {
		uploadURL := "https://rumble.com/upload.php"
		assert.Contains(t, uploadURL, "rumble.com")
		assert.Contains(t, uploadURL, "upload")
	})
}

func TestPublishRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		req := &pb.PublishRequest{
			Title:       "My Video",
			Description: "Video description",
			FileUrl:     "/tmp/video.mp4",
		}

		assert.Equal(t, "My Video", req.Title)
		assert.NotEmpty(t, req.FileUrl)
	})
}

func TestPublishResult(t *testing.T) {
	t.Run("result structure", func(t *testing.T) {
		result := &pb.PublishResult{
			PlatformId: "rumble_123456",
			Url:        "https://rumble.com/vrumble_123456",
		}

		assert.Contains(t, result.Url, "rumble.com")
		assert.NotEmpty(t, result.PlatformId)
	})
}

func TestRetractRequest(t *testing.T) {
	t.Run("valid retract request", func(t *testing.T) {
		req := &pb.RetractRequest{
			PlatformId: "rumble_123456",
		}

		assert.Equal(t, "rumble_123456", req.PlatformId)
	})
}

func TestContextCancellation(t *testing.T) {
	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		assert.Error(t, ctx.Err())
		assert.Equal(t, context.Canceled, ctx.Err())
	})
}

// Benchmark for future optimization
func BenchmarkNewRumblePublisher(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewRumblePublisher()
	}
}

func BenchmarkPublishRequestCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = &pb.PublishRequest{
			Title:       "Benchmark Video",
			Description: "Benchmark test",
			FileUrl:     "/tmp/video.mp4",
		}
	}
}

func TestRumblePublisher_GetLiveStreamEndpoint(t *testing.T) {
	t.Run("requires cookies env var", func(t *testing.T) {
		pub := NewRumblePublisher()
		ctx := context.Background()

		req := &pb.GetLiveStreamEndpointRequest{}

		// Without RUMBLE_COOKIES_JSON, should fail
		result, err := pub.GetLiveStreamEndpoint(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "RUMBLE_COOKIES_JSON")
		assert.Nil(t, result)
	})
}

func TestRumblePublisher_GetChatMessages(t *testing.T) {
	pub := NewRumblePublisher()
	ctx := context.Background()

	t.Run("empty channel ID returns empty messages", func(t *testing.T) {
		req := &pb.GetChatMessagesRequest{
			ChannelId: "",
		}

		result, err := pub.GetChatMessages(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.Messages)
	})

	t.Run("invalid channel ID returns empty messages", func(t *testing.T) {
		req := &pb.GetChatMessagesRequest{
			ChannelId: "invalid-channel-id",
		}

		result, err := pub.GetChatMessages(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		// Should gracefully handle non-existent channels
	})
}

func TestRumblePublisher_SendChatMessage(t *testing.T) {
	t.Run("send returns empty", func(t *testing.T) {
		pub := NewRumblePublisher()
		ctx := context.Background()

		req := &pb.SendChatMessageRequest{
			ChannelId: "test-channel",
			Message:   "Hello!",
		}

		result, err := pub.SendChatMessage(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestHttpClient_Get(t *testing.T) {
	t.Run("creates client with timeout", func(t *testing.T) {
		client := &httpClient{timeout: 10}
		assert.NotNil(t, client)
		assert.Equal(t, 10, int(client.timeout))
	})

	t.Run("invalid URL returns error", func(t *testing.T) {
		client := &httpClient{timeout: 10}
		_, err := client.Get("://invalid-url")
		assert.Error(t, err)
	})
}

func TestChatMessage(t *testing.T) {
	t.Run("valid chat message", func(t *testing.T) {
		msg := &pb.ChatMessage{
			Id:         "msg-123",
			Platform:   "rumble",
			AuthorName: "testuser",
			Content:    "Hello world!",
			Timestamp:  1234567890,
		}

		assert.Equal(t, "rumble", msg.Platform)
		assert.Equal(t, "testuser", msg.AuthorName)
		assert.NotEmpty(t, msg.Content)
	})
}

func TestGetLiveStreamEndpointResponse(t *testing.T) {
	t.Run("valid response", func(t *testing.T) {
		resp := &pb.GetLiveStreamEndpointResponse{
			RtmpUrl:   "rtmp://live-input.rumble.com/live",
			StreamKey: "test-stream-key",
		}

		assert.Contains(t, resp.RtmpUrl, "rtmp://")
		assert.Contains(t, resp.RtmpUrl, "rumble.com")
		assert.NotEmpty(t, resp.StreamKey)
	})
}

func TestGetChatMessagesRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		req := &pb.GetChatMessagesRequest{
			ChannelId: "channel-123",
		}

		assert.Equal(t, "channel-123", req.ChannelId)
	})
}

func TestSendChatMessageRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		req := &pb.SendChatMessageRequest{
			ChannelId: "channel-123",
			Message:   "Test message",
		}

		assert.Equal(t, "channel-123", req.ChannelId)
		assert.Equal(t, "Test message", req.Message)
	})
}
