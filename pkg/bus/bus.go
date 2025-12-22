package bus

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/rennerdo30/webencode/pkg/logger"
)

type Bus struct {
	nc     *nats.Conn
	js     jetstream.JetStream
	logger *logger.Logger
}

const (
	StreamWork   = "WEBENCODE_WORK"
	StreamEvents = "WEBENCODE_EVENTS"
	StreamLive   = "WEBENCODE_LIVE"

	SubjectJobCreated      = "jobs.created"
	SubjectJobDispatch     = "jobs.dispatch"
	SubjectJobEvents       = "jobs.events"
	SubjectWorkerHeartbeat = "workers.heartbeat"
	SubjectErrorEvents     = "events.error"
	SubjectAuditUser       = "audit.user_action"
	SubjectAuditSystem     = "audit.system"
	SubjectLiveTelemetry   = "live.telemetry" // Pattern: live.telemetry.{stream_id}
	SubjectLiveLifecycle   = "live.lifecycle" // Pattern: live.lifecycle.{stream_id}
)

func Connect(url string, l *logger.Logger) (*Bus, error) {
	nc, err := nats.Connect(url, nats.Name("webencode-kernel"), nats.RetryOnFailedConnect(true), nats.MaxReconnects(-1))
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("jetstream init: %w", err)
	}

	b := &Bus{nc: nc, js: js, logger: l}
	return b, nil
}

func (b *Bus) InitStreams(ctx context.Context) error {
	// WorkQueue Stream for Jobs
	_, err := b.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:      StreamWork,
		Subjects:  []string{"jobs.*", "tasks.*"},
		Retention: jetstream.WorkQueuePolicy,
	})
	if err != nil {
		return fmt.Errorf("create stream %s: %w", StreamWork, err)
	}

	// Fanout Stream for Events (Logs, Progress, Webhooks)
	_, err = b.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:      StreamEvents,
		Subjects:  []string{"events.*", "workers.*", "audit.*"},
		Retention: jetstream.LimitsPolicy,
		MaxAge:    90 * 24 * time.Hour, // 90 days for audit
	})
	if err != nil {
		return fmt.Errorf("create stream %s: %w", StreamEvents, err)
	}

	// Live Stream for Telemetry (ephemeral, memory-backed)
	_, err = b.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:      StreamLive,
		Subjects:  []string{"live.telemetry.*", "live.lifecycle.*"},
		Retention: jetstream.LimitsPolicy,
		MaxAge:    10 * time.Second, // Ephemeral - only recent data
		Storage:   jetstream.MemoryStorage,
	})
	if err != nil {
		return fmt.Errorf("create stream %s: %w", StreamLive, err)
	}

	return nil
}

func (b *Bus) Publish(ctx context.Context, subject string, data []byte) error {
	_, err := b.js.Publish(ctx, subject, data)
	return err
}

func (b *Bus) Close() {
	b.nc.Close()
}

func (b *Bus) JetStream() jetstream.JetStream {
	return b.js
}

func (b *Bus) Conn() *nats.Conn {
	return b.nc
}
