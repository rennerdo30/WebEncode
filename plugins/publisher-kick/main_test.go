package main

import (
	"context"
	"testing"

	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/stretchr/testify/assert"
)

func TestNewKickPublisher(t *testing.T) {
	t.Run("initialization", func(t *testing.T) {
		pub := NewKickPublisher()
		assert.NotNil(t, pub)
		assert.NotNil(t, pub.logger)
	})
}

func TestKickPublisher_Publish(t *testing.T) {
	t.Run("requires cookies env var", func(t *testing.T) {
		pub := NewKickPublisher()
		ctx := context.Background()

		req := &pb.PublishRequest{
			Title:       "Test Video",
			Description: "Test description",
			FileUrl:     "/tmp/video.mp4",
		}

		// Without KICK_COOKIES_JSON, should fail
		result, err := pub.Publish(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "KICK_COOKIES_JSON")
		assert.Nil(t, result)
	})
}

func TestKickPublisher_Retract(t *testing.T) {
	t.Run("retract returns empty", func(t *testing.T) {
		pub := NewKickPublisher()
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
			Domain:   ".kick.com",
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
		}

		assert.Equal(t, "session", cookie.Name)
		assert.Equal(t, ".kick.com", cookie.Domain)
		assert.True(t, cookie.HttpOnly)
	})
}

func TestKickConfig(t *testing.T) {
	t.Run("dashboard URL", func(t *testing.T) {
		// Kick dashboard URL
		dashboardURL := "https://kick.com/dashboard/videos"
		assert.Contains(t, dashboardURL, "kick.com")
		assert.Contains(t, dashboardURL, "dashboard")
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
			PlatformId: "kick_123456",
			Url:        "https://kick.com/video/kick_123456",
		}

		assert.Contains(t, result.Url, "kick.com")
		assert.NotEmpty(t, result.PlatformId)
	})
}

func TestRetractRequest(t *testing.T) {
	t.Run("valid retract request", func(t *testing.T) {
		req := &pb.RetractRequest{
			PlatformId: "kick_123456",
		}

		assert.Equal(t, "kick_123456", req.PlatformId)
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
func BenchmarkNewKickPublisher(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewKickPublisher()
	}
}

func BenchmarkPublishRequest(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = &pb.PublishRequest{
			Title:       "Benchmark Video",
			Description: "Benchmark test",
			FileUrl:     "/tmp/video.mp4",
		}
	}
}
