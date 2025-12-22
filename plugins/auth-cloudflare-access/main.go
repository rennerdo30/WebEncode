// Package main implements the Cloudflare Zero Trust Access authentication plugin.
// This plugin validates JWT tokens from Cloudflare Access, enabling Google SSO
// authentication through Cloudflare's Zero Trust platform.
//
// Authentication Flow:
// 1. User visits the application behind Cloudflare Access
// 2. Cloudflare redirects to Google login if not authenticated
// 3. After authentication, Cloudflare adds JWT headers/cookies to requests:
//   - Cf-Access-Jwt-Assertion: Full JWT token
//   - CF_Authorization: Cookie with JWT
//   - cf-access-authenticated-user-email: User's email
//
// 4. This plugin validates the JWT against Cloudflare's public keys
// 5. User info is extracted and returned for session creation
package main

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hashicorp/go-plugin"
	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/rennerdo30/webencode/pkg/pluginsdk"
)

// Config holds the Cloudflare Access configuration
type Config struct {
	TeamDomain string // Team domain (e.g., "renner" for renner.cloudflareaccess.com)
	Audience   string // Application audience tag from CF Access settings
	DevMode    bool   // Enable dev mode for testing
}

// CloudflareAccessClaims represents the claims in a Cloudflare Access JWT
type CloudflareAccessClaims struct {
	jwt.RegisteredClaims
	Email         string   `json:"email"`
	Type          string   `json:"type"`
	IdentityNonce string   `json:"identity_nonce"`
	Country       string   `json:"country"`
	Aud           []string `json:"aud"`
}

// JWKS represents the JSON Web Key Set response from Cloudflare
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a single JSON Web Key
type JWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// KeyCache caches Cloudflare's public keys
type KeyCache struct {
	keys      map[string]*rsa.PublicKey
	expiresAt time.Time
	mu        sync.RWMutex
}

// AuthPlugin implements the Cloudflare Access authentication plugin
type AuthPlugin struct {
	pb.UnimplementedAuthServiceServer
	logger   *logger.Logger
	config   Config
	keyCache *KeyCache
}

// NewAuthPlugin creates a new Cloudflare Access auth plugin
func NewAuthPlugin() *AuthPlugin {
	cfg := Config{
		TeamDomain: os.Getenv("CF_ACCESS_TEAM_DOMAIN"),
		Audience:   os.Getenv("CF_ACCESS_AUDIENCE"),
		DevMode:    os.Getenv("CF_ACCESS_DEV_MODE") == "true" || os.Getenv("CF_ACCESS_DEV_MODE") == "1",
	}

	return &AuthPlugin{
		logger:   logger.New("plugin-auth-cloudflare-access"),
		config:   cfg,
		keyCache: &KeyCache{keys: make(map[string]*rsa.PublicKey)},
	}
}

// fetchJWKS fetches the JSON Web Key Set from Cloudflare
func (p *AuthPlugin) fetchJWKS(ctx context.Context) error {
	p.keyCache.mu.Lock()
	defer p.keyCache.mu.Unlock()

	// Check if cache is still valid
	if time.Now().Before(p.keyCache.expiresAt) && len(p.keyCache.keys) > 0 {
		return nil
	}

	url := fmt.Sprintf("https://%s.cloudflareaccess.com/cdn-cgi/access/certs", p.config.TeamDomain)
	p.logger.Info("Fetching JWKS from Cloudflare", "url", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("failed to decode JWKS: %w", err)
	}

	// Parse and cache the keys
	p.keyCache.keys = make(map[string]*rsa.PublicKey)
	for _, key := range jwks.Keys {
		if key.Kty != "RSA" {
			continue
		}

		publicKey, err := parseRSAPublicKey(key)
		if err != nil {
			p.logger.Warn("Failed to parse RSA key", "kid", key.Kid, "error", err)
			continue
		}
		p.keyCache.keys[key.Kid] = publicKey
	}

	// Cache for 1 hour
	p.keyCache.expiresAt = time.Now().Add(1 * time.Hour)
	p.logger.Info("JWKS cached successfully", "keyCount", len(p.keyCache.keys))

	return nil
}

// getPublicKey returns the public key for the given kid
func (p *AuthPlugin) getPublicKey(kid string) (*rsa.PublicKey, error) {
	p.keyCache.mu.RLock()
	defer p.keyCache.mu.RUnlock()

	key, ok := p.keyCache.keys[kid]
	if !ok {
		return nil, fmt.Errorf("key not found: %s", kid)
	}
	return key, nil
}

