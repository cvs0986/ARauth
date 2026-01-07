// +build integration

package mfa

import (
	"context"
	"testing"

	"github.com/google/uuid"
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

func TestService_Enroll_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	// Create repositories
	userRepo := postgres.NewUserRepository(db)
	credentialRepo := postgres.NewCredentialRepository(db)
	mfaRecoveryCodeRepo := postgres.NewMFARecoveryCodeRepository(db)
	tenantRepo := postgres.NewTenantRepository(db)

	// Create tenant
	tenantID := uuid.New()
	tenant := &models.Tenant{
		ID:     tenantID,
		Name:   "Test Tenant",
		Domain: "test-" + tenantID.String() + ".example.com",
		Status: "active",
	}
	err := tenantRepo.Create(context.Background(), tenant)
	require.NoError(t, err)

	// Create user
	userID := uuid.New()
	user := &models.User{
		ID:       userID,
		TenantID: tenantID,
		Username: "mfauser",
		Email:    "mfa@example.com",
		Status:   "active",
	}
	err = userRepo.Create(context.Background(), user)
	require.NoError(t, err)

	// Create password credential
	hasher := password.NewHasher()
	hashedPassword, err := hasher.Hash("TestPassword123!")
	require.NoError(t, err)

	cred := &credential.Credential{
		UserID:       userID,
		PasswordHash: hashedPassword,
	}
	err = credentialRepo.Create(context.Background(), cred)
	require.NoError(t, err)

	// Setup MFA service dependencies
	encryptionKey := make([]byte, 32)
	copy(encryptionKey, "test-encryption-key-32-bytes-long!")
	encryptor, err := encryption.NewEncryptor(encryptionKey)
	require.NoError(t, err)

	totpGenerator := totp.NewGenerator("Test Issuer")
	cacheClient := cache.NewCache(nil) // In-memory cache for tests
	sessionManager := NewSessionManager(cacheClient)

	service := NewService(userRepo, credentialRepo, mfaRecoveryCodeRepo, totpGenerator, encryptor, sessionManager)

	// Test enrollment
	req := &EnrollRequest{
		UserID: userID,
	}

	response, err := service.Enroll(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Secret)
	assert.NotEmpty(t, response.QRCode)
	assert.NotEmpty(t, response.RecoveryCodes)
	assert.Equal(t, 10, len(response.RecoveryCodes))

	// Verify user has MFA enabled
	updatedUser, err := userRepo.GetByID(context.Background(), userID)
	require.NoError(t, err)
	assert.True(t, updatedUser.MFAEnabled)
	assert.NotNil(t, updatedUser.MFASecretEncrypted)
}

func TestService_Verify_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	userRepo := postgres.NewUserRepository(db)
	credentialRepo := postgres.NewCredentialRepository(db)
	mfaRecoveryCodeRepo := postgres.NewMFARecoveryCodeRepository(db)
	tenantRepo := postgres.NewTenantRepository(db)

	tenantID := uuid.New()
	tenant := &models.Tenant{
		ID:     tenantID,
		Name:   "Test Tenant",
		Domain: "test-" + tenantID.String() + ".example.com",
		Status: "active",
	}
	err := tenantRepo.Create(context.Background(), tenant)
	require.NoError(t, err)

	userID := uuid.New()
	user := &models.User{
		ID:       userID,
		TenantID: tenantID,
		Username: "verifyuser",
		Email:    "verify@example.com",
		Status:   "active",
	}
	err = userRepo.Create(context.Background(), user)
	require.NoError(t, err)

	hasher := password.NewHasher()
	hashedPassword, err := hasher.Hash("TestPassword123!")
	require.NoError(t, err)

	cred := &credential.Credential{
		UserID:       userID,
		PasswordHash: hashedPassword,
	}
	err = credentialRepo.Create(context.Background(), cred)
	require.NoError(t, err)

	encryptionKey := make([]byte, 32)
	copy(encryptionKey, "test-encryption-key-32-bytes-long!")
	encryptor, err := encryption.NewEncryptor(encryptionKey)
	require.NoError(t, err)

	totpGenerator := totp.NewGenerator("Test Issuer")
	cacheClient := cache.NewCache(nil)
	sessionManager := NewSessionManager(cacheClient)

	service := NewService(userRepo, credentialRepo, mfaRecoveryCodeRepo, totpGenerator, encryptor, sessionManager)

	// Enroll user first
	enrollReq := &EnrollRequest{
		UserID: userID,
	}
	enrollResp, err := service.Enroll(context.Background(), enrollReq)
	require.NoError(t, err)

	// Test verification with recovery code (more reliable for testing)
	// Get a recovery code from enrollment
	recoveryCode := enrollResp.RecoveryCodes[0]

	verifyReq := &VerifyRequest{
		UserID:      userID,
		RecoveryCode: recoveryCode,
	}

	valid, err := service.Verify(context.Background(), verifyReq)
	require.NoError(t, err)
	assert.True(t, valid)
}

func TestService_CreateChallenge_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	userRepo := postgres.NewUserRepository(db)
	credentialRepo := postgres.NewCredentialRepository(db)
	mfaRecoveryCodeRepo := postgres.NewMFARecoveryCodeRepository(db)
	tenantRepo := postgres.NewTenantRepository(db)

	tenantID := uuid.New()
	tenant := &models.Tenant{
		ID:     tenantID,
		Name:   "Test Tenant",
		Domain: "test-" + tenantID.String() + ".example.com",
		Status: "active",
	}
	err := tenantRepo.Create(context.Background(), tenant)
	require.NoError(t, err)

	userID := uuid.New()
	user := &models.User{
		ID:         userID,
		TenantID:   tenantID,
		Username:   "challengeuser",
		Email:      "challenge@example.com",
		Status:     "active",
		MFAEnabled: true, // Enable MFA for this user
	}
	err = userRepo.Create(context.Background(), user)
	require.NoError(t, err)

	hasher := password.NewHasher()
	hashedPassword, err := hasher.Hash("TestPassword123!")
	require.NoError(t, err)

	cred := &credential.Credential{
		UserID:       userID,
		PasswordHash: hashedPassword,
	}
	err = credentialRepo.Create(context.Background(), cred)
	require.NoError(t, err)

	encryptionKey := make([]byte, 32)
	copy(encryptionKey, "test-encryption-key-32-bytes-long!")
	encryptor, err := encryption.NewEncryptor(encryptionKey)
	require.NoError(t, err)

	totpGenerator := totp.NewGenerator("Test Issuer")
	cacheClient := cache.NewCache(nil)
	sessionManager := NewSessionManager(cacheClient)

	service := NewService(userRepo, credentialRepo, mfaRecoveryCodeRepo, totpGenerator, encryptor, sessionManager)

	// Test challenge creation
	req := &ChallengeRequest{
		UserID:   userID,
		TenantID: tenantID,
	}

	response, err := service.CreateChallenge(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.SessionID)
	assert.Greater(t, response.ExpiresIn, 0)
}

