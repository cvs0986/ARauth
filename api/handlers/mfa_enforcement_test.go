package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arauth-identity/iam/auth/claims"
	"github.com/arauth-identity/iam/auth/login"
	"github.com/arauth-identity/iam/auth/token"
	"github.com/arauth-identity/iam/config"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/storage/interfaces"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test 1: Verify Login returns NO tokens when MFA is required
func TestMFA_Enforcement_Login_NoTokens(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockLoginService := new(MockLoginService)
	// mockMFAService is defined in mfa_handler_test.go in the same package
	mockMFAService := new(MockMFAService)
	mockAuditService := new(MockAuditService)

	// Create handler with mocks
	handler := NewAuthHandler(mockLoginService, nil, nil, mockAuditService, mockMFAService)

	router := gin.New()
	router.POST("/auth/login", handler.Login)

	// Setup: User requires MFA
	loginResp := &login.LoginResponse{
		MFARequired: true,
		UserID:      uuid.New().String(),
		TenantID:    uuid.New().String(),
	}

	reqBody := login.LoginRequest{Username: "user", Password: "password"}
	mockLoginService.On("Login", mock.Anything, &reqBody).Return(loginResp, nil)

	// Expect MFA Session Creation
	sessionID := "mfa-session-123"
	mockMFAService.On("CreateSession", mock.Anything, mock.Anything, mock.Anything).Return(sessionID, nil)

	// Expect Audit Log
	mockAuditService.On("LogMFAChallengeCreated", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Execute
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	// CRITICAL: Check MFA flag and Session ID
	assert.Equal(t, true, resp["mfa_required"])
	assert.Equal(t, sessionID, resp["mfa_session_id"])

	// CRITICAL: Ensure NO TOKENS are issued
	assert.Empty(t, resp["access_token"], "Access Token must be empty when MFA is required")
	assert.Empty(t, resp["refresh_token"], "Refresh Token must be empty when MFA is required")
	assert.Empty(t, resp["id_token"], "ID Token must be empty when MFA is required")
}

// Test 2: Verify Refresh Flow Enforces MFA (Gap Fixed)
func TestMFA_Enforcement_Refresh_Enforced(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Refresh_Blocked_If_Not_Verified", func(t *testing.T) {
		// Dependencies
		mockTokenService := new(MockTokenService)
		mockUserRepo := new(MockUserRepo)
		mockRefreshTokenRepo := new(MockRefreshTokenRepository)
		mockAuditService := new(MockAuditService)
		mockMFAService := new(MockMFAService)
		mockCapabilityService := new(MockCapabilityService)

		// Mocks for ClaimsBuilder
		mockRoleRepo := new(MockRoleRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockSysRoleRepo := new(MockSystemRoleRepository)

		// Mocks for LifetimeResolver
		mockTenantSettingsRepo := new(MockTenantSettingsRepository)

		// Real Services constructed with mocks
		claimsBuilder := claims.NewBuilder(mockRoleRepo, mockPermRepo, mockSysRoleRepo, mockCapabilityService, nil)

		// Correct Config structure
		secConfig := &config.SecurityConfig{
			JWT: config.JWTConfig{
				AccessTokenTTL:  15 * time.Minute,
				RefreshTokenTTL: 7 * 24 * time.Hour,
			},
		}
		lifetimeResolver := token.NewLifetimeResolver(secConfig, mockTenantSettingsRepo)

		refreshService := token.NewRefreshService(
			mockTokenService,
			mockRefreshTokenRepo,
			mockUserRepo,
			claimsBuilder,
			lifetimeResolver,
		)

		// Create handler with mocks
		handler := NewAuthHandler(nil, refreshService, mockTokenService, mockAuditService, mockMFAService)

		router := gin.New()
		router.POST("/auth/refresh", handler.RefreshToken)

		// Setup: User with MFA enabled
		userID := uuid.New()
		tenantID := uuid.New()
		refreshTokenHash := "hashed_refresh_token"
		rawRefreshToken := "raw_refresh_token"

		user := &models.User{
			ID:            userID,
			TenantID:      &tenantID,
			Username:      "mfa_user",
			MFAEnabled:    true, // User has MFA enabled
			PrincipalType: "USER",
			Status:        "active",
		}

		// Setup Token Service mocks
		mockTokenService.On("HashRefreshToken", rawRefreshToken).Return(refreshTokenHash, nil)
		// NOTE: GenerateAccessToken should NOT be called if logic works
		// But in case of bug, it might be.

		// Setup Refresh Repo
		rt := &interfaces.RefreshToken{
			TokenHash:   refreshTokenHash,
			UserID:      userID,
			TenantID:    tenantID,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			CreatedAt:   time.Now(),
			MFAVerified: false, // NOT VERIFIED
		}
		mockRefreshTokenRepo.On("GetByTokenHash", mock.Anything, refreshTokenHash).Return(rt, nil)
		// Revoke might still happen if we consider it "invalid usage" but currently logic returns error before revoke?
		// Logic: Hash -> Get -> Check Exp -> Get User -> Check Active -> MFA Check.
		// So checking user happens.

		// Setup User Repo
		mockUserRepo.On("GetByID", mock.Anything, userID).Return(user, nil)

		// Setup Capability Service Mocks (Prevent Panic)
		mockCapabilityService.On("GetEnabledFeaturesForTenant", mock.Anything, tenantID).Return(map[string]bool{}, nil)
		mockCapabilityService.On("GetAllowedCapabilitiesForTenant", mock.Anything, tenantID).Return(map[string]bool{}, nil)

		// Execute refresh request
		reqBody := map[string]string{"refresh_token": rawRefreshToken}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// ASSERTION: Should be Unauthorized
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)

		assert.Equal(t, "token_refresh_failed", resp["error"])
		assert.Contains(t, resp["message"], "MFA required")
	})

	t.Run("Refresh_Allowed_If_Verified", func(t *testing.T) {
		// Dependencies
		mockTokenService := new(MockTokenService)
		mockUserRepo := new(MockUserRepo)
		mockRefreshTokenRepo := new(MockRefreshTokenRepository)
		mockAuditService := new(MockAuditService)
		mockMFAService := new(MockMFAService)
		mockCapabilityService := new(MockCapabilityService)

		// Mocks for ClaimsBuilder
		mockRoleRepo := new(MockRoleRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockSysRoleRepo := new(MockSystemRoleRepository)
		mockTenantSettingsRepo := new(MockTenantSettingsRepository)

		// Real Services
		claimsBuilder := claims.NewBuilder(mockRoleRepo, mockPermRepo, mockSysRoleRepo, mockCapabilityService, nil)
		secConfig := &config.SecurityConfig{JWT: config.JWTConfig{AccessTokenTTL: 15 * time.Minute, RefreshTokenTTL: 7 * 24 * time.Hour}}
		lifetimeResolver := token.NewLifetimeResolver(secConfig, mockTenantSettingsRepo)
		refreshService := token.NewRefreshService(mockTokenService, mockRefreshTokenRepo, mockUserRepo, claimsBuilder, lifetimeResolver)
		handler := NewAuthHandler(nil, refreshService, mockTokenService, mockAuditService, mockMFAService)

		router := gin.New()
		router.POST("/auth/refresh", handler.RefreshToken)

		// Setup
		userID := uuid.New()
		tenantID := uuid.New()
		refreshTokenHash := "hash_verified"
		rawRefreshToken := "raw_verified"

		user := &models.User{
			ID:            userID,
			TenantID:      &tenantID,
			Username:      "mfa_verified_user",
			MFAEnabled:    true,
			PrincipalType: "USER",
			Status:        "active",
		}

		// Token Service Mocks
		mockTokenService.On("HashRefreshToken", rawRefreshToken).Return(refreshTokenHash, nil)
		mockTokenService.On("GenerateAccessToken", mock.Anything, mock.Anything).Return("new_access_token", nil)
		mockTokenService.On("GenerateRefreshToken").Return("new_refresh_token_rotated", nil)
		mockTokenService.On("HashRefreshToken", "new_refresh_token_rotated").Return("new_hash_rotated", nil)
		// ValidateAccessToken called in handler for audit logging
		mockTokenService.On("ValidateAccessToken", "new_access_token").Return(&claims.Claims{
			Subject:  userID.String(),
			Username: "mfa_verified_user",
		}, nil)

		// Refresh Repo Mocks
		oldToken := &interfaces.RefreshToken{
			TokenHash:   refreshTokenHash,
			UserID:      userID,
			TenantID:    tenantID,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			MFAVerified: true, // VERIFIED!
		}
		mockRefreshTokenRepo.On("GetByTokenHash", mock.Anything, refreshTokenHash).Return(oldToken, nil)

		// Expect Revoke of old token
		mockRefreshTokenRepo.On("RevokeByTokenHash", mock.Anything, refreshTokenHash).Return(nil)

		// Expect Create of new token - MUST PRESERVE MFAVerified=true
		mockRefreshTokenRepo.On("Create", mock.Anything, mock.MatchedBy(func(rt *interfaces.RefreshToken) bool {
			return rt.MFAVerified == true
		})).Return(nil)

		// User Repo Mocks
		mockUserRepo.On("GetByID", mock.Anything, userID).Return(user, nil)

		// Capability Service Mocks
		mockCapabilityService.On("GetEnabledFeaturesForTenant", mock.Anything, tenantID).Return(map[string]bool{}, nil)
		mockCapabilityService.On("GetAllowedCapabilitiesForTenant", mock.Anything, tenantID).Return(map[string]bool{}, nil)

		// Audit Mocks
		mockAuditService.On("LogTokenIssued", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		// Execute
		reqBody := map[string]string{"refresh_token": rawRefreshToken}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Verify
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "new_access_token", resp["access_token"])
	})
}

