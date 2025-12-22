package live

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rennerdo30/webencode/internal/plugin_manager"
	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/logger"
)

// Publisher defines the interface for publishing events
type Publisher interface {
	Publish(ctx context.Context, subject string, data []byte) error
}

// RestreamDestination defines the JSON structure for stream destinations
type RestreamDestination struct {
	PluginID    string `json:"plugin_id"`
	AccessToken string `json:"access_token"`
	Enabled     bool   `json:"enabled"`
	URL         string `json:"url"` // Manual RTMP URL override
}

// MonitorService polls active streams and publishes telemetry
type MonitorService struct {
	db            store.Querier
	bus           Publisher
	pm            *plugin_manager.Manager
	logger        *logger.Logger
	stopCh        chan struct{}
	running       bool
	activeStreams map[string]bool // streamID -> isLive
}

func NewMonitorService(db store.Querier, b Publisher, pm *plugin_manager.Manager, l *logger.Logger) *MonitorService {
	return &MonitorService{
		db:            db,
		bus:           b,
		pm:            pm,
		logger:        l,
		stopCh:        make(chan struct{}),
		activeStreams: make(map[string]bool),
	}
}

func (s *MonitorService) Start() {
	if s.running {
		return
	}
	s.running = true
	go s.loop()
	s.logger.Info("Live Monitor Service started")
}

func (s *MonitorService) Stop() {
	if !s.running {
		return
	}
	close(s.stopCh)
	s.running = false
	s.logger.Info("Live Monitor Service stopped")
}

func (s *MonitorService) loop() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.poll()
		}
	}
}

func (s *MonitorService) poll() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Get all streams to check their status
	// Ideally we should filtering by "expected to be live" or similar for scale,
	// but list all is fine for now.
	streams, err := s.db.ListStreams(ctx, store.ListStreamsParams{Limit: 100, Offset: 0})
	if err != nil {
		s.logger.Error("Failed to list streams", "error", err)
		return
	}

	for _, stream := range streams {
		streamID := stream.ID.String()

		// 2. Get Live Plugin (MediaMTX)
		// We use the default live plugin "live-mediamtx"
		rawPlugin, ok := s.pm.Live["live-mediamtx"]
		if !ok {
			// Try to find ANY live plugin
			for _, p := range s.pm.Live {
				rawPlugin = p
				break
			}
			if rawPlugin == nil {
				continue
			}
		}

		livePlugin, ok := rawPlugin.(pb.LiveServiceClient)
		if !ok {
			continue
		}

		// 3. Get Telemetry
		telem, err := livePlugin.GetTelemetry(ctx, &pb.SessionID{Id: stream.StreamKey})
		if err != nil {
			s.logger.Warn("Failed to get telemetry", "stream_id", streamID, "error", err)
			continue
		}

		// 4. Update internal state and handle transitions
		wasLive := s.activeStreams[streamID]
		isLive := telem.IsLive

		if isLive && !wasLive {
			s.logger.Info("Stream started", "stream_id", streamID)
			go s.handleStreamStart(stream)
		} else if !isLive && wasLive {
			s.logger.Info("Stream ended", "stream_id", streamID)
			go s.handleStreamEnd(stream)
		}

		s.activeStreams[streamID] = isLive

		// 5. Publish to NATS if live
		if isLive {
			payload := map[string]interface{}{
				"stream_id": streamID,
				"fps":       telem.Fps,
				"bitrate":   telem.Bitrate,
				"viewers":   telem.Viewers,
				"timestamp": time.Now(),
			}

			data, _ := json.Marshal(payload)
			subject := fmt.Sprintf("live.telemetry.%s", streamID)
			s.bus.Publish(ctx, subject, data)

			// Also update DB status to reflect reality
			if !stream.IsLive {
				s.db.UpdateStreamLive(ctx, store.UpdateStreamLiveParams{
					ID:     stream.ID,
					IsLive: true,
				})
			}
		} else {
			if stream.IsLive {
				s.db.UpdateStreamLive(ctx, store.UpdateStreamLiveParams{
					ID:     stream.ID,
					IsLive: false,
				})
			}
		}
	}
}

