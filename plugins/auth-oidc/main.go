package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/hashicorp/go-plugin"
	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/rennerdo30/webencode/pkg/pluginsdk"
	"golang.org/x/oauth2"
)

// Config holds the OIDC configuration
type Config struct {
	IssuerURL    string // OIDC issuer URL (e.g., https://keycloak.example.com/realms/myrealm)
	ClientID     string // OAuth2 client ID
	ClientSecret string // OAuth2 client secret (optional for public clients)
	DevMode      bool   // Enable dev mode for testing (accepts dev-* tokens)
}

// Claims represents the claims extracted from an OIDC token
type Claims struct {
	Subject           string   `json:"sub"`
	Email             string   `json:"email"`
	EmailVerified     bool     `json:"email_verified"`
	Name              string   `json:"name"`
	PreferredUsername string   `json:"preferred_username"`
	GivenName         string   `json:"given_name"`
	FamilyName        string   `json:"family_name"`
	Groups            []string `json:"groups"`
	Roles             []string `json:"roles"`
	RealmAccess       struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
	ResourceAccess map[string]struct {
		Roles []string `json:"roles"`
	} `json:"resource_access"`
}

// AuthPlugin implements the OIDC authentication plugin
type AuthPlugin struct {
	pb.UnimplementedAuthServiceServer
	logger   *logger.Logger
	config   Config
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
	oauth2   *oauth2.Config
	mu       sync.RWMutex
	initOnce sync.Once
	initErr  error
}

// NewAuthPlugin creates a new OIDC auth plugin
func NewAuthPlugin() *AuthPlugin {
	cfg := Config{
		IssuerURL:    os.Getenv("OIDC_ISSUER_URL"),
		ClientID:     os.Getenv("OIDC_CLIENT_ID"),
		ClientSecret: os.Getenv("OIDC_CLIENT_SECRET"),
		DevMode:      os.Getenv("OIDC_DEV_MODE") == "true" || os.Getenv("OIDC_DEV_MODE") == "1",
	}

	return &AuthPlugin{
		logger: logger.New("plugin-auth-oidc"),
		config: cfg,
	}
}

// initProvider initializes the OIDC provider (lazy initialization)
func (p *AuthPlugin) initProvider(ctx context.Context) error {
	p.initOnce.Do(func() {
		if p.config.IssuerURL == "" {
			if p.config.DevMode {
				p.logger.Warn("OIDC not configured, running in dev mode only")
				return
			}
			p.initErr = fmt.Errorf("OIDC_ISSUER_URL is required")
			return
		}

		p.logger.Info("Initializing OIDC provider", "issuer", p.config.IssuerURL)

		// Create context with timeout for discovery
		discoverCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		provider, err := oidc.NewProvider(discoverCtx, p.config.IssuerURL)
		if err != nil {
			p.initErr = fmt.Errorf("failed to create OIDC provider: %w", err)
			p.logger.Error("Failed to initialize OIDC provider", "error", err)
			return
		}

		p.provider = provider
		p.verifier = provider.Verifier(&oidc.Config{
			ClientID: p.config.ClientID,
		})

		p.oauth2 = &oauth2.Config{
			ClientID:     p.config.ClientID,
			ClientSecret: p.config.ClientSecret,
			Endpoint:     provider.Endpoint(),
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		}

		p.logger.Info("OIDC provider initialized successfully")
	})

	return p.initErr
}

// ValidateToken validates an OIDC token and returns user session info
func (p *AuthPlugin) ValidateToken(ctx context.Context, req *pb.TokenRequest) (*pb.UserSession, error) {
	token := req.Token

	// Check for dev mode tokens first
	if p.config.DevMode && strings.HasPrefix(token, "dev-") {
		return p.handleDevToken(token)
	}

	// Initialize provider if not already done
	if err := p.initProvider(ctx); err != nil {
		// If OIDC not configured but dev mode is on, we already handled dev tokens above
		if p.config.DevMode {
			p.logger.Warn("OIDC not configured, rejecting non-dev token")
			return &pb.UserSession{Valid: false}, nil
		}
		return nil, fmt.Errorf("OIDC provider not initialized: %w", err)
	}

	if p.verifier == nil {
		return &pb.UserSession{Valid: false}, nil
	}

	// Verify the token
	idToken, err := p.verifier.Verify(ctx, token)
	if err != nil {
		p.logger.Debug("Token verification failed", "error", err)
		return &pb.UserSession{Valid: false}, nil
	}

	// Extract claims
	var claims Claims
	if err := idToken.Claims(&claims); err != nil {
		p.logger.Error("Failed to parse claims", "error", err)
		return &pb.UserSession{Valid: false}, nil
	}

	// Determine roles
	roles := p.extractRoles(&claims)
	primaryRole := "user"
	if contains(roles, "admin") {
		primaryRole = "admin"
	}

	// Build display name
	displayName := claims.PreferredUsername
	if displayName == "" {
		displayName = claims.Name
	}
	if displayName == "" {
		displayName = claims.Email
	}

	return &pb.UserSession{
		Valid:     true,
		UserId:    claims.Subject,
		Username:  claims.PreferredUsername,
		Email:     claims.Email,
		Role:      primaryRole,
		Roles:     roles,
		ExpiresAt: idToken.Expiry.Unix(),
	}, nil
}

