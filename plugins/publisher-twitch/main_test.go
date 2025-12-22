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

func TestPublish_Twitch(t *testing.T) {
	// Source mock
	sourceServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("fake-video-content"))
	}))
	defer sourceServer.Close()

	// Upload Server Mock
	uploadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify PUT upload
		if r.Method == "PUT" {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer uploadServer.Close()

	// API Mock
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Create Video
		if r.Method == "POST" && r.URL.Path == "/videos" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"data": [{
					"id": "11223344",
					"upload_url": "` + uploadServer.URL + `",
					"url": "https://twitch.tv/videos/11223344"
				}]
			}`))
			return
		}

		// 2. Complete Upload
		if r.Method == "POST" && r.URL.Path == "/videos/11223344/complete" {
			w.WriteHeader(http.StatusOK)
			return
		}

		http.NotFound(w, r)
	}))
	defer apiServer.Close()

	// Config
	os.Setenv("TWITCH_CLIENT_ID", "test-client")
	os.Setenv("TWITCH_API_URL", apiServer.URL)

	p := NewTwitchPublisher()

	req := &pb.PublishRequest{
		FileUrl:     sourceServer.URL + "/video.mp4",
		AccessToken: "fake-token",
		Title:       "Test Video",
		Description: "Test Description",
	}

	result, err := p.Publish(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, "11223344", result.PlatformId)
	assert.Contains(t, result.Url, "11223344")
}

func TestRetract_Twitch(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" && r.URL.Path == "/videos" {
			if r.URL.Query().Get("id") == "11223344" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		http.NotFound(w, r)
	}))
	defer apiServer.Close()

	os.Setenv("TWITCH_API_URL", apiServer.URL)
	p := NewTwitchPublisher()

	req := &pb.RetractRequest{
		PlatformId:  "11223344",
		AccessToken: "fake-token",
	}

	_, err := p.Retract(context.Background(), req)
	assert.NoError(t, err)
}

func TestGetLiveStreamEndpoint_Twitch(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/streams/key" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"data": [{
					"stream_key": "live_12345_abcde"
				}]
			}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer apiServer.Close()

	os.Setenv("TWITCH_API_URL", apiServer.URL)
	p := NewTwitchPublisher()

	req := &pb.GetLiveStreamEndpointRequest{
		AccessToken: "fake-token",
	}

	res, err := p.GetLiveStreamEndpoint(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, "live_12345_abcde", res.StreamKey)
	assert.NotEmpty(t, res.RtmpUrl)
}
