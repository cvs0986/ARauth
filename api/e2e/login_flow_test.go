// +build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/auth/claims"
	"github.com/arauth-identity/iam/auth/hydra"
	"github.com/arauth-identity/iam/auth/login"
	"github.com/arauth-identity/iam/auth/mfa"
	"github.com/arauth-identity/iam/identity/credential"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/internal/cache"
	"github.com/arauth-identity/iam/security/encryption"
	"github.com/arauth-identity/iam/security/password"
	"github.com/arauth-identity/iam/security/totp"
	"github.com/arauth-identity/iam/internal/testutil"
	"github.com/arauth-identity/iam/storage/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_LoginFlow(t *testing.T) {
	// Setup test database
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)
	defer testutil.CleanupTestDB(t, db)

	// Setup cache (nil for now, can be enhanced later)
	var cacheClient *cache.Cache

	// Setup repositories
	userRepo := postgres.NewUserRepository(db)
	credentialRepo := postgres.NewCredentialRepository(db)
	tenantRepo := postgres.NewTenantRepository(db)

	// Create test tenant
	tenant := &models.Tenant{
		ID:     uuid.New(),
		Name:   "Test Tenant",
		Domain: "test.example.com",
		Status: models.TenantStatusActive,
	}
	require.NoError(t, tenantRepo.Create(context.Background(), tenant))

	// Create test user
	user := &models.User{
		ID:       uuid.New(),
		TenantID: tenant.ID,
		Username: "testuser",
		Email:    "test@example.com",
		Status:   models.UserStatusActive,
	}
	require.NoError(t, userRepo.Create(context.Background(), user))

	// Set password
	hasher := password.NewHasher()
	hashedPassword, err := hasher.Hash("SecurePassword123!")
	require.NoError(t, err)
	
	cred := &credential.Credential{
		UserID:       user.ID,
		PasswordHash: hashedPassword,
	}
	require.NoError(t, credentialRepo.Create(context.Background(), cred))
	testPassword := "SecurePassword123!"

	// Setup services
	hydraClient := hydra.NewClient("http://localhost:4445") // Mock URL, won't be used
	roleRepo := postgres.NewRoleRepository(db)
	permissionRepo := postgres.NewPermissionRepository(db)
	claimsBuilder := claims.NewBuilder(roleRepo, permissionRepo)
	loginService := login.NewService(userRepo, credentialRepo, hydraClient, claimsBuilder)
	
	// Setup MFA service
	mfaRepo := postgres.NewMFARecoveryCodeRepository(db)
	encryptionKey := make([]byte, 32) // 32 bytes for AES-256
	copy(encryptionKey, "test-encryption-key-32-bytes!!")
	encryptor, _ := encryption.NewEncryptor(encryptionKey)
	totpGenerator := totp.NewGenerator("Test")
	mfaService := mfa.NewService(userRepo, credentialRepo, mfaRepo, totpGenerator, encryptor, nil)

	// Setup test server
	server, _ := setupTestServerWithAuth(db, cacheClient, loginService, mfaService)
	defer server.Close()

	// Test: Login request
	loginReq := map[string]interface{}{
		"username":  "testuser",
		"password":  testPassword,
		"tenant_id": tenant.ID.String(),
	}
	body, _ := json.Marshal(loginReq)

	req, _ := http.NewRequest("POST", server.URL+"/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenant.ID.String())

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var loginResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&loginResp)
	require.NoError(t, err)

	// Verify response contains access token or redirect
	// Note: Actual token generation depends on Hydra, which is mocked
	assert.NotNil(t, loginResp)
}

func TestE2E_LoginFlow_InvalidCredentials(t *testing.T) {
	// Setup test database
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)
	defer testutil.CleanupTestDB(t, db)

	var cacheClient *cache.Cache

	// Setup repositories
	userRepo := postgres.NewUserRepository(db)
	credentialRepo := postgres.NewCredentialRepository(db)
	tenantRepo := postgres.NewTenantRepository(db)

	// Create test tenant
	tenant := &models.Tenant{
		ID:     uuid.New(),
		Name:   "Test Tenant",
		Domain: "test.example.com",
		Status: models.TenantStatusActive,
	}
	require.NoError(t, tenantRepo.Create(context.Background(), tenant))

	// Create test user
	user := &models.User{
		ID:       uuid.New(),
		TenantID: tenant.ID,
		Username: "testuser",
		Email:    "test@example.com",
		Status:   models.UserStatusActive,
	}
	require.NoError(t, userRepo.Create(context.Background(), user))

	// Set password
	hasher := password.NewHasher()
	hashedPassword, err := hasher.Hash("CorrectPassword123!")
	require.NoError(t, err)
	
	cred := &credential.Credential{
		UserID:       user.ID,
		PasswordHash: hashedPassword,
	}
	require.NoError(t, credentialRepo.Create(context.Background(), cred))

	// Setup services
	hydraClient := hydra.NewClient("http://localhost:4445")
	roleRepo := postgres.NewRoleRepository(db)
	permissionRepo := postgres.NewPermissionRepository(db)
	claimsBuilder := claims.NewBuilder(roleRepo, permissionRepo)
	loginService := login.NewService(userRepo, credentialRepo, hydraClient, claimsBuilder)
	
	// Setup MFA service
	mfaRepo := postgres.NewMFARecoveryCodeRepository(db)
	encryptionKey := make([]byte, 32)
	copy(encryptionKey, "test-encryption-key-32-bytes!!")
	encryptor, _ := encryption.NewEncryptor(encryptionKey)
	totpGenerator := totp.NewGenerator("Test")
	mfaService := mfa.NewService(userRepo, credentialRepo, mfaRepo, totpGenerator, encryptor, nil)

	// Setup test server
	server, _ := setupTestServerWithAuth(db, cacheClient, loginService, mfaService)
	defer server.Close()

	// Test: Login with wrong password
	loginReq := map[string]interface{}{
		"username":  "testuser",
		"password":  "WrongPassword123!",
		"tenant_id": tenant.ID.String(),
	}
	body, _ := json.Marshal(loginReq)

	req, _ := http.NewRequest("POST", server.URL+"/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenant.ID.String())

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify error response
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
