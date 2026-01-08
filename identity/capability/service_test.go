package capability

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/arauth-identity/iam/identity/models"
)

// Mock repositories
type MockSystemCapabilityRepository struct {
	mock.Mock
}

func (m *MockSystemCapabilityRepository) GetByKey(ctx context.Context, key string) (*models.SystemCapability, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SystemCapability), args.Error(1)
}

func (m *MockSystemCapabilityRepository) GetAll(ctx context.Context) ([]*models.SystemCapability, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.SystemCapability), args.Error(1)
}

func (m *MockSystemCapabilityRepository) Update(ctx context.Context, capability *models.SystemCapability) error {
	args := m.Called(ctx, capability)
	return args.Error(0)
}

func (m *MockSystemCapabilityRepository) GetEnabled(ctx context.Context) ([]*models.SystemCapability, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.SystemCapability), args.Error(1)
}

func (m *MockSystemCapabilityRepository) Create(ctx context.Context, capability *models.SystemCapability) error {
	args := m.Called(ctx, capability)
	return args.Error(0)
}

func (m *MockSystemCapabilityRepository) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

type MockTenantCapabilityRepository struct {
	mock.Mock
}

func (m *MockTenantCapabilityRepository) GetByTenantIDAndKey(ctx context.Context, tenantID uuid.UUID, key string) (*models.TenantCapability, error) {
	args := m.Called(ctx, tenantID, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TenantCapability), args.Error(1)
}

func (m *MockTenantCapabilityRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*models.TenantCapability, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]*models.TenantCapability), args.Error(1)
}

func (m *MockTenantCapabilityRepository) Create(ctx context.Context, capability *models.TenantCapability) error {
	args := m.Called(ctx, capability)
	return args.Error(0)
}

func (m *MockTenantCapabilityRepository) Update(ctx context.Context, capability *models.TenantCapability) error {
	args := m.Called(ctx, capability)
	return args.Error(0)
}

func (m *MockTenantCapabilityRepository) Delete(ctx context.Context, tenantID uuid.UUID, key string) error {
	args := m.Called(ctx, tenantID, key)
	return args.Error(0)
}

func (m *MockTenantCapabilityRepository) GetEnabledByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*models.TenantCapability, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]*models.TenantCapability), args.Error(1)
}

func (m *MockTenantCapabilityRepository) DeleteByTenantID(ctx context.Context, tenantID uuid.UUID) error {
	args := m.Called(ctx, tenantID)
	return args.Error(0)
}

type MockTenantFeatureEnablementRepository struct {
	mock.Mock
}

func (m *MockTenantFeatureEnablementRepository) GetByTenantIDAndKey(ctx context.Context, tenantID uuid.UUID, key string) (*models.TenantFeatureEnablement, error) {
	args := m.Called(ctx, tenantID, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TenantFeatureEnablement), args.Error(1)
}

func (m *MockTenantFeatureEnablementRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*models.TenantFeatureEnablement, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]*models.TenantFeatureEnablement), args.Error(1)
}

func (m *MockTenantFeatureEnablementRepository) Create(ctx context.Context, enablement *models.TenantFeatureEnablement) error {
	args := m.Called(ctx, enablement)
	return args.Error(0)
}

func (m *MockTenantFeatureEnablementRepository) Update(ctx context.Context, enablement *models.TenantFeatureEnablement) error {
	args := m.Called(ctx, enablement)
	return args.Error(0)
}

func (m *MockTenantFeatureEnablementRepository) Delete(ctx context.Context, tenantID uuid.UUID, key string) error {
	args := m.Called(ctx, tenantID, key)
	return args.Error(0)
}

func (m *MockTenantFeatureEnablementRepository) GetEnabledByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*models.TenantFeatureEnablement, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]*models.TenantFeatureEnablement), args.Error(1)
}

