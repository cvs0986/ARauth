package oidc

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/arauth-identity/iam/identity/federation"
)

// Client handles OIDC authentication flows
type Client struct {
	config *federation.OIDCConfiguration
	httpClient *http.Client
}

// NewClient creates a new OIDC client
func NewClient(config *federation.OIDCConfiguration) *Client {
	return &Client{
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// DiscoveryResponse represents the OIDC discovery document
type DiscoveryResponse struct {
	Issuer                string   `json:"issuer"`
	AuthorizationEndpoint string   `json:"authorization_endpoint"`
	TokenEndpoint         string   `json:"token_endpoint"`
	UserInfoEndpoint      string   `json:"userinfo_endpoint"`
	JWKSURI               string   `json:"jwks_uri"`
	ResponseTypesSupported []string `json:"response_types_supported"`
	SubjectTypesSupported  []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`
}

// Discover performs OIDC provider discovery
func (c *Client) Discover(ctx context.Context) (*DiscoveryResponse, error) {
	discoveryURL := c.config.IssuerURL
	if discoveryURL[len(discoveryURL)-1] != '/' {
		discoveryURL += "/"
	}
	discoveryURL += ".well-known/openid-configuration"

	req, err := http.NewRequestWithContext(ctx, "GET", discoveryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform discovery: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discovery failed with status: %d", resp.StatusCode)
	}

	var discovery DiscoveryResponse
	if err := json.NewDecoder(resp.Body).Decode(&discovery); err != nil {
		return nil, fmt.Errorf("failed to decode discovery response: %w", err)
	}

	// Update config with discovered endpoints if not already set
	if c.config.AuthURL == "" {
		c.config.AuthURL = discovery.AuthorizationEndpoint
	}
	if c.config.TokenURL == "" {
		c.config.TokenURL = discovery.TokenEndpoint
	}
	if c.config.UserInfoURL == "" {
		c.config.UserInfoURL = discovery.UserInfoEndpoint
	}

	return &discovery, nil
}

// GenerateState generates a random state parameter for OAuth flow
func GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GenerateAuthorizationURL generates the authorization URL for OIDC login
func (c *Client) GenerateAuthorizationURL(redirectURI, state string) (string, error) {
	authURL, err := url.Parse(c.config.AuthURL)
	if err != nil {
		return "", fmt.Errorf("invalid authorization URL: %w", err)
	}

	params := url.Values{}
	params.Set("client_id", c.config.ClientID)
	params.Set("response_type", "code")
	params.Set("scope", "openid profile email")
	if len(c.config.Scopes) > 0 {
		params.Set("scope", "openid profile email "+joinScopes(c.config.Scopes))
	}
	params.Set("redirect_uri", redirectURI)
	params.Set("state", state)
	params.Set("nonce", generateNonce())

	authURL.RawQuery = params.Encode()
	return authURL.String(), nil
}

// generateNonce generates a random nonce for OIDC
func generateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// joinScopes joins scope strings
func joinScopes(scopes []string) string {
	result := ""
	for i, scope := range scopes {
		if i > 0 {
			result += " "
		}
		result += scope
	}
	return result
}

// TokenResponse represents the OAuth token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token"`
	Scope        string `json:"scope"`
}

// ExchangeCode exchanges an authorization code for tokens
func (c *Client) ExchangeCode(ctx context.Context, code, redirectURI string) (*TokenResponse, error) {
	tokenURL, err := url.Parse(c.config.TokenURL)
	if err != nil {
		return nil, fmt.Errorf("invalid token URL: %w", err)
	}

	params := url.Values{}
	params.Set("grant_type", "authorization_code")
	params.Set("code", code)
	params.Set("redirect_uri", redirectURI)
	params.Set("client_id", c.config.ClientID)
	params.Set("client_secret", c.config.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.URL.RawQuery = params.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed with status: %d", resp.StatusCode)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenResp, nil
}

// IDTokenClaims represents the claims in an ID token
type IDTokenClaims struct {
	Iss           string `json:"iss"`
	Sub           string `json:"sub"`
	Aud           string `json:"aud"`
	Exp           int64  `json:"exp"`
	Iat           int64  `json:"iat"`
	Nonce         string `json:"nonce,omitempty"`
	Email         string `json:"email,omitempty"`
	EmailVerified bool   `json:"email_verified,omitempty"`
	Name          string `json:"name,omitempty"`
	GivenName     string `json:"given_name,omitempty"`
	FamilyName    string `json:"family_name,omitempty"`
	Picture       string `json:"picture,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
}

// ValidateIDToken validates an ID token (basic validation - full JWT validation would require JWKS)
func (c *Client) ValidateIDToken(ctx context.Context, idToken string) (*IDTokenClaims, error) {
	// Basic validation: decode and check structure
	// In production, you should:
	// 1. Verify JWT signature using JWKS
	// 2. Verify issuer matches
	// 3. Verify audience matches client_id
	// 4. Verify expiration
	// 5. Verify nonce (if provided)

	// For now, we'll do a simple decode
	// Split JWT into parts
	parts := splitJWT(idToken)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid ID token format")
	}

	// Decode payload (second part)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode ID token payload: %w", err)
	}

	var claims IDTokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ID token claims: %w", err)
	}

	// Basic validation
	if claims.Iss != c.config.IssuerURL {
		return nil, fmt.Errorf("invalid issuer: expected %s, got %s", c.config.IssuerURL, claims.Iss)
	}

	if claims.Aud != c.config.ClientID {
		return nil, fmt.Errorf("invalid audience: expected %s, got %s", c.config.ClientID, claims.Aud)
	}

	// Check expiration
	if claims.Exp > 0 {
		expTime := time.Unix(claims.Exp, 0)
		if time.Now().After(expTime) {
			return nil, fmt.Errorf("ID token has expired")
		}
	}

	return &claims, nil
}

// splitJWT splits a JWT into its parts
func splitJWT(token string) []string {
	var parts []string
	start := 0
	for i, char := range token {
		if char == '.' {
			parts = append(parts, token[start:i])
			start = i + 1
		}
	}
	if start < len(token) {
		parts = append(parts, token[start:])
	}
	return parts
}

// UserInfo represents user information from the UserInfo endpoint
type UserInfo struct {
	Sub                string `json:"sub"`
	Name               string `json:"name,omitempty"`
	GivenName          string `json:"given_name,omitempty"`
	FamilyName         string `json:"family_name,omitempty"`
	Email              string `json:"email,omitempty"`
	EmailVerified      bool   `json:"email_verified,omitempty"`
	Picture            string `json:"picture,omitempty"`
	PreferredUsername  string `json:"preferred_username,omitempty"`
}

// GetUserInfo retrieves user information from the UserInfo endpoint
func (c *Client) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	if c.config.UserInfoURL == "" {
		return nil, fmt.Errorf("userinfo endpoint not configured")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.config.UserInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create userinfo request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo request failed with status: %d", resp.StatusCode)
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode userinfo response: %w", err)
	}

	return &userInfo, nil
}