// Additional Mocks needed

type MockRoleRepository struct{ mock.Mock }

func (m *MockRoleRepository) Create(ctx context.Context, role *models.Role) error { return nil }
func (m *MockRoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	return nil, nil
}
func (m *MockRoleRepository) GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Role, error) {
	return nil, nil
}
func (m *MockRoleRepository) Update(ctx context.Context, role *models.Role) error { return nil }
func (m *MockRoleRepository) Delete(ctx context.Context, id uuid.UUID) error      { return nil }
func (m *MockRoleRepository) List(ctx context.Context, tenantID uuid.UUID, filters *interfaces.RoleFilters) ([]*models.Role, error) {
	return nil, nil
}
func (m *MockRoleRepository) Count(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *MockRoleRepository) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	return nil
}
func (m *MockRoleRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	return nil
}
func (m *MockRoleRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*models.Role, error) {
	return []*models.Role{}, nil
}
func (m *MockRoleRepository) GetSystemRoles(ctx context.Context) ([]*models.Role, error) {
	return nil, nil
}

type MockPermissionRepository struct{ mock.Mock }

func (m *MockPermissionRepository) Create(ctx context.Context, perm *models.Permission) error {
	return nil
}
func (m *MockPermissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Permission, error) {
	return nil, nil
}
func (m *MockPermissionRepository) GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Permission, error) {
	return nil, nil
}

