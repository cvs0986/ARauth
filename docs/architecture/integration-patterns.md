# Hydra Integration Patterns

This document describes how Nuage Identity integrates with ORY Hydra.

## 🎯 Integration Philosophy

**Key Principle**: Hydra is NEVER exposed directly to clients. All Hydra interactions go through the IAM API.

```
Client Apps → IAM API → Hydra Admin API
```

## 🔌 Integration Architecture

### Hydra as OAuth2/OIDC Provider

Hydra handles:
- OAuth2 flows (Authorization Code, Client Credentials)
- Token issuance and validation
- Token refresh
- Client management
- Consent management (if needed)

IAM API handles:
- User authentication
- Credential validation
- MFA
- Claims building
- Business logic

## 📋 Integration Patterns

### Pattern 1: Direct Login (Simplified)

**Use Case**: Simple username/password login without full OAuth2 flow.

**Flow**:

```go
// 1. Client calls IAM API
POST /auth/login
{
  "username": "user@example.com",
  "password": "password",
  "tenant_id": "tenant-123"
}

// 2. IAM validates credentials
user, err := identityService.GetUserByUsername(username, tenantID)
if err != nil {
    return error
}

// 3. IAM builds claims
claims := policyService.BuildClaims(user, tenant, roles, permissions)

// 4. IAM creates OAuth2 client (if not exists)
client, err := hydraClient.GetOAuth2Client(clientID)
if err == NotFound {
    client = &OAuth2Client{
        ClientID:     clientID,
        ClientSecret: generateSecret(),
        GrantTypes:   []string{"authorization_code", "refresh_token"},
        RedirectURIs: []string{redirectURI},
    }
    hydraClient.CreateOAuth2Client(client)
}

// 5. IAM calls Hydra to issue tokens directly
tokenResponse, err := hydraClient.IssueToken(&TokenRequest{
    GrantType:    "authorization_code",
    ClientID:     clientID,
    ClientSecret: clientSecret,
    Subject:      user.ID,
    Claims:       claims,
})

// 6. Return tokens to client
return tokenResponse
```

**Implementation**:

```go
type HydraClient interface {
    IssueToken(req *TokenRequest) (*TokenResponse, error)
    CreateOAuth2Client(client *OAuth2Client) error
    GetOAuth2Client(clientID string) (*OAuth2Client, error)
}
```

### Pattern 2: OAuth2 Authorization Code Flow

**Use Case**: Full OAuth2 compliance with PKCE.

**Flow**:

```go
// 1. Client initiates OAuth2 flow
GET /oauth2/auth?client_id=...&redirect_uri=...&code_challenge=...

// 2. Hydra calls IAM login challenge callback
POST /oauth2/auth/requests/login
{
  "login_challenge": "challenge-123"
}

// 3. IAM returns login_challenge to client
// Client shows login UI

// 4. Client sends credentials
POST /auth/login
{
  "login_challenge": "challenge-123",
  "username": "user@example.com",
  "password": "password"
}

// 5. IAM validates and accepts login
loginRequest, err := hydraClient.GetLoginRequest(loginChallenge)
if err != nil {
    return error
}

// Validate credentials
user, err := identityService.GetUserByUsername(username, tenantID)

// Build claims
claims := policyService.BuildClaims(user, tenant, roles, permissions)

// Accept login
acceptResponse, err := hydraClient.AcceptLoginRequest(loginChallenge, &AcceptLoginRequest{
    Subject: user.ID,
    Context: claims,
})

// 6. Redirect client to Hydra with authorization code
redirect(acceptResponse.RedirectTo)

// 7. Client exchanges code for tokens
POST /oauth2/token
{
  "grant_type": "authorization_code",
  "code": "auth-code-123",
  "code_verifier": "...",
  "client_id": "...",
  "redirect_uri": "..."
}

// 8. Hydra returns tokens
```

**Implementation**:

```go
type HydraClient interface {
    GetLoginRequest(challenge string) (*LoginRequest, error)
    AcceptLoginRequest(challenge string, req *AcceptLoginRequest) (*AcceptLoginResponse, error)
    GetConsentRequest(challenge string) (*ConsentRequest, error)
    AcceptConsentRequest(challenge string, req *AcceptConsentRequest) (*AcceptConsentResponse, error)
}
```

### Pattern 3: Client Credentials Flow

**Use Case**: Service-to-service authentication.

**Flow**:

```go
// 1. Service requests token
POST /oauth2/token
{
  "grant_type": "client_credentials",
  "client_id": "service-client",
  "client_secret": "secret"
}

// 2. IAM validates client credentials
// (Can be stored in IAM DB or Hydra)

// 3. IAM calls Hydra to issue token
tokenResponse, err := hydraClient.IssueToken(&TokenRequest{
    GrantType:    "client_credentials",
    ClientID:     clientID,
    ClientSecret: clientSecret,
    Scope:        "service:read service:write",
})

// 4. Return token
return tokenResponse
```

### Pattern 4: Token Refresh

**Flow**:

