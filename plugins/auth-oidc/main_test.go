package main

import (
	"context"
	"os"
	"testing"

	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/stretchr/testify/assert"
)

// newDevModePlugin creates a plugin with dev mode enabled for testing
func newDevModePlugin() *AuthPlugin {
	return &AuthPlugin{
		logger: logger.New("plugin-auth-oidc-test"),
		config: Config{
			DevMode: true,
		},
	}
}

func TestNewAuthPlugin(t *testing.T) {
	t.Run("initialization", func(t *testing.T) {
		p := NewAuthPlugin()
		assert.NotNil(t, p)
		assert.NotNil(t, p.logger)
	})

	t.Run("reads env variables", func(t *testing.T) {
		os.Setenv("OIDC_ISSUER_URL", "https://example.com")
		os.Setenv("OIDC_CLIENT_ID", "test-client")
		os.Setenv("OIDC_CLIENT_SECRET", "test-secret")
		os.Setenv("OIDC_DEV_MODE", "true")
		defer func() {
			os.Unsetenv("OIDC_ISSUER_URL")
			os.Unsetenv("OIDC_CLIENT_ID")
			os.Unsetenv("OIDC_CLIENT_SECRET")
			os.Unsetenv("OIDC_DEV_MODE")
		}()

		p := NewAuthPlugin()
		assert.Equal(t, "https://example.com", p.config.IssuerURL)
		assert.Equal(t, "test-client", p.config.ClientID)
		assert.Equal(t, "test-secret", p.config.ClientSecret)
		assert.True(t, p.config.DevMode)
	})

	t.Run("reads dev mode with 1", func(t *testing.T) {
		os.Setenv("OIDC_DEV_MODE", "1")
		defer os.Unsetenv("OIDC_DEV_MODE")

		p := NewAuthPlugin()
		assert.True(t, p.config.DevMode)
	})
}

