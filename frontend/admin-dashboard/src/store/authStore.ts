/**
 * Authentication Store (Zustand)
 */

import { create } from 'zustand';

interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;
  tenantId: string | null;
  isAuthenticated: boolean;
  setTokens: (accessToken: string, refreshToken?: string) => void;
  setTenantId: (tenantId: string) => void;
  clearAuth: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  accessToken: localStorage.getItem('access_token'),
  refreshToken: localStorage.getItem('refresh_token'),
  tenantId: localStorage.getItem('tenant_id'),
  isAuthenticated: !!localStorage.getItem('access_token'),
  
  setTokens: (accessToken, refreshToken) => {
    set({
      accessToken,
      refreshToken: refreshToken || null,
      isAuthenticated: !!accessToken,
    });
    // Store in localStorage for API client
    localStorage.setItem('access_token', accessToken);
    if (refreshToken) {
      localStorage.setItem('refresh_token', refreshToken);
    }
  },
  
  setTenantId: (tenantId) => {
    set({ tenantId });
    localStorage.setItem('tenant_id', tenantId);
  },
  
  clearAuth: () => {
    set({
      accessToken: null,
      refreshToken: null,
      tenantId: null,
      isAuthenticated: false,
    });
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('tenant_id');
  },
}));