func (s *MonitorService) publishEvent(stream store.Stream, eventType string) {
	// Publish event to the bus so webhooks can pick it up
	payload := map[string]interface{}{
		"stream_id":  stream.ID.String(),
		"stream_key": stream.StreamKey,
		"user_id":    stream.UserID.String(),
		"title":      stream.Title.String,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		s.logger.Error("Failed to marshal stream event payload", "error", err)
		return
	}

	// Subject format: events.stream.started
	subject := fmt.Sprintf("events.%s", eventType)
	if err := s.bus.Publish(context.Background(), subject, data); err != nil {
		s.logger.Error("Failed to publish stream event", "subject", subject, "error", err)
	} else {
		s.logger.Info("Published stream event", "subject", subject)
	}
}

func (s *MonitorService) handleStreamStart(stream store.Stream) {
	// Publish generic event
	s.publishEvent(stream, "stream.started")

	// 5. Notify user
	title := "Stream Started"
	message := fmt.Sprintf("Your stream '%s' is now live!", stream.Title.String)
	if !stream.Title.Valid || stream.Title.String == "" {
		message = "Your stream is now live!"
	}

	_, err := s.db.CreateNotification(context.Background(), store.CreateNotificationParams{
		UserID:  stream.UserID,
		Title:   title,
		Message: message,
		Type:    pgtype.Text{String: "info", Valid: true},
		Link:    pgtype.Text{String: fmt.Sprintf("/studio"), Valid: true}, // Assuming studio link
	})
	if err != nil {
		s.logger.Error("Failed to create stream start notification", "error", err)
	}

	if len(stream.RestreamDestinations) == 0 {
		return
	}

	var dests []RestreamDestination
	if err := json.Unmarshal(stream.RestreamDestinations, &dests); err != nil {
		s.logger.Error("Failed to parse restream destinations", "stream_id", stream.ID, "error", err)
		return
	}

	for _, dest := range dests {
		if !dest.Enabled {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var targetURL string

		if dest.URL != "" {
			targetURL = dest.URL
		} else {
			// 1. Get Publisher Plugin
			rawPub, ok := s.pm.Publisher[dest.PluginID]
			if !ok {
				s.logger.Error("Publisher plugin not found", "plugin_id", dest.PluginID)
				continue
			}
			pubPlugin, ok := rawPub.(pb.PublisherServiceClient)
			if !ok {
				continue
			}

			// 2. Resolve RTMP Endpoint
			endpoint, err := pubPlugin.GetLiveStreamEndpoint(ctx, &pb.GetLiveStreamEndpointRequest{
				AccessToken: dest.AccessToken,
			})
			if err != nil {
				s.logger.Error("Failed to get live endpoint", "plugin_id", dest.PluginID, "error", err)
				continue
			}

			targetURL = fmt.Sprintf("%s/%s", endpoint.RtmpUrl, endpoint.StreamKey)
			// Handle trailing slashes in URL if needed
			if endpoint.RtmpUrl[len(endpoint.RtmpUrl)-1] == '/' {
				targetURL = fmt.Sprintf("%s%s", endpoint.RtmpUrl, endpoint.StreamKey)
			}
		}

		// 3. Get Live Plugin (again, or pass it down)
		rawLive, ok := s.pm.Live["live-mediamtx"]
		if !ok {
			continue
		}
		livePlugin := rawLive.(pb.LiveServiceClient)

		// 4. Add Relay Target
		_, err = livePlugin.AddOutputTarget(ctx, &pb.AddOutputTargetRequest{
			SessionId: stream.StreamKey,
			TargetUrl: targetURL,
		})
		if err != nil {
			s.logger.Error("Failed to add output target", "target", targetURL, "error", err)
		} else {
			s.logger.Info("Started restream relay", "stream_id", stream.ID, "target", dest.PluginID)
		}
	}
}

func (s *MonitorService) handleStreamEnd(stream store.Stream) {
	// Publish generic event
	s.publishEvent(stream, "stream.ended")

	// Notify user
	title := "Stream Ended"
	message := fmt.Sprintf("Your stream '%s' has ended.", stream.Title.String)
	if !stream.Title.Valid || stream.Title.String == "" {
		message = "Your stream has ended."
	}

	_, err := s.db.CreateNotification(context.Background(), store.CreateNotificationParams{
		UserID:  stream.UserID,
		Title:   title,
		Message: message,
		Type:    pgtype.Text{String: "info", Valid: true},
		Link:    pgtype.Text{String: fmt.Sprintf("/studio"), Valid: true},
	})
	if err != nil {
		s.logger.Error("Failed to create stream end notification", "error", err)
	}
}
