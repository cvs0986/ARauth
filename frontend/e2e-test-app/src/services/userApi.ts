/**
 * User API Service for E2E Testing App
 */

import { apiClient, handleApiError } from '../../../shared/utils/api-client';
import { API_ENDPOINTS } from '../../../shared/constants/api';
import type { User, Role, Permission } from '../../../shared/types/api';

export const userApi = {
  getCurrentUser: async (): Promise<User> => {
    const response = await apiClient.get<User>(`${API_ENDPOINTS.USERS.BASE}/me`);
    return response.data;
  },

  getUserRoles: async (userId: string): Promise<Role[]> => {
    const response = await apiClient.get<Role[]>(`${API_ENDPOINTS.USERS.BASE}/${userId}/roles`);
    return response.data;
  },

  getUserPermissions: async (userId: string): Promise<Permission[]> => {
    const response = await apiClient.get<Permission[]>(`${API_ENDPOINTS.USERS.BASE}/${userId}/permissions`);
    return response.data;
  },
};

export { handleApiError };

