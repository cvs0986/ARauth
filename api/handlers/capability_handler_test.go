package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arauth-identity/iam/identity/capability"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCapabilityService is a mock implementation of capability.ServiceInterface
type MockCapabilityService struct {
	mock.Mock
}

func (m *MockCapabilityService) IsCapabilitySupported(ctx context.Context, capabilityKey string) (bool, error) {
	args := m.Called(ctx, capabilityKey)
	return args.Bool(0), args.Error(1)
}

func (m *MockCapabilityService) GetSystemCapability(ctx context.Context, capabilityKey string) (*models.SystemCapability, error) {
	args := m.Called(ctx, capabilityKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SystemCapability), args.Error(1)
}

func (m *MockCapabilityService) GetAllSystemCapabilities(ctx context.Context) ([]*models.SystemCapability, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.SystemCapability), args.Error(1)
}

func (m *MockCapabilityService) UpdateSystemCapability(ctx context.Context, capability *models.SystemCapability) error {
	args := m.Called(ctx, capability)
	return args.Error(0)
}

func (m *MockCapabilityService) IsCapabilityAllowedForTenant(ctx context.Context, tenantID uuid.UUID, capabilityKey string) (bool, error) {
	args := m.Called(ctx, tenantID, capabilityKey)
	return args.Bool(0), args.Error(1)
}

func (m *MockCapabilityService) GetAllowedCapabilitiesForTenant(ctx context.Context, tenantID uuid.UUID) (map[string]bool, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).(map[string]bool), args.Error(1)
}

func (m *MockCapabilityService) SetTenantCapability(ctx context.Context, tenantID uuid.UUID, capabilityKey string, enabled bool, value *json.RawMessage, configuredBy uuid.UUID) error {
	args := m.Called(ctx, tenantID, capabilityKey, enabled, value, configuredBy)
	return args.Error(0)
}

func (m *MockCapabilityService) DeleteTenantCapability(ctx context.Context, tenantID uuid.UUID, capabilityKey string) error {
	args := m.Called(ctx, tenantID, capabilityKey)
	return args.Error(0)
}

func (m *MockCapabilityService) IsFeatureEnabledByTenant(ctx context.Context, tenantID uuid.UUID, featureKey string) (bool, error) {
	args := m.Called(ctx, tenantID, featureKey)
	return args.Bool(0), args.Error(1)
}

func (m *MockCapabilityService) GetEnabledFeaturesForTenant(ctx context.Context, tenantID uuid.UUID) (map[string]bool, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).(map[string]bool), args.Error(1)
}

func (m *MockCapabilityService) EnableFeatureForTenant(ctx context.Context, tenantID uuid.UUID, featureKey string, config *json.RawMessage, enabledBy uuid.UUID) error {
	args := m.Called(ctx, tenantID, featureKey, config, enabledBy)
	return args.Error(0)
}

func (m *MockCapabilityService) DisableFeatureForTenant(ctx context.Context, tenantID uuid.UUID, featureKey string) error {
	args := m.Called(ctx, tenantID, featureKey)
	return args.Error(0)
}

func (m *MockCapabilityService) IsUserEnrolled(ctx context.Context, userID uuid.UUID, capabilityKey string) (bool, error) {
	args := m.Called(ctx, userID, capabilityKey)
	return args.Bool(0), args.Error(1)
}

func (m *MockCapabilityService) GetUserCapabilityState(ctx context.Context, userID uuid.UUID, capabilityKey string) (*models.UserCapabilityState, error) {
	args := m.Called(ctx, userID, capabilityKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserCapabilityState), args.Error(1)
}

func (m *MockCapabilityService) GetUserCapabilityStates(ctx context.Context, userID uuid.UUID) ([]*models.UserCapabilityState, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*models.UserCapabilityState), args.Error(1)
}

func (m *MockCapabilityService) EnrollUserInCapability(ctx context.Context, userID uuid.UUID, capabilityKey string, stateData *json.RawMessage) error {
	args := m.Called(ctx, userID, capabilityKey, stateData)
	return args.Error(0)
}

func (m *MockCapabilityService) UnenrollUserFromCapability(ctx context.Context, userID uuid.UUID, capabilityKey string) error {
	args := m.Called(ctx, userID, capabilityKey)
	return args.Error(0)
}

func (m *MockCapabilityService) EvaluateCapability(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, capabilityKey string) (*capability.CapabilityEvaluation, error) {
	args := m.Called(ctx, tenantID, userID, capabilityKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*capability.CapabilityEvaluation), args.Error(1)
}

func (m *MockCapabilityService) GetTenantCapabilities(ctx context.Context, tenantID uuid.UUID) ([]*models.TenantCapability, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TenantCapability), args.Error(1)
}

func (m *MockCapabilityService) GetTenantFeatureEnablements(ctx context.Context, tenantID uuid.UUID) ([]*models.TenantFeatureEnablement, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TenantFeatureEnablement), args.Error(1)
}

func TestCapabilityHandler_ListSystemCapabilities(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupMock  func(*MockCapabilityService)
		wantStatus int
	}{
		{
			name: "success",
			setupMock: func(m *MockCapabilityService) {
				m.On("GetAllSystemCapabilities", mock.Anything).Return([]*models.SystemCapability{
					{
						CapabilityKey: "mfa",
						Enabled:       true,
					},
				}, nil)
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockCapabilityService)
			tt.setupMock(mockService)

			handler := NewCapabilityHandler(mockService)

			router := gin.New()
			router.GET("/system/capabilities", handler.ListSystemCapabilities)

			req, _ := http.NewRequest("GET", "/system/capabilities", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestCapabilityHandler_GetSystemCapability(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		capKey     string
		setupMock  func(*MockCapabilityService)
		wantStatus int
	}{
		{
			name:   "success",
			capKey: "mfa",
			setupMock: func(m *MockCapabilityService) {
				m.On("GetSystemCapability", mock.Anything, "mfa").Return(&models.SystemCapability{
					CapabilityKey: "mfa",
					Enabled:       true,
				}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "not found",
			capKey: "unknown",
			setupMock: func(m *MockCapabilityService) {
				m.On("GetSystemCapability", mock.Anything, "unknown").Return(nil, assert.AnError)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockCapabilityService)
			tt.setupMock(mockService)

			handler := NewCapabilityHandler(mockService)

			router := gin.New()
			router.GET("/system/capabilities/:key", handler.GetSystemCapability)

			req, _ := http.NewRequest("GET", "/system/capabilities/"+tt.capKey, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestCapabilityHandler_UpdateSystemCapability(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		capKey     string
		body       map[string]interface{}
		setupMock  func(*MockCapabilityService)
		wantStatus int
	}{
		{
			name:   "success",
			capKey: "mfa",
			body: map[string]interface{}{
				"name":        "MFA",
				"description": "Multi-factor authentication",
			},
			setupMock: func(m *MockCapabilityService) {
				m.On("GetSystemCapability", mock.Anything, "mfa").Return(&models.SystemCapability{
					CapabilityKey: "mfa",
					Enabled:       true,
				}, nil)
				m.On("UpdateSystemCapability", mock.Anything, mock.AnythingOfType("*models.SystemCapability")).Return(nil)
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockCapabilityService)
			tt.setupMock(mockService)

			handler := NewCapabilityHandler(mockService)

			router := gin.New()
			router.PUT("/system/capabilities/:key", handler.UpdateSystemCapability)

			bodyBytes, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest("PUT", "/system/capabilities/"+tt.capKey, bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}