// Fix List signature
func (m *MockPermissionRepository) List(ctx context.Context, tenantID uuid.UUID, filters *interfaces.PermissionFilters) ([]*models.Permission, error) {
	return nil, nil
}
func (m *MockPermissionRepository) Update(ctx context.Context, perm *models.Permission) error {
	return nil
}
func (m *MockPermissionRepository) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *MockPermissionRepository) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*models.Permission, error) {
	return []*models.Permission{}, nil
}
func (m *MockPermissionRepository) Count(ctx context.Context) (int64, error) { return 0, nil }
func (m *MockPermissionRepository) AssignPermissionToRole(ctx context.Context, permissionID, roleID uuid.UUID) error {
	return nil
}
func (m *MockPermissionRepository) RemovePermissionFromRole(ctx context.Context, permissionID, roleID uuid.UUID) error {
	return nil
}

type MockSystemRoleRepository struct{ mock.Mock }

// Fixed GetByID signature
func (m *MockSystemRoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*interfaces.SystemRole, error) {
	return nil, nil
}

// Fixed GetByName signature
func (m *MockSystemRoleRepository) GetByName(ctx context.Context, name string) (*interfaces.SystemRole, error) {
	return nil, nil
}

// Fixed GetUserSystemRoles signature
func (m *MockSystemRoleRepository) GetUserSystemRoles(ctx context.Context, userID uuid.UUID) ([]*interfaces.SystemRole, error) {
	return []*interfaces.SystemRole{}, nil
}
func (m *MockSystemRoleRepository) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID, tenantID *uuid.UUID) error {
	return nil
}
func (m *MockSystemRoleRepository) RemoveSystemRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	return nil
}

