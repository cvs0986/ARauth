/**
 * API Service Layer
 */

import { apiClient, handleApiError } from '../../../shared/utils/api-client';
import { API_ENDPOINTS } from '../../../shared/constants/api';
import type {
  LoginRequest,
  LoginResponse,
  Tenant,
  CreateTenantRequest,
  User,
  CreateUserRequest,
  UpdateUserRequest,
  Role,
  CreateRoleRequest,
  UpdateRoleRequest,
  Permission,
  CreatePermissionRequest,
  UpdatePermissionRequest,
} from '../../shared/types/api';

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

// Tenant API
export const tenantApi = {
  list: async (): Promise<Tenant[]> => {
    const response = await apiClient.get<{ tenants: Tenant[]; page: number; page_size: number }>(API_ENDPOINTS.TENANTS.BASE);
    return response.data.tenants || [];
  },
  
  getById: async (id: string): Promise<Tenant> => {
    const response = await apiClient.get<Tenant>(API_ENDPOINTS.TENANTS.BY_ID(id));
    return response.data;
  },
  
  getByDomain: async (domain: string): Promise<Tenant> => {
    const response = await apiClient.get<Tenant>(
      API_ENDPOINTS.TENANTS.BY_DOMAIN(domain)
    );
    return response.data;
  },
  
  create: async (data: CreateTenantRequest): Promise<Tenant> => {
    const response = await apiClient.post<Tenant>(
      API_ENDPOINTS.TENANTS.BASE,
      data
    );
    return response.data;
  },
  
  update: async (id: string, data: Partial<CreateTenantRequest>): Promise<Tenant> => {
    const response = await apiClient.put<Tenant>(
      API_ENDPOINTS.TENANTS.BY_ID(id),
      data
    );
    return response.data;
  },
  
  delete: async (id: string): Promise<void> => {
    await apiClient.delete(API_ENDPOINTS.TENANTS.BY_ID(id));
  },
};

// User API
export const userApi = {
  list: async (): Promise<User[]> => {
    const response = await apiClient.get<{ users: User[]; page: number; page_size: number; total: number }>(API_ENDPOINTS.USERS.BASE);
    return response.data.users || [];
  },
  
  getById: async (id: string): Promise<User> => {
    const response = await apiClient.get<User>(API_ENDPOINTS.USERS.BY_ID(id));
    return response.data;
  },
  
  create: async (data: CreateUserRequest): Promise<User> => {
    const response = await apiClient.post<User>(
      API_ENDPOINTS.USERS.BASE,
      data
    );
    return response.data;
  },
  
  update: async (id: string, data: UpdateUserRequest): Promise<User> => {
    const response = await apiClient.put<User>(
      API_ENDPOINTS.USERS.BY_ID(id),
      data
    );
    return response.data;
  },
  
  delete: async (id: string): Promise<void> => {
    await apiClient.delete(API_ENDPOINTS.USERS.BY_ID(id));
  },
};

// Role API
export const roleApi = {
  list: async (): Promise<Role[]> => {
    const response = await apiClient.get<{ roles: Role[]; page: number; page_size: number; total: number }>(API_ENDPOINTS.ROLES.BASE);
    return response.data.roles || [];
  },
  
  getById: async (id: string): Promise<Role> => {
    const response = await apiClient.get<Role>(API_ENDPOINTS.ROLES.BY_ID(id));
    return response.data;
  },
  
  create: async (data: CreateRoleRequest): Promise<Role> => {
    const response = await apiClient.post<Role>(
      API_ENDPOINTS.ROLES.BASE,
      data
    );
    return response.data;
  },
  
  update: async (id: string, data: UpdateRoleRequest): Promise<Role> => {
    const response = await apiClient.put<Role>(
      API_ENDPOINTS.ROLES.BY_ID(id),
      data
    );
    return response.data;
  },
  
  delete: async (id: string): Promise<void> => {
    await apiClient.delete(API_ENDPOINTS.ROLES.BY_ID(id));
  },
  
  getPermissions: async (roleId: string): Promise<Permission[]> => {
    const response = await apiClient.get<Permission[]>(
      API_ENDPOINTS.ROLES.PERMISSIONS(roleId)
    );
    return response.data;
  },
  
  assignPermission: async (roleId: string, permissionId: string): Promise<void> => {
    await apiClient.post(
      API_ENDPOINTS.ROLES.ASSIGN_PERMISSION(roleId, permissionId)
    );
  },
  
  removePermission: async (roleId: string, permissionId: string): Promise<void> => {
    await apiClient.delete(
      API_ENDPOINTS.ROLES.ASSIGN_PERMISSION(roleId, permissionId)
    );
  },
};

// Permission API
export const permissionApi = {
  list: async (): Promise<Permission[]> => {
    const response = await apiClient.get<{ permissions: Permission[]; page: number; page_size: number; total: number }>(
      API_ENDPOINTS.PERMISSIONS.BASE
    );
    return response.data.permissions || [];
  },
  
  getById: async (id: string): Promise<Permission> => {
    const response = await apiClient.get<Permission>(
      API_ENDPOINTS.PERMISSIONS.BY_ID(id)
    );
    return response.data;
  },
  
  create: async (data: CreatePermissionRequest): Promise<Permission> => {
    const response = await apiClient.post<Permission>(
      API_ENDPOINTS.PERMISSIONS.BASE,
      data
    );
    return response.data;
  },
  
  update: async (id: string, data: UpdatePermissionRequest): Promise<Permission> => {
    const response = await apiClient.put<Permission>(
      API_ENDPOINTS.PERMISSIONS.BY_ID(id),
      data
    );
    return response.data;
  },
  
  delete: async (id: string): Promise<void> => {
    await apiClient.delete(API_ENDPOINTS.PERMISSIONS.BY_ID(id));
  },
};

// Export error handler
export { handleApiError };

