/**
 * API Client Utility
 * Base configuration for API requests
 */

import axios, { AxiosInstance, AxiosRequestConfig, AxiosError } from 'axios';
import { API_BASE_URL, API_ENDPOINTS } from '../constants/api';

/**
 * Helper function to update auth tokens
 */
function updateAuthTokens(accessToken: string, refreshToken?: string) {
  localStorage.setItem('access_token', accessToken);
  if (refreshToken) {
    localStorage.setItem('refresh_token', refreshToken);
  }
  
  // Dispatch custom event for auth store to listen to
  window.dispatchEvent(new CustomEvent('auth:tokens-updated', {
    detail: { accessToken, refreshToken },
  }));
}

/**
 * Helper function to handle logout - clears auth state and redirects to login
 */
function handleLogout() {
  // Clear localStorage
  localStorage.removeItem('access_token');
  localStorage.removeItem('refresh_token');
  localStorage.removeItem('tenant_id');
  localStorage.removeItem('principal_type');
  localStorage.removeItem('system_permissions');
  localStorage.removeItem('permissions');
  localStorage.removeItem('selected_tenant_id');

  // Dispatch custom event for auth store to listen to
  window.dispatchEvent(new CustomEvent('auth:logout'));

  // Redirect to login
  window.location.href = '/login';
}

/**
 * Create configured axios instance
 */
export function createApiClient(baseURL: string = API_BASE_URL): AxiosInstance {
  const client = axios.create({
    baseURL,
    timeout: 30000,
    headers: {
      'Content-Type': 'application/json',
    },
  });

  // Request interceptor - Add auth token and tenant ID
  client.interceptors.request.use(
    (config) => {
      // Get token from localStorage
      const token = localStorage.getItem('access_token');
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }

      // Get principal type and tenant ID from localStorage
      const principalType = localStorage.getItem('principal_type');
      const tenantId = localStorage.getItem('tenant_id');
      const selectedTenantId = localStorage.getItem('selected_tenant_id');

      // For SYSTEM users:
      // - Use selectedTenantId if available (for tenant-scoped operations)
      // - Don't add X-Tenant-ID header if no tenant is selected (for system-level operations)
      // For TENANT users:
      // - Always use tenantId from localStorage
      if (principalType === 'SYSTEM') {
        // SYSTEM users: only add X-Tenant-ID if they've selected a tenant context
        if (selectedTenantId) {
          config.headers['X-Tenant-ID'] = selectedTenantId;
        }
        // If no selectedTenantId, don't add X-Tenant-ID header (for system-level APIs)
      } else if (principalType === 'TENANT' && tenantId) {
        // TENANT users: always add their tenant_id
        config.headers['X-Tenant-ID'] = tenantId;
      } else if (tenantId) {
        // Fallback: if principal_type is not set, use tenantId (for backward compatibility)
        config.headers['X-Tenant-ID'] = tenantId;
      }

      return config;
    },
    (error) => {
      return Promise.reject(error);
    }
  );

  // Response interceptor - Handle errors and token refresh
  client.interceptors.response.use(
    (response) => response,
    async (error: AxiosError) => {
      const originalRequest = error.config as AxiosRequestConfig & { _retry?: boolean };

      // Handle 401 Unauthorized
      if (error.response?.status === 401 && !originalRequest._retry) {
        originalRequest._retry = true;

        // Try to refresh token (if refresh token exists)
        const refreshToken = localStorage.getItem('refresh_token');
        if (refreshToken) {
          try {
            // Use axios directly (not apiClient) to avoid interceptor loops
            const response = await axios.post(
              `${API_BASE_URL}${API_ENDPOINTS.AUTH.REFRESH}`,
              { refresh_token: refreshToken },
              {
                headers: {
                  'Content-Type': 'application/json',
                },
              }
            );

            const { access_token, refresh_token: newRefreshToken } = response.data;

            // Update tokens in localStorage and notify auth store
            updateAuthTokens(access_token, newRefreshToken);

            // Retry original request with new token
            originalRequest.headers = originalRequest.headers || {};
            originalRequest.headers.Authorization = `Bearer ${access_token}`;
            return client(originalRequest);
          } catch (refreshError) {
            // Refresh failed - clear tokens and redirect to login
            handleLogout();
            return Promise.reject(refreshError);
          }
        } else {
          // No refresh token - redirect to login
          handleLogout();
        }
      }

      return Promise.reject(error);
    }
  );

  return client;
}

/**
 * Default API client instance
 */
export const apiClient = createApiClient();

/**
 * Helper function to handle API errors
 */
export function handleApiError(error: unknown): string {
  if (axios.isAxiosError(error)) {
    const axiosError = error as AxiosError<{ message?: string; error?: string }>;
    
    if (axiosError.response) {
      // Server responded with error
      return (
        axiosError.response.data?.message ||
        axiosError.response.data?.error ||
        `Error: ${axiosError.response.status} ${axiosError.response.statusText}`
      );
    } else if (axiosError.request) {
      // Request made but no response
      return 'Network error: Please check your connection';
    }
  }
  
  return 'An unexpected error occurred';
}

