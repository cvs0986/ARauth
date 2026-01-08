// +build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/internal/testutil"
	"github.com/arauth-identity/iam/storage/postgres"
)

func TestE2E_CapabilityFlow(t *testing.T) {
	// Setup test database
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)
	defer testutil.CleanupTestDB(t, db)

	// Setup repositories
	tenantRepo := postgres.NewTenantRepository(db)
	userRepo := postgres.NewUserRepository(db)
	systemCapabilityRepo := postgres.NewSystemCapabilityRepository(db)
	tenantCapabilityRepo := postgres.NewTenantCapabilityRepository(db)
	tenantFeatureRepo := postgres.NewTenantFeatureEnablementRepository(db)
	userCapabilityRepo := postgres.NewUserCapabilityStateRepository(db)

	// Create test tenant
	tenant := &models.Tenant{
		ID:     uuid.New(),
		Name:   "Test Tenant",
		Domain: "test.example.com",
		Status: models.TenantStatusActive,
	}
	require.NoError(t, tenantRepo.Create(context.Background(), tenant))

	// Create test system user (for system admin operations)
	systemUser := &models.User{
		ID:           uuid.New(),
		Username:     "systemadmin",
		Email:        "admin@system.com",
		Status:       models.UserStatusActive,
		PrincipalType: models.PrincipalTypeSystem,
	}
	require.NoError(t, userRepo.Create(context.Background(), systemUser))

	// Create test tenant user
	tenantUser := &models.User{
		ID:           uuid.New(),
		TenantID:     &tenant.ID,
		Username:     "tenantuser",
		Email:        "user@test.example.com",
		Status:       models.UserStatusActive,
		PrincipalType: models.PrincipalTypeTenant,
	}
	require.NoError(t, userRepo.Create(context.Background(), tenantUser))

	// Step 1: System admin creates a system capability
	t.Run("System admin creates system capability", func(t *testing.T) {
		systemCap := &models.SystemCapability{
			CapabilityKey: "mfa",
			Enabled:       true,
		}
		err := systemCapabilityRepo.Create(context.Background(), systemCap)
		require.NoError(t, err)

		// Verify it was created
		retrieved, err := systemCapabilityRepo.GetByKey(context.Background(), "mfa")
		require.NoError(t, err)
		assert.Equal(t, "mfa", retrieved.CapabilityKey)
		assert.True(t, retrieved.Enabled)
	})

	// Step 2: System admin assigns capability to tenant
	t.Run("System admin assigns capability to tenant", func(t *testing.T) {
		tenantCap := &models.TenantCapability{
			TenantID:      tenant.ID,
			CapabilityKey: "mfa",
			Enabled:       true,
			ConfiguredBy:  &systemUser.ID,
		}
		err := tenantCapabilityRepo.Create(context.Background(), tenantCap)
		require.NoError(t, err)

		// Verify it was assigned
		retrieved, err := tenantCapabilityRepo.GetByTenantIDAndKey(context.Background(), tenant.ID, "mfa")
		require.NoError(t, err)
		assert.Equal(t, tenant.ID, retrieved.TenantID)
		assert.Equal(t, "mfa", retrieved.CapabilityKey)
		assert.True(t, retrieved.Enabled)
	})

	// Step 3: Tenant admin enables the feature
	t.Run("Tenant admin enables feature", func(t *testing.T) {
		feature := &models.TenantFeatureEnablement{
			TenantID:   tenant.ID,
			FeatureKey: "mfa",
			Enabled:    true,
			EnabledBy:  &tenantUser.ID,
		}
		err := tenantFeatureRepo.Create(context.Background(), feature)
		require.NoError(t, err)

		// Verify it was enabled
		retrieved, err := tenantFeatureRepo.GetByTenantIDAndKey(context.Background(), tenant.ID, "mfa")
		require.NoError(t, err)
		assert.Equal(t, tenant.ID, retrieved.TenantID)
		assert.Equal(t, "mfa", retrieved.FeatureKey)
		assert.True(t, retrieved.Enabled)
	})

	// Step 4: User enrolls in capability
	t.Run("User enrolls in capability", func(t *testing.T) {
		userState := &models.UserCapabilityState{
			UserID:       tenantUser.ID,
			CapabilityKey: "mfa",
			Enrolled:     true,
		}
		err := userCapabilityRepo.Create(context.Background(), userState)
		require.NoError(t, err)

		// Verify user is enrolled
		retrieved, err := userCapabilityRepo.GetByUserIDAndKey(context.Background(), tenantUser.ID, "mfa")
		require.NoError(t, err)
		assert.Equal(t, tenantUser.ID, retrieved.UserID)
		assert.Equal(t, "mfa", retrieved.CapabilityKey)
		assert.True(t, retrieved.Enrolled)
	})

	// Step 5: Verify full capability flow works
	t.Run("Verify capability evaluation", func(t *testing.T) {
		// This would use the capability service to evaluate
		// For now, we'll verify the data exists at each layer
		
		// System level
		systemCap, err := systemCapabilityRepo.GetByKey(context.Background(), "mfa")
		require.NoError(t, err)
		assert.True(t, systemCap.IsSupported())

		// Tenant level
		tenantCap, err := tenantCapabilityRepo.GetByTenantIDAndKey(context.Background(), tenant.ID, "mfa")
		require.NoError(t, err)
		assert.True(t, tenantCap.IsAllowed())

		// Feature level
		feature, err := tenantFeatureRepo.GetByTenantIDAndKey(context.Background(), tenant.ID, "mfa")
		require.NoError(t, err)
		assert.True(t, feature.IsEnabled())

		// User level
		userState, err := userCapabilityRepo.GetByUserIDAndKey(context.Background(), tenantUser.ID, "mfa")
		require.NoError(t, err)
		assert.True(t, userState.IsEnrolled())
	})
}

func TestE2E_CapabilityEnforcement(t *testing.T) {
	// Setup test database
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)
	defer testutil.CleanupTestDB(t, db)

	// Setup repositories
	tenantRepo := postgres.NewTenantRepository(db)
	userRepo := postgres.NewUserRepository(db)
	systemCapabilityRepo := postgres.NewSystemCapabilityRepository(db)
	tenantCapabilityRepo := postgres.NewTenantCapabilityRepository(db)

	// Create test tenant
	tenant := &models.Tenant{
		ID:     uuid.New(),
		Name:   "Test Tenant",
		Domain: "test.example.com",
		Status: models.TenantStatusActive,
	}
	require.NoError(t, tenantRepo.Create(context.Background(), tenant))

	// Create system capability but don't assign to tenant
	t.Run("Tenant cannot use unassigned capability", func(t *testing.T) {
		systemCap := &models.SystemCapability{
			CapabilityKey: "saml",
			Enabled:       true,
		}
		err := systemCapabilityRepo.Create(context.Background(), systemCap)
		require.NoError(t, err)

		// Try to get tenant capability (should not exist)
		_, err = tenantCapabilityRepo.GetByTenantIDAndKey(context.Background(), tenant.ID, "saml")
		assert.Error(t, err) // Should not be found
	})
}