// handleDevToken handles dev-mode tokens
func (p *AuthPlugin) handleDevToken(token string) (*pb.UserSession, error) {
	role := "user"
	if strings.Contains(token, "admin") {
		role = "admin"
	}

	userID := "00000000-0000-0000-0000-000000000001"
	if role == "user" {
		userID = "00000000-0000-0000-0000-000000000002"
	}

	p.logger.Debug("Dev mode token accepted", "role", role)

	return &pb.UserSession{
		Valid:     true,
		UserId:    userID,
		Username:  "dev-user",
		Email:     "dev@example.com",
		Role:      role,
		Roles:     []string{role},
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}, nil
}

// extractRoles extracts roles from various claim locations (Keycloak compatible)
func (p *AuthPlugin) extractRoles(claims *Claims) []string {
	roleSet := make(map[string]bool)

	// Add direct roles claim
	for _, r := range claims.Roles {
		roleSet[r] = true
	}

	// Add groups (often used as roles)
	for _, g := range claims.Groups {
		roleSet[g] = true
	}

	// Add realm roles (Keycloak)
	for _, r := range claims.RealmAccess.Roles {
		roleSet[r] = true
	}

	// Add client-specific roles (Keycloak)
	if clientRoles, ok := claims.ResourceAccess[p.config.ClientID]; ok {
		for _, r := range clientRoles.Roles {
			roleSet[r] = true
		}
	}

	// Convert to slice
	roles := make([]string, 0, len(roleSet))
	for r := range roleSet {
		roles = append(roles, r)
	}

	// Ensure at least "user" role
	if len(roles) == 0 {
		roles = []string{"user"}
	}

	return roles
}

// Authorize checks if a user is authorized to perform an action
func (p *AuthPlugin) Authorize(ctx context.Context, req *pb.AuthZRequest) (*pb.AuthZResponse, error) {
	// Get user to determine their roles
	user, err := p.GetUser(ctx, &pb.UserRequest{UserId: req.UserId})
	if err != nil {
		return &pb.AuthZResponse{
			Allowed: false,
			Reason:  fmt.Sprintf("failed to get user: %v", err),
		}, nil
	}

	roles := user.Roles
	if len(roles) == 0 {
		roles = []string{user.Role}
	}

	// Simple RBAC: admin can do anything, users have limited access
	if contains(roles, "admin") {
		return &pb.AuthZResponse{
			Allowed: true,
			Reason:  "admin role has full access",
		}, nil
	}

	// Define allowed actions for regular users
	userAllowedActions := map[string]bool{
		"read":   true,
		"list":   true,
		"create": true, // Users can create their own resources
	}

	// Check if action is in the allowed list
	if userAllowedActions[req.Action] {
		return &pb.AuthZResponse{
			Allowed: true,
			Reason:  "action allowed for user role",
		}, nil
	}

	// Deny by default
	return &pb.AuthZResponse{
		Allowed: false,
		Reason:  fmt.Sprintf("action '%s' not allowed for roles: %v", req.Action, roles),
	}, nil
}

