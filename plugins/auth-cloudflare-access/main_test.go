package main

import (
	"context"
	"testing"
	"time"

	pb "github.com/rennerdo30/webencode/pkg/api/v1"
)

// TestNewAuthPlugin tests plugin initialization
func TestNewAuthPlugin(t *testing.T) {
	plugin := NewAuthPlugin()

	if plugin == nil {
		t.Fatal("NewAuthPlugin returned nil")
	}

	if plugin.logger == nil {
		t.Error("Plugin logger is nil")
	}

	if plugin.keyCache == nil {
		t.Error("Plugin keyCache is nil")
	}
}

// TestDevModeTokenValidation tests dev mode token handling
func TestDevModeTokenValidation(t *testing.T) {
	plugin := NewAuthPlugin()
	plugin.config.DevMode = true

	tests := []struct {
		name          string
		token         string
		expectedValid bool
		expectedRole  string
	}{
		{
			name:          "Dev admin token",
			token:         "dev-admin-token",
			expectedValid: true,
			expectedRole:  "admin",
		},
		{
			name:          "Dev user token",
			token:         "dev-user-token",
			expectedValid: true,
			expectedRole:  "user",
		},
		{
			name:          "Non-dev token in dev mode (no CF config)",
			token:         "bearer-some-other-token",
			expectedValid: false,
			expectedRole:  "",
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &pb.TokenRequest{Token: tt.token}
			session, err := plugin.ValidateToken(ctx, req)

			if err != nil {
				t.Fatalf("ValidateToken returned error: %v", err)
			}

			if session.Valid != tt.expectedValid {
				t.Errorf("Expected valid=%v, got valid=%v", tt.expectedValid, session.Valid)
			}

			if tt.expectedValid && session.Role != tt.expectedRole {
				t.Errorf("Expected role=%s, got role=%s", tt.expectedRole, session.Role)
			}
		})
	}
}

// TestAuthorize tests authorization logic
func TestAuthorize(t *testing.T) {
	plugin := NewAuthPlugin()
	plugin.config.DevMode = true

	ctx := context.Background()

	tests := []struct {
		name           string
		userId         string
		action         string
		expectedResult bool
	}{
		{
			name:           "User can read",
			userId:         "user-123",
			action:         "read",
			expectedResult: true,
		},
		{
			name:           "User can list",
			userId:         "user-123",
			action:         "list",
			expectedResult: true,
		},
		{
			name:           "User can create",
			userId:         "user-123",
			action:         "create",
			expectedResult: true,
		},
		{
			name:           "User cannot delete",
			userId:         "user-123",
			action:         "delete",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &pb.AuthZRequest{
				UserId: tt.userId,
				Action: tt.action,
			}
			resp, err := plugin.Authorize(ctx, req)

			if err != nil {
				t.Fatalf("Authorize returned error: %v", err)
			}

			if resp.Allowed != tt.expectedResult {
				t.Errorf("Expected allowed=%v, got allowed=%v (reason: %s)",
					tt.expectedResult, resp.Allowed, resp.Reason)
			}
		})
	}
}

// TestGetUser tests user retrieval in dev mode
func TestGetUser(t *testing.T) {
	plugin := NewAuthPlugin()
	plugin.config.DevMode = true

	ctx := context.Background()
	req := &pb.UserRequest{UserId: "test-user-id"}

	user, err := plugin.GetUser(ctx, req)
	if err != nil {
		t.Fatalf("GetUser returned error: %v", err)
	}

	if user.Id != "test-user-id" {
		t.Errorf("Expected user ID 'test-user-id', got '%s'", user.Id)
	}

	if user.Username != "dev-user" {
		t.Errorf("Expected username 'dev-user', got '%s'", user.Username)
	}
}

// TestListUsers tests user listing in dev mode
func TestListUsers(t *testing.T) {
	plugin := NewAuthPlugin()
	plugin.config.DevMode = true

	ctx := context.Background()
	req := &pb.ListUsersRequest{Page: 1, PageSize: 10}

	resp, err := plugin.ListUsers(ctx, req)
	if err != nil {
		t.Fatalf("ListUsers returned error: %v", err)
	}

	if resp.TotalCount != 2 {
		t.Errorf("Expected 2 users in dev mode, got %d", resp.TotalCount)
	}

	if len(resp.Users) != 2 {
		t.Errorf("Expected 2 user objects, got %d", len(resp.Users))
	}
}

// TestRefreshToken tests that refresh is not supported
func TestRefreshToken(t *testing.T) {
	plugin := NewAuthPlugin()

	ctx := context.Background()
	req := &pb.RefreshTokenRequest{RefreshToken: "some-token"}

	session, err := plugin.RefreshToken(ctx, req)
	if err != nil {
		t.Fatalf("RefreshToken returned error: %v", err)
	}

	if session.Valid {
		t.Error("Expected RefreshToken to return invalid session (not supported)")
	}
}

// TestLogout tests logout functionality
func TestLogout(t *testing.T) {
	plugin := NewAuthPlugin()

	ctx := context.Background()
	req := &pb.LogoutRequest{
		UserId:      "test-user",
		AllSessions: true,
	}

	_, err := plugin.Logout(ctx, req)
	if err != nil {
		t.Fatalf("Logout returned error: %v", err)
	}
}

// TestExtractUsername tests username extraction from email
func TestExtractUsername(t *testing.T) {
	tests := []struct {
		email    string
		expected string
	}{
		{"john.doe@example.com", "john.doe"},
		{"admin@test.org", "admin"},
		{"user", "user"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := extractUsername(tt.email)
			if result != tt.expected {
				t.Errorf("extractUsername(%q) = %q, expected %q", tt.email, result, tt.expected)
			}
		})
	}
}

// TestContains tests the contains helper function
func TestContains(t *testing.T) {
	slice := []string{"admin", "user", "moderator"}

	if !contains(slice, "admin") {
		t.Error("Expected contains to find 'admin'")
	}

	if !contains(slice, "user") {
		t.Error("Expected contains to find 'user'")
	}

	if contains(slice, "guest") {
		t.Error("Expected contains to not find 'guest'")
	}

	if contains(nil, "anything") {
		t.Error("Expected contains to return false for nil slice")
	}

	if contains([]string{}, "anything") {
		t.Error("Expected contains to return false for empty slice")
	}
}

// TestKeyCacheInitialization tests that the key cache is properly initialized
func TestKeyCacheInitialization(t *testing.T) {
	plugin := NewAuthPlugin()

	if plugin.keyCache.keys == nil {
		t.Error("Key cache keys map should be initialized")
	}

	if !plugin.keyCache.expiresAt.IsZero() {
		t.Error("Key cache should not have an expiration time initially")
	}
}

// TestHandleDevToken tests the dev token handler directly
func TestHandleDevToken(t *testing.T) {
	plugin := NewAuthPlugin()
	plugin.config.DevMode = true

	// Test admin token
	session, err := plugin.handleDevToken("dev-admin-test")
	if err != nil {
		t.Fatalf("handleDevToken returned error: %v", err)
	}

	if !session.Valid {
		t.Error("Expected valid session for dev token")
	}

	if session.Role != "admin" {
		t.Errorf("Expected role 'admin', got '%s'", session.Role)
	}

	if session.ExpiresAt <= time.Now().Unix() {
		t.Error("Expected future expiration time")
	}

	// Test user token
	session, err = plugin.handleDevToken("dev-user-test")
	if err != nil {
		t.Fatalf("handleDevToken returned error: %v", err)
	}

	if session.Role != "user" {
		t.Errorf("Expected role 'user', got '%s'", session.Role)
	}
}