// Fixed GetRolePermissions signature
func (m *MockSystemRoleRepository) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*interfaces.SystemPermission, error) {
	return nil, nil
}

// Added List method
func (m *MockSystemRoleRepository) List(ctx context.Context) ([]*interfaces.SystemRole, error) {
	return nil, nil
}

// Added RemoveRoleFromUser (missing)
func (m *MockSystemRoleRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	return nil
}

type MockTenantSettingsRepository struct{ mock.Mock }

func (m *MockTenantSettingsRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*interfaces.TenantSettings, error) {
	return &interfaces.TenantSettings{
		AccessTokenTTLMinutes: 15,
		RefreshTokenTTLDays:   7,
	}, nil
}
func (m *MockTenantSettingsRepository) Update(ctx context.Context, settings *interfaces.TenantSettings) error {
	return nil
}
func (m *MockTenantSettingsRepository) Create(ctx context.Context, settings *interfaces.TenantSettings) error {
	return nil
}
func (m *MockTenantSettingsRepository) Delete(ctx context.Context, tenantID uuid.UUID) error {
	return nil
}

type MockUserRepo struct{ mock.Mock }

func (m *MockUserRepo) Create(ctx context.Context, user *models.User) error { return nil }
func (m *MockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.User), args.Error(1)
}

// Fixed GetByUsername signature
func (m *MockUserRepo) GetByUsername(ctx context.Context, username string, tenantID uuid.UUID) (*models.User, error) {
	return nil, nil
}

// Fixed GetByEmail signature
func (m *MockUserRepo) GetByEmail(ctx context.Context, email string, tenantID uuid.UUID) (*models.User, error) {
	return nil, nil
}

// Added GetByEmailSystem
func (m *MockUserRepo) GetByEmailSystem(ctx context.Context, email string) (*models.User, error) {
	return nil, nil
}

// Added GetSystemUserByUsername (missing method)
func (m *MockUserRepo) GetSystemUserByUsername(ctx context.Context, username string) (*models.User, error) {
	return nil, nil
}

func (m *MockUserRepo) Update(ctx context.Context, user *models.User) error { return nil }
func (m *MockUserRepo) Delete(ctx context.Context, id uuid.UUID) error      { return nil }

// Corrected List signature
func (m *MockUserRepo) List(ctx context.Context, tenantID uuid.UUID, filters *interfaces.UserFilters) ([]*models.User, error) {
	return nil, nil
}

// Corrected ListSystem signature (removed extra args)
func (m *MockUserRepo) ListSystem(ctx context.Context, filters *interfaces.UserFilters) ([]*models.User, error) {
	return nil, nil
}

func (m *MockUserRepo) Count(ctx context.Context, tenantID uuid.UUID, filters *interfaces.UserFilters) (int, error) {
	return 0, nil
}

// Added CountSystem
func (m *MockUserRepo) CountSystem(ctx context.Context, filters *interfaces.UserFilters) (int, error) {
	return 0, nil
}

type MockRefreshTokenRepository struct{ mock.Mock }

func (m *MockRefreshTokenRepository) Create(ctx context.Context, token *interfaces.RefreshToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}
func (m *MockRefreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*interfaces.RefreshToken, error) {
	args := m.Called(ctx, tokenHash)
	return args.Get(0).(*interfaces.RefreshToken), args.Error(1)
}
func (m *MockRefreshTokenRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*interfaces.RefreshToken, error) {
	return nil, nil
}
func (m *MockRefreshTokenRepository) Revoke(ctx context.Context, tokenID uuid.UUID) error { return nil }

func (m *MockRefreshTokenRepository) RevokeByTokenHash(ctx context.Context, tokenHash string) error {
	args := m.Called(ctx, tokenHash)
	return args.Error(0)
}
func (m *MockRefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	return nil
}
func (m *MockRefreshTokenRepository) DeleteExpired(ctx context.Context) error { return nil }
func (m *MockRefreshTokenRepository) RevokeByClientID(ctx context.Context, clientID string) (int, error) {
	args := m.Called(ctx, clientID)
	return args.Int(0), args.Error(1)
}
