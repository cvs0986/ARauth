/**
 * JWT Token Decoder
 * Decodes JWT tokens to extract claims (principal_type, permissions, etc.)
 */

export interface JWTPayload {
  sub: string; // User ID
  tenant_id?: string;
  principal_type?: 'SYSTEM' | 'TENANT' | 'SERVICE';
  system_permissions?: string[];
  permissions?: string[];
  roles?: string[];
  system_roles?: string[];
  exp?: number;
  iat?: number;
  iss?: string;
  [key: string]: unknown;
}

/**
 * Decodes a JWT token without verification (client-side only)
 * Note: This does NOT verify the signature. Backend should always verify.
 */
export function decodeJWT(token: string): JWTPayload | null {
  try {
    const parts = token.split('.');
    if (parts.length !== 3) {
      return null;
    }

    // Decode the payload (second part)
    const payload = parts[1];
    const decoded = atob(payload.replace(/-/g, '+').replace(/_/g, '/'));
    return JSON.parse(decoded) as JWTPayload;
  } catch (error) {
    console.error('Failed to decode JWT:', error);
    return null;
  }
}

/**
 * Extracts user information from JWT token
 */
export function extractUserInfo(token: string): {
  userId: string | null;
  tenantId: string | null;
  principalType: 'SYSTEM' | 'TENANT' | 'SERVICE' | null;
  systemPermissions: string[];
  permissions: string[];
} {
  const payload = decodeJWT(token);
  
  if (!payload) {
    return {
      userId: null,
      tenantId: null,
      principalType: null,
      systemPermissions: [],
      permissions: [],
    };
  }

  return {
    userId: payload.sub || null,
    tenantId: payload.tenant_id || null,
    principalType: (payload.principal_type as 'SYSTEM' | 'TENANT' | 'SERVICE') || null,
    systemPermissions: payload.system_permissions || [],
    permissions: payload.permissions || [],
  };
}

