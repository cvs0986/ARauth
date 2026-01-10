/**
 * API Service Layer
 */

import { apiClient, handleApiError } from '../../../shared/utils/api-client';
import { API_ENDPOINTS, API_PREFIX } from '../../../shared/constants/api';
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
  SystemCapability,
  TenantCapability,
  TenantFeature,
  UserCapabilityState,
  CapabilityEvaluation,
  UpdateSystemCapabilityRequest,
  SetTenantCapabilityRequest,
  EnableTenantFeatureRequest,
  EnrollUserCapabilityRequest,
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

  refresh: async (refreshToken: string): Promise<{ access_token: string; refresh_token: string; token_type: string; expires_in: number }> => {
    const response = await apiClient.post<{ access_token: string; refresh_token: string; token_type: string; expires_in: number }>(
      API_ENDPOINTS.AUTH.REFRESH,
      { refresh_token: refreshToken }
    );
    return response.data;
  },

  revoke: async (token: string, tokenTypeHint?: 'access_token' | 'refresh_token'): Promise<void> => {
    await apiClient.post(
      API_ENDPOINTS.AUTH.REVOKE,
      {
        token,
        ...(tokenTypeHint && { token_type_hint: tokenTypeHint })
      }
    );
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

  // Tenant Settings API (for TENANT users to access their own settings)
  getSettings: async (): Promise<any> => {
    const response = await apiClient.get('/api/v1/tenant/settings');
    return response.data;
  },

  updateSettings: async (data: any): Promise<any> => {
    const response = await apiClient.put('/api/v1/tenant/settings', data);
    return response.data;
  },
};

// System API (for SYSTEM users only)
export const systemApi = {
  tenants: {
    list: async (): Promise<Tenant[]> => {
      const response = await apiClient.get<{ tenants: Tenant[]; page: number; page_size: number }>('/system/tenants');
      return response.data.tenants || [];
    },

    getById: async (id: string): Promise<Tenant> => {
      const response = await apiClient.get<Tenant>(`/system/tenants/${id}`);
      return response.data;
    },

    create: async (data: CreateTenantRequest): Promise<Tenant> => {
      const response = await apiClient.post<Tenant>('/system/tenants', data);
      return response.data;
    },

    update: async (id: string, data: Partial<CreateTenantRequest>): Promise<Tenant> => {
      const response = await apiClient.put<Tenant>(`/system/tenants/${id}`, data);
      return response.data;
    },

    delete: async (id: string): Promise<void> => {
      await apiClient.delete(`/system/tenants/${id}`);
    },

    suspend: async (id: string): Promise<Tenant> => {
      const response = await apiClient.post<Tenant>(`/system/tenants/${id}/suspend`, {});
      return response.data;
    },

    resume: async (id: string): Promise<Tenant> => {
      const response = await apiClient.post<Tenant>(`/system/tenants/${id}/resume`, {});
      return response.data;
    },

    // Tenant Settings Management (SYSTEM users only)
    getSettings: async (id: string): Promise<any> => {
      const response = await apiClient.get(`/system/tenants/${id}/settings`);
      return response.data;
    },

    updateSettings: async (id: string, data: any): Promise<any> => {
      const response = await apiClient.put(`/system/tenants/${id}/settings`, data);
      return response.data;
    },
  },
};

