package logger

import (
	"context"
	"log/slog"
	"os"
)

// Logger is a wrapper around slog.Logger
type Logger struct {
	*slog.Logger
}

// New creates a new JSON logger with the service name
func New(serviceName string) *Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := slog.New(handler).With("service", serviceName)

	return &Logger{logger}
}

// WithContext adds tracing info from context
func (l *Logger) WithContext(ctx context.Context) *slog.Logger {
	logger := l.Logger

	// Extract request ID if present
	if reqID, ok := ctx.Value("request_id").(string); ok && reqID != "" {
		logger = logger.With("request_id", reqID)
	}

	// Extract trace ID if present
	if traceID, ok := ctx.Value("trace_id").(string); ok && traceID != "" {
		logger = logger.With("trace_id", traceID)
	}

	// Extract user ID if present
	if userID, ok := ctx.Value("user_id").(string); ok && userID != "" {
		logger = logger.With("user_id", userID)
	}

	return logger
}
