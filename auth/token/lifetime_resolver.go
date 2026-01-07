package token

import (
	"context"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/config"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// LifetimeResolver resolves token lifetimes from multiple sources
type LifetimeResolver struct {
	config      *config.SecurityConfig
	settingsRepo interfaces.TenantSettingsRepository
}

// NewLifetimeResolver creates a new lifetime resolver
func NewLifetimeResolver(cfg *config.SecurityConfig, settingsRepo interfaces.TenantSettingsRepository) *LifetimeResolver {
	return &LifetimeResolver{
		config:      cfg,
		settingsRepo: settingsRepo,
	}
}

// TokenLifetimes holds all token lifetime values
type TokenLifetimes struct {
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	IDTokenTTL      time.Duration
}

// GetAccessTokenTTL resolves access token TTL with priority:
// 1. Per-tenant settings (if remember_me and enabled)
// 2. Environment variables
// 3. Config file
// 4. System defaults
func (r *LifetimeResolver) GetAccessTokenTTL(ctx context.Context, tenantID uuid.UUID, rememberMe bool) time.Duration {
	// 1. Check per-tenant settings
	if r.settingsRepo != nil {
		settings, err := r.settingsRepo.GetByTenantID(ctx, tenantID)
		if err == nil && settings != nil {
			if rememberMe && settings.RememberMeEnabled {
				return time.Duration(settings.RememberMeAccessTokenTTLMinutes) * time.Minute
			}
			return time.Duration(settings.AccessTokenTTLMinutes) * time.Minute
		}
	}

	// 2. Check environment variables
	if envTTL := os.Getenv("JWT_ACCESS_TOKEN_TTL"); envTTL != "" {
		if rememberMe {
			if rememberTTL := os.Getenv("JWT_REMEMBER_ME_ACCESS_TTL"); rememberTTL != "" {
				if ttl, err := time.ParseDuration(rememberTTL); err == nil {
					return ttl
				}
			}
		}
		if ttl, err := time.ParseDuration(envTTL); err == nil {
			return ttl
		}
	}

	// 3. Check config file
	if r.config != nil && r.config.JWT.AccessTokenTTL > 0 {
		if rememberMe && r.config.JWT.RememberMe.Enabled {
			if r.config.JWT.RememberMe.AccessTokenTTL > 0 {
				return r.config.JWT.RememberMe.AccessTokenTTL
			}
		}
		return r.config.JWT.AccessTokenTTL
	}

	// 4. System defaults
	if rememberMe {
		return 60 * time.Minute
	}
	return 15 * time.Minute
}

// GetRefreshTokenTTL resolves refresh token TTL
func (r *LifetimeResolver) GetRefreshTokenTTL(ctx context.Context, tenantID uuid.UUID, rememberMe bool) time.Duration {
	// 1. Check per-tenant settings
	if r.settingsRepo != nil {
		settings, err := r.settingsRepo.GetByTenantID(ctx, tenantID)
		if err == nil && settings != nil {
			if rememberMe && settings.RememberMeEnabled {
				return time.Duration(settings.RememberMeRefreshTokenTTLDays) * 24 * time.Hour
			}
			return time.Duration(settings.RefreshTokenTTLDays) * 24 * time.Hour
		}
	}

	// 2. Check environment variables
	if envTTL := os.Getenv("JWT_REFRESH_TOKEN_TTL"); envTTL != "" {
		if rememberMe {
			if rememberTTL := os.Getenv("JWT_REMEMBER_ME_REFRESH_TTL"); rememberTTL != "" {
				if ttl, err := time.ParseDuration(rememberTTL); err == nil {
					return ttl
				}
			}
		}
		if ttl, err := time.ParseDuration(envTTL); err == nil {
			return ttl
		}
	}

	// 3. Check config file
	if r.config != nil && r.config.JWT.RefreshTokenTTL > 0 {
		if rememberMe && r.config.JWT.RememberMe.Enabled {
			if r.config.JWT.RememberMe.RefreshTokenTTL > 0 {
				return r.config.JWT.RememberMe.RefreshTokenTTL
			}
		}
		return r.config.JWT.RefreshTokenTTL
	}

	// 4. System defaults
	if rememberMe {
		return 90 * 24 * time.Hour // 90 days
	}
	return 30 * 24 * time.Hour // 30 days
}

// GetIDTokenTTL resolves ID token TTL
func (r *LifetimeResolver) GetIDTokenTTL(ctx context.Context, tenantID uuid.UUID) time.Duration {
	// 1. Check per-tenant settings
	if r.settingsRepo != nil {
		settings, err := r.settingsRepo.GetByTenantID(ctx, tenantID)
		if err == nil && settings != nil {
			return time.Duration(settings.IDTokenTTLMinutes) * time.Minute
		}
	}

	// 2. Check environment variables
	if envTTL := os.Getenv("JWT_ID_TOKEN_TTL"); envTTL != "" {
		if ttl, err := time.ParseDuration(envTTL); err == nil {
			return ttl
		}
	}

	// 3. Check config file
	if r.config != nil && r.config.JWT.IDTokenTTL > 0 {
		return r.config.JWT.IDTokenTTL
	}

	// 4. System default
	return 1 * time.Hour
}

// GetAllLifetimes gets all token lifetimes at once
func (r *LifetimeResolver) GetAllLifetimes(ctx context.Context, tenantID uuid.UUID, rememberMe bool) *TokenLifetimes {
	return &TokenLifetimes{
		AccessTokenTTL:  r.GetAccessTokenTTL(ctx, tenantID, rememberMe),
		RefreshTokenTTL: r.GetRefreshTokenTTL(ctx, tenantID, rememberMe),
		IDTokenTTL:      r.GetIDTokenTTL(ctx, tenantID),
	}
}

