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

func TestLDAPConfigFields(t *testing.T) {
	t.Run("all config fields", func(t *testing.T) {
		config := &LDAPConfig{
			URL:        "ldap://localhost:389",
			BaseDN:     "dc=example,dc=com",
			BindDN:     "cn=admin,dc=example,dc=com",
			BindPass:   "secretpassword",
			UserFilter: "(&(objectClass=person)(uid=%s))",
			GroupDN:    "ou=groups,dc=example,dc=com",
		}

		assert.Equal(t, "ldap://localhost:389", config.URL)
		assert.Equal(t, "dc=example,dc=com", config.BaseDN)
		assert.Equal(t, "cn=admin,dc=example,dc=com", config.BindDN)
		assert.Equal(t, "secretpassword", config.BindPass)
		assert.Contains(t, config.UserFilter, "objectClass=person")
		assert.Equal(t, "ou=groups,dc=example,dc=com", config.GroupDN)
	})

	t.Run("empty config", func(t *testing.T) {
		config := &LDAPConfig{}

		assert.Empty(t, config.URL)
		assert.Empty(t, config.BaseDN)
		assert.Empty(t, config.BindDN)
		assert.Empty(t, config.BindPass)
	})
}

func TestLDAPURLSchemes(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		isLDAPS  bool
	}{
		{"plain LDAP", "ldap://ldap.example.com:389", false},
		{"LDAPS", "ldaps://ldap.example.com:636", true},
		{"LDAP with IP", "ldap://192.168.1.100:389", false},
		{"LDAPS with IP", "ldaps://192.168.1.100:636", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			config := &LDAPConfig{URL: tc.url}
			if tc.isLDAPS {
				assert.True(t, len(config.URL) >= 8 && config.URL[:8] == "ldaps://")
			} else {
				assert.True(t, len(config.URL) >= 7 && config.URL[:7] == "ldap://")
			}
		})
	}
}

func TestUserSessionStructure(t *testing.T) {
	t.Run("valid session", func(t *testing.T) {
		session := &pb.UserSession{
			Valid:     true,
			UserId:    "cn=john,ou=users,dc=example,dc=com",
			Username:  "john",
			Email:     "john@example.com",
			Role:      "admin",
			Roles:     []string{"admin", "user"},
			ExpiresAt: 1700000000,
		}

		assert.True(t, session.Valid)
		assert.Contains(t, session.UserId, "cn=john")
		assert.Equal(t, "john", session.Username)
		assert.Equal(t, "john@example.com", session.Email)
		assert.Equal(t, "admin", session.Role)
		assert.Len(t, session.Roles, 2)
		assert.Greater(t, session.ExpiresAt, int64(0))
	})

	t.Run("invalid session", func(t *testing.T) {
		session := &pb.UserSession{
			Valid: false,
		}

		assert.False(t, session.Valid)
		assert.Empty(t, session.UserId)
	})
}

func TestUserStructure(t *testing.T) {
	t.Run("full user info", func(t *testing.T) {
		user := &pb.User{
			Id:          "cn=jane,ou=users,dc=example,dc=com",
			Username:    "jane",
			Email:       "jane@example.com",
			FullName:    "Jane Doe",
			DisplayName: "Jane D.",
			Role:        "user",
			Roles:       []string{"user"},
		}

		assert.Contains(t, user.Id, "cn=jane")
		assert.Equal(t, "jane", user.Username)
		assert.Equal(t, "jane@example.com", user.Email)
		assert.Equal(t, "Jane Doe", user.FullName)
		assert.Equal(t, "Jane D.", user.DisplayName)
	})
}

func TestListUsersRequest(t *testing.T) {
	t.Run("pagination parameters", func(t *testing.T) {
		req := &pb.ListUsersRequest{
			Page:     1,
			PageSize: 50,
		}

		assert.Equal(t, int32(1), req.Page)
		assert.Equal(t, int32(50), req.PageSize)
	})

	t.Run("zero values use defaults", func(t *testing.T) {
		req := &pb.ListUsersRequest{}

		pageSize := int(req.PageSize)
		if pageSize == 0 {
			pageSize = 50
		}
		page := int(req.Page)
		if page == 0 {
			page = 1
		}

		assert.Equal(t, 50, pageSize)
		assert.Equal(t, 1, page)
	})
}

func TestGroupMembershipDetection(t *testing.T) {
	tests := []struct {
		name         string
		groups       []string
		expectedRole string
	}{
		{
			name: "admin in admins group",
			groups: []string{
				"cn=developers,ou=groups,dc=example,dc=com",
				"cn=admins,ou=groups,dc=example,dc=com",
			},
			expectedRole: "admin",
		},
		{
			name: "user in no admin group",
			groups: []string{
				"cn=developers,ou=groups,dc=example,dc=com",
				"cn=users,ou=groups,dc=example,dc=com",
			},
			expectedRole: "user",
		},
		{
			name:         "no groups",
			groups:       []string{},
			expectedRole: "user",
		},
		{
			name: "admin case sensitive",
			groups: []string{
				"cn=Admins,ou=groups,dc=example,dc=com", // Different case
			},
			expectedRole: "user", // Won't match because strings.Contains is case-sensitive
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			role := "user"
			for _, group := range tc.groups {
				if contains(group, "admins") {
					role = "admin"
				}
			}
			assert.Equal(t, tc.expectedRole, role)
		})
	}
}

func TestAuthZActions(t *testing.T) {
	p := NewLDAPAuthPlugin()
	ctx := context.Background()

	actions := []string{"create", "read", "update", "delete", "list", "admin"}

	for _, action := range actions {
		t.Run("action_"+action, func(t *testing.T) {
			req := &pb.AuthZRequest{
				UserId:       "user-123",
				Action:       action,
				ResourceType: "job",
			}

			resp, err := p.Authorize(ctx, req)
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			// Default implementation allows all
			assert.True(t, resp.Allowed)
		})
	}
}

func TestTokenRequestStructure(t *testing.T) {
	t.Run("bearer token", func(t *testing.T) {
		req := &pb.TokenRequest{
			Token: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		}

		assert.True(t, len(req.Token) > 7 && req.Token[:7] == "Bearer ")
	})

	t.Run("basic auth token", func(t *testing.T) {
		req := &pb.TokenRequest{
			Token: "Basic dXNlcm5hbWU6cGFzc3dvcmQ=",
		}

		assert.True(t, len(req.Token) > 6 && req.Token[:6] == "Basic ")
	})
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

func BenchmarkRoleMapping(b *testing.B) {
	groups := []string{
		"cn=developers,ou=groups,dc=example,dc=com",
		"cn=admins,ou=groups,dc=example,dc=com",
		"cn=users,ou=groups,dc=example,dc=com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		role := "user"
		for _, group := range groups {
			if contains(group, "admins") {
				role = "admin"
				break
			}
		}
		_ = role
	}
}