func TestAuthPlugin_ValidateToken(t *testing.T) {
	p := newDevModePlugin()
	ctx := context.Background()

	t.Run("dev admin token", func(t *testing.T) {
		req := &pb.TokenRequest{
			Token: "dev-admin",
		}

		session, err := p.ValidateToken(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.True(t, session.Valid)
		assert.Equal(t, "admin", session.Role)
		assert.Equal(t, "00000000-0000-0000-0000-000000000001", session.UserId)
	})

	t.Run("dev user token", func(t *testing.T) {
		req := &pb.TokenRequest{
			Token: "dev-user",
		}

		session, err := p.ValidateToken(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.True(t, session.Valid)
		assert.Equal(t, "user", session.Role)
		assert.Equal(t, "00000000-0000-0000-0000-000000000002", session.UserId)
	})

	t.Run("non-dev token in dev mode rejected when no OIDC", func(t *testing.T) {
		req := &pb.TokenRequest{
			Token: "invalid-token",
		}

		session, err := p.ValidateToken(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.False(t, session.Valid)
	})

	t.Run("empty token", func(t *testing.T) {
		req := &pb.TokenRequest{
			Token: "",
		}

		session, err := p.ValidateToken(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.False(t, session.Valid)
	})

	t.Run("dev mode disabled without OIDC returns error", func(t *testing.T) {
		pNoDevMode := &AuthPlugin{
			logger: logger.New("plugin-auth-oidc-test"),
			config: Config{
				DevMode: false,
			},
		}

		req := &pb.TokenRequest{
			Token: "any-token",
		}

		_, err := pNoDevMode.ValidateToken(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "OIDC_ISSUER_URL is required")
	})
}

func TestAuthPlugin_HandleDevToken(t *testing.T) {
	p := newDevModePlugin()

	t.Run("admin token", func(t *testing.T) {
		session, err := p.handleDevToken("dev-admin")
		assert.NoError(t, err)
		assert.True(t, session.Valid)
		assert.Equal(t, "admin", session.Role)
		assert.Equal(t, "00000000-0000-0000-0000-000000000001", session.UserId)
		assert.Contains(t, session.Roles, "admin")
	})

	t.Run("user token", func(t *testing.T) {
		session, err := p.handleDevToken("dev-user")
		assert.NoError(t, err)
		assert.True(t, session.Valid)
		assert.Equal(t, "user", session.Role)
		assert.Equal(t, "00000000-0000-0000-0000-000000000002", session.UserId)
		assert.Contains(t, session.Roles, "user")
	})
}

func TestAuthPlugin_ExtractRoles(t *testing.T) {
	p := newDevModePlugin()
	p.config.ClientID = "test-client"

	t.Run("extracts direct roles", func(t *testing.T) {
		claims := &Claims{
			Roles: []string{"role1", "role2"},
		}
		roles := p.extractRoles(claims)
		assert.Contains(t, roles, "role1")
		assert.Contains(t, roles, "role2")
	})

	t.Run("extracts groups", func(t *testing.T) {
		claims := &Claims{
			Groups: []string{"group1", "group2"},
		}
		roles := p.extractRoles(claims)
		assert.Contains(t, roles, "group1")
		assert.Contains(t, roles, "group2")
	})

	t.Run("extracts realm access roles", func(t *testing.T) {
		claims := &Claims{}
		claims.RealmAccess.Roles = []string{"realm-role1", "realm-role2"}
		roles := p.extractRoles(claims)
		assert.Contains(t, roles, "realm-role1")
		assert.Contains(t, roles, "realm-role2")
	})

	t.Run("extracts client-specific roles", func(t *testing.T) {
		claims := &Claims{
			ResourceAccess: map[string]struct {
				Roles []string `json:"roles"`
			}{
				"test-client": {Roles: []string{"client-role1"}},
			},
		}
		roles := p.extractRoles(claims)
		assert.Contains(t, roles, "client-role1")
	})

	t.Run("returns user role when no roles found", func(t *testing.T) {
		claims := &Claims{}
		roles := p.extractRoles(claims)
		assert.Equal(t, []string{"user"}, roles)
	})

	t.Run("deduplicates roles", func(t *testing.T) {
		claims := &Claims{
			Roles:  []string{"admin"},
			Groups: []string{"admin"},
		}
		roles := p.extractRoles(claims)
		adminCount := 0
		for _, r := range roles {
			if r == "admin" {
				adminCount++
			}
		}
		assert.Equal(t, 1, adminCount)
	})
}

func TestAuthPlugin_Authorize(t *testing.T) {
	p := newDevModePlugin()
	ctx := context.Background()

	t.Run("user allowed actions", func(t *testing.T) {
		allowedActions := []string{"read", "list", "create"}
		for _, action := range allowedActions {
			req := &pb.AuthZRequest{
				UserId:       "user-123",
				Action:       action,
				ResourceType: "job",
			}

			resp, err := p.Authorize(ctx, req)
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.True(t, resp.Allowed, "action %s should be allowed", action)
		}
	})

	t.Run("user denied actions", func(t *testing.T) {
		deniedActions := []string{"delete", "update", "execute"}
		for _, action := range deniedActions {
			req := &pb.AuthZRequest{
				UserId:       "user-123",
				Action:       action,
				ResourceType: "job",
			}

			resp, err := p.Authorize(ctx, req)
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.False(t, resp.Allowed, "action %s should be denied", action)
		}
	})
}

func TestAuthPlugin_GetUser(t *testing.T) {
	p := newDevModePlugin()
	ctx := context.Background()

	t.Run("returns dev user", func(t *testing.T) {
		req := &pb.UserRequest{
			UserId: "user-123",
		}

		user, err := p.GetUser(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "user-123", user.Id)
		assert.Equal(t, "dev-user", user.Username)
		assert.Equal(t, "dev@example.com", user.Email)
		assert.Equal(t, "Dev User", user.FullName)
	})
}

func TestAuthPlugin_ListUsers(t *testing.T) {
	p := newDevModePlugin()
	ctx := context.Background()

	t.Run("returns dev users", func(t *testing.T) {
		req := &pb.ListUsersRequest{
			Page:     1,
			PageSize: 10,
		}

		resp, err := p.ListUsers(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int32(2), resp.TotalCount)
		assert.Len(t, resp.Users, 2)

		// Check admin user
		admin := resp.Users[0]
		assert.Equal(t, "admin", admin.Username)
		assert.Equal(t, "admin", admin.Role)

		// Check regular user
		user := resp.Users[1]
		assert.Equal(t, "user", user.Username)
		assert.Equal(t, "user", user.Role)
	})
}

func TestAuthPlugin_RefreshToken(t *testing.T) {
	p := newDevModePlugin()
	ctx := context.Background()

	t.Run("returns refreshed session in dev mode", func(t *testing.T) {
		req := &pb.RefreshTokenRequest{
			RefreshToken: "old-refresh-token",
		}

		session, err := p.RefreshToken(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.True(t, session.Valid)
		assert.Equal(t, "dev-refreshed-token", session.AccessToken)
		assert.Equal(t, "old-refresh-token", session.RefreshToken)
	})
}

func TestAuthPlugin_Logout(t *testing.T) {
	p := newDevModePlugin()
	ctx := context.Background()

	t.Run("logout single session", func(t *testing.T) {
		req := &pb.LogoutRequest{
			UserId:      "user-123",
			AllSessions: false,
		}

		result, err := p.Logout(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("logout all sessions", func(t *testing.T) {
		req := &pb.LogoutRequest{
			UserId:      "user-123",
			AllSessions: true,
		}

		result, err := p.Logout(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestContains(t *testing.T) {
	t.Run("returns true when value exists", func(t *testing.T) {
		slice := []string{"a", "b", "c"}
		assert.True(t, contains(slice, "b"))
	})

	t.Run("returns false when value not exists", func(t *testing.T) {
		slice := []string{"a", "b", "c"}
		assert.False(t, contains(slice, "d"))
	})

	t.Run("returns false for empty slice", func(t *testing.T) {
		slice := []string{}
		assert.False(t, contains(slice, "a"))
	})
}

// Benchmark tests
func BenchmarkValidateToken(b *testing.B) {
	p := newDevModePlugin()
	ctx := context.Background()
	req := &pb.TokenRequest{Token: "dev-admin"}

	for i := 0; i < b.N; i++ {
		_, _ = p.ValidateToken(ctx, req)
	}
}

func BenchmarkAuthorize(b *testing.B) {
	p := newDevModePlugin()
	ctx := context.Background()
	req := &pb.AuthZRequest{
		UserId:       "user-123",
		Action:       "read",
		ResourceType: "job",
	}

	for i := 0; i < b.N; i++ {
		_, _ = p.Authorize(ctx, req)
	}
}
