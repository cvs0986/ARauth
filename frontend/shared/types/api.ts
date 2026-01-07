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
  access_token: string;
  refresh_token?: string;
  id_token?: string;
  token_type: string;
  expires_in: number;
  refresh_expires_in?: number;
  remember_me?: boolean;
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

// User types
export interface User {
  id: string;
  tenant_id: string;
  username: string;
  email: string;
  first_name?: string;
  last_name?: string;
  status: 'active' | 'inactive' | 'locked';
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

// Error types
export interface ApiError {
  message: string;
  code?: string;
  details?: Record<string, unknown>;
}

