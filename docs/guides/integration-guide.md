# Client Integration Guide

This guide helps client applications integrate with ARauth Identity.

## 🎯 Integration Options

### Option 1: Direct Login (Simplified)

**Use Case**: Simple username/password login.

**Flow**:
1. Client collects username/password
2. Client calls `/auth/login`
3. Client receives tokens
4. Client uses access token for API calls

**Example**:

```javascript
// Login
const response = await fetch('https://iam.example.com/api/v1/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    username: 'user@example.com',
    password: 'password123',
    tenant_id: 'tenant-123',
    client_id: 'client-123'
  })
});

const { access_token, refresh_token, id_token } = await response.json();

// Use token
const apiResponse = await fetch('https://api.example.com/users', {
  headers: {
    'Authorization': `Bearer ${access_token}`
  }
});
```

### Option 2: OAuth2 Authorization Code Flow

**Use Case**: Full OAuth2 compliance.

**Flow**:
1. Client initiates OAuth2 flow
2. User authenticates
3. Client receives authorization code
4. Client exchanges code for tokens

**Example**:

```javascript
// 1. Initiate OAuth2 flow
const codeVerifier = generateCodeVerifier();
const codeChallenge = await generateCodeChallenge(codeVerifier);

const authURL = new URL('https://hydra.example.com/oauth2/auth');
authURL.searchParams.set('client_id', 'client-123');
authURL.searchParams.set('redirect_uri', 'https://app.example.com/callback');
authURL.searchParams.set('response_type', 'code');
authURL.searchParams.set('code_challenge', codeChallenge);
authURL.searchParams.set('code_challenge_method', 'S256');

window.location.href = authURL.toString();

// 2. Handle callback
const code = new URLSearchParams(window.location.search).get('code');

// 3. Exchange code for tokens
const tokenResponse = await fetch('https://hydra.example.com/oauth2/token', {
  method: 'POST',
  headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
  body: new URLSearchParams({
    grant_type: 'authorization_code',
    code: code,
    redirect_uri: 'https://app.example.com/callback',
    client_id: 'client-123',
    code_verifier: codeVerifier
  })
});

const { access_token, refresh_token } = await tokenResponse.json();
```

## 🔄 Token Refresh

```javascript
const refreshResponse = await fetch('https://iam.example.com/api/v1/auth/refresh', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    refresh_token: refreshToken,
    client_id: 'client-123'
  })
});

const { access_token, refresh_token: newRefreshToken } = await refreshResponse.json();
```

## 🔐 JWT Validation

### Client-Side (Basic)

```javascript
// Decode JWT (don't verify signature on client)
function decodeJWT(token) {
  const base64Url = token.split('.')[1];
  const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
  const jsonPayload = decodeURIComponent(atob(base64).split('').map(c => {
    return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
  }).join(''));
  return JSON.parse(jsonPayload);
}

const claims = decodeJWT(accessToken);
console.log(claims.sub, claims.roles, claims.permissions);
```

### Server-Side (Recommended)

```go
import "github.com/golang-jwt/jwt/v5"

// Get JWKS from IAM
jwksURL := "https://iam.example.com/.well-known/jwks.json"

// Validate token
token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
    // Verify signing method
    if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
        return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
    }
    
    // Get public key from JWKS
    // ... fetch and parse JWKS
    
    return publicKey, nil
})

if err != nil {
    return err
}

claims := token.Claims.(jwt.MapClaims)
userID := claims["sub"].(string)
roles := claims["roles"].([]interface{})
```

## 📚 Related Documentation

- [API Design](../technical/api-design.md) - API specifications
- [Architecture Overview](../architecture/overview.md) - System architecture

