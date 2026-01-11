package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/arauth-identity/iam/auth/claims"
	"github.com/arauth-identity/iam/identity/ratelimit"
	"github.com/arauth-identity/iam/observability/security_events"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockSecurityEventRepo implements security_events.Repository for testing
type mockSecurityEventRepo struct {
	events []*security_events.SecurityEvent
}

func (m *mockSecurityEventRepo) Create(ctx context.Context, event *security_events.SecurityEvent) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockSecurityEventRepo) CreateBatch(ctx context.Context, events []*security_events.SecurityEvent) error {
	m.events = append(m.events, events...)
	return nil
}

func (m *mockSecurityEventRepo) Find(ctx context.Context, filters security_events.EventFilters) ([]*security_events.SecurityEvent, error) {
	return m.events, nil
}

func (m *mockSecurityEventRepo) Count(ctx context.Context, filters security_events.EventFilters) (int, error) {
	return len(m.events), nil
}

func (m *mockSecurityEventRepo) DeleteOlderThan(ctx context.Context, olderThan time.Time) (int, error) {
	return 0, nil
}

// TestJWTAuthMiddleware_LogsTokenValidationFailure tests that failed token validations are logged
func TestJWTAuthMiddleware_LogsTokenValidationFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockSecurityEventRepo{events: make([]*security_events.SecurityEvent, 0)}
	mockLogger := security_events.NewAsyncLogger(mockRepo, zap.NewNop(), 10, 100*time.Millisecond)
	defer mockLogger.Close()

	mockTokenService := new(MockTokenService)
	mockTokenService.On("ValidateAccessToken", "invalid-token").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer invalid-token")

	middleware := JWTAuthMiddleware(mockTokenService, mockLogger)
	middleware(c)

	// Wait for async logging
	time.Sleep(200 * time.Millisecond)

	// Verify event was logged
	events, err := mockRepo.Find(context.Background(), security_events.EventFilters{})
	require.NoError(t, err)
	require.Len(t, events, 1)

	event := events[0]
	assert.Equal(t, security_events.EventTokenValidationFailed, event.EventType)
	assert.Equal(t, security_events.SeverityWarning, event.Severity)
	assert.Equal(t, "/test", event.Resource)
	assert.Equal(t, "GET", event.Action)
	assert.Equal(t, "failure", event.Result)
	assert.NotEmpty(t, event.IP)
}

// TestJWTAuthMiddleware_LogsBlacklistedTokenUsage tests that blacklisted token usage is logged as CRITICAL
func TestJWTAuthMiddleware_LogsBlacklistedTokenUsage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockSecurityEventRepo{events: make([]*security_events.SecurityEvent, 0)}
	mockLogger := security_events.NewAsyncLogger(mockRepo, zap.NewNop(), 10, 100*time.Millisecond)
	defer mockLogger.Close()

	mockTokenService := new(MockTokenService)
	testClaims := &claims.Claims{
		Subject:  "user-123",
		TenantID: uuid.New().String(),
		ID:       "jti-blacklisted",
	}
	mockTokenService.On("ValidateAccessToken", "blacklisted-token").Return(testClaims, nil)
	mockTokenService.On("IsAccessTokenRevoked", mock.Anything, "jti-blacklisted").Return(true, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/users", nil)
	c.Request.Header.Set("Authorization", "Bearer blacklisted-token")

	middleware := JWTAuthMiddleware(mockTokenService, mockLogger)
	middleware(c)

	// Wait for async logging
	time.Sleep(200 * time.Millisecond)

	// Verify event was logged
	events, err := mockRepo.Find(context.Background(), security_events.EventFilters{})
	require.NoError(t, err)
	require.Len(t, events, 1)

	event := events[0]
	assert.Equal(t, security_events.EventBlacklistedTokenUsed, event.EventType)
	assert.Equal(t, security_events.SeverityCritical, event.Severity)
	assert.Equal(t, "/api/v1/users", event.Resource)
	assert.Equal(t, "GET", event.Action)
	assert.Equal(t, "blocked", event.Result)
	assert.NotNil(t, event.UserID)
	assert.NotNil(t, event.TenantID)
	assert.NotEmpty(t, event.IP)
	assert.Equal(t, "jti-blacklisted", event.Details["token_id"])
}

