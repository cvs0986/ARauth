// +build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/auth/claims"
	"github.com/nuage-identity/iam/auth/hydra"
	"github.com/nuage-identity/iam/auth/login"
	"github.com/nuage-identity/iam/auth/mfa"
	"github.com/nuage-identity/iam/identity/credential"
	"github.com/nuage-identity/iam/identity/models"
	"github.com/nuage-identity/iam/internal/cache"
	"github.com/nuage-identity/iam/internal/testutil"
	"github.com/nuage-identity/iam/security/encryption"
	"github.com/nuage-identity/iam/security/password"
	"github.com/nuage-identity/iam/security/totp"
	"github.com/nuage-identity/iam/storage/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_MFAFlow(t *testing.T) {
	// Setup test database
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)
	defer testutil.CleanupTestDB(t, db)

	var cacheClient *cache.Cache

	// Setup repositories
	userRepo := postgres.NewUserRepository(db)
	credentialRepo := postgres.NewCredentialRepository(db)
	tenantRepo := postgres.NewTenantRepository(db)
	mfaRepo := postgres.NewMFARecoveryCodeRepository(db)

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

	// Setup services
	hydraClient := hydra.NewClient("http://localhost:4445")
	roleRepo := postgres.NewRoleRepository(db)
	permissionRepo := postgres.NewPermissionRepository(db)
	claimsBuilder := claims.NewBuilder(roleRepo, permissionRepo)
	loginService := login.NewService(userRepo, credentialRepo, hydraClient, claimsBuilder)
	
	// Setup MFA service
	encryptionKey := make([]byte, 32) // 32 bytes for AES-256
	copy(encryptionKey, "test-encryption-key-32-bytes!!")
	encryptor, _ := encryption.NewEncryptor(encryptionKey)
	totpGenerator := totp.NewGenerator("Test")
	mfaService := mfa.NewService(userRepo, credentialRepo, mfaRepo, totpGenerator, encryptor, nil)

	// Setup test server
	server, _ := testutil.SetupTestServerWithAuth(db, cacheClient, loginService, mfaService)
	defer server.Close()

	// Step 1: Enroll MFA
	enrollReq := map[string]interface{}{
		"user_id": user.ID.String(),
	}
	body, _ := json.Marshal(enrollReq)

	req, _ := http.NewRequest("POST", server.URL+"/api/v1/mfa/enroll", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenant.ID.String())

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var enrollResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&enrollResp)
	require.NoError(t, err)

	// Verify response contains secret and QR code
	assert.NotEmpty(t, enrollResp["secret"])
	assert.NotEmpty(t, enrollResp["qr_code"])
	
	// Note: Full MFA verification would require TOTP code generation
	// which is tested in integration tests. This E2E test verifies
	// the enrollment flow works end-to-end through the API.
}
