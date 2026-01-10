/**
 * Permission Gate Component
 * 
 * GUARDRAIL #1: Backend Is Law
 * - Permission checks based on backend JWT claims only
 * 
 * GUARDRAIL #6: UI Quality Bar
 * - Conditionally render based on permissions (no disabled buttons)
 */

import { usePrincipalContext } from '../contexts/PrincipalContext';

interface PermissionGateProps {
    permission: string;
    systemPermission?: boolean;
    fallback?: React.ReactNode;
    children: React.ReactNode;
}

/**
 * Conditionally renders children based on permission check
 * 
 * Usage:
 * <PermissionGate permission="users:create">
 *   <Button>Create User</Button>
 * </PermissionGate>
 */
export function PermissionGate({
    permission,
    systemPermission = false,
    fallback = null,
    children,
}: PermissionGateProps) {
    const { hasPermission, hasSystemPermission } = usePrincipalContext();

    const hasAccess = systemPermission
        ? hasSystemPermission(permission)
        : hasPermission(permission);

    return hasAccess ? <>{children}</> : <>{fallback}</>;
}