// TestRequirePermission_LogsPermissionDenial tests that permission denials are logged
func TestRequirePermission_LogsPermissionDenial(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockSecurityEventRepo{events: make([]*security_events.SecurityEvent, 0)}
	mockLogger := security_events.NewAsyncLogger(mockRepo, zap.NewNop(), 10, 100*time.Millisecond)
	defer mockLogger.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("DELETE", "/api/v1/users/123", nil)
	c.Set("user_id", uuid.New())
	c.Set("tenant_id", uuid.New())
	c.Set("user_permissions", []string{"users:read", "users:update"}) // Missing users:delete

	middleware := RequirePermission("users", "delete", mockLogger)
	middleware(c)

	// Wait for async logging
	time.Sleep(200 * time.Millisecond)

	// Verify event was logged
	events, err := mockRepo.Find(context.Background(), security_events.EventFilters{})
	require.NoError(t, err)
	require.Len(t, events, 1)

	event := events[0]
	assert.Equal(t, security_events.EventPermissionDenied, event.EventType)
	assert.Equal(t, security_events.SeverityWarning, event.Severity)
	assert.Equal(t, "users", event.Resource)
	assert.Equal(t, "delete", event.Action)
	assert.Equal(t, "denied", event.Result)
	assert.NotNil(t, event.UserID)
	assert.NotNil(t, event.TenantID)
	assert.NotEmpty(t, event.IP)
	assert.Equal(t, "users:delete", event.Details["required_permission"])
}

// TestMultiTierRateLimit_LogsViolationWarning tests that general rate limit violations are logged as WARNING
func TestMultiTierRateLimit_LogsViolationWarning(t *testing.T) {
	mockRepo := &mockSecurityEventRepo{events: make([]*security_events.SecurityEvent, 0)}
	mockLogger := security_events.NewAsyncLogger(mockRepo, zap.NewNop(), 10, 100*time.Millisecond)
	defer mockLogger.Close()

	// Setup rate limiter with very low limits
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	config := &ratelimit.Config{
		UserRequestsPerMinute:    1,
		UserBurstSize:            0,
		ClientRequestsPerMinute:  1,
		ClientBurstSize:          0,
		AdminIPRequestsPerMinute: 1,
		AdminIPBurstSize:         0,
		WindowDuration:           time.Minute,
	}

	limiter := ratelimit.NewRedisLimiter(client, config)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(MultiTierRateLimit(limiter, mockLogger))
	router.GET("/api/v1/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// First request succeeds
	req1 := httptest.NewRequest("GET", "/api/v1/users", nil)
	req1.RemoteAddr = "192.168.1.100:12345"
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request should be rate limited
	req2 := httptest.NewRequest("GET", "/api/v1/users", nil)
	req2.RemoteAddr = "192.168.1.100:12345"
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)

	// Wait for async logging
	time.Sleep(200 * time.Millisecond)

	// Verify event was logged
	events, err := mockRepo.Find(context.Background(), security_events.EventFilters{})
	require.NoError(t, err)
	require.Len(t, events, 1)

	event := events[0]
	assert.Equal(t, security_events.EventRateLimitExceeded, event.EventType)
	assert.Equal(t, security_events.SeverityWarning, event.Severity) // General endpoint = WARNING
	assert.Equal(t, "/api/v1/users", event.Resource)
	assert.Equal(t, "GET", event.Action)
	assert.Equal(t, "blocked", event.Result)
	assert.NotEmpty(t, event.IP)
	assert.Equal(t, "general", event.Details["category"])
}

// TestMultiTierRateLimit_LogsViolationCritical tests that sensitive endpoint violations are logged as CRITICAL
func TestMultiTierRateLimit_LogsViolationCritical(t *testing.T) {
	mockRepo := &mockSecurityEventRepo{events: make([]*security_events.SecurityEvent, 0)}
	mockLogger := security_events.NewAsyncLogger(mockRepo, zap.NewNop(), 10, 100*time.Millisecond)
	defer mockLogger.Close()

	// Setup rate limiter with very low limits
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	config := &ratelimit.Config{
		UserRequestsPerMinute:      1,
		UserBurstSize:              0,
		SensitiveRequestsPerMinute: 1,
		SensitiveBurstSize:         0,
		WindowDuration:             time.Minute,
	}

	limiter := ratelimit.NewRedisLimiter(client, config)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(MultiTierRateLimit(limiter, mockLogger))
	router.POST("/api/v1/auth/mfa/enroll", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// First request succeeds
	req1 := httptest.NewRequest("POST", "/api/v1/auth/mfa/enroll", nil)
	req1.RemoteAddr = "192.168.1.100:12345"
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request should be rate limited
	req2 := httptest.NewRequest("POST", "/api/v1/auth/mfa/enroll", nil)
	req2.RemoteAddr = "192.168.1.100:12345"
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)

	// Wait for async logging
	time.Sleep(200 * time.Millisecond)

	// Verify event was logged
	events, err := mockRepo.Find(context.Background(), security_events.EventFilters{})
	require.NoError(t, err)
	require.Len(t, events, 1)

	event := events[0]
	assert.Equal(t, security_events.EventRateLimitExceeded, event.EventType)
	assert.Equal(t, security_events.SeverityCritical, event.Severity) // Sensitive endpoint = CRITICAL
	assert.Equal(t, "/api/v1/auth/mfa/enroll", event.Resource)
	assert.Equal(t, "POST", event.Action)
	assert.Equal(t, "blocked", event.Result)
	assert.NotEmpty(t, event.IP)
	assert.Equal(t, "sensitive", event.Details["category"])
}
