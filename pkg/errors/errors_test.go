package errors

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebEncodeError_Error(t *testing.T) {
	err := &WebEncodeError{
		Code:       "TEST_ERROR",
		Message:    "Test error message",
		HTTPStatus: 500,
	}

	if err.Error() != "Test error message" {
		t.Errorf("expected 'Test error message', got '%s'", err.Error())
	}
}

func TestResponse(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "WebEncodeError",
			err:            ErrNotFound,
			expectedStatus: 404,
			expectedCode:   "NOT_FOUND",
		},
		{
			name:           "StandardError",
			err:            errors.New("standard error"),
			expectedStatus: 500,
			expectedCode:   "INTERNAL_ERROR",
		},
		{
			name:           "UnknownWebEncodeError",
			err:            &WebEncodeError{Code: "CUSTOM", Message: "Custom", HTTPStatus: 418},
			expectedStatus: 418,
			expectedCode:   "CUSTOM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/test", nil)
			Response(w, r, tt.err)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if resp.Header.Get("Content-Type") != "application/json" {
				t.Errorf("expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
			}

			var body WebEncodeError
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}

			if body.Code != tt.expectedCode {
				t.Errorf("expected code %s, got %s", tt.expectedCode, body.Code)
			}
		})
	}
}
