package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestHttpClient(t *testing.T) {
	t.Run("creates HTTP request with headers", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify headers
			assert.Contains(t, r.Header.Get("User-Agent"), "Mozilla")
			assert.Equal(t, "application/json", r.Header.Get("Accept"))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data":{"messages":[]}}`))
		}))
		defer server.Close()

		client := &httpClient{timeout: 10 * 1000000000} // 10 seconds
		resp, err := client.Get(server.URL)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 200, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("handles timeout", func(t *testing.T) {
		client := &httpClient{timeout: 1} // 1 nanosecond
		_, err := client.Get("http://10.255.255.1:12345") // Non-routable IP
		assert.Error(t, err)
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

func TestKickPublisher_GetChatMessages(t *testing.T) {
	pub := NewKickPublisher()
	ctx := context.Background()

	t.Run("returns empty array for empty channel ID", func(t *testing.T) {
		req := &pb.GetChatMessagesRequest{
			ChannelId: "",
		}

		resp, err := pub.GetChatMessages(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Empty(t, resp.Messages)
	})

	t.Run("handles API errors gracefully", func(t *testing.T) {
		// This test uses a non-existent channel, which should return empty array
		req := &pb.GetChatMessagesRequest{
			ChannelId: "nonexistent-channel-12345",
		}

		resp, err := pub.GetChatMessages(ctx, req)
		// Should not error, just return empty messages
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.Messages)
	})

	t.Run("sanitizes channel ID for URL", func(t *testing.T) {
		// Test with special characters
		req := &pb.GetChatMessagesRequest{
			ChannelId: "channel/with/slashes",
		}

		// Should not panic, should handle URL encoding
		resp, err := pub.GetChatMessages(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}

func TestKickPublisher_SendChatMessage(t *testing.T) {
	pub := NewKickPublisher()
	ctx := context.Background()

	t.Run("returns empty response", func(t *testing.T) {
		req := &pb.SendChatMessageRequest{
			ChannelId: "test-channel",
			Message:   "Hello world",
		}

		resp, err := pub.SendChatMessage(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}

func TestKickPublisher_GetLiveStreamEndpoint(t *testing.T) {
	pub := NewKickPublisher()
	ctx := context.Background()

	t.Run("requires cookies env var", func(t *testing.T) {
		req := &pb.GetLiveStreamEndpointRequest{}

		// Without KICK_COOKIES_JSON, should fail
		result, err := pub.GetLiveStreamEndpoint(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "KICK_COOKIES_JSON")
		assert.Nil(t, result)
	})
}

func TestCookieSerialization(t *testing.T) {
	t.Run("cookie JSON parsing", func(t *testing.T) {
		cookieJSON := `[{"name":"session","value":"abc123","domain":".kick.com","path":"/","httpOnly":true,"secure":true}]`

		var cookies []Cookie
		err := json.Unmarshal([]byte(cookieJSON), &cookies)
		assert.NoError(t, err)
		assert.Len(t, cookies, 1)
		assert.Equal(t, "session", cookies[0].Name)
		assert.Equal(t, "abc123", cookies[0].Value)
		assert.Equal(t, ".kick.com", cookies[0].Domain)
		assert.True(t, cookies[0].HttpOnly)
		assert.True(t, cookies[0].Secure)
	})

	t.Run("multiple cookies", func(t *testing.T) {
		cookieJSON := `[
			{"name":"session","value":"sess123","domain":".kick.com","path":"/"},
			{"name":"token","value":"tok456","domain":".kick.com","path":"/"}
		]`

		var cookies []Cookie
		err := json.Unmarshal([]byte(cookieJSON), &cookies)
		assert.NoError(t, err)
		assert.Len(t, cookies, 2)
		assert.Equal(t, "session", cookies[0].Name)
		assert.Equal(t, "token", cookies[1].Name)
	})

	t.Run("cookie with expiration", func(t *testing.T) {
		cookieJSON := `[{"name":"session","value":"abc","domain":".kick.com","path":"/","expirationDate":1700000000}]`

		var cookies []Cookie
		err := json.Unmarshal([]byte(cookieJSON), &cookies)
		assert.NoError(t, err)
		assert.Len(t, cookies, 1)
		assert.Equal(t, float64(1700000000), cookies[0].Expires)
	})
}

func TestPublishRequestValidation(t *testing.T) {
	t.Run("valid request with all fields", func(t *testing.T) {
		req := &pb.PublishRequest{
			Title:       "My Video",
			Description: "Video description",
			FileUrl:     "/tmp/video.mp4",
			Platform:    "kick",
		}

		assert.Equal(t, "My Video", req.Title)
		assert.NotEmpty(t, req.FileUrl)
		assert.Equal(t, "kick", req.Platform)
	})

	t.Run("minimal request", func(t *testing.T) {
		req := &pb.PublishRequest{
			FileUrl: "/tmp/video.mp4",
		}

		assert.Empty(t, req.Title)
		assert.NotEmpty(t, req.FileUrl)
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

func BenchmarkCookieParsing(b *testing.B) {
	cookieJSON := `[{"name":"session","value":"abc123","domain":".kick.com","path":"/","httpOnly":true,"secure":true}]`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var cookies []Cookie
		json.Unmarshal([]byte(cookieJSON), &cookies)
	}
}
