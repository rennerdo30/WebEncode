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

func TestStartIngest(t *testing.T) {
	p := NewMediaMTXPlugin()

	req := &pb.IngestConfig{StreamKey: "test-stream"}
	resp, err := p.StartIngest(context.Background(), req)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Id)
	assert.Contains(t, resp.IngestUrl, "test-stream")
	assert.Contains(t, resp.PlaybackUrl, "test-stream")
}

func TestGetTelemetry(t *testing.T) {
	// Mock MediaMTX API
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/paths/list", r.URL.Path)

		// resp := PathListResponse{}
		// Needs to match the struct definition in main.go
		// We have to redefine it or make it public. It IS public in main.go

		// Wait, we need to populate generic items
		// Can't use struct literal easily if structs define inline...
		// Using map for JSON marshal simplicity

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"items": [
				{
					"name": "live/test-stream",
					"source": { "type": "rtmp" },
					"ready": true,
					"readers": [{}, {}]
				}
			]
		}`))
	}))
	defer ts.Close()

	os.Setenv("MEDIAMTX_API_URL", ts.URL)
	p := NewMediaMTXPlugin()

	// Test Match
	telem, err := p.GetTelemetry(context.Background(), &pb.SessionID{Id: "test-stream"})
	assert.NoError(t, err)
	assert.True(t, telem.IsLive)
	assert.Equal(t, int64(2), telem.Viewers)

	// Test No Match
	telem, err = p.GetTelemetry(context.Background(), &pb.SessionID{Id: "missing"})
	assert.NoError(t, err)
	assert.False(t, telem.IsLive)
}
