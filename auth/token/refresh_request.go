package token

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshTokenResponse represents a token refresh response
type RefreshTokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"` // New refresh token (rotated)
	TokenType        string `json:"token_type"`
	ExpiresIn        int    `json:"expires_in"`         // Access token expiry in seconds
	RefreshExpiresIn int    `json:"refresh_expires_in"` // Refresh token expiry in seconds
}

