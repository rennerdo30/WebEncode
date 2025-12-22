package middleware

import (
	"context"
	"net/http"
	"strings"

	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/rennerdo30/webencode/pkg/errors"
)

// UserContextKey is the key for user information in request context
type contextKey string

const (
	UserContextKey contextKey = "user"
)

// UserContext contains authenticated user information
type UserContext struct {
	UserID   string
	Username string
	Email    string
	Role     string
	Roles    []string
}

// GetUser retrieves the user from the request context
func GetUser(ctx context.Context) *UserContext {
	if user, ok := ctx.Value(UserContextKey).(*UserContext); ok {
		return user
	}
	return nil
}

// GetUserID retrieves just the user ID from context, with a fallback default
func GetUserID(ctx context.Context) string {
	if user := GetUser(ctx); user != nil {
		return user.UserID
	}
	// Fallback for unauthenticated requests (dev mode)
	return "00000000-0000-0000-0000-000000000001"
}

// AuthValidator is an interface for validating tokens
type AuthValidator interface {
	ValidateToken(ctx context.Context, req *pb.TokenRequest) (*pb.UserSession, error)
}

// AuthMiddlewareConfig configures the auth middleware
type AuthMiddlewareConfig struct {
	Validator        AuthValidator
	SkipPaths        []string // Paths that don't require auth
	DevMode          bool     // If true, allow unauthenticated requests with default user
	DefaultDevUserID string   // Default user ID for dev mode
}

// DefaultAuthConfig returns a development-friendly auth config
func DefaultAuthConfig() AuthMiddlewareConfig {
	return AuthMiddlewareConfig{
		DevMode:          true,
		DefaultDevUserID: "00000000-0000-0000-0000-000000000001",
		SkipPaths: []string{
			"/v1/system/health",
			"/v1/events", // SSE endpoint
		},
	}
}

// Auth creates an authentication middleware with the given config
func Auth(config AuthMiddlewareConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if path should skip auth
			for _, path := range config.SkipPaths {
				if strings.HasPrefix(r.URL.Path, path) {
					next.ServeHTTP(w, r)
					return
				}
			}

			authHeader := r.Header.Get("Authorization")

			// No auth header
			if authHeader == "" {
				if config.DevMode {
					// In dev mode, use default user
					ctx := context.WithValue(r.Context(), UserContextKey, &UserContext{
						UserID:   config.DefaultDevUserID,
						Username: "dev-user",
						Email:    "dev@localhost",
						Role:     "admin",
						Roles:    []string{"admin"},
					})
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
				errors.Response(w, r, errors.ErrUnauthorized)
				return
			}

			// Parse auth header
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 {
				errors.Response(w, r, errors.ErrUnauthorized)
				return
			}

			scheme := strings.ToLower(parts[0])
			token := parts[1]

			// If we have a validator, use it
			if config.Validator != nil {
				session, err := config.Validator.ValidateToken(r.Context(), &pb.TokenRequest{
					Token:     token,
					Scheme:    scheme,
					ClientIp:  getClientIP(r),
					UserAgent: r.UserAgent(),
				})
				if err != nil || !session.Valid {
					errors.Response(w, r, errors.ErrUnauthorized)
					return
				}

				ctx := context.WithValue(r.Context(), UserContextKey, &UserContext{
					UserID:   session.UserId,
					Username: session.Username,
					Email:    session.Email,
					Role:     session.Role,
					Roles:    session.Roles,
				})
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Fallback: simple token validation for dev mode
			if config.DevMode {
				// Accept any token starting with "dev-" or "Bearer dev-"
				if strings.HasPrefix(token, "dev-") || strings.HasPrefix(token, "basic-") {
					role := "user"
					userID := "00000000-0000-0000-0000-000000000002"
					if strings.Contains(token, "admin") {
						role = "admin"
						userID = "00000000-0000-0000-0000-000000000001"
					}

					ctx := context.WithValue(r.Context(), UserContextKey, &UserContext{
						UserID:   userID,
						Username: "dev-user",
						Email:    "dev@localhost",
						Role:     role,
						Roles:    []string{role},
					})
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			errors.Response(w, r, errors.ErrUnauthorized)
		})
	}
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For first (for proxied requests)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	// Check X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// RequireRole creates a middleware that requires a specific role
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUser(r.Context())
			if user == nil {
				errors.Response(w, r, errors.ErrUnauthorized)
				return
			}

			// Check if user has any of the required roles
			for _, required := range roles {
				for _, userRole := range user.Roles {
					if userRole == required {
						next.ServeHTTP(w, r)
						return
					}
				}
			}

			errors.Response(w, r, errors.ErrForbidden)
		})
	}
}

// Legacy middleware for backward compatibility
func AuthMiddleware(next http.Handler) http.Handler {
	config := DefaultAuthConfig()
	return Auth(config)(next)
}
