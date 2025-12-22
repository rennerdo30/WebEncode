package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/logger"
)

// MockQuerier embeds store.Querier to satisfy the interface.
// Unimplemented methods will panic if called (which is fine, we only want CreateErrorEvent).
type MockQuerier struct {
	store.Querier
	CreateErrorEventFunc func(ctx context.Context, arg store.CreateErrorEventParams) (store.ErrorEvent, error)
}

func (m *MockQuerier) CreateErrorEvent(ctx context.Context, arg store.CreateErrorEventParams) (store.ErrorEvent, error) {
	if m.CreateErrorEventFunc != nil {
		return m.CreateErrorEventFunc(ctx, arg)
	}
	return store.ErrorEvent{}, nil
}

func TestRecovery(t *testing.T) {
	// Setup
	l := logger.New("test")
	captured := false
	mockDb := &MockQuerier{
		CreateErrorEventFunc: func(ctx context.Context, arg store.CreateErrorEventParams) (store.ErrorEvent, error) {
			captured = true
			if arg.Message != "test panic" {
				t.Errorf("expected message 'test panic', got '%s'", arg.Message)
			}
			if arg.SourceComponent != "backend:panic" {
				t.Errorf("expected source 'backend:panic', got '%s'", arg.SourceComponent)
			}
			return store.ErrorEvent{}, nil
		},
	}

	handler := Recovery(mockDb, l)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(errors.New("test panic"))
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	// Execute
	// The middleware should recover the panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Middleware did not recover panic: %v", r)
			}
		}()
		handler.ServeHTTP(rec, req)
	}()

	// Verify
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}

	if !captured {
		t.Error("CreateErrorEvent was not called")
	}
}