// User API
export const userApi = {
  list: async (tenantId?: string | null): Promise<User[]> => {
    // For SYSTEM users, tenantId is passed as query param or header
    // For TENANT users, tenant context is automatically set via middleware
    const config = tenantId ? { headers: { 'X-Tenant-ID': tenantId } } : undefined;
    const response = await apiClient.get<{ users: User[]; page: number; page_size: number; total: number }>(
      API_ENDPOINTS.USERS.BASE,
      config
    );
    return response.data.users || [];
  },

  listSystem: async (): Promise<User[]> => {
    // List system users (principal_type = 'SYSTEM')
    const response = await apiClient.get<{ users: User[]; page: number; page_size: number; total: number }>(
      API_ENDPOINTS.SYSTEM_USERS.BASE
    );
    return response.data.users || [];
  },

  createSystem: async (data: CreateUserRequest): Promise<User> => {
    // Create system user (no tenant required)
    const response = await apiClient.post<User>(
      API_ENDPOINTS.SYSTEM_USERS.BASE,
      data
    );
    return response.data;
  },

  getById: async (id: string, tenantId?: string): Promise<User> => {
    // For SYSTEM users, tenantId can be passed as query parameter if not set in header
    const config = tenantId ? { params: { tenant_id: tenantId } } : undefined;
    const response = await apiClient.get<User>(API_ENDPOINTS.USERS.BY_ID(id), config);
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

  getUserPermissions: async (userId: string): Promise<Permission[]> => {
    const response = await apiClient.get<{ permissions: Permission[] }>(
      `${API_ENDPOINTS.USERS.BASE}/${userId}/permissions`
    );
    return response.data.permissions || [];
  },
};

// Role API
export const roleApi = {
  list: async (tenantId?: string | null): Promise<Role[]> => {
    // For SYSTEM users, tenantId is passed as header
    // For TENANT users, tenant context is automatically set via middleware
    const config = tenantId ? { headers: { 'X-Tenant-ID': tenantId } } : undefined;
    const response = await apiClient.get<{ roles: Role[]; page: number; page_size: number; total: number }>(
      API_ENDPOINTS.ROLES.BASE,
      config
    );
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

  getUserRoles: async (userId: string): Promise<Role[]> => {
    const response = await apiClient.get<{ roles: Role[] }>(`${API_ENDPOINTS.USERS.BASE}/${userId}/roles`);
    return response.data.roles || [];
  },

  assignRoleToUser: async (userId: string, roleId: string, tenantId?: string): Promise<void> => {
    // For SYSTEM users, tenantId can be passed as header
    const config = tenantId ? { headers: { 'X-Tenant-ID': tenantId } } : undefined;
    await apiClient.post(
      `${API_ENDPOINTS.USERS.BASE}/${userId}/roles/${roleId}`,
      {},
      config
    );
  },

  removePermission: async (roleId: string, permissionId: string): Promise<void> => {
    await apiClient.delete(
      API_ENDPOINTS.ROLES.ASSIGN_PERMISSION(roleId, permissionId)
    );
  },

  listSystem: async (): Promise<Role[]> => {
    // List system roles (is_system = true)
    const response = await apiClient.get<{ roles: Role[]; page: number; page_size: number; total: number }>(
      API_ENDPOINTS.SYSTEM_ROLES.BASE
    );
    return response.data.roles || [];
  },
};

// Permission API
export const permissionApi = {
  list: async (tenantId?: string | null): Promise<Permission[]> => {
    // For SYSTEM users, tenantId is passed as header
    // For TENANT users, tenant context is automatically set via middleware
    const config = tenantId ? { headers: { 'X-Tenant-ID': tenantId } } : undefined;
    const response = await apiClient.get<{ permissions: Permission[]; page: number; page_size: number; total: number }>(
      API_ENDPOINTS.PERMISSIONS.BASE,
      config
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

  listSystem: async (): Promise<Permission[]> => {
    // List system permissions (predefined)
    const response = await apiClient.get<{ permissions: Permission[]; page: number; page_size: number; total: number }>(
      API_ENDPOINTS.SYSTEM_PERMISSIONS.BASE
    );
    return response.data.permissions || [];
  },
};

// MFA API
export const mfaApi = {
  enroll: async (): Promise<any> => {
    const response = await apiClient.post(API_ENDPOINTS.MFA.ENROLL);
    return response.data;
  },

  enrollForLogin: async (data: { session_id: string }): Promise<any> => {
    const response = await apiClient.post(`${API_PREFIX}/mfa/enroll/login`, data);
    return response.data;
  },

  verify: async (data: { code: string }): Promise<void> => {
    await apiClient.post(API_ENDPOINTS.MFA.VERIFY, data);
  },

  challenge: async (data: { user_id: string; tenant_id: string }): Promise<{ session_id: string; challenge_id?: string }> => {
    const response = await apiClient.post<{ session_id: string; challenge_id?: string }>(
      API_ENDPOINTS.MFA.CHALLENGE,
      data
    );
    // Backend returns session_id, but we also support challenge_id for compatibility
    return {
      session_id: response.data.session_id,
      challenge_id: response.data.session_id, // Use session_id as challenge_id for compatibility
    };
  },

  verifyChallenge: async (data: { challenge_id: string; code: string }): Promise<{ access_token: string }> => {
    const response = await apiClient.post<{ access_token: string }>(
      API_ENDPOINTS.MFA.VERIFY_CHALLENGE,
      data
    );
    return response.data;
  },
};

// Capability API
export const systemCapabilityApi = {
  list: async (): Promise<SystemCapability[]> => {
    // Check user type to determine which endpoint to use
    const principalType = localStorage.getItem('principal_type');

    const url = principalType === 'SYSTEM'
      ? API_ENDPOINTS.SYSTEM_CAPABILITIES.BASE
      : API_ENDPOINTS.SYSTEM_CAPABILITIES_READ.BASE;

    const response = await apiClient.get<{ capabilities: SystemCapability[] }>(url);
    return response.data.capabilities || [];
  },

  getByKey: async (key: string): Promise<SystemCapability> => {
    // Check user type to determine which endpoint to use
    const principalType = localStorage.getItem('principal_type');

    const url = principalType === 'SYSTEM'
      ? API_ENDPOINTS.SYSTEM_CAPABILITIES.BY_KEY(key)
      : API_ENDPOINTS.SYSTEM_CAPABILITIES_READ.BY_KEY(key);

    const response = await apiClient.get<SystemCapability>(url);
    return response.data;
  },

  update: async (key: string, data: UpdateSystemCapabilityRequest): Promise<SystemCapability> => {
    // Only SYSTEM users can update system capabilities
    const response = await apiClient.put<SystemCapability>(
      API_ENDPOINTS.SYSTEM_CAPABILITIES.BY_KEY(key),
      data
    );
    return response.data;
  },
};

// Tenant Capability API
export const tenantCapabilityApi = {
  list: async (tenantId: string): Promise<TenantCapability[]> => {
    // Check user type to determine which endpoint to use
    const principalType = localStorage.getItem('principal_type');

    let url: string;
    if (principalType === 'SYSTEM') {
      // SYSTEM users use the system endpoint
      url = API_ENDPOINTS.TENANT_CAPABILITIES.BASE(tenantId);
    } else {
      // TENANT users use the tenant-scoped endpoint (tenantId from context)
      url = API_ENDPOINTS.TENANT_CAPABILITIES_READ.BASE;
    }

    const response = await apiClient.get<{ capabilities: TenantCapability[] }>(url);
    const capabilities = response.data?.capabilities;
    return Array.isArray(capabilities) ? capabilities : [];
  },

  set: async (tenantId: string, key: string, data: SetTenantCapabilityRequest): Promise<any> => {
    const response = await apiClient.put(
      API_ENDPOINTS.TENANT_CAPABILITIES.BY_KEY(tenantId, key),
      data
    );
    return response.data;
  },

  delete: async (tenantId: string, key: string): Promise<void> => {
    await apiClient.delete(API_ENDPOINTS.TENANT_CAPABILITIES.BY_KEY(tenantId, key));
  },

  evaluate: async (tenantId: string, userId?: string): Promise<CapabilityEvaluation[]> => {
    // Check user type to determine which endpoint to use
    const principalType = localStorage.getItem('principal_type');

    // Validate tenantId for SYSTEM users
    if (principalType === 'SYSTEM' && !tenantId) {
      throw new Error('Tenant ID is required for SYSTEM users');
    }

    let url: string;
    if (principalType === 'SYSTEM') {
      // SYSTEM users use the system endpoint
      url = userId
        ? `${API_ENDPOINTS.TENANT_CAPABILITIES.EVALUATION(tenantId)}?user_id=${userId}`
        : API_ENDPOINTS.TENANT_CAPABILITIES.EVALUATION(tenantId);
    } else {
      // TENANT users use the tenant-scoped endpoint
      // Use direct string to avoid build/cache issues
      const baseUrl = `${API_PREFIX}/tenant/capabilities/evaluation`;
      url = userId ? `${baseUrl}?user_id=${userId}` : baseUrl;
    }

    console.log('[tenantCapabilityApi.evaluate] Calling:', url, { principalType, tenantId, userId });
    const response = await apiClient.get<{ evaluations: CapabilityEvaluation[] }>(url);
    console.log('[tenantCapabilityApi.evaluate] Response:', response.data);
    return response.data.evaluations || [];
  },
};

// Tenant Feature API (Tenant users)
export const tenantFeatureApi = {
  list: async (tenantId?: string): Promise<TenantFeature[]> => {
    // For SYSTEM users, add X-Tenant-ID header if tenantId is provided
    const config = tenantId ? { headers: { 'X-Tenant-ID': tenantId } } : undefined;
    const response = await apiClient.get<{ features: TenantFeature[] | null }>(
      API_ENDPOINTS.TENANT_FEATURES.BASE,
      config
    );
    const features = response.data?.features;
    // Handle both null and undefined, ensure we always return an array
    return Array.isArray(features) ? features : [];
  },

  enable: async (key: string, data?: EnableTenantFeatureRequest): Promise<any> => {
    const response = await apiClient.put(
      API_ENDPOINTS.TENANT_FEATURES.BY_KEY(key),
      data || {}
    );
    return response.data;
  },

  disable: async (key: string): Promise<void> => {
    await apiClient.delete(API_ENDPOINTS.TENANT_FEATURES.BY_KEY(key));
  },
};

// User Capability API (Tenant users)
export const userCapabilityApi = {
  list: async (userId: string): Promise<UserCapabilityState[]> => {
    const response = await apiClient.get<{ states: UserCapabilityState[] }>(
      API_ENDPOINTS.USER_CAPABILITIES.BASE(userId)
    );
    return response.data.states || [];
  },

  getByKey: async (userId: string, key: string): Promise<UserCapabilityState> => {
    const response = await apiClient.get<UserCapabilityState>(
      API_ENDPOINTS.USER_CAPABILITIES.BY_KEY(userId, key)
    );
    return response.data;
  },

  enroll: async (userId: string, key: string, data?: EnrollUserCapabilityRequest): Promise<any> => {
    const response = await apiClient.post(
      API_ENDPOINTS.USER_CAPABILITIES.ENROLL(userId, key),
      data || {}
    );
    return response.data;
  },

  unenroll: async (userId: string, key: string): Promise<void> => {
    await apiClient.delete(API_ENDPOINTS.USER_CAPABILITIES.BY_KEY(userId, key));
  },
};

// Export error handler
export { handleApiError };