// ValidateToken validates a Cloudflare Access JWT and returns user session info
func (p *AuthPlugin) ValidateToken(ctx context.Context, req *pb.TokenRequest) (*pb.UserSession, error) {
	token := req.Token

	// Check for dev mode tokens first
	if p.config.DevMode && strings.HasPrefix(token, "dev-") {
		return p.handleDevToken(token)
	}

	// Validate configuration
	if p.config.TeamDomain == "" || p.config.Audience == "" {
		if p.config.DevMode {
			p.logger.Warn("CF Access not configured, running in dev mode only")
			return &pb.UserSession{Valid: false}, nil
		}
		return nil, errors.New("CF_ACCESS_TEAM_DOMAIN and CF_ACCESS_AUDIENCE are required")
	}

	// Fetch JWKS if needed
	if err := p.fetchJWKS(ctx); err != nil {
		p.logger.Error("Failed to fetch JWKS", "error", err)
		return &pb.UserSession{Valid: false}, nil
	}

	// Parse and validate the token
	parsedToken, err := jwt.ParseWithClaims(token, &CloudflareAccessClaims{}, func(t *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		// Get the key ID
		kid, ok := t.Header["kid"].(string)
		if !ok {
			return nil, errors.New("missing kid in token header")
		}

		// Get the public key
		return p.getPublicKey(kid)
	})

	if err != nil {
		p.logger.Debug("Token validation failed", "error", err)
		return &pb.UserSession{Valid: false}, nil
	}

	claims, ok := parsedToken.Claims.(*CloudflareAccessClaims)
	if !ok || !parsedToken.Valid {
		p.logger.Debug("Invalid token claims")
		return &pb.UserSession{Valid: false}, nil
	}

	// Verify audience
	audienceValid := false
	for _, aud := range claims.Aud {
		if aud == p.config.Audience {
			audienceValid = true
			break
		}
	}
	if !audienceValid {
		p.logger.Debug("Invalid audience in token")
		return &pb.UserSession{Valid: false}, nil
	}

	// Verify issuer
	expectedIssuer := fmt.Sprintf("https://%s.cloudflareaccess.com", p.config.TeamDomain)
	if claims.Issuer != expectedIssuer {
		p.logger.Debug("Invalid issuer in token", "expected", expectedIssuer, "got", claims.Issuer)
		return &pb.UserSession{Valid: false}, nil
	}

	// Build user session
	return &pb.UserSession{
		Valid:     true,
		UserId:    claims.Subject,
		Username:  extractUsername(claims.Email),
		Email:     claims.Email,
		Role:      "user", // Default role, can be enhanced with group mapping
		Roles:     []string{"user"},
		ExpiresAt: claims.ExpiresAt.Time.Unix(),
		Metadata: map[string]string{
			"country": claims.Country,
			"type":    claims.Type,
		},
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
		"create": true,
	}

	if userAllowedActions[req.Action] {
		return &pb.AuthZResponse{
			Allowed: true,
			Reason:  "action allowed for user role",
		}, nil
	}

	return &pb.AuthZResponse{
		Allowed: false,
		Reason:  fmt.Sprintf("action '%s' not allowed for roles: %v", req.Action, roles),
	}, nil
}

// GetUser retrieves user information
func (p *AuthPlugin) GetUser(ctx context.Context, req *pb.UserRequest) (*pb.User, error) {
	// In dev mode or when provider is not set, return stub user
	if p.config.DevMode || p.config.TeamDomain == "" {
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

	// For CF Access, we don't have a user info endpoint
	// User info comes from the JWT claims during authentication
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

// ListUsers lists users (not supported by CF Access)
func (p *AuthPlugin) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	// CF Access doesn't provide user listing
	if p.config.DevMode || p.config.TeamDomain == "" {
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

	p.logger.Warn("ListUsers not supported for CF Access")
	return &pb.ListUsersResponse{
		Users:      []*pb.User{},
		TotalCount: 0,
		Page:       req.Page,
	}, nil
}

// RefreshToken is not supported by CF Access (tokens are managed by Cloudflare)
func (p *AuthPlugin) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.UserSession, error) {
	return &pb.UserSession{Valid: false}, nil
}

// Logout handles user logout
func (p *AuthPlugin) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.Empty, error) {
	p.logger.Info("User logged out", "user_id", req.UserId, "all_sessions", req.AllSessions)
	return &pb.Empty{}, nil
}

// Helper functions

// extractUsername extracts username from email
func extractUsername(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) > 0 {
		return parts[0]
	}
	return email
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

// parseRSAPublicKey parses a JWK into an RSA public key
func parseRSAPublicKey(key JWK) (*rsa.PublicKey, error) {
	// Decode the modulus and exponent from base64url
	nBytes, err := jwt.NewParser().DecodeSegment(key.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	eBytes, err := jwt.NewParser().DecodeSegment(key.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// Convert to big integers
	n := new(big.Int).SetBytes(nBytes)

	// Convert exponent bytes to int
	var e int
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}

	return &rsa.PublicKey{N: n, E: e}, nil
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
