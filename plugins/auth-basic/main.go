package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-plugin"
	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/rennerdo30/webencode/pkg/pluginsdk"
)

type BasicAuthPlugin struct {
	pb.UnimplementedAuthServiceServer
	logger *logger.Logger
	users  map[string]*User
	mu     sync.RWMutex
}

type User struct {
	ID           string
	Username     string
	PasswordHash string
	Email        string
	Role         string
}

func NewBasicAuthPlugin() *BasicAuthPlugin {
	p := &BasicAuthPlugin{
		logger: logger.New("plugin-auth-basic"),
		users:  make(map[string]*User),
	}

	// Load users from environment or use defaults
	adminPass := os.Getenv("AUTH_ADMIN_PASSWORD")
	if adminPass == "" {
		adminPass = "admin123"
	}

	userPass := os.Getenv("AUTH_USER_PASSWORD")
	if userPass == "" {
		userPass = "user123"
	}

	// Add default users
	p.users["admin"] = &User{
		ID:           "00000000-0000-0000-0000-000000000001",
		Username:     "admin",
		PasswordHash: hashPassword(adminPass),
		Email:        "admin@localhost",
		Role:         "admin",
	}

	p.users["user"] = &User{
		ID:           "00000000-0000-0000-0000-000000000002",
		Username:     "user",
		PasswordHash: hashPassword(userPass),
		Email:        "user@localhost",
		Role:         "user",
	}

	return p
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func (p *BasicAuthPlugin) ValidateToken(ctx context.Context, req *pb.TokenRequest) (*pb.UserSession, error) {
	token := req.Token

	// Handle Basic auth header
	if req.Scheme == "basic" || strings.HasPrefix(token, "Basic ") {
		token = strings.TrimPrefix(token, "Basic ")
		decoded, err := base64.StdEncoding.DecodeString(token)
		if err != nil {
			return &pb.UserSession{Valid: false}, nil
		}

		parts := strings.SplitN(string(decoded), ":", 2)
		if len(parts) != 2 {
			return &pb.UserSession{Valid: false}, nil
		}

		username, password := parts[0], parts[1]
		return p.authenticate(username, password)
	}

	// Handle simple token format: "basic-<username>-<password>"
	if strings.HasPrefix(token, "basic-") {
		parts := strings.SplitN(strings.TrimPrefix(token, "basic-"), "-", 2)
		if len(parts) == 2 {
			return p.authenticate(parts[0], parts[1])
		}
	}

	return &pb.UserSession{Valid: false}, nil
}

func (p *BasicAuthPlugin) authenticate(username, password string) (*pb.UserSession, error) {
	p.mu.RLock()
	user, exists := p.users[username]
	p.mu.RUnlock()

	if !exists {
		p.logger.Warn("User not found", "username", username)
		return &pb.UserSession{Valid: false}, nil
	}

	if user.PasswordHash != hashPassword(password) {
		p.logger.Warn("Invalid password", "username", username)
		return &pb.UserSession{Valid: false}, nil
	}

	p.logger.Info("User authenticated", "username", username, "role", user.Role)
	return &pb.UserSession{
		Valid:     true,
		UserId:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		Roles:     []string{user.Role},
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}, nil
}

func (p *BasicAuthPlugin) Authorize(ctx context.Context, req *pb.AuthZRequest) (*pb.AuthZResponse, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Find user by ID
	var user *User
	for _, u := range p.users {
		if u.ID == req.UserId {
			user = u
			break
		}
	}

	if user == nil {
		return &pb.AuthZResponse{Allowed: false, Reason: "User not found"}, nil
	}

	// Admin can do everything
	if user.Role == "admin" {
		return &pb.AuthZResponse{Allowed: true}, nil
	}

	// Regular users can read and create, but not delete workers/plugins
	switch req.Action {
	case "read":
		return &pb.AuthZResponse{Allowed: true}, nil
	case "create":
		if req.ResourceType != "plugin" && req.ResourceType != "worker" {
			return &pb.AuthZResponse{Allowed: true}, nil
		}
	case "update":
		if req.ResourceType == "job" || req.ResourceType == "stream" {
			return &pb.AuthZResponse{Allowed: true}, nil
		}
	case "delete":
		if req.ResourceType == "job" || req.ResourceType == "stream" {
			return &pb.AuthZResponse{Allowed: true}, nil
		}
	}

	return &pb.AuthZResponse{
		Allowed: false,
		Reason:  "Permission denied for action " + req.Action + " on " + req.ResourceType,
	}, nil
}

func (p *BasicAuthPlugin) GetUser(ctx context.Context, req *pb.UserRequest) (*pb.User, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, user := range p.users {
		if user.ID == req.UserId {
			return &pb.User{
				Id:          user.ID,
				Username:    user.Username,
				Email:       user.Email,
				FullName:    user.Username,
				DisplayName: user.Username,
				Role:        user.Role,
				Roles:       []string{user.Role},
			}, nil
		}
	}

	return nil, nil
}

func (p *BasicAuthPlugin) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var users []*pb.User
	for _, user := range p.users {
		if req.RoleFilter != "" && user.Role != req.RoleFilter {
			continue
		}
		users = append(users, &pb.User{
			Id:          user.ID,
			Username:    user.Username,
			Email:       user.Email,
			FullName:    user.Username,
			DisplayName: user.Username,
			Role:        user.Role,
			Roles:       []string{user.Role},
		})
	}

	return &pb.ListUsersResponse{
		Users:      users,
		TotalCount: int32(len(users)),
		Page:       1,
	}, nil
}

func (p *BasicAuthPlugin) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.UserSession, error) {
	// Basic auth doesn't really have refresh tokens
	return &pb.UserSession{Valid: false}, nil
}

func (p *BasicAuthPlugin) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.Empty, error) {
	p.logger.Info("User logout", "user_id", req.UserId)
	return &pb.Empty{}, nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: pluginsdk.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"webencode": &pluginsdk.Plugin{
				AuthImpl: NewBasicAuthPlugin(),
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