```go
// 1. Client requests refresh
POST /auth/refresh
{
  "refresh_token": "refresh-token-123"
}

// 2. IAM validates refresh token
// Check blacklist in Redis

// 3. IAM calls Hydra to refresh
tokenResponse, err := hydraClient.RefreshToken(&RefreshTokenRequest{
    RefreshToken: refreshToken,
    ClientID:     clientID,
    ClientSecret: clientSecret,
})

// 4. Hydra rotates refresh token (if configured)
// Returns new access_token and refresh_token

// 5. Return new tokens
return tokenResponse
```

## 🔧 Hydra Client Implementation

### HTTP Client

```go
type HydraAdminClient struct {
    baseURL    string
    httpClient *http.Client
}

func (c *HydraAdminClient) AcceptLoginRequest(
    ctx context.Context,
    challenge string,
    req *AcceptLoginRequest,
) (*AcceptLoginResponse, error) {
    url := fmt.Sprintf("%s/admin/oauth2/auth/requests/login/accept?login_challenge=%s",
        c.baseURL, challenge)
    
    resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(reqJSON))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var acceptResp AcceptLoginResponse
    if err := json.NewDecoder(resp.Body).Decode(&acceptResp); err != nil {
        return nil, err
    }
    
    return &acceptResp, nil
}
```

### Error Handling

```go
type HydraError struct {
    ErrorCode    string `json:"error"`
    ErrorDescription string `json:"error_description"`
    StatusCode   int
}

func (e *HydraError) Error() string {
    return fmt.Sprintf("hydra error: %s - %s", e.ErrorCode, e.ErrorDescription)
}
```

## 🎨 Claims Injection

### Custom Claims in Tokens

```go
func BuildClaims(user *User, tenant *Tenant, roles []*Role, permissions []string) map[string]interface{} {
    claims := map[string]interface{}{
        "sub":        user.ID,
        "tenant":     tenant.ID,
        "roles":      extractRoleNames(roles),
        "permissions": permissions,
        "acr":        "mfa", // Authentication Context Class Reference
        "email":      user.Email,
        "username":   user.Username,
    }
    
    // Add custom claims
    if tenant.CustomClaims != nil {
        for k, v := range tenant.CustomClaims {
            claims[k] = v
        }
    }
    
    return claims
}
```

### Claims in AcceptLoginRequest

```go
acceptReq := &AcceptLoginRequest{
    Subject: user.ID,
    Context: map[string]interface{}{
        "claims": claims,
    },
    Remember: true,
    RememberFor: 3600,
}
```

## 🔐 Security Considerations

### 1. Admin API Security

- **Authentication**: Use API key or mTLS
- **Network**: Internal network only, never exposed to internet
- **Rate Limiting**: Limit admin API calls

### 2. Token Security

- **Short-lived access tokens**: 15 minutes
- **Refresh token rotation**: Enabled
- **JWT signing**: RS256 with key rotation

### 3. Client Management

- **Client secrets**: Stored securely, hashed
- **Redirect URIs**: Whitelist validation
- **PKCE**: Required for public clients

## 📊 Monitoring Integration

### Metrics to Track

- Hydra API call latency
- Token issuance rate
- Login acceptance rate
- Error rates

### Logging

```go
logger.Info("hydra_login_accepted",
    "login_challenge", challenge,
    "subject", user.ID,
    "tenant", tenant.ID,
)
```

## 🧪 Testing Integration

### Mock Hydra Client

```go
type MockHydraClient struct {
    AcceptLoginRequestFunc func(challenge string, req *AcceptLoginRequest) (*AcceptLoginResponse, error)
}

func (m *MockHydraClient) AcceptLoginRequest(
    challenge string,
    req *AcceptLoginRequest,
) (*AcceptLoginResponse, error) {
    if m.AcceptLoginRequestFunc != nil {
        return m.AcceptLoginRequestFunc(challenge, req)
    }
    return &AcceptLoginResponse{
        RedirectTo: "http://example.com/callback?code=mock-code",
    }, nil
}
```

### Integration Tests

```go
func TestHydraIntegration(t *testing.T) {
    // Start test Hydra instance
    hydra := startTestHydra(t)
    defer hydra.Stop()
    
    // Test login flow
    client := NewHydraClient(hydra.AdminURL)
    
    challenge := "test-challenge"
    acceptReq := &AcceptLoginRequest{
        Subject: "user-123",
        Context: map[string]interface{}{},
    }
    
    resp, err := client.AcceptLoginRequest(challenge, acceptReq)
    assert.NoError(t, err)
    assert.NotEmpty(t, resp.RedirectTo)
}
```

## 🔄 Hydra Configuration

### Required Hydra Settings

```yaml
# hydra.yml
serve:
  admin:
    port: 4445
  public:
    port: 4444

urls:
  self:
    issuer: https://hydra.example.com
  admin: http://localhost:4445
  public: http://localhost:4444

oauth2:
  expose_internal_errors: false
  hashers:
    algorithm: bcrypt
  grant_allowed_grant_types:
    - authorization_code
    - refresh_token
    - client_credentials
  refresh_token_hook: ""
  
strategies:
  access_token: jwt
  scope: exact

ttl:
  access_token: 15m
  refresh_token: 30d
  id_token: 1h
```

## 📚 Related Documentation

- [Architecture Overview](./overview.md)
- [Data Flow](./data-flow.md)
- [Components](./components.md)