func (m *MockTenantFeatureEnablementRepository) DeleteByTenantID(ctx context.Context, tenantID uuid.UUID) error {
	args := m.Called(ctx, tenantID)
	return args.Error(0)
}

type MockUserCapabilityStateRepository struct {
	mock.Mock
}

func (m *MockUserCapabilityStateRepository) GetByUserIDAndKey(ctx context.Context, userID uuid.UUID, key string) (*models.UserCapabilityState, error) {
	args := m.Called(ctx, userID, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserCapabilityState), args.Error(1)
}

func (m *MockUserCapabilityStateRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.UserCapabilityState, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*models.UserCapabilityState), args.Error(1)
}

func (m *MockUserCapabilityStateRepository) Create(ctx context.Context, state *models.UserCapabilityState) error {
	args := m.Called(ctx, state)
	return args.Error(0)
}

func (m *MockUserCapabilityStateRepository) Update(ctx context.Context, state *models.UserCapabilityState) error {
	args := m.Called(ctx, state)
	return args.Error(0)
}

func (m *MockUserCapabilityStateRepository) Delete(ctx context.Context, userID uuid.UUID, key string) error {
	args := m.Called(ctx, userID, key)
	return args.Error(0)
}

func (m *MockUserCapabilityStateRepository) GetEnrolledByUserID(ctx context.Context, userID uuid.UUID) ([]*models.UserCapabilityState, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*models.UserCapabilityState), args.Error(1)
}

