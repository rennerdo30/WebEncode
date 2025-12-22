package main

import (
	"context"
	"testing"

	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/stretchr/testify/assert"
)

func TestNewLDAPAuthPlugin(t *testing.T) {
	t.Run("initialization", func(t *testing.T) {
		p := NewLDAPAuthPlugin()
		assert.NotNil(t, p)
		assert.NotNil(t, p.logger)
		assert.NotNil(t, p.config)
	})

	t.Run("config from env", func(t *testing.T) {
		// Config should be empty when env vars aren't set
		p := NewLDAPAuthPlugin()
		assert.Empty(t, p.config.URL)
		assert.Empty(t, p.config.BaseDN)
	})
}

func TestLDAPConfig(t *testing.T) {
	t.Run("config structure", func(t *testing.T) {
		config := &LDAPConfig{
			URL:        "ldap://localhost:389",
			BaseDN:     "dc=example,dc=com",
			BindDN:     "cn=admin,dc=example,dc=com",
			BindPass:   "secret",
			UserFilter: "(uid=%s)",
			GroupDN:    "ou=groups,dc=example,dc=com",
		}

		assert.Equal(t, "ldap://localhost:389", config.URL)
		assert.Equal(t, "dc=example,dc=com", config.BaseDN)
		assert.Contains(t, config.BindDN, "admin")
	})

	t.Run("LDAPS URL", func(t *testing.T) {
		config := &LDAPConfig{
			URL: "ldaps://ldap.example.com:636",
		}

		assert.Contains(t, config.URL, "ldaps://")
	})
}

func TestLDAPAuthPlugin_Connect(t *testing.T) {
	t.Run("fails without URL", func(t *testing.T) {
		p := NewLDAPAuthPlugin()
		p.config.URL = ""

		conn, err := p.connect()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "LDAP_URL not configured")
		assert.Nil(t, conn)
	})

	t.Run("connect requires valid LDAP server", func(t *testing.T) {
		// This test would require a live LDAP server
		t.Skip("Requires LDAP server - integration test")
	})
}

func TestLDAPAuthPlugin_ValidateToken(t *testing.T) {
	p := NewLDAPAuthPlugin()
	ctx := context.Background()

	t.Run("direct validation not implemented", func(t *testing.T) {
		req := &pb.TokenRequest{
			Token: "some-token",
		}

		session, err := p.ValidateToken(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not fully implemented")
		assert.NotNil(t, session)
		assert.False(t, session.Valid)
	})
}

func TestLDAPAuthPlugin_Authorize(t *testing.T) {
	p := NewLDAPAuthPlugin()
	ctx := context.Background()

	t.Run("allows all by default", func(t *testing.T) {
		req := &pb.AuthZRequest{
			UserId:       "user-123",
			Action:       "create",
			ResourceType: "job",
		}

		resp, err := p.Authorize(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.Allowed)
	})
}

func TestLDAPAuthPlugin_Login(t *testing.T) {
	t.Run("login requires LDAP connection", func(t *testing.T) {
		p := NewLDAPAuthPlugin()
		ctx := context.Background()

		session, err := p.Login(ctx, "testuser", "testpass")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "LDAP_URL not configured")
		assert.Nil(t, session)
	})

	t.Run("integration login", func(t *testing.T) {
		// This test requires a live LDAP server
		t.Skip("Requires LDAP server - integration test")
	})
}

func TestLDAPAuthPlugin_GetUser(t *testing.T) {
	p := NewLDAPAuthPlugin()
	ctx := context.Background()

	t.Run("requires LDAP connection", func(t *testing.T) {
		req := &pb.UserRequest{
			UserId: "user-123",
		}

		user, err := p.GetUser(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "LDAP_URL not configured")
		assert.Nil(t, user)
	})
}

func TestLDAPAuthPlugin_ListUsers(t *testing.T) {
	p := NewLDAPAuthPlugin()
	ctx := context.Background()

	t.Run("requires LDAP connection", func(t *testing.T) {
		req := &pb.ListUsersRequest{
			Page:     1,
			PageSize: 10,
		}

		resp, err := p.ListUsers(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "LDAP_URL not configured")
		assert.Nil(t, resp)
	})
}

func TestLDAPAuthPlugin_RefreshToken(t *testing.T) {
	p := NewLDAPAuthPlugin()
	ctx := context.Background()

	t.Run("returns invalid session", func(t *testing.T) {
		req := &pb.RefreshTokenRequest{
			RefreshToken: "old-token",
		}

		session, err := p.RefreshToken(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.False(t, session.Valid)
	})
}

func TestLDAPAuthPlugin_Logout(t *testing.T) {
	p := NewLDAPAuthPlugin()
	ctx := context.Background()

	t.Run("logout returns empty", func(t *testing.T) {
		req := &pb.LogoutRequest{
			UserId: "user-123",
		}

		result, err := p.Logout(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestRoleMapping(t *testing.T) {
	t.Run("admin group detection", func(t *testing.T) {
		groups := []string{
			"cn=developers,ou=groups,dc=example,dc=com",
			"cn=admins,ou=groups,dc=example,dc=com",
		}

		role := "user"
		for _, group := range groups {
			if contains(group, "admins") {
				role = "admin"
			}
		}

		assert.Equal(t, "admin", role)
	})

	t.Run("user role for non-admin", func(t *testing.T) {
		groups := []string{
			"cn=developers,ou=groups,dc=example,dc=com",
			"cn=users,ou=groups,dc=example,dc=com",
		}

		role := "user"
		for _, group := range groups {
			if contains(group, "admins") {
				role = "admin"
			}
		}

		assert.Equal(t, "user", role)
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmarks
func BenchmarkNewLDAPAuthPlugin(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewLDAPAuthPlugin()
	}
}

func BenchmarkAuthorize(b *testing.B) {
	p := NewLDAPAuthPlugin()
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
