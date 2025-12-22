package workers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nats-io/nats.go"
	"github.com/rennerdo30/webencode/pkg/bus"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/logger"
)

type HeartbeatListener struct {
	bus    *bus.Bus
	db     store.Querier
	logger *logger.Logger
}

func NewHeartbeatListener(b *bus.Bus, db store.Querier, l *logger.Logger) *HeartbeatListener {
	return &HeartbeatListener{
		bus:    b,
		db:     db,
		logger: l,
	}
}

type HeartbeatPayload struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Capacity  map[string]interface{} `json:"capacity"`
	Status    string                 `json:"status"`
}

func (h *HeartbeatListener) Start(ctx context.Context) error {
	// Subscribe to worker heartbeats
	sub, err := h.bus.Conn().Subscribe(bus.SubjectWorkerHeartbeat, func(msg *nats.Msg) {
		h.HandleHeartbeat(ctx, msg.Data)
	})
	if err != nil {
		return err
	}

	h.logger.Info("Heartbeat listener started")

	// Wait for context cancellation
	<-ctx.Done()
	sub.Unsubscribe()
	return nil
}

func (h *HeartbeatListener) HandleHeartbeat(ctx context.Context, data []byte) {
	var payload HeartbeatPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		h.logger.Error("Invalid heartbeat payload", "error", err)
		return
	}

	// Marshal capacity back to JSON for storage
	capacityJSON, _ := json.Marshal(payload.Capacity)

	// Upsert worker in database using RegisterWorker (which does ON CONFLICT UPDATE)
	err := h.db.RegisterWorker(ctx, store.RegisterWorkerParams{
		ID:           payload.ID,
		Hostname:     payload.ID, // Use ID as hostname for now
		Version:      "1.0.0",
		Status:       payload.Status,
		Capacity:     capacityJSON,
		IpAddress:    nil, // Optional
		Port:         pgtype.Int4{Valid: false},
		Capabilities: []byte("{}"),
	})

	if err != nil {
		h.logger.Error("Failed to update worker", "id", payload.ID, "error", err)
		return
	}

	h.logger.Debug("Worker heartbeat received", "id", payload.ID, "status", payload.Status)
}
