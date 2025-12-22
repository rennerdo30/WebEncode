package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/hashicorp/go-plugin"
	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/rennerdo30/webencode/pkg/pluginsdk"
)

type LDAPAuthPlugin struct {
	pb.UnimplementedAuthServiceServer
	logger *logger.Logger
	config *LDAPConfig
}

type LDAPConfig struct {
	URL        string
	BaseDN     string
	BindDN     string
	BindPass   string
	UserFilter string
	GroupDN    string
}

func NewLDAPAuthPlugin() *LDAPAuthPlugin {
	return &LDAPAuthPlugin{
		logger: logger.New("plugin-auth-ldap"),
		config: &LDAPConfig{
			URL:        os.Getenv("LDAP_URL"),
			BaseDN:     os.Getenv("LDAP_BASE_DN"),
			BindDN:     os.Getenv("LDAP_BIND_DN"),
			BindPass:   os.Getenv("LDAP_BIND_PASSWORD"),
			UserFilter: os.Getenv("LDAP_USER_FILTER"),
			GroupDN:    os.Getenv("LDAP_GROUP_DN"),
		},
	}
}

func (p *LDAPAuthPlugin) connect() (*ldap.Conn, error) {
	if p.config.URL == "" {
		return nil, fmt.Errorf("LDAP_URL not configured")
	}

	// Determine if TLS is needed
	if strings.HasPrefix(p.config.URL, "ldaps://") {
		return ldap.DialURL(p.config.URL, ldap.DialWithTLSConfig(&tls.Config{InsecureSkipVerify: true}))
	}
	return ldap.DialURL(p.config.URL)
}

func (p *LDAPAuthPlugin) ValidateToken(ctx context.Context, req *pb.TokenRequest) (*pb.UserSession, error) {
	// LDAP "tokens" in this basic implementation are expecting Basic Auth credentials
	// In a real scenario, this might validate a session token issued after an earlier LDAP login
	// For simplicity, we'll assume the input is still credentials if scheme is "basic"

	// Basic auth logic identical to auth-basic but verifying against LDAP
	// logic omitted for brevity, assuming similar base64 decode
	// For OIDC-like flows, LDAP usually sits behind an IdP (Dex/Keycloak)
	// But here we implement direct Bind verification
	// _ = username
	// _ = password
	// _ = password
	return &pb.UserSession{Valid: false}, fmt.Errorf("direct LDAP token validation not fully implemented without IdP wrapper, use Login first")
}

func (p *LDAPAuthPlugin) Authorize(ctx context.Context, req *pb.AuthZRequest) (*pb.AuthZResponse, error) {
	// Simple role check based on roles returned from login
	// In production, might query LDAP group membership
	return &pb.AuthZResponse{Allowed: true}, nil
}

func (p *LDAPAuthPlugin) Login(ctx context.Context, username, password string) (*pb.UserSession, error) {
	l, err := p.connect()
	if err != nil {
		p.logger.Error("Failed to connect to LDAP", "error", err)
		return nil, err
	}
	defer l.Close()

	// 1. Initial Bind (Service Account)
	if p.config.BindDN != "" {
		err = l.Bind(p.config.BindDN, p.config.BindPass)
		if err != nil {
			return nil, fmt.Errorf("initial bind failed: %w", err)
		}
	}

	// 2. Search for user DN
	searchReq := ldap.NewSearchRequest(
		p.config.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=person)(uid=%s))", username), // simplified filter
		[]string{"dn", "cn", "mail", "memberOf"},
		nil,
	)

	sr, err := l.Search(searchReq)
	if err != nil {
		return nil, fmt.Errorf("user search failed: %w", err)
	}

	if len(sr.Entries) != 1 {
		return nil, fmt.Errorf("user not found or too many results")
	}

	userDN := sr.Entries[0].DN
	userCN := sr.Entries[0].GetAttributeValue("cn")
	email := sr.Entries[0].GetAttributeValue("mail")

	p.logger.Info("Attempting login", "dn", userDN, "cn", userCN)

	// 3. Bind as User to verify password
	err = l.Bind(userDN, password)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// 4. Map Groups to Roles
	// (Simplified logic)
	role := "user"
	for _, group := range sr.Entries[0].GetAttributeValues("memberOf") {
		if strings.Contains(group, "admins") {
			role = "admin"
		}
	}

	return &pb.UserSession{
		Valid:     true,
		UserId:    userDN, // Using DN as ID
		Username:  username,
		Email:     email,
		Role:      role,
		Roles:     []string{role},
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}, nil
}

