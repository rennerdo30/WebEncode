package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInstrumentHandler(t *testing.T) {
	// Create a dummy handler
	handler := InstrumentHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest("GET", "/v1/jobs/123", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTeapot {
		t.Errorf("expected status %d, got %d", http.StatusTeapot, w.Code)
	}

	// We can't easily verify the metrics were recorded without scraping the registry,
	// but ensuring it doesn't panic and passes through is good.
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "root",
			path:     "/",
			expected: "/",
		},
		{
			name:     "simple",
			path:     "/v1/health",
			expected: "/v1/health",
		},
		{
			name:     "with uuid",
			path:     "/v1/jobs/123e4567-e89b-12d3-a456-426614174000",
			expected: "/v1/jobs/:id",
		},
		{
			name:     "with numeric id",
			path:     "/v1/users/123/profile",
			expected: "/v1/users/:id/profile",
		},
		{
			name:     "complex",
			path:     "/v1/jobs/123e4567-e89b-12d3-a456-426614174000/tasks/5",
			expected: "/v1/jobs/:id/tasks/:id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePath(tt.path)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestIsUUID(t *testing.T) {
	if !isUUID("123e4567-e89b-12d3-a456-426614174000") {
		t.Error("expected true for valid UUID")
	}
	if isUUID("invalid") {
		t.Error("expected false for invalid UUID")
	}
}

func TestIsNumeric(t *testing.T) {
	if !isNumeric("123") {
		t.Error("expected true for numeric string")
	}
	if isNumeric("123a") {
		t.Error("expected false for alphanumeric string")
	}
}
