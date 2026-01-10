/**
 * Sidebar Component
 * 
 * GUARDRAIL #1: Backend Is Law
 * - Permission filtering based on PrincipalContext
 * 
 * GUARDRAIL #6: UI Quality Bar
 * - Professional IAM vendor aesthetic
 * - Grouped navigation with clear hierarchy
 * - No disabled items - hide what user can't access
 */

import { Link, useLocation } from 'react-router-dom';
import { cn } from '@/lib/utils';
import { usePrincipalContext } from '@/contexts/PrincipalContext';
import { systemNavigation, tenantNavigation, type NavigationGroup } from '@/config/navigation';

export function Sidebar() {
  const location = useLocation();
  const { consoleMode, hasPermission, hasSystemPermission } = usePrincipalContext();

  // Select navigation based on console mode
  const navigationGroups: NavigationGroup[] =
    consoleMode === 'SYSTEM' ? systemNavigation : tenantNavigation;

  // Filter navigation items based on permissions
  const filteredGroups = navigationGroups.map(group => ({
    ...group,
    items: group.items.filter(item => {
      // No permission required - always show
      if (!item.permission) return true;

      // Check permission based on console mode
      if (consoleMode === 'SYSTEM') {
        return hasSystemPermission(item.permission);
      } else {
        return hasPermission(item.permission);
      }
    }),
  })).filter(group => group.items.length > 0); // Remove empty groups

  return (
    <aside className="w-64 bg-gradient-to-b from-gray-900 to-gray-800 text-white min-h-screen shadow-xl border-r border-gray-700">
      <div className="p-4 space-y-6">
        {filteredGroups.map((group, groupIndex) => (
          <div key={group.name}>
            {/* Group separator (except for first group) */}
            {groupIndex > 0 && (
              <div className="border-t border-gray-700 mb-4" />
            )}

            {/* Group label */}
            <div className="px-3 mb-2">
              <h3 className="text-xs font-semibold text-gray-400 uppercase tracking-wider">
                {group.name}
              </h3>
            </div>

            {/* Group items */}
            <nav className="space-y-1">
              {group.items.map((item) => {
                const Icon = item.icon;
                const isActive = location.pathname === item.href ||
                  (item.href !== '/' && location.pathname.startsWith(item.href));

                return (
                  <Link
                    key={item.name}
                    to={item.href}
                    className={cn(
                      'flex items-center justify-between px-3 py-2.5 rounded-lg transition-all duration-200 group',
                      isActive
                        ? 'bg-gradient-to-r from-blue-600 to-blue-700 text-white shadow-lg'
                        : 'text-gray-300 hover:bg-gray-800 hover:text-white'
                    )}
                  >
                    <div className="flex items-center gap-3">
                      <Icon className={cn(
                        'h-5 w-5',
                        isActive ? 'text-white' : 'text-gray-400 group-hover:text-white'
                      )} />
                      <span className="font-medium text-sm">{item.name}</span>
                    </div>

                    {/* Badge for coming soon features */}
                    {item.badge && (
                      <span className="px-2 py-0.5 text-xs font-medium bg-gray-700 text-gray-300 rounded">
                        {item.badge}
                      </span>
                    )}
                  </Link>
                );
              })}
            </nav>
          </div>
        ))}
      </div>
    </aside>
  );
}
