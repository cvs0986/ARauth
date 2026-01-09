/**
 * Protected Route Component
 * Wraps routes that require authentication and admin access
 */

import { Navigate } from 'react-router-dom';
import { useAuthStore } from '../store/authStore';

interface ProtectedRouteProps {
  children: React.ReactNode;
}

export function ProtectedRoute({ children }: ProtectedRouteProps) {
  const { isAuthenticated, hasPermission, hasSystemPermission, isSystemUser } = useAuthStore();

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  // Check for admin access permission
  // SYSTEM users have admin access by default (they can access system admin dashboard)
  // TENANT users need tenant.admin.access permission
  const hasAdminAccess = isSystemUser() || hasPermission('tenant.admin.access');

  if (!hasAdminAccess) {
    return <Navigate to="/no-access" replace />;
  }

  return <>{children}</>;
}

