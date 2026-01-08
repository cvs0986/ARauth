/**
 * Sidebar Component
 * Shows different menu items for SYSTEM vs TENANT users
 */

import { Link, useLocation } from 'react-router-dom';
import { cn } from '@/lib/utils';
import { useAuthStore } from '@/store/authStore';

// System navigation items (for SYSTEM users)
const systemNavigation = [
  { name: 'Dashboard', href: '/', icon: '📊', permission: null },
  { name: 'Tenants', href: '/tenants', icon: '🏢', permission: 'tenant:read' },
  { name: 'Users', href: '/users', icon: '👤', permission: null }, // Shows all users across tenants
  { name: 'Roles', href: '/roles', icon: '🔑', permission: null }, // Shows all roles
  { name: 'Permissions', href: '/permissions', icon: '🛡️', permission: null }, // Shows all permissions
  { name: 'Audit Logs', href: '/audit', icon: '📋', permission: 'system:audit' },
  { name: 'Settings', href: '/settings', icon: '⚙️', permission: 'system:settings' },
];

// Tenant navigation items (for TENANT users)
const tenantNavigation = [
  { name: 'Dashboard', href: '/', icon: '📊', permission: null },
  { name: 'Users', href: '/users', icon: '👤', permission: 'users:read' },
  { name: 'Roles', href: '/roles', icon: '🔑', permission: 'roles:read' },
  { name: 'Permissions', href: '/permissions', icon: '🛡️', permission: 'permissions:read' },
  { name: 'Audit Logs', href: '/audit', icon: '📋', permission: null },
  { name: 'Settings', href: '/settings', icon: '⚙️', permission: null },
];

export function Sidebar() {
  const location = useLocation();
  const { isSystemUser, hasSystemPermission, hasPermission } = useAuthStore();

  // Select navigation based on user type
  const navigation = isSystemUser() ? systemNavigation : tenantNavigation;

  // Filter navigation items based on permissions
  const filteredNavigation = navigation.filter((item) => {
    if (!item.permission) return true; // No permission required
    
    if (isSystemUser()) {
      return hasSystemPermission(item.permission);
    } else {
      return hasPermission(item.permission);
    }
  });

  return (
    <aside className="w-64 bg-gray-900 text-white min-h-screen">
      <div className="p-4">
        <div className="mb-4 px-4 py-2 bg-gray-800 rounded-lg">
          <div className="text-xs text-gray-400 uppercase tracking-wider">
            {isSystemUser() ? 'System Admin' : 'Tenant Admin'}
          </div>
        </div>
        <nav className="space-y-2">
          {filteredNavigation.map((item) => {
            const isActive = location.pathname === item.href;
            return (
              <Link
                key={item.name}
                to={item.href}
                className={cn(
                  'flex items-center space-x-3 px-4 py-3 rounded-lg transition-colors',
                  isActive
                    ? 'bg-gray-800 text-white'
                    : 'text-gray-300 hover:bg-gray-800 hover:text-white'
                )}
              >
                <span>{item.icon}</span>
                <span>{item.name}</span>
              </Link>
            );
          })}
        </nav>
      </div>
    </aside>
  );
}

