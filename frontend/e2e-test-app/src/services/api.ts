/**
 * API Service Layer for E2E Testing App
 */

import { apiClient, handleApiError } from '../../../shared/utils/api-client';
import { API_ENDPOINTS } from '../../../shared/constants/api';
import type {
  LoginRequest,
  LoginResponse,
  User,
  CreateUserRequest,
} from '../../../shared/types/api';

// Auth API
export const authApi = {
  login: async (data: LoginRequest): Promise<LoginResponse> => {
    // Extract tenant_id from data and send as header
    const { tenant_id, remember_me, ...loginData } = data;
    
    // Create request config with X-Tenant-ID header if tenant_id is provided
    const config = tenant_id
      ? { headers: { 'X-Tenant-ID': tenant_id } }
      : {};
    
    // Include remember_me in request body
    const requestBody = {
      ...loginData,
      ...(remember_me !== undefined && { remember_me }),
    };
    
    const response = await apiClient.post<LoginResponse>(
      API_ENDPOINTS.AUTH.LOGIN,
      requestBody,
      config
    );
    return response.data;
  },
};

// User API (for registration)
export const userApi = {
  create: async (data: CreateUserRequest): Promise<User> => {
    const response = await apiClient.post<User>(
      API_ENDPOINTS.USERS.BASE,
      data
    );
    return response.data;
  },
};

// Export error handler
export { handleApiError };

