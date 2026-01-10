/**
 * Sidebar Component
 * Shows different menu items for SYSTEM vs TENANT users
 */

import { Link, useLocation } from 'react-router-dom';
import { cn } from '@/lib/utils';
import { useAuthStore } from '@/store/authStore';

// System navigation items (for SYSTEM users)
const systemNavigation = [
  { name: 'Dashboard', href: '/', icon: '📊', permission: 'admin:access' },
  { name: 'Tenants', href: '/tenants', icon: '🏢', permission: 'tenant:read' },
  { name: 'Users', href: '/users', icon: '👤', permission: 'users:read' },
  { name: 'Roles', href: '/roles', icon: '🔑', permission: 'roles:read' },
  { name: 'Permissions', href: '/permissions', icon: '🛡️', permission: 'permissions:read' },
  { name: 'System Capabilities', href: '/capabilities/system', icon: '🛠️', permission: 'system:configure' },
  { name: 'Tenant Capabilities', href: '/capabilities/tenant-assignment', icon: '🔧', permission: 'tenant:configure' },
  { name: 'MFA', href: '/mfa', icon: '🔐', permission: 'admin:access' }, // MFA management - requires admin access
  { name: 'Audit Logs', href: '/audit', icon: '📋', permission: 'audit:read' },
  { name: 'Settings', href: '/settings', icon: '⚙️', permission: 'tenant:settings:read' },
];

// Tenant navigation items (for TENANT users)
// Using tenant.* namespace for all tenant permissions
const tenantNavigation = [
  { name: 'Dashboard', href: '/', icon: '📊', permission: 'tenant.admin.access' },
  { name: 'Users', href: '/users', icon: '👤', permission: 'tenant.users.read' },
  { name: 'Roles', href: '/roles', icon: '🔑', permission: 'tenant.roles.read' },
  { name: 'Permissions', href: '/permissions', icon: '🛡️', permission: 'tenant.permissions.read' },
  { name: 'Features', href: '/capabilities/features', icon: '✨', permission: 'tenant.admin.access' },
  { name: 'User Capabilities', href: '/capabilities/user-enrollment', icon: '👥', permission: 'tenant.admin.access' },
  { name: 'MFA', href: '/mfa', icon: '🔐', permission: 'tenant.admin.access' }, // MFA management - requires admin access
  { name: 'Audit Logs', href: '/audit', icon: '📋', permission: 'tenant.audit.read' },
  { name: 'Settings', href: '/settings', icon: '⚙️', permission: 'tenant.settings.read' },
];

export function Sidebar() {
  const location = useLocation();
  const { isSystemUser, hasSystemPermission, hasPermission } = useAuthStore();

  // Select navigation based on user type
  const navigation = isSystemUser() ? systemNavigation : tenantNavigation;

  // Filter navigation items based on permissions
  // Core features (permission: null) are always visible
  // Features with specific permissions are filtered based on user permissions
  const filteredNavigation = navigation.filter((item) => {
    if (!item.permission) return true; // No permission required - always show
    
    // Only filter items that have specific permission requirements
    if (isSystemUser()) {
      return hasSystemPermission(item.permission);
    } else {
      return hasPermission(item.permission);
    }
  });

  return (
    <aside className="w-64 bg-gradient-to-b from-gray-900 to-gray-800 text-white min-h-screen shadow-xl">
      <div className="p-4">
        <nav className="space-y-1">
          {filteredNavigation.map((item) => {
            const isActive = location.pathname === item.href || 
              (item.href !== '/' && location.pathname.startsWith(item.href));
            return (
              <Link
                key={item.name}
                to={item.href}
                className={cn(
                  'flex items-center space-x-3 px-4 py-3 rounded-lg transition-all duration-200',
                  isActive
                    ? 'bg-gradient-to-r from-primary-600 to-primary-700 text-white shadow-md'
                    : 'text-gray-300 hover:bg-gray-800 hover:text-white hover:translate-x-1'
                )}
              >
                <span className="text-lg">{item.icon}</span>
                <span className="font-medium">{item.name}</span>
              </Link>
            );
          })}
        </nav>
      </div>
    </aside>
  );
}

