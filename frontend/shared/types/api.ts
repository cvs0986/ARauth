/**
 * API Types and Interfaces
 */

// Common types
export interface ApiResponse<T = unknown> {
  data?: T;
  message?: string;
  error?: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  pagination?: {
    page: number;
    limit: number;
    total: number;
    total_pages: number;
  };
}

// Auth types
export interface LoginRequest {
  username: string;
  password: string;
  tenant_id?: string;
  remember_me?: boolean;
}

export interface LoginResponse {
  access_token?: string;
  refresh_token?: string;
  id_token?: string;
  token_type?: string;
  expires_in?: number;
  refresh_expires_in?: number;
  remember_me?: boolean;
  mfa_required?: boolean;
  mfa_session_id?: string;
  user_id?: string;   // Returned when MFA is required
  tenant_id?: string; // Returned when MFA is required
  redirect_to?: string;
}

// Tenant types
export interface Tenant {
  id: string;
  name: string;
  domain: string;
  status: 'active' | 'inactive' | 'suspended';
  created_at: string;
  updated_at: string;
}

export interface CreateTenantRequest {
  name: string;
  domain: string;
  status?: 'active' | 'inactive';
}

// Tenant Settings types
export interface TenantSettings {
  id?: string;
  tenant_id: string;
  access_token_ttl_minutes: number;
  refresh_token_ttl_days: number;
  id_token_ttl_minutes: number;
  remember_me_enabled: boolean;
  remember_me_refresh_token_ttl_days: number;
  remember_me_access_token_ttl_minutes: number;
  token_rotation_enabled: boolean;
  require_mfa_for_extended_sessions: boolean;
  // Security Settings
  min_password_length: number;
  require_uppercase: boolean;
  require_lowercase: boolean;
  require_numbers: boolean;
  require_special_chars: boolean;
  password_expiry_days?: number | null;
  mfa_required: boolean;
  rate_limit_requests: number;
  rate_limit_window_seconds: number;
}

export interface UpdateTenantSettingsRequest {
  access_token_ttl_minutes?: number;
  refresh_token_ttl_days?: number;
  id_token_ttl_minutes?: number;
  remember_me_enabled?: boolean;
  remember_me_refresh_token_ttl_days?: number;
  remember_me_access_token_ttl_minutes?: number;
  token_rotation_enabled?: boolean;
  require_mfa_for_extended_sessions?: boolean;
  min_password_length?: number;
  require_uppercase?: boolean;
  require_lowercase?: boolean;
  require_numbers?: boolean;
  require_special_chars?: boolean;
  password_expiry_days?: number | null;
  mfa_required?: boolean;
  rate_limit_requests?: number;
  rate_limit_window_seconds?: number;
}

// User types
export interface User {
  id: string;
  tenant_id: string;
  username: string;
  email: string;
  first_name?: string;
  last_name?: string;
  status: 'active' | 'inactive' | 'locked';
  mfa_enabled?: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateUserRequest {
  username: string;
  email: string;
  password: string;
  first_name?: string;
  last_name?: string;
  tenant_id?: string;
}

export interface UpdateUserRequest {
  email?: string;
  first_name?: string;
  last_name?: string;
  status?: 'active' | 'inactive' | 'locked';
}

// Role types
export interface Role {
  id: string;
  tenant_id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateRoleRequest {
  name: string;
  description?: string;
}

export interface UpdateRoleRequest {
  name?: string;
  description?: string;
}

// Permission types
export interface Permission {
  id: string;
  tenant_id: string;
  resource: string;
  action: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

export interface CreatePermissionRequest {
  resource: string;
  action: string;
  description?: string;
}

export interface UpdatePermissionRequest {
  resource?: string;
  action?: string;
  description?: string;
}

// MFA types
export interface MFAEnrollResponse {
  secret: string;
  qr_code: string;
  recovery_codes: string[];
}

export interface MFAVerifyRequest {
  code: string;
}

export interface MFAChallengeRequest {
  username: string;
  password: string;
}

export interface MFAVerifyChallengeRequest {
  challenge_id: string;
  code: string;
}

// Capability types
export interface SystemCapability {
  capability_key: string;
  enabled: boolean;
  default_value?: Record<string, unknown>;
  description?: string;
  created_at: string;
  updated_at: string;
  created_by?: string;
  updated_by?: string;
}

export interface TenantCapability {
  tenant_id: string;
  capability_key: string;
  enabled: boolean;
  value?: Record<string, unknown>;
  configured_at: string;
  configured_by?: string;
}

export interface TenantFeature {
  tenant_id: string;
  capability_key: string;
  enabled: boolean;
  configuration?: Record<string, unknown>;
  enabled_at: string;
  enabled_by?: string;
}

export interface UserCapabilityState {
  user_id: string;
  capability_key: string;
  enrolled: boolean;
  state_data?: Record<string, unknown>;
  enrolled_at?: string;
  last_verified_at?: string;
}

export interface CapabilityEvaluation {
  capability_key: string;
  can_use: boolean;
  reason?: string;
  system_supported: boolean;
  tenant_allowed: boolean;
  tenant_enabled: boolean;
  user_enrolled: boolean;
  system_value?: Record<string, unknown>;
  tenant_value?: Record<string, unknown>;
  tenant_configuration?: Record<string, unknown>;
}

export interface UpdateSystemCapabilityRequest {
  enabled?: boolean;
  default_value?: Record<string, unknown>;
  description?: string;
}

export interface SetTenantCapabilityRequest {
  enabled: boolean;
  value?: Record<string, unknown>;
}

export interface EnableTenantFeatureRequest {
  configuration?: Record<string, unknown>;
}

export interface EnrollUserCapabilityRequest {
  state_data?: Record<string, unknown>;
}

// Error types
export interface ApiError {
  message: string;
  code?: string;
  details?: Record<string, unknown>;
}

