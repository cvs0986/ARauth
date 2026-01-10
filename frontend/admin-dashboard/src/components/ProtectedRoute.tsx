/**
 * Protected Route Component
 * 
 * GUARDRAIL #1: Backend Is Law
 * - Route protection based on backend permissions only
 * 
 * GUARDRAIL #6: UI Quality Bar
 * - No disabled routes - hide or show based on permissions
 */

import { Navigate } from 'react-router-dom';
import { usePrincipalContext } from '../contexts/PrincipalContext';

interface ProtectedRouteProps {
  children: React.ReactNode;
  requiredPermission?: string;
  systemOnly?: boolean;
  fallback?: React.ReactElement;
}

export function ProtectedRoute({
  children,
  requiredPermission,
  systemOnly,
  fallback = <Navigate to="/no-access" replace />,
}: ProtectedRouteProps) {
  const {
    principalType,
    hasPermission,
    hasSystemPermission,
    isAuthenticated,
  } = usePrincipalContext();

  // Check authentication first
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  // Check if route is system-only
  if (systemOnly && principalType !== 'SYSTEM') {
    return fallback;
  }

  // Check required permission
  if (requiredPermission) {
    const hasAccess = principalType === 'SYSTEM'
      ? hasSystemPermission(requiredPermission)
      : hasPermission(requiredPermission);

    if (!hasAccess) {
      return fallback;
    }
  }

  // Default: check for admin access
  // SYSTEM users have admin access by default
  // TENANT users need tenant.admin.access permission
  if (!requiredPermission && !systemOnly) {
    const hasAdminAccess = principalType === 'SYSTEM' || hasPermission('tenant.admin.access');
    if (!hasAdminAccess) {
      return fallback;
    }
  }

  return <>{children}</>;
}