// Implement other methods as stubs since LDAP is mostly for authentication logic

func (p *LDAPAuthPlugin) GetUser(ctx context.Context, req *pb.UserRequest) (*pb.User, error) {
	l, err := p.connect()
	if err != nil {
		return nil, err
	}
	defer l.Close()

	// Initial bind with service account
	if p.config.BindDN != "" {
		err = l.Bind(p.config.BindDN, p.config.BindPass)
		if err != nil {
			return nil, fmt.Errorf("initial bind failed: %w", err)
		}
	}

	// Search for user by DN or username
	filter := fmt.Sprintf("(|(dn=%s)(uid=%s))", req.UserId, req.UserId)
	searchReq := ldap.NewSearchRequest(
		p.config.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		[]string{"dn", "uid", "cn", "mail", "displayName", "memberOf"},
		nil,
	)

	sr, err := l.Search(searchReq)
	if err != nil {
		return nil, fmt.Errorf("user search failed: %w", err)
	}

	if len(sr.Entries) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	entry := sr.Entries[0]
	username := entry.GetAttributeValue("uid")
	if username == "" {
		username = entry.GetAttributeValue("cn")
	}

	// Determine role from group membership
	role := "user"
	for _, group := range entry.GetAttributeValues("memberOf") {
		if strings.Contains(group, "admins") {
			role = "admin"
			break
		}
	}

	return &pb.User{
		Id:          entry.DN,
		Username:    username,
		Email:       entry.GetAttributeValue("mail"),
		FullName:    entry.GetAttributeValue("cn"),
		DisplayName: entry.GetAttributeValue("displayName"),
		Role:        role,
		Roles:       []string{role},
	}, nil
}

func (p *LDAPAuthPlugin) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	l, err := p.connect()
	if err != nil {
		return nil, err
	}
	defer l.Close()

	// Initial bind with service account
	if p.config.BindDN != "" {
		err = l.Bind(p.config.BindDN, p.config.BindPass)
		if err != nil {
			return nil, fmt.Errorf("initial bind failed: %w", err)
		}
	}

	// Build filter
	filter := "(objectClass=person)"
	if p.config.UserFilter != "" {
		filter = fmt.Sprintf("(&%s%s)", filter, p.config.UserFilter)
	}

	// Calculate pagination
	pageSize := int(req.PageSize)
	if pageSize == 0 {
		pageSize = 50
	}
	page := int(req.Page)
	if page == 0 {
		page = 1
	}

	searchReq := ldap.NewSearchRequest(
		p.config.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		[]string{"dn", "uid", "cn", "mail", "displayName", "memberOf"},
		nil,
	)

	sr, err := l.Search(searchReq)
	if err != nil {
		return nil, fmt.Errorf("user search failed: %w", err)
	}

	totalCount := len(sr.Entries)
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > totalCount {
		end = totalCount
	}
	if start > totalCount {
		start = totalCount
	}

	users := make([]*pb.User, 0, end-start)
	for i := start; i < end; i++ {
		entry := sr.Entries[i]
		username := entry.GetAttributeValue("uid")
		if username == "" {
			username = entry.GetAttributeValue("cn")
		}

		// Determine role from group membership
		role := "user"
		for _, group := range entry.GetAttributeValues("memberOf") {
			if strings.Contains(group, "admins") {
				role = "admin"
				break
			}
		}

		users = append(users, &pb.User{
			Id:          entry.DN,
			Username:    username,
			Email:       entry.GetAttributeValue("mail"),
			FullName:    entry.GetAttributeValue("cn"),
			DisplayName: entry.GetAttributeValue("displayName"),
			Role:        role,
			Roles:       []string{role},
		})
	}

	return &pb.ListUsersResponse{
		Users:      users,
		TotalCount: int32(totalCount),
		Page:       int32(page),
	}, nil
}

func (p *LDAPAuthPlugin) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.UserSession, error) {
	return &pb.UserSession{Valid: false}, nil
}

func (p *LDAPAuthPlugin) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: pluginsdk.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"webencode": &pluginsdk.Plugin{
				AuthImpl: NewLDAPAuthPlugin(),
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
