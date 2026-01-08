/**
 * API Configuration Constants
 */

export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
export const API_VERSION = 'v1';
export const API_PREFIX = `/api/${API_VERSION}`;

/**
 * API Endpoints
 */
export const API_ENDPOINTS = {
  // Auth
  AUTH: {
    LOGIN: `${API_PREFIX}/auth/login`,
  },
  
  // Tenants
  TENANTS: {
    BASE: `${API_PREFIX}/tenants`,
    BY_ID: (id: string) => `${API_PREFIX}/tenants/${id}`,
    BY_DOMAIN: (domain: string) => `${API_PREFIX}/tenants/domain/${domain}`,
  },
  
  // Users
  USERS: {
    BASE: `${API_PREFIX}/users`,
    BY_ID: (id: string) => `${API_PREFIX}/users/${id}`,
  },
  
  // Roles
  ROLES: {
    BASE: `${API_PREFIX}/roles`,
    BY_ID: (id: string) => `${API_PREFIX}/roles/${id}`,
    PERMISSIONS: (roleId: string) => `${API_PREFIX}/roles/${roleId}/permissions`,
    ASSIGN_PERMISSION: (roleId: string, permissionId: string) => 
      `${API_PREFIX}/roles/${roleId}/permissions/${permissionId}`,
  },
  
  // Permissions
  PERMISSIONS: {
    BASE: `${API_PREFIX}/permissions`,
    BY_ID: (id: string) => `${API_PREFIX}/permissions/${id}`,
  },
  
  // MFA
  MFA: {
    ENROLL: `${API_PREFIX}/mfa/enroll`,
    VERIFY: `${API_PREFIX}/mfa/verify`,
    CHALLENGE: `${API_PREFIX}/mfa/challenge`,
    VERIFY_CHALLENGE: `${API_PREFIX}/mfa/challenge/verify`,
  },
  
  // Health
  HEALTH: {
    BASE: '/health',
    LIVE: '/health/live',
    READY: '/health/ready',
  },
  
  // System Capabilities (SYSTEM users only)
  SYSTEM_CAPABILITIES: {
    BASE: '/system/capabilities',
    BY_KEY: (key: string) => `/system/capabilities/${key}`,
  },
  
  // Tenant Capabilities (SYSTEM users only)
  TENANT_CAPABILITIES: {
    BASE: (tenantId: string) => `/system/tenants/${tenantId}/capabilities`,
    BY_KEY: (tenantId: string, key: string) => `/system/tenants/${tenantId}/capabilities/${key}`,
    EVALUATION: (tenantId: string) => `/system/tenants/${tenantId}/capabilities/evaluation`,
  },
  
  // Tenant Features (TENANT users)
  TENANT_FEATURES: {
    BASE: `${API_PREFIX}/tenant/features`,
    BY_KEY: (key: string) => `${API_PREFIX}/tenant/features/${key}`,
  },
  
  // User Capabilities (TENANT users)
  USER_CAPABILITIES: {
    BASE: (userId: string) => `${API_PREFIX}/users/${userId}/capabilities`,
    BY_KEY: (userId: string, key: string) => `${API_PREFIX}/users/${userId}/capabilities/${key}`,
    ENROLL: (userId: string, key: string) => `${API_PREFIX}/users/${userId}/capabilities/${key}/enroll`,
  },
} as const;

