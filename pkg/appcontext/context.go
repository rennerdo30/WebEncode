package appcontext

import (
	"context"

	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/logger"
)

type contextKey string

const (
	LoggerKey  contextKey = "logger"
	QuerierKey contextKey = "querier"
)

// GetLogger retrieves the logger from the context
func GetLogger(ctx context.Context) *logger.Logger {
	if l, ok := ctx.Value(LoggerKey).(*logger.Logger); ok {
		return l
	}
	return logger.New("unknown") // Fallback
}

// GetQuerier retrieves the store querier from the context
func GetQuerier(ctx context.Context) store.Querier {
	if q, ok := ctx.Value(QuerierKey).(store.Querier); ok {
		return q
	}
	return nil
}

// WithLogger returns a context with the logger
func WithLogger(ctx context.Context, l *logger.Logger) context.Context {
	return context.WithValue(ctx, LoggerKey, l)
}

// WithQuerier returns a context with the querier
func WithQuerier(ctx context.Context, q store.Querier) context.Context {
	return context.WithValue(ctx, QuerierKey, q)
}
