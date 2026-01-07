package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/arauth-identity/iam/auth/mfa"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMFAService is a mock implementation of mfa.ServiceInterface
type MockMFAService struct {
	mock.Mock
}

func (m *MockMFAService) Enroll(ctx context.Context, req *mfa.EnrollRequest) (*mfa.EnrollResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mfa.EnrollResponse), args.Error(1)
}

func (m *MockMFAService) Verify(ctx context.Context, req *mfa.VerifyRequest) (bool, error) {
	args := m.Called(ctx, req)
	return args.Bool(0), args.Error(1)
}

func (m *MockMFAService) Disable(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) error {
	args := m.Called(ctx, userID, tenantID)
	return args.Error(0)
}

func (m *MockMFAService) GenerateRecoveryCodes(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) ([]string, error) {
	args := m.Called(ctx, userID, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockMFAService) VerifyRecoveryCode(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID, code string) error {
	args := m.Called(ctx, userID, tenantID, code)
	return args.Error(0)
}

func (m *MockMFAService) CreateChallenge(ctx context.Context, req *mfa.ChallengeRequest) (*mfa.ChallengeResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mfa.ChallengeResponse), args.Error(1)
}

func (m *MockMFAService) VerifyChallenge(ctx context.Context, req *mfa.VerifyChallengeRequest) (*mfa.VerifyChallengeResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mfa.VerifyChallengeResponse), args.Error(1)
}

func TestMFAHandler_Enroll(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockMFAService)
	// Note: Audit logger is not easily mockable, but tests can still verify handler behavior
	handler := NewMFAHandler(mockService, nil)

	router := gin.New()
	router.POST("/api/v1/mfa/enroll", handler.Enroll)

	userID := uuid.New()
	reqBody := mfa.EnrollRequest{
		UserID: userID,
	}
	body, _ := json.Marshal(reqBody)

	expectedResponse := &mfa.EnrollResponse{
		Secret:        "test-secret",
		QRCode:        "data:image/png;base64,test",
		RecoveryCodes: []string{"code1", "code2"},
	}

	mockService.On("Enroll", mock.Anything, mock.AnythingOfType("*mfa.EnrollRequest")).Return(expectedResponse, nil)

	req, _ := http.NewRequest("POST", "/api/v1/mfa/enroll", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestMFAHandler_Challenge(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockMFAService)
	handler := NewMFAHandler(mockService, nil)

	router := gin.New()
	router.POST("/api/v1/mfa/challenge", handler.Challenge)

	userID := uuid.New()
	tenantID := uuid.New()
	reqBody := mfa.ChallengeRequest{
		UserID:   userID,
		TenantID: tenantID,
	}
	body, _ := json.Marshal(reqBody)

	expectedResponse := &mfa.ChallengeResponse{
		SessionID: "test-session-id",
	}

	mockService.On("CreateChallenge", mock.Anything, mock.AnythingOfType("*mfa.ChallengeRequest")).Return(expectedResponse, nil)

	req, _ := http.NewRequest("POST", "/api/v1/mfa/challenge", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestMFAHandler_Enroll_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockMFAService)
	handler := NewMFAHandler(mockService, nil)

	router := gin.New()
	router.POST("/api/v1/mfa/enroll", handler.Enroll)

	// Invalid JSON
	req, _ := http.NewRequest("POST", "/api/v1/mfa/enroll", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

