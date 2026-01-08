/**
 * API Client Utility
 * Base configuration for API requests
 */

import axios, { AxiosInstance, AxiosRequestConfig, AxiosError } from 'axios';
import { API_BASE_URL } from '../constants/api';

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
            // TODO: Implement token refresh endpoint
            // const response = await axios.post(`${API_BASE_URL}/auth/refresh`, {
            //   refresh_token: refreshToken,
            // });
            // const { access_token } = response.data;
            // localStorage.setItem('access_token', access_token);
            // originalRequest.headers = originalRequest.headers || {};
            // originalRequest.headers.Authorization = `Bearer ${access_token}`;
            // return client(originalRequest);
          } catch (refreshError) {
            // Refresh failed - clear tokens and redirect to login
            localStorage.removeItem('access_token');
            localStorage.removeItem('refresh_token');
            localStorage.removeItem('tenant_id');
            window.location.href = '/login';
            return Promise.reject(refreshError);
          }
        } else {
          // No refresh token - redirect to login
          localStorage.removeItem('access_token');
          localStorage.removeItem('tenant_id');
          window.location.href = '/login';
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

