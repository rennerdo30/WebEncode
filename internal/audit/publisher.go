package audit

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rennerdo30/webencode/pkg/bus"
	"github.com/rennerdo30/webencode/pkg/logger"
)

// EventType defines the type of audit event
type EventType string

const (
	// User actions
	EventTypeJobCreated      EventType = "job.created"
	EventTypeJobCancelled    EventType = "job.cancelled"
	EventTypeJobDeleted      EventType = "job.deleted"
	EventTypeStreamCreated   EventType = "stream.created"
	EventTypeStreamDeleted   EventType = "stream.deleted"
	EventTypeRestreamCreated EventType = "restream.created"
	EventTypeRestreamStarted EventType = "restream.started"
	EventTypeRestreamStopped EventType = "restream.stopped"
	EventTypeProfileCreated  EventType = "profile.created"
	EventTypeProfileDeleted  EventType = "profile.deleted"
	EventTypeWebhookCreated  EventType = "webhook.created"
	EventTypeWebhookDeleted  EventType = "webhook.deleted"
	EventTypePluginEnabled   EventType = "plugin.enabled"
	EventTypePluginDisabled  EventType = "plugin.disabled"
	EventTypeUserLogin       EventType = "user.login"
	EventTypeUserLogout      EventType = "user.logout"

	// System events
	EventTypeWorkerJoined   EventType = "worker.joined"
	EventTypeWorkerLeft     EventType = "worker.left"
	EventTypePluginCrashed  EventType = "plugin.crashed"
	EventTypePluginStarted  EventType = "plugin.started"
	EventTypeSystemStartup  EventType = "system.startup"
	EventTypeSystemShutdown EventType = "system.shutdown"
)

// Event represents an audit event
type Event struct {
	ID         string            `json:"id"`
	Type       EventType         `json:"type"`
	UserID     string            `json:"user_id,omitempty"`
	ResourceID string            `json:"resource_id,omitempty"`
	Action     string            `json:"action"`
	Details    map[string]string `json:"details,omitempty"`
	ClientIP   string            `json:"client_ip,omitempty"`
	UserAgent  string            `json:"user_agent,omitempty"`
	Timestamp  time.Time         `json:"timestamp"`
}

// EventBus defines the interface for publishing events
type EventBus interface {
	Publish(ctx context.Context, subject string, data []byte) error
}

// Publisher publishes audit events to NATS
type Publisher struct {
	bus    EventBus
	logger *logger.Logger
}

// NewPublisher creates a new audit event publisher
func NewPublisher(b EventBus, l *logger.Logger) *Publisher {
	return &Publisher{
		bus:    b,
		logger: l,
	}
}

// PublishUserAction publishes a user-initiated action
func (p *Publisher) PublishUserAction(ctx context.Context, event Event) error {
	event.Timestamp = time.Now()
	if event.ID == "" {
		event.ID = generateEventID()
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	subject := bus.SubjectAuditUser + "." + string(event.Type)
	if err := p.bus.Publish(ctx, subject, data); err != nil {
		p.logger.Error("Failed to publish audit event", "type", event.Type, "error", err)
		return err
	}

	p.logger.Debug("Audit event published", "type", event.Type, "user_id", event.UserID, "resource_id", event.ResourceID)
	return nil
}

// PublishSystemEvent publishes a system-initiated event
func (p *Publisher) PublishSystemEvent(ctx context.Context, event Event) error {
	event.Timestamp = time.Now()
	if event.ID == "" {
		event.ID = generateEventID()
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	subject := bus.SubjectAuditSystem + "." + string(event.Type)
	if err := p.bus.Publish(ctx, subject, data); err != nil {
		p.logger.Error("Failed to publish system audit event", "type", event.Type, "error", err)
		return err
	}

	p.logger.Debug("System audit event published", "type", event.Type)
	return nil
}

// Convenience methods for common events

func (p *Publisher) JobCreated(ctx context.Context, userID, jobID, sourceURL string) {
	_ = p.PublishUserAction(ctx, Event{
		Type:       EventTypeJobCreated,
		UserID:     userID,
		ResourceID: jobID,
		Action:     "create",
		Details: map[string]string{
			"source_url": sourceURL,
		},
	})
}

func (p *Publisher) JobCancelled(ctx context.Context, userID, jobID string) {
	_ = p.PublishUserAction(ctx, Event{
		Type:       EventTypeJobCancelled,
		UserID:     userID,
		ResourceID: jobID,
		Action:     "cancel",
	})
}

func (p *Publisher) JobDeleted(ctx context.Context, userID, jobID string) {
	_ = p.PublishUserAction(ctx, Event{
		Type:       EventTypeJobDeleted,
		UserID:     userID,
		ResourceID: jobID,
		Action:     "delete",
	})
}

func (p *Publisher) StreamCreated(ctx context.Context, userID, streamID, title string) {
	_ = p.PublishUserAction(ctx, Event{
		Type:       EventTypeStreamCreated,
		UserID:     userID,
		ResourceID: streamID,
		Action:     "create",
		Details: map[string]string{
			"title": title,
		},
	})
}

func (p *Publisher) WorkerJoined(ctx context.Context, workerID, hostname string) {
	_ = p.PublishSystemEvent(ctx, Event{
		Type:       EventTypeWorkerJoined,
		ResourceID: workerID,
		Action:     "join",
		Details: map[string]string{
			"hostname": hostname,
		},
	})
}

func (p *Publisher) WorkerLeft(ctx context.Context, workerID string) {
	_ = p.PublishSystemEvent(ctx, Event{
		Type:       EventTypeWorkerLeft,
		ResourceID: workerID,
		Action:     "leave",
	})
}

func (p *Publisher) PluginCrashed(ctx context.Context, pluginID string, errorMsg string) {
	_ = p.PublishSystemEvent(ctx, Event{
		Type:       EventTypePluginCrashed,
		ResourceID: pluginID,
		Action:     "crash",
		Details: map[string]string{
			"error": errorMsg,
		},
	})
}

func (p *Publisher) UserLogin(ctx context.Context, userID, clientIP, userAgent string) {
	_ = p.PublishUserAction(ctx, Event{
		Type:      EventTypeUserLogin,
		UserID:    userID,
		Action:    "login",
		ClientIP:  clientIP,
		UserAgent: userAgent,
	})
}

// Helper function to generate unique event IDs
func generateEventID() string {
	return time.Now().Format("20060102150405.000000")
}