// GetUser retrieves user information
func (p *AuthPlugin) GetUser(ctx context.Context, req *pb.UserRequest) (*pb.User, error) {
	// In dev mode, return stub user
	if p.config.DevMode || p.provider == nil {
		return &pb.User{
			Id:          req.UserId,
			Username:    "dev-user",
			Email:       "dev@example.com",
			FullName:    "Dev User",
			DisplayName: "Dev",
			Role:        "user",
			Roles:       []string{"user"},
		}, nil
	}

	// For real OIDC, we would need to call the userinfo endpoint with a valid access token
	// This typically requires storing the access token from the original authentication
	// For now, return basic info based on the user ID
	p.logger.Debug("GetUser called", "user_id", req.UserId)

	return &pb.User{
		Id:          req.UserId,
		Username:    "unknown",
		Email:       "",
		FullName:    "",
		DisplayName: "",
		Role:        "user",
		Roles:       []string{"user"},
	}, nil
}

// ListUsers lists users (limited functionality without admin access to OIDC provider)
func (p *AuthPlugin) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	// OIDC providers typically don't expose user listing via standard protocols
	// This would require provider-specific admin APIs (e.g., Keycloak Admin REST API)

	if p.config.DevMode || p.provider == nil {
		return &pb.ListUsersResponse{
			Users: []*pb.User{
				{
					Id:          "00000000-0000-0000-0000-000000000001",
					Username:    "admin",
					Email:       "admin@example.com",
					FullName:    "Admin User",
					DisplayName: "Admin",
					Role:        "admin",
					Roles:       []string{"admin"},
				},
				{
					Id:          "00000000-0000-0000-0000-000000000002",
					Username:    "user",
					Email:       "user@example.com",
					FullName:    "Regular User",
					DisplayName: "User",
					Role:        "user",
					Roles:       []string{"user"},
				},
			},
			TotalCount: 2,
			Page:       1,
		}, nil
	}

	// In production, this would require Keycloak Admin API integration
	p.logger.Warn("ListUsers not fully implemented for OIDC - requires admin API access")
	return &pb.ListUsersResponse{
		Users:      []*pb.User{},
		TotalCount: 0,
		Page:       req.Page,
	}, nil
}

// RefreshToken refreshes an access token using a refresh token
func (p *AuthPlugin) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.UserSession, error) {
	// In dev mode, return a new dev session
	if p.config.DevMode || p.oauth2 == nil {
		return &pb.UserSession{
			Valid:        true,
			UserId:       "00000000-0000-0000-0000-000000000001",
			Username:     "dev-user",
			Email:        "dev@example.com",
			Role:         "user",
			Roles:        []string{"user"},
			ExpiresAt:    time.Now().Add(24 * time.Hour).Unix(),
			AccessToken:  "dev-refreshed-token",
			RefreshToken: req.RefreshToken,
		}, nil
	}

	// Use OAuth2 token source to refresh
	token := &oauth2.Token{
		RefreshToken: req.RefreshToken,
	}

	tokenSource := p.oauth2.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		p.logger.Error("Failed to refresh token", "error", err)
		return &pb.UserSession{Valid: false}, nil
	}

	// Extract ID token from the new token
	rawIDToken, ok := newToken.Extra("id_token").(string)
	if !ok {
		p.logger.Error("No id_token in refreshed token response")
		return &pb.UserSession{Valid: false}, nil
	}

	// Verify the new ID token
	idToken, err := p.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		p.logger.Error("Failed to verify refreshed token", "error", err)
		return &pb.UserSession{Valid: false}, nil
	}

	// Extract claims
	var claims Claims
	if err := idToken.Claims(&claims); err != nil {
		p.logger.Error("Failed to parse claims", "error", err)
		return &pb.UserSession{Valid: false}, nil
	}

	roles := p.extractRoles(&claims)
	primaryRole := "user"
	if contains(roles, "admin") {
		primaryRole = "admin"
	}

	return &pb.UserSession{
		Valid:        true,
		UserId:       claims.Subject,
		Username:     claims.PreferredUsername,
		Email:        claims.Email,
		Role:         primaryRole,
		Roles:        roles,
		ExpiresAt:    idToken.Expiry.Unix(),
		AccessToken:  newToken.AccessToken,
		RefreshToken: newToken.RefreshToken,
	}, nil
}

// Logout handles user logout
func (p *AuthPlugin) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.Empty, error) {
	p.logger.Info("User logged out", "user_id", req.UserId, "all_sessions", req.AllSessions)

	// For OIDC, we could implement RP-initiated logout if the provider supports it
	// This would redirect to the provider's end_session_endpoint

	return &pb.Empty{}, nil
}

// contains checks if a string slice contains a value
func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: pluginsdk.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"webencode": &pluginsdk.Plugin{
				AuthImpl: NewAuthPlugin(),
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
