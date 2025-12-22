package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUser_NoContext(t *testing.T) {
	ctx := context.Background()
	user := GetUser(ctx)
	assert.Nil(t, user)
}

func TestGetUser_WithContext(t *testing.T) {
	user := &UserContext{
		UserID:   "123",
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "admin",
		Roles:    []string{"admin"},
	}
	ctx := context.WithValue(context.Background(), UserContextKey, user)

	retrieved := GetUser(ctx)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "123", retrieved.UserID)
	assert.Equal(t, "testuser", retrieved.Username)
}

func TestGetUserID_NoContext(t *testing.T) {
	ctx := context.Background()
	userID := GetUserID(ctx)
	// Should return default dev user ID
	assert.Equal(t, "00000000-0000-0000-0000-000000000001", userID)
}

func TestGetUserID_WithContext(t *testing.T) {
	user := &UserContext{
		UserID: "custom-id",
	}
	ctx := context.WithValue(context.Background(), UserContextKey, user)

	userID := GetUserID(ctx)
	assert.Equal(t, "custom-id", userID)
}

func TestDefaultAuthConfig(t *testing.T) {
	config := DefaultAuthConfig()

	assert.True(t, config.DevMode)
	assert.NotEmpty(t, config.DefaultDevUserID)
	assert.Contains(t, config.SkipPaths, "/v1/system/health")
}

func TestAuth_SkipPaths(t *testing.T) {
	config := AuthMiddlewareConfig{
		DevMode:   false,
		SkipPaths: []string{"/health"},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := Auth(config)(handler)

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAuth_NoAuthHeader_DevMode(t *testing.T) {
	config := DefaultAuthConfig()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUser(r.Context())
		assert.NotNil(t, user)
		assert.Equal(t, config.DefaultDevUserID, user.UserID)
		w.WriteHeader(http.StatusOK)
	})

	middleware := Auth(config)(handler)

	req := httptest.NewRequest("GET", "/v1/jobs", nil)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAuth_NoAuthHeader_ProductionMode(t *testing.T) {
	config := AuthMiddlewareConfig{
		DevMode: false,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := Auth(config)(handler)

	req := httptest.NewRequest("GET", "/v1/jobs", nil)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuth_DevToken(t *testing.T) {
	config := DefaultAuthConfig()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUser(r.Context())
		assert.NotNil(t, user)
		assert.Equal(t, "user", user.Role)
		w.WriteHeader(http.StatusOK)
	})

	middleware := Auth(config)(handler)

	req := httptest.NewRequest("GET", "/v1/jobs", nil)
	req.Header.Set("Authorization", "Bearer dev-user")
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAuth_DevAdminToken(t *testing.T) {
	config := DefaultAuthConfig()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUser(r.Context())
		assert.NotNil(t, user)
		assert.Equal(t, "admin", user.Role)
		w.WriteHeader(http.StatusOK)
	})

	middleware := Auth(config)(handler)

	req := httptest.NewRequest("GET", "/v1/jobs", nil)
	req.Header.Set("Authorization", "Bearer dev-admin")
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRequireRole_HasRole(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequireRole("admin")(handler)

	user := &UserContext{
		UserID: "123",
		Roles:  []string{"admin"},
	}
	ctx := context.WithValue(context.Background(), UserContextKey, user)

	req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRequireRole_MissingRole(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequireRole("admin")(handler)

	user := &UserContext{
		UserID: "123",
		Roles:  []string{"user"},
	}
	ctx := context.WithValue(context.Background(), UserContextKey, user)

	req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestRequireRole_NoUser(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequireRole("admin")(handler)

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestGetClientIP_Direct(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	ip := getClientIP(req)
	assert.Equal(t, "192.168.1.1:12345", ip)
}

func TestGetClientIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1, 10.0.0.2")
	req.RemoteAddr = "192.168.1.1:12345"

	ip := getClientIP(req)
	assert.Equal(t, "10.0.0.1", ip)
}

func TestGetClientIP_XRealIP(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Real-IP", "10.0.0.5")
	req.RemoteAddr = "192.168.1.1:12345"

	ip := getClientIP(req)
	assert.Equal(t, "10.0.0.5", ip)
}
