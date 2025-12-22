package webhooks

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/rennerdo30/webencode/pkg/bus"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/logger"
)

// WebhookPayload is the standard payload sent to webhook endpoints
type WebhookPayload struct {
	Event     string          `json:"event"`
	Timestamp string          `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

type Manager struct {
	bus    *bus.Bus
	db     store.Querier
	logger *logger.Logger
	client *http.Client
}

func New(b *bus.Bus, db store.Querier, l *logger.Logger) *Manager {
	return &Manager{
		bus:    b,
		db:     db,
		logger: l,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (m *Manager) Start(ctx context.Context) error {
	// Subscribe to events using Pull Consumer
	consumer, err := m.bus.JetStream().CreateOrUpdateConsumer(ctx, bus.StreamEvents, jetstream.ConsumerConfig{
		Durable:   "webhooks-dispatcher",
		AckPolicy: jetstream.AckExplicitPolicy,
	})
	if err != nil {
		m.logger.Error("Failed to create webhooks consumer", "error", err)
		return err
	}

	cons, err := consumer.Consume(func(msg jetstream.Msg) {
		m.handleEvent(ctx, msg)
	})
	if err != nil {
		return err
	}

	m.logger.Info("Webhooks manager started")

	<-ctx.Done()
	cons.Stop()
	return nil
}

func (m *Manager) handleEvent(ctx context.Context, msg jetstream.Msg) {
	defer msg.Ack()

	// Extract event type from subject (e.g., "events.job.completed" -> "job.completed")
	eventType := strings.TrimPrefix(msg.Subject(), "events.")

	// Find webhooks that listen for this event type
	// The query uses $1 = ANY(events), ensuring atomic text match
	webhooks, err := m.db.ListActiveWebhooksForEvent(ctx, eventType)
	if err != nil {
		m.logger.Error("Failed to list webhooks", "error", err)
		return
	}

	if len(webhooks) == 0 {
		return
	}

	// Build payload
	payload := WebhookPayload{
		Event:     eventType,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Data:      msg.Data(),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		m.logger.Error("Failed to marshal webhook payload", "error", err)
		return
	}

	// Dispatch to all matching webhooks
	for _, wh := range webhooks {
		go m.dispatchWithRetry(ctx, wh, body)
	}
}

func (m *Manager) dispatchWithRetry(ctx context.Context, wh store.Webhook, body []byte) {
	maxRetries := 3
	backoff := 5 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := m.send(wh.Url, wh.Secret, body)
		if err == nil {
			m.db.UpdateWebhookTriggered(ctx, wh.ID)
			m.logger.Info("Webhook delivered", "url", wh.Url)
			return
		}

		m.logger.Warn("Webhook delivery failed",
			"url", wh.Url,
			"attempt", attempt,
			"error", err,
		)

		if attempt < maxRetries {
			time.Sleep(backoff)
			backoff *= 2 // Exponential backoff
		}
	}

	// All retries failed
	m.db.IncrementWebhookFailure(ctx, wh.ID)

	// Get updated failure count
	updatedWh, _ := m.db.GetWebhook(ctx, wh.ID)
	if updatedWh.FailureCount.Int32 >= 10 {
		m.db.DeactivateWebhook(ctx, wh.ID)
		m.logger.Warn("Webhook deactivated after repeated failures", "url", wh.Url)
	}
}

func (m *Manager) send(url string, secret string, body []byte) error {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "WebEncode/1.0")
	req.Header.Set("X-WebEncode-Event", "true")

	// Sign the payload if secret is provided
	if secret != "" {
		signature := computeHMAC(body, secret)
		req.Header.Set("X-WebEncode-Signature", "sha256="+signature)
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

func computeHMAC(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}
