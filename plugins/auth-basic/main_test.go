package main

import (
	"context"
	"encoding/base64"
	"os"
	"testing"

	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/stretchr/testify/assert"
)

func TestValidateToken_Success(t *testing.T) {
	os.Setenv("AUTH_ADMIN_PASSWORD", "secret")
	p := NewBasicAuthPlugin()

	token := base64.StdEncoding.EncodeToString([]byte("admin:secret"))

	resp, err := p.ValidateToken(context.Background(), &pb.TokenRequest{
		Token: "Basic " + token,
	})
	assert.NoError(t, err)
	assert.True(t, resp.Valid)
	assert.Equal(t, "admin", resp.Username)
	assert.Equal(t, "admin", resp.Role)
}

func TestValidateToken_Fail_Password(t *testing.T) {
	os.Setenv("AUTH_ADMIN_PASSWORD", "secret")
	p := NewBasicAuthPlugin()

	token := base64.StdEncoding.EncodeToString([]byte("admin:wrong"))

	resp, err := p.ValidateToken(context.Background(), &pb.TokenRequest{
		Token: "Basic " + token,
	})
	assert.NoError(t, err)
	assert.False(t, resp.Valid)
}

func TestValidateToken_Fail_User(t *testing.T) {
	p := NewBasicAuthPlugin()

	token := base64.StdEncoding.EncodeToString([]byte("nobody:secret"))

	resp, err := p.ValidateToken(context.Background(), &pb.TokenRequest{
		Token: "Basic " + token,
	})
	assert.NoError(t, err)
	assert.False(t, resp.Valid)
}

func TestAuthorize_Admin(t *testing.T) {
	p := NewBasicAuthPlugin()

	// Admin ID
	resp, err := p.Authorize(context.Background(), &pb.AuthZRequest{
		UserId:       "00000000-0000-0000-0000-000000000001",
		Action:       "delete",
		ResourceType: "plugin",
	})
	assert.NoError(t, err)
	assert.True(t, resp.Allowed)
}

func TestAuthorize_User(t *testing.T) {
	p := NewBasicAuthPlugin()

	// User ID
	userID := "00000000-0000-0000-0000-000000000002"

	// Should be able to read
	resp, err := p.Authorize(context.Background(), &pb.AuthZRequest{
		UserId:       userID,
		Action:       "read",
		ResourceType: "job",
	})
	assert.NoError(t, err)
	assert.True(t, resp.Allowed)

	// Should NOT be able to create plugin
	resp, err = p.Authorize(context.Background(), &pb.AuthZRequest{
		UserId:       userID,
		Action:       "create",
		ResourceType: "plugin",
	})
	assert.NoError(t, err)
	assert.False(t, resp.Allowed)
}

func TestListUsers(t *testing.T) {
	p := NewBasicAuthPlugin()

	resp, err := p.ListUsers(context.Background(), &pb.ListUsersRequest{})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(resp.Users), 2)
}

func TestGetUser(t *testing.T) {
	p := NewBasicAuthPlugin()
	resp, err := p.GetUser(context.Background(), &pb.UserRequest{UserId: "00000000-0000-0000-0000-000000000001"})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "admin", resp.Username)
}
