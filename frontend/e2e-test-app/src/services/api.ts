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
    const response = await apiClient.post<LoginResponse>(
      API_ENDPOINTS.AUTH.LOGIN,
      data
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

