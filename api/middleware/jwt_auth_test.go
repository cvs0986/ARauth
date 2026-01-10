package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arauth-identity/iam/auth/claims"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTokenService for middleware tests
type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) ValidateAccessToken(tokenString string) (*claims.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*claims.Claims), args.Error(1)
}

func (m *MockTokenService) IsAccessTokenRevoked(ctx context.Context, jti string) (bool, error) {
	args := m.Called(ctx, jti)
	return args.Bool(0), args.Error(1)
}

// Stubs for interface satisfaction
func (m *MockTokenService) GenerateAccessToken(claimsObj *claims.Claims, expiresIn time.Duration) (string, error) {
	return "", nil
}
func (m *MockTokenService) GenerateRefreshToken() (string, error)         { return "", nil }
func (m *MockTokenService) HashRefreshToken(token string) (string, error) { return "", nil }
func (m *MockTokenService) VerifyRefreshToken(token, hash string) bool    { return true }
func (m *MockTokenService) RevokeAccessToken(ctx context.Context, tokenString string) error {
	return nil
}
func (m *MockTokenService) GetPublicKey() interface{} { return nil }

func TestJWTAuthMiddleware_Revocation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		token          string
		claims         *claims.Claims
		isRevoked      bool
		revokeErr      error
		expectedStatus int
	}{
		{
			name:  "Valid Token Not Revoked",
			token: "valid-token",
			claims: &claims.Claims{
				Subject: "user-1",
				ID:      "jti-1",
			},
			isRevoked:      false,
			revokeErr:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:  "Valid Token Revoked",
			token: "revoked-token",
			claims: &claims.Claims{
				Subject: "user-2",
				ID:      "jti-2",
			},
			isRevoked:      true,
			revokeErr:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:  "Redis Failure (Fail Closed)",
			token: "error-token",
			claims: &claims.Claims{
				Subject: "user-3",
				ID:      "jti-3",
			},
			isRevoked:      false,
			revokeErr:      errors.New("redis error"),
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockTokenService)

			// Setup expectations
			mockService.On("ValidateAccessToken", tt.token).Return(tt.claims, nil)
			if tt.claims != nil && tt.claims.ID != "" {
				mockService.On("IsAccessTokenRevoked", mock.Anything, tt.claims.ID).Return(tt.isRevoked, tt.revokeErr)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("GET", "/", nil)
			c.Request.Header.Set("Authorization", "Bearer "+tt.token)

			middleware := JWTAuthMiddleware(mockService)
			middleware(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