func (m *MockUserCapabilityStateRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func TestService_IsCapabilitySupported(t *testing.T) {
	tests := []struct {
		name          string
		capabilityKey string
		setupMock     func(*MockSystemCapabilityRepository)
		want          bool
		wantErr       bool
	}{
		{
			name:          "capability supported",
			capabilityKey: "mfa",
			setupMock: func(m *MockSystemCapabilityRepository) {
				m.On("GetByKey", mock.Anything, "mfa").Return(&models.SystemCapability{
					CapabilityKey: "mfa",
					Enabled:       true,
				}, nil)
			},
			want:    true,
			wantErr: false,
		},
		{
			name:          "capability not supported",
			capabilityKey: "saml",
			setupMock: func(m *MockSystemCapabilityRepository) {
				m.On("GetByKey", mock.Anything, "saml").Return(&models.SystemCapability{
					CapabilityKey: "saml",
					Enabled:       false,
				}, nil)
			},
			want:    false,
			wantErr: false,
		},
		{
			name:          "capability not found",
			capabilityKey: "unknown",
			setupMock: func(m *MockSystemCapabilityRepository) {
				m.On("GetByKey", mock.Anything, "unknown").Return(nil, assert.AnError)
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSystemRepo := new(MockSystemCapabilityRepository)
			mockTenantRepo := new(MockTenantCapabilityRepository)
			mockFeatureRepo := new(MockTenantFeatureEnablementRepository)
			mockUserRepo := new(MockUserCapabilityStateRepository)

			tt.setupMock(mockSystemRepo)

			service := NewService(mockSystemRepo, mockTenantRepo, mockFeatureRepo, mockUserRepo)

			got, err := service.IsCapabilitySupported(context.Background(), tt.capabilityKey)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mockSystemRepo.AssertExpectations(t)
		})
	}
}

func TestService_EvaluateCapability(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	capabilityKey := "mfa"

	tests := []struct {
		name      string
		setupMocks func(*MockSystemCapabilityRepository, *MockTenantCapabilityRepository, *MockTenantFeatureEnablementRepository, *MockUserCapabilityStateRepository)
		want      *CapabilityEvaluation
		wantErr   bool
	}{
		{
			name: "full evaluation - all layers pass",
			setupMocks: func(sys *MockSystemCapabilityRepository, tenant *MockTenantCapabilityRepository, feature *MockTenantFeatureEnablementRepository, user *MockUserCapabilityStateRepository) {
				// System level: supported
				sys.On("GetByKey", mock.Anything, capabilityKey).Return(&models.SystemCapability{
					CapabilityKey: capabilityKey,
					Enabled:       true,
				}, nil)

				// Tenant level: allowed
				tenant.On("GetByTenantIDAndKey", mock.Anything, tenantID, capabilityKey).Return(&models.TenantCapability{
					TenantID:      tenantID,
					CapabilityKey: capabilityKey,
					Enabled:       true,
				}, nil)

				// Feature level: enabled
				feature.On("GetByTenantIDAndKey", mock.Anything, tenantID, capabilityKey).Return(&models.TenantFeatureEnablement{
					TenantID:   tenantID,
					FeatureKey: capabilityKey,
					Enabled:    true,
				}, nil)

				// User level: enrolled (MFA requires enrollment)
				now := time.Now()
				user.On("GetByUserIDAndKey", mock.Anything, userID, capabilityKey).Return(&models.UserCapabilityState{
					UserID:       userID,
					CapabilityKey: capabilityKey,
					Enrolled:     true,
					EnrolledAt:   &now,
				}, nil)
			},
			want: &CapabilityEvaluation{
				CapabilityKey:    capabilityKey,
				SystemSupported:  true,
				TenantAllowed:    true,
				TenantEnabled:    true,
				UserEnrolled:     true,
				CanUse:           true,
			},
			wantErr: false,
		},
		{
			name: "system not supported",
			setupMocks: func(sys *MockSystemCapabilityRepository, tenant *MockTenantCapabilityRepository, feature *MockTenantFeatureEnablementRepository, user *MockUserCapabilityStateRepository) {
				sys.On("GetByKey", mock.Anything, capabilityKey).Return(&models.SystemCapability{
					CapabilityKey: capabilityKey,
					Enabled:       false,
				}, nil)
			},
			want: &CapabilityEvaluation{
				CapabilityKey:    capabilityKey,
				SystemSupported:  false,
				CanUse:           false,
				Reason:           "capability mfa is not supported by the system",
			},
			wantErr: false,
		},
		{
			name: "tenant not allowed",
			setupMocks: func(sys *MockSystemCapabilityRepository, tenant *MockTenantCapabilityRepository, feature *MockTenantFeatureEnablementRepository, user *MockUserCapabilityStateRepository) {
				sys.On("GetByKey", mock.Anything, capabilityKey).Return(&models.SystemCapability{
					CapabilityKey: capabilityKey,
					Enabled:       true,
				}, nil)
				tenant.On("GetByTenantIDAndKey", mock.Anything, tenantID, capabilityKey).Return(nil, assert.AnError)
			},
			want: &CapabilityEvaluation{
				CapabilityKey:    capabilityKey,
				SystemSupported:  true,
				TenantAllowed:    false,
				CanUse:           false,
				Reason:           "capability mfa is not allowed for this tenant",
			},
			wantErr: false,
		},
		{
			name: "feature not enabled",
			setupMocks: func(sys *MockSystemCapabilityRepository, tenant *MockTenantCapabilityRepository, feature *MockTenantFeatureEnablementRepository, user *MockUserCapabilityStateRepository) {
				sys.On("GetByKey", mock.Anything, capabilityKey).Return(&models.SystemCapability{
					CapabilityKey: capabilityKey,
					Enabled:       true,
				}, nil)
				tenant.On("GetByTenantIDAndKey", mock.Anything, tenantID, capabilityKey).Return(&models.TenantCapability{
					TenantID:      tenantID,
					CapabilityKey: capabilityKey,
					Enabled:       true,
				}, nil)
				feature.On("GetByTenantIDAndKey", mock.Anything, tenantID, capabilityKey).Return(nil, assert.AnError)
			},
			want: &CapabilityEvaluation{
				CapabilityKey:    capabilityKey,
				SystemSupported:  true,
				TenantAllowed:    true,
				TenantEnabled:    false,
				CanUse:           false,
				Reason:           "feature mfa is not enabled by this tenant",
			},
			wantErr: false,
		},
		{
			name: "user not enrolled",
			setupMocks: func(sys *MockSystemCapabilityRepository, tenant *MockTenantCapabilityRepository, feature *MockTenantFeatureEnablementRepository, user *MockUserCapabilityStateRepository) {
				sys.On("GetByKey", mock.Anything, capabilityKey).Return(&models.SystemCapability{
					CapabilityKey: capabilityKey,
					Enabled:       true,
				}, nil)
				tenant.On("GetByTenantIDAndKey", mock.Anything, tenantID, capabilityKey).Return(&models.TenantCapability{
					TenantID:      tenantID,
					CapabilityKey: capabilityKey,
					Enabled:       true,
				}, nil)
				feature.On("GetByTenantIDAndKey", mock.Anything, tenantID, capabilityKey).Return(&models.TenantFeatureEnablement{
					TenantID:   tenantID,
					FeatureKey: capabilityKey,
					Enabled:    true,
				}, nil)
				user.On("GetByUserIDAndKey", mock.Anything, userID, capabilityKey).Return(nil, assert.AnError)
			},
			want: &CapabilityEvaluation{
				CapabilityKey:    capabilityKey,
				SystemSupported:  true,
				TenantAllowed:    true,
				TenantEnabled:    true,
				UserEnrolled:     false,
				CanUse:           false,
				Reason:           "user is not enrolled in capability mfa",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSystemRepo := new(MockSystemCapabilityRepository)
			mockTenantRepo := new(MockTenantCapabilityRepository)
			mockFeatureRepo := new(MockTenantFeatureEnablementRepository)
			mockUserRepo := new(MockUserCapabilityStateRepository)

			tt.setupMocks(mockSystemRepo, mockTenantRepo, mockFeatureRepo, mockUserRepo)

			service := NewService(mockSystemRepo, mockTenantRepo, mockFeatureRepo, mockUserRepo)

			got, err := service.EvaluateCapability(context.Background(), tenantID, userID, capabilityKey)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.CapabilityKey, got.CapabilityKey)
				assert.Equal(t, tt.want.SystemSupported, got.SystemSupported)
				assert.Equal(t, tt.want.TenantAllowed, got.TenantAllowed)
				assert.Equal(t, tt.want.TenantEnabled, got.TenantEnabled)
				assert.Equal(t, tt.want.UserEnrolled, got.UserEnrolled)
				assert.Equal(t, tt.want.CanUse, got.CanUse)
				if tt.want.Reason != "" {
					assert.Contains(t, got.Reason, tt.want.Reason)
				}
			}

			mockSystemRepo.AssertExpectations(t)
			mockTenantRepo.AssertExpectations(t)
			mockFeatureRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestService_IsCapabilityAllowedForTenant(t *testing.T) {
	tenantID := uuid.New()
	capabilityKey := "mfa"

	tests := []struct {
		name      string
		setupMocks func(*MockSystemCapabilityRepository, *MockTenantCapabilityRepository)
		want      bool
		wantErr   bool
	}{
		{
			name: "capability allowed",
			setupMocks: func(sys *MockSystemCapabilityRepository, tenant *MockTenantCapabilityRepository) {
				sys.On("GetByKey", mock.Anything, capabilityKey).Return(&models.SystemCapability{
					CapabilityKey: capabilityKey,
					Enabled:       true,
				}, nil)
				tenant.On("GetByTenantIDAndKey", mock.Anything, tenantID, capabilityKey).Return(&models.TenantCapability{
					TenantID:      tenantID,
					CapabilityKey: capabilityKey,
					Enabled:       true,
				}, nil)
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "system not supported",
			setupMocks: func(sys *MockSystemCapabilityRepository, tenant *MockTenantCapabilityRepository) {
				sys.On("GetByKey", mock.Anything, capabilityKey).Return(&models.SystemCapability{
					CapabilityKey: capabilityKey,
					Enabled:       false,
				}, nil)
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "tenant not assigned",
			setupMocks: func(sys *MockSystemCapabilityRepository, tenant *MockTenantCapabilityRepository) {
				sys.On("GetByKey", mock.Anything, capabilityKey).Return(&models.SystemCapability{
					CapabilityKey: capabilityKey,
					Enabled:       true,
				}, nil)
				tenant.On("GetByTenantIDAndKey", mock.Anything, tenantID, capabilityKey).Return(nil, assert.AnError)
			},
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSystemRepo := new(MockSystemCapabilityRepository)
			mockTenantRepo := new(MockTenantCapabilityRepository)
			mockFeatureRepo := new(MockTenantFeatureEnablementRepository)
			mockUserRepo := new(MockUserCapabilityStateRepository)

			tt.setupMocks(mockSystemRepo, mockTenantRepo)

			service := NewService(mockSystemRepo, mockTenantRepo, mockFeatureRepo, mockUserRepo)

			got, err := service.IsCapabilityAllowedForTenant(context.Background(), tenantID, capabilityKey)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mockSystemRepo.AssertExpectations(t)
			mockTenantRepo.AssertExpectations(t)
		})
	}
}

func TestService_EnableFeatureForTenant(t *testing.T) {
	tenantID := uuid.New()
	enabledBy := uuid.New()
	featureKey := "mfa"
	config := json.RawMessage(`{"required_for_admins": true}`)

	tests := []struct {
		name      string
		setupMocks func(*MockSystemCapabilityRepository, *MockTenantCapabilityRepository, *MockTenantFeatureEnablementRepository)
		wantErr   bool
	}{
		{
			name: "successfully enable feature",
			setupMocks: func(sys *MockSystemCapabilityRepository, tenant *MockTenantCapabilityRepository, feature *MockTenantFeatureEnablementRepository) {
				// Check if allowed
				sys.On("GetByKey", mock.Anything, featureKey).Return(&models.SystemCapability{
					CapabilityKey: featureKey,
					Enabled:       true,
				}, nil)
				tenant.On("GetByTenantIDAndKey", mock.Anything, tenantID, featureKey).Return(&models.TenantCapability{
					TenantID:      tenantID,
					CapabilityKey: featureKey,
					Enabled:       true,
				}, nil)

				// Create feature enablement
				feature.On("GetByTenantIDAndKey", mock.Anything, tenantID, featureKey).Return(nil, assert.AnError)
				feature.On("Create", mock.Anything, mock.MatchedBy(func(e *models.TenantFeatureEnablement) bool {
					return e.TenantID == tenantID && e.FeatureKey == featureKey && e.Enabled == true
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "feature not allowed for tenant",
			setupMocks: func(sys *MockSystemCapabilityRepository, tenant *MockTenantCapabilityRepository, feature *MockTenantFeatureEnablementRepository) {
				sys.On("GetByKey", mock.Anything, featureKey).Return(&models.SystemCapability{
					CapabilityKey: featureKey,
					Enabled:       true,
				}, nil)
				tenant.On("GetByTenantIDAndKey", mock.Anything, tenantID, featureKey).Return(nil, assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSystemRepo := new(MockSystemCapabilityRepository)
			mockTenantRepo := new(MockTenantCapabilityRepository)
			mockFeatureRepo := new(MockTenantFeatureEnablementRepository)
			mockUserRepo := new(MockUserCapabilityStateRepository)

			tt.setupMocks(mockSystemRepo, mockTenantRepo, mockFeatureRepo)

			service := NewService(mockSystemRepo, mockTenantRepo, mockFeatureRepo, mockUserRepo)

			err := service.EnableFeatureForTenant(context.Background(), tenantID, featureKey, &config, enabledBy)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockSystemRepo.AssertExpectations(t)
			mockTenantRepo.AssertExpectations(t)
			mockFeatureRepo.AssertExpectations(t)
		})
	}
}

