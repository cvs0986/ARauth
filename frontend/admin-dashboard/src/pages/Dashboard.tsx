/**
 * Dashboard Home Page
 * Shows different statistics for SYSTEM vs TENANT users
 */

import { useQuery } from '@tanstack/react-query';
import { tenantApi, userApi, roleApi, permissionApi, systemApi } from '@/services/api';
import { useAuthStore } from '@/store/authStore';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { useNavigate } from 'react-router-dom';
import { Users, Building2, Shield, Key, TrendingUp, Activity, Globe, Server } from 'lucide-react';

export function Dashboard() {
  const navigate = useNavigate();
  const { isSystemUser, selectedTenantId, tenantId } = useAuthStore();

  // SYSTEM users see all tenants, TENANT users see their tenant
  const { data: tenants, isLoading: tenantsLoading } = useQuery({
    queryKey: isSystemUser() ? ['system', 'tenants'] : ['tenant', tenantId],
    queryFn: async () => {
      if (isSystemUser()) {
        return systemApi.tenants.list();
      } else {
        // TENANT users: fetch their own tenant
        if (tenantId) {
          const tenant = await tenantApi.getById(tenantId);
          return [tenant];
        }
        return [];
      }
    },
    enabled: !isSystemUser() ? !!tenantId : true,
  });

  // For SYSTEM users: show all users or filtered by selected tenant
  // For TENANT users: show only their tenant's users
  const currentTenantId = isSystemUser() ? selectedTenantId : tenantId;
  const { data: users, isLoading: usersLoading } = useQuery({
    queryKey: ['users', currentTenantId],
    queryFn: () => userApi.list(currentTenantId || undefined),
    enabled: isSystemUser() ? true : !!tenantId, // For SYSTEM users, allow even if no tenant selected (will show empty)
  });

  const { data: roles, isLoading: rolesLoading } = useQuery({
    queryKey: ['roles', currentTenantId],
    queryFn: () => roleApi.list(currentTenantId || undefined),
    enabled: isSystemUser() ? true : !!tenantId, // For SYSTEM users, allow even if no tenant selected (will show empty)
  });

  const { data: permissions, isLoading: permissionsLoading } = useQuery({
    queryKey: ['permissions', currentTenantId],
    queryFn: () => permissionApi.list(currentTenantId || undefined),
    enabled: isSystemUser() ? true : !!tenantId, // For SYSTEM users, allow even if no tenant selected (will show empty)
  });

  const isLoading = tenantsLoading || usersLoading || rolesLoading || permissionsLoading;

  // Calculate statistics - ensure data is always an array
  const stats = {
    tenants: {
      total: Array.isArray(tenants) ? tenants.length : 0,
      active: Array.isArray(tenants) ? tenants.filter((t) => t.status === 'active').length : 0,
    },
    users: {
      total: Array.isArray(users) ? users.length : 0,
      active: Array.isArray(users) ? users.filter((u) => u.status === 'active').length : 0,
    },
    roles: {
      total: Array.isArray(roles) ? roles.length : 0,
    },
    permissions: {
      total: Array.isArray(permissions) ? permissions.length : 0,
    },
  };

  if (isLoading) {
    return (
      <div className="space-y-4">
        <h1 className="text-3xl font-bold">Dashboard</h1>
        <div className="text-center py-8">Loading statistics...</div>
      </div>
    );
  }

  // Determine dashboard title and description
  const dashboardTitle = isSystemUser() ? 'System Dashboard' : 'Tenant Dashboard';
  const dashboardDescription = isSystemUser()
    ? 'Overview of all tenants and system-wide statistics'
    : 'Overview of your tenant and resources';

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">{dashboardTitle}</h1>
          <p className="text-gray-600 mt-1">{dashboardDescription}</p>
          {isSystemUser() && selectedTenantId && (
            <p className="text-sm text-blue-600 mt-1">
              Viewing data for selected tenant context
            </p>
          )}
        </div>
      </div>

      {/* Statistics Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Tenants Card - Only show for SYSTEM users */}
        {isSystemUser() && (
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Tenants</CardTitle>
              <Globe className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.tenants.total}</div>
              <p className="text-xs text-muted-foreground">
                {stats.tenants.active} active
              </p>
              <Button
                variant="link"
                className="p-0 h-auto mt-2"
                onClick={() => navigate('/tenants')}
              >
                View all →
              </Button>
            </CardContent>
          </Card>
        )}

        {/* Current Tenant Info - Only show for TENANT users */}
        {!isSystemUser() && tenants && tenants.length > 0 && (
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Tenant</CardTitle>
              <Building2 className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-lg font-bold">{tenants[0]?.name || 'N/A'}</div>
              <p className="text-xs text-muted-foreground">
                {tenants[0]?.domain || 'N/A'}
              </p>
              <p className="text-xs text-muted-foreground mt-1">
                Status: <span className="capitalize">{tenants[0]?.status || 'N/A'}</span>
              </p>
            </CardContent>
          </Card>
        )}

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Users</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.users.total}</div>
            <p className="text-xs text-muted-foreground">
              {stats.users.active} active
            </p>
            <Button
              variant="link"
              className="p-0 h-auto mt-2"
              onClick={() => navigate('/users')}
            >
              View all →
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Roles</CardTitle>
            <Shield className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.roles.total}</div>
            <p className="text-xs text-muted-foreground">Total roles</p>
            <Button
              variant="link"
              className="p-0 h-auto mt-2"
              onClick={() => navigate('/roles')}
            >
              View all →
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Permissions</CardTitle>
            <Key className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.permissions.total}</div>
            <p className="text-xs text-muted-foreground">Total permissions</p>
            <Button
              variant="link"
              className="p-0 h-auto mt-2"
              onClick={() => navigate('/permissions')}
            >
              View all →
            </Button>
          </CardContent>
        </Card>
      </div>

      {/* Quick Actions and Recent Activity */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Quick Actions */}
        <Card>
          <CardHeader>
            <CardTitle>Quick Actions</CardTitle>
            <CardDescription>Common management tasks</CardDescription>
          </CardHeader>
          <CardContent className="space-y-2">
            {isSystemUser() && (
              <Button
                variant="outline"
                className="w-full justify-start"
                onClick={() => navigate('/tenants')}
              >
                <Building2 className="mr-2 h-4 w-4" />
                Manage Tenants
              </Button>
            )}
            <Button
              variant="outline"
              className="w-full justify-start"
              onClick={() => navigate('/users')}
            >
              <Users className="mr-2 h-4 w-4" />
              Manage Users
            </Button>
            <Button
              variant="outline"
              className="w-full justify-start"
              onClick={() => navigate('/roles')}
            >
              <Shield className="mr-2 h-4 w-4" />
              Manage Roles
            </Button>
            <Button
              variant="outline"
              className="w-full justify-start"
              onClick={() => navigate('/permissions')}
            >
              <Key className="mr-2 h-4 w-4" />
              Manage Permissions
            </Button>
            {isSystemUser() && (
              <Button
                variant="outline"
                className="w-full justify-start"
                onClick={() => navigate('/settings')}
              >
                <Server className="mr-2 h-4 w-4" />
                System Settings
              </Button>
            )}
          </CardContent>
        </Card>

        {/* System/Tenant Overview */}
        <Card>
          <CardHeader>
            <CardTitle>{isSystemUser() ? 'System Overview' : 'Tenant Overview'}</CardTitle>
            <CardDescription>
              {isSystemUser() ? 'Current system status' : 'Current tenant status'}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Activity className="h-4 w-4 text-green-500" />
                <span className="text-sm">{isSystemUser() ? 'System Status' : 'Tenant Status'}</span>
              </div>
              <span className="px-2 py-1 bg-green-100 text-green-800 rounded text-xs font-medium">
                Operational
              </span>
            </div>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <TrendingUp className="h-4 w-4 text-blue-500" />
                <span className="text-sm">Total Resources</span>
              </div>
              <span className="text-sm font-medium">
                {isSystemUser()
                  ? stats.tenants.total + stats.users.total + stats.roles.total + stats.permissions.total
                  : stats.users.total + stats.roles.total + stats.permissions.total}
              </span>
            </div>
            <div className="pt-2 border-t">
              {isSystemUser() && (
                <p className="text-xs text-gray-600">
                  Active Tenants: {stats.tenants.active} / {stats.tenants.total}
                </p>
              )}
              <p className="text-xs text-gray-600">
                Active Users: {stats.users.active} / {stats.users.total}
              </p>
              <p className="text-xs text-gray-600">
                Total Roles: {stats.roles.total}
              </p>
              <p className="text-xs text-gray-600">
                Total Permissions: {stats.permissions.total}
              </p>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Recent Activity Placeholder */}
      <Card>
        <CardHeader>
          <CardTitle>Recent Activity</CardTitle>
          <CardDescription>Latest system events and changes</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="text-center py-8 text-gray-500">
            <Activity className="h-8 w-8 mx-auto mb-2 opacity-50" />
            <p className="text-sm">Recent activity will be displayed here</p>
            <p className="text-xs mt-1">Audit logs feature coming soon</p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

