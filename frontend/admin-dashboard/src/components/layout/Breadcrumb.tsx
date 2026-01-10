/**
 * Breadcrumb Component for Navigation Hierarchy
 */

import { Link, useLocation } from 'react-router-dom';
import { ChevronRight, Home } from 'lucide-react';
import { useAuthStore } from '@/store/authStore';
import { cn } from '@/lib/utils';

interface BreadcrumbItem {
  label: string;
  href?: string;
}

export function Breadcrumb() {
  const location = useLocation();
  const { isSystemUser, selectedTenantId, tenantId } = useAuthStore();

  // Parse the current path to build breadcrumbs
  const buildBreadcrumbs = (): BreadcrumbItem[] => {
    const path = location.pathname;
    const segments = path.split('/').filter(Boolean);
    const items: BreadcrumbItem[] = [];

    // Always start with home
    items.push({ label: 'Dashboard', href: '/' });

    if (segments.length === 0) {
      return items;
    }

    // System level navigation
    if (isSystemUser()) {
      if (segments[0] === 'tenants' && segments[1]) {
        // System > Tenants > [Tenant Name]
        items.push({ label: 'Tenants', href: '/tenants' });
        if (segments[2]) {
          // System > Tenants > [Tenant] > [Resource]
          items.push({ label: 'Tenant Details', href: `/tenants/${segments[1]}` });
          if (segments[2] === 'users' && segments[3]) {
            items.push({ label: 'Users', href: `/tenants/${segments[1]}/users` });
            items.push({ label: segments[3], href: `/tenants/${segments[1]}/users/${segments[3]}` });
          } else if (segments[2]) {
            items.push({ label: segments[2].charAt(0).toUpperCase() + segments[2].slice(1) });
          }
        } else {
          items.push({ label: 'Tenant Details' });
        }
      } else if (segments[0] === 'users' && segments[1]) {
        // System > System Users > [User]
        items.push({ label: 'System Users', href: '/users' });
        items.push({ label: 'User Details' });
      } else if (segments[0]) {
        // System > [Resource]
        const resourceName = segments[0].charAt(0).toUpperCase() + segments[0].slice(1).replace(/-/g, ' ');
        items.push({ label: resourceName });
      }
    } else {
      // Tenant level navigation
      if (segments[0] === 'users' && segments[1]) {
        // Tenant > Users > [User]
        items.push({ label: 'Users', href: '/users' });
        items.push({ label: 'User Details' });
      } else if (segments[0]) {
        // Tenant > [Resource]
        const resourceName = segments[0].charAt(0).toUpperCase() + segments[0].slice(1).replace(/-/g, ' ');
        items.push({ label: resourceName });
      }
    }

    return items;
  };

  const breadcrumbs = buildBreadcrumbs();

  if (breadcrumbs.length <= 1) {
    return null; // Don't show breadcrumb if we're at the root
  }

  return (
    <nav className="flex items-center space-x-2 text-sm text-gray-600 mb-4" aria-label="Breadcrumb">
      {breadcrumbs.map((item, index) => {
        const isLast = index === breadcrumbs.length - 1;
        
        return (
          <div key={index} className="flex items-center">
            {index === 0 ? (
              <Link
                to={item.href || '#'}
                className={cn(
                  "flex items-center hover:text-primary-600 transition-colors",
                  isLast && "text-gray-900 font-medium"
                )}
              >
                <Home className="h-4 w-4" />
              </Link>
            ) : (
              <>
                <ChevronRight className="h-4 w-4 mx-2 text-gray-400" />
                {item.href && !isLast ? (
                  <Link
                    to={item.href}
                    className="hover:text-primary-600 transition-colors"
                  >
                    {item.label}
                  </Link>
                ) : (
                  <span className={cn("text-gray-900", isLast && "font-medium")}>
                    {item.label}
                  </span>
                )}
              </>
            )}
          </div>
        );
      })}
    </nav>
  );
}


