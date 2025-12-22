package events

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nats-io/nats.go"
	"github.com/rennerdo30/webencode/internal/orchestrator"
	"github.com/rennerdo30/webencode/pkg/bus"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/logger"
)

type EventListener struct {
	bus    *bus.Bus
	db     store.Querier
	orch   orchestrator.Service
	logger *logger.Logger
}

func NewEventListener(b *bus.Bus, db store.Querier, orch orchestrator.Service, l *logger.Logger) *EventListener {
	return &EventListener{
		bus:    b,
		db:     db,
		orch:   orch,
		logger: l,
	}
}

type TaskEvent struct {
	TaskID string          `json:"task_id"`
	Event  string          `json:"event"` // "completed", "failed", "progress", "log"
	Result json.RawMessage `json:"result"`
}

func (e *EventListener) Start(ctx context.Context) error {
	// Subscribe to job events
	sub, err := e.bus.Conn().Subscribe(bus.SubjectJobEvents, func(msg *nats.Msg) {
		e.handleEvent(ctx, msg.Data)
	})
	if err != nil {
		return err
	}

	e.logger.Info("Job event listener started")

	// Subscribe to error events
	subErrors, err := e.bus.Conn().Subscribe(bus.SubjectErrorEvents, func(msg *nats.Msg) {
		e.handleErrorEvent(ctx, msg.Data)
	})
	if err != nil {
		e.logger.Error("Failed to subscribe to error events", "error", err)
	} else {
		e.logger.Info("Error event listener started")
	}

	// Wait for context cancellation
	<-ctx.Done()
	sub.Unsubscribe()
	if subErrors != nil {
		subErrors.Unsubscribe()
	}
	return nil
}

func (e *EventListener) handleErrorEvent(ctx context.Context, data []byte) {
	var event struct {
		Source      string                 `json:"source"`
		Severity    string                 `json:"severity"`
		Message     string                 `json:"message"`
		StackTrace  string                 `json:"stack_trace"`
		ContextData map[string]interface{} `json:"context_data"`
	}

	if err := json.Unmarshal(data, &event); err != nil {
		e.logger.Error("Invalid error event payload", "error", err)
		return
	}

	// Default severity
	severity := store.ErrorSeverityError
	switch event.Severity {
	case "warning":
		severity = store.ErrorSeverityWarning
	case "critical":
		severity = store.ErrorSeverityCritical
	case "fatal":
		severity = store.ErrorSeverityFatal
	}

	ctxDataBytes, _ := json.Marshal(event.ContextData)

	_, err := e.db.CreateErrorEvent(ctx, store.CreateErrorEventParams{
		SourceComponent: event.Source,
		Column2:         severity,
		Message:         event.Message,
		StackTrace:      pgtype.Text{String: event.StackTrace, Valid: event.StackTrace != ""},
		ContextData:     ctxDataBytes,
	})
	if err != nil {
		e.logger.Error("Failed to persist error event from bus", "error", err)
	}
}

func (e *EventListener) handleEvent(ctx context.Context, data []byte) {
	var event TaskEvent
	if err := json.Unmarshal(data, &event); err != nil {
		e.logger.Error("Invalid event payload", "error", err)
		return
	}

	// Delegate to Orchestrator workflow logic
	if err := e.orch.HandleTaskEvent(ctx, event.TaskID, event.Event, event.Result); err != nil {
		e.logger.Error("Failed to handle task event", "task_id", event.TaskID, "event", event.Event, "error", err)
	}
}
