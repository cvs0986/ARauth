/**
 * Authentication Store (Zustand)
 * Supports both SYSTEM and TENANT users
 */

import { create } from 'zustand';

export type PrincipalType = 'SYSTEM' | 'TENANT';

interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;
  tenantId: string | null;
  principalType: PrincipalType | null;
  systemPermissions: string[];
  permissions: string[];
  systemRoles: string[]; // System roles from JWT
  username: string | null; // Username from JWT
  email: string | null; // Email from JWT
  userId: string | null; // User ID from JWT
  isAuthenticated: boolean;
  selectedTenantId: string | null; // For SYSTEM users to select tenant context
  
  setAuth: (data: {
    accessToken: string;
    refreshToken?: string;
    tenantId?: string | null;
    principalType: PrincipalType;
    systemPermissions?: string[];
    permissions?: string[];
    systemRoles?: string[];
    username?: string;
    email?: string;
    userId?: string;
  }) => void;
  setTokens: (accessToken: string, refreshToken?: string) => void;
  setTenantId: (tenantId: string) => void;
  setSelectedTenantId: (tenantId: string | null) => void;
  clearAuth: () => void;
  
  // Helper getters
  isSystemUser: () => boolean;
  isTenantUser: () => boolean;
  hasSystemPermission: (permission: string) => boolean;
  hasPermission: (permission: string) => boolean;
  getCurrentTenantId: () => string | null; // Returns selectedTenantId for SYSTEM, tenantId for TENANT
}

const getStoredPrincipalType = (): PrincipalType | null => {
  const stored = localStorage.getItem('principal_type');
  return stored === 'SYSTEM' || stored === 'TENANT' ? stored : null;
};

export const useAuthStore = create<AuthState>((set, get) => ({
  accessToken: localStorage.getItem('access_token'),
  refreshToken: localStorage.getItem('refresh_token'),
  tenantId: localStorage.getItem('tenant_id'),
  principalType: getStoredPrincipalType(),
  systemPermissions: JSON.parse(localStorage.getItem('system_permissions') || '[]'),
  permissions: JSON.parse(localStorage.getItem('permissions') || '[]'),
  systemRoles: JSON.parse(localStorage.getItem('system_roles') || '[]'),
  username: localStorage.getItem('username'),
  email: localStorage.getItem('email'),
  userId: localStorage.getItem('user_id'),
  isAuthenticated: !!localStorage.getItem('access_token'),
  selectedTenantId: localStorage.getItem('selected_tenant_id'),
  
  setAuth: (data) => {
    set({
      accessToken: data.accessToken,
      refreshToken: data.refreshToken || null,
      tenantId: data.tenantId || null,
      principalType: data.principalType,
      systemPermissions: data.systemPermissions || [],
      permissions: data.permissions || [],
      systemRoles: data.systemRoles || [],
      username: data.username || null,
      email: data.email || null,
      userId: data.userId || null,
      isAuthenticated: !!data.accessToken,
    });
    
    // Store in localStorage
    localStorage.setItem('access_token', data.accessToken);
    if (data.refreshToken) {
      localStorage.setItem('refresh_token', data.refreshToken);
    }
    if (data.tenantId) {
      localStorage.setItem('tenant_id', data.tenantId);
    }
    localStorage.setItem('principal_type', data.principalType);
    localStorage.setItem('system_permissions', JSON.stringify(data.systemPermissions || []));
    localStorage.setItem('permissions', JSON.stringify(data.permissions || []));
    localStorage.setItem('system_roles', JSON.stringify(data.systemRoles || []));
    if (data.username) {
      localStorage.setItem('username', data.username);
    }
    if (data.email) {
      localStorage.setItem('email', data.email);
    }
    if (data.userId) {
      localStorage.setItem('user_id', data.userId);
    }
  },
  
  setTokens: (accessToken, refreshToken) => {
    set({
      accessToken,
      refreshToken: refreshToken || null,
      isAuthenticated: !!accessToken,
    });
    localStorage.setItem('access_token', accessToken);
    if (refreshToken) {
      localStorage.setItem('refresh_token', refreshToken);
    }
  },
  
  setTenantId: (tenantId) => {
    set({ tenantId });
    localStorage.setItem('tenant_id', tenantId);
  },
  
  setSelectedTenantId: (tenantId) => {
    set({ selectedTenantId: tenantId });
    if (tenantId) {
      localStorage.setItem('selected_tenant_id', tenantId);
    } else {
      localStorage.removeItem('selected_tenant_id');
    }
  },
  
  clearAuth: () => {
    set({
      accessToken: null,
      refreshToken: null,
      tenantId: null,
      principalType: null,
      systemPermissions: [],
      permissions: [],
      systemRoles: [],
      username: null,
      email: null,
      userId: null,
      isAuthenticated: false,
      selectedTenantId: null,
    });
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('tenant_id');
    localStorage.removeItem('principal_type');
    localStorage.removeItem('system_permissions');
    localStorage.removeItem('permissions');
    localStorage.removeItem('system_roles');
    localStorage.removeItem('username');
    localStorage.removeItem('email');
    localStorage.removeItem('user_id');
    localStorage.removeItem('selected_tenant_id');
  },
  
  // Helper methods
  isSystemUser: () => {
    return get().principalType === 'SYSTEM';
  },
  
  isTenantUser: () => {
    return get().principalType === 'TENANT';
  },
  
  hasSystemPermission: (permission: string) => {
    const { systemPermissions } = get();
    return systemPermissions.includes(permission) || 
           systemPermissions.includes('*:*') ||
           systemPermissions.some(p => p.startsWith(permission.split(':')[0] + ':*'));
  },
  
  hasPermission: (permission: string) => {
    const { permissions } = get();
    return permissions.includes(permission) || 
           permissions.includes('*:*') ||
           permissions.some(p => p.startsWith(permission.split(':')[0] + ':*'));
  },
  
  getCurrentTenantId: () => {
    const { principalType, tenantId, selectedTenantId } = get();
    if (principalType === 'SYSTEM') {
      return selectedTenantId; // SYSTEM users can select tenant context
    }
    return tenantId; // TENANT users are locked to their tenant
  },
}));

// Listen to custom events from API client for token updates and logout
if (typeof window !== 'undefined') {
  // Listen for token refresh events
  window.addEventListener('auth:tokens-updated', ((event: CustomEvent<{ accessToken: string; refreshToken?: string }>) => {
    const { accessToken, refreshToken } = event.detail;
    useAuthStore.getState().setTokens(accessToken, refreshToken);
  }) as EventListener);

  // Listen for logout events
  window.addEventListener('auth:logout', () => {
    useAuthStore.getState().clearAuth();
  });
}

