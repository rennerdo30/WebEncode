package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/stretchr/testify/assert"
)

func TestPublish_YouTube(t *testing.T) {
	// 1. Mock Video Source
	sourceServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("fake-video-content"))
	}))
	defer sourceServer.Close()

	// 2. Mock Google API
	googleServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock Upload
		if r.Method == "POST" && r.URL.Path == "/upload/youtube/v3/videos" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"id": "yt-123456",
				"snippet": { "title": "Test Video" },
				"status": { "privacyStatus": "private" }
			}`))
			return
		}

		http.NotFound(w, r)
	}))
	defer googleServer.Close()

	// Setup Env
	os.Setenv("YOUTUBE_CLIENT_ID", "test-client")
	os.Setenv("YOUTUBE_CLIENT_SECRET", "test-secret")
	os.Setenv("YOUTUBE_API_ENDPOINT", googleServer.URL) // Use Mock

	p := NewYouTubePublisher()

	req := &pb.PublishRequest{
		FileUrl:     sourceServer.URL + "/video.mp4",
		AccessToken: "fake-token",
		Title:       "Test Video",
		Description: "Test Description",
	}

	result, err := p.Publish(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, "yt-123456", result.PlatformId)
	assert.Contains(t, result.Url, "yt-123456")
}

func TestRetract_YouTube(t *testing.T) {
	// Mock Google API
	googleServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" && r.URL.Path == "/youtube/v3/videos" {
			// Query param id=...
			if r.URL.Query().Get("id") == "yt-123456" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		http.NotFound(w, r)
	}))
	defer googleServer.Close()

	os.Setenv("YOUTUBE_API_ENDPOINT", googleServer.URL)
	p := NewYouTubePublisher()

	req := &pb.RetractRequest{
		PlatformId:  "yt-123456",
		AccessToken: "fake-token",
	}

	_, err := p.Retract(context.Background(), req)
	assert.NoError(t, err)
}
