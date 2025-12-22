package logger

import (
	"context"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("creates logger with service name", func(t *testing.T) {
		l := New("test-service")
		if l == nil {
			t.Fatal("expected logger to be non-nil")
		}
		if l.Logger == nil {
			t.Fatal("expected underlying slog.Logger to be non-nil")
		}
	})

	t.Run("creates different logger instances", func(t *testing.T) {
		l1 := New("service1")
		l2 := New("service2")
		if l1 == l2 {
			t.Fatal("expected different logger instances")
		}
	})
}

func TestLogger_WithContext(t *testing.T) {
	t.Run("returns valid logger with context", func(t *testing.T) {
		l := New("test-service")
		ctx := context.Background()
		ctxLogger := l.WithContext(ctx)
		if ctxLogger == nil {
			t.Fatal("expected context logger to be non-nil")
		}
	})

	t.Run("works with cancelled context", func(t *testing.T) {
		l := New("test-service")
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		ctxLogger := l.WithContext(ctx)
		if ctxLogger == nil {
			t.Fatal("expected context logger to be non-nil even with cancelled context")
		}
	})
}

func TestLogger_Methods(t *testing.T) {
	// Test that the logger can be used for basic logging operations
	l := New("test-service")

	t.Run("Info does not panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Info panicked: %v", r)
			}
		}()
		l.Info("test message", "key", "value")
	})

	t.Run("Error does not panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Error panicked: %v", r)
			}
		}()
		l.Error("test error", "key", "value")
	})

	t.Run("Warn does not panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Warn panicked: %v", r)
			}
		}()
		l.Warn("test warning", "key", "value")
	})

	t.Run("Debug does not panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Debug panicked: %v", r)
			}
		}()
		l.Debug("test debug", "key", "value")
	})
}
