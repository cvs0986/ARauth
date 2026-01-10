/**
 * Principal Context - Single Source of Truth for Identity & Authority
 * 
 * GUARDRAIL #1: Backend Is Law
 * - All authority decisions come from JWT claims via authStore
 * - No invented security logic
 * - No assumptions about permissions
 * 
 * This context provides:
 * - Identity (userId, username, email, principalType)
 * - Authority (permissions, roles)
 * - Scope (tenantId, selectedTenantId)
 * - Computed properties (consoleMode, effectivePermissions)
 */

import React, { createContext, useContext, useMemo } from 'react';
import { useAuthStore } from '../store/authStore';
import type { PrincipalType } from '../store/authStore';

export type ConsoleMode = 'SYSTEM' | 'TENANT';

export interface Tenant {
    id: string;
    name: string;
    domain: string;
    status: string;
}

export interface PrincipalContextValue {
    // Identity
    userId: string | null;
    username: string | null;
    email: string | null;
    principalType: PrincipalType | null;

    // Authority
    systemPermissions: string[];
    tenantPermissions: string[];
    systemRoles: string[];
    tenantRoles: string[]; // Future: when we add tenant roles to JWT

    // Scope
    homeTenantId: string | null; // User's home tenant (TENANT users)
    selectedTenantId: string | null; // Currently selected tenant (SYSTEM users)
    availableTenants: Tenant[]; // Tenants SYSTEM user can manage (future: fetch from API)

    // Computed
    effectivePermissions: string[]; // Combined permissions for current context
    consoleMode: ConsoleMode; // SYSTEM or TENANT mode
    currentTenant: Tenant | null; // Currently active tenant

    // Actions
    selectTenant: (tenantId: string | null) => void;
    hasPermission: (permission: string) => boolean;
    hasSystemPermission: (permission: string) => boolean;
    canAccessRoute: (route: string) => boolean;

    // Auth state
    isAuthenticated: boolean;
}

const PrincipalContext = createContext<PrincipalContextValue | undefined>(undefined);

interface PrincipalProviderProps {
    children: React.ReactNode;
}

export function PrincipalProvider({ children }: PrincipalProviderProps) {
    const authState = useAuthStore();

    // Compute console mode
    // GUARDRAIL #1: This logic comes from backend JWT claims, not invented
    const consoleMode: ConsoleMode = useMemo(() => {
        if (authState.principalType === 'SYSTEM' && authState.selectedTenantId === null) {
            return 'SYSTEM';
        }
        return 'TENANT';
    }, [authState.principalType, authState.selectedTenantId]);

    // Compute effective permissions based on console mode
    const effectivePermissions = useMemo(() => {
        if (consoleMode === 'SYSTEM') {
            // In SYSTEM mode, use system permissions
            return authState.systemPermissions;
        } else {
            // In TENANT mode, use tenant permissions
            return authState.permissions;
        }
    }, [consoleMode, authState.systemPermissions, authState.permissions]);

    // Current tenant (stub for now - will fetch from API in future)
    // GUARDRAIL #4: Data gap - tenant details not available yet
    const currentTenant: Tenant | null = useMemo(() => {
        const tenantId = authState.getCurrentTenantId();
        if (!tenantId) return null;

        // TODO: Fetch tenant details from API
        // For now, return stub data
        return {
            id: tenantId,
            name: 'Loading...', // Will be replaced with API call
            domain: '',
            status: 'active',
        };
    }, [authState.getCurrentTenantId()]);

    // Available tenants (stub for now - will fetch from API in future)
    // GUARDRAIL #4: Data gap - available tenants list not available yet
    const availableTenants: Tenant[] = useMemo(() => {
        if (authState.principalType !== 'SYSTEM') return [];

        // TODO: Fetch available tenants from API
        // For now, return empty array
        return [];
    }, [authState.principalType]);

    const value: PrincipalContextValue = {
        // Identity
        userId: authState.userId,
        username: authState.username,
        email: authState.email,
        principalType: authState.principalType,

        // Authority
        systemPermissions: authState.systemPermissions,
        tenantPermissions: authState.permissions,
        systemRoles: authState.systemRoles,
        tenantRoles: [], // Future: extract from JWT

        // Scope
        homeTenantId: authState.tenantId,
        selectedTenantId: authState.selectedTenantId,
        availableTenants,

        // Computed
        effectivePermissions,
        consoleMode,
        currentTenant,

        // Actions
        selectTenant: authState.setSelectedTenantId,
        hasPermission: authState.hasPermission,
        hasSystemPermission: authState.hasSystemPermission,
        canAccessRoute: (_route: string) => {
            // TODO: Implement route permission mapping
            // For now, allow all routes if authenticated
            return authState.isAuthenticated;
        },

        // Auth state
        isAuthenticated: authState.isAuthenticated,
    };

    return (
        <PrincipalContext.Provider value={value}>
            {children}
        </PrincipalContext.Provider>
    );
}

/**
 * Hook to access Principal Context
 * 
 * GUARDRAIL #1: This is the ONLY way components should access user identity/authority
 * Components must NEVER read JWT directly or access authStore directly
 */
export function usePrincipalContext(): PrincipalContextValue {
    const context = useContext(PrincipalContext);
    if (context === undefined) {
        throw new Error('usePrincipalContext must be used within a PrincipalProvider');
    }
    return context;
}
