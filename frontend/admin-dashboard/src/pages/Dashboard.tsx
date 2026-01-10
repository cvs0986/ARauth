/**
 * Dashboard Home Page
 * Shows different statistics for SYSTEM vs TENANT users
 */

import { useQuery } from '@tanstack/react-query';
import { tenantApi, userApi, roleApi, permissionApi, systemApi, systemCapabilityApi, tenantCapabilityApi, tenantFeatureApi, userCapabilityApi } from '@/services/api';
import { useAuthStore } from '@/store/authStore';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { useNavigate } from 'react-router-dom';
import { Users, Building2, Shield, Key, TrendingUp, Activity, Globe, Server, Info, Settings, Zap } from 'lucide-react';
import { CapabilityInheritanceVisualization } from '@/components/capabilities/CapabilityInheritanceVisualization';
import { StatCard } from '@/components/dashboard/StatCard';

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
  
  // For SYSTEM users: when "All Tenants" is selected (selectedTenantId === null), aggregate data
  // For TENANT users: always fetch (they have tenantId)
  const shouldFetchTenantData = isSystemUser() ? !!selectedTenantId : !!tenantId;
  const shouldAggregateAllTenants = isSystemUser() && !selectedTenantId;
  
  // Fetch all tenants for aggregation when "All Tenants" is selected
  const { data: allTenantsForAggregation } = useQuery({
    queryKey: ['system', 'tenants', 'all'],
    queryFn: () => systemApi.tenants.list(),
    enabled: shouldAggregateAllTenants,
  });

  // Aggregate users from all tenants when "All Tenants" is selected
  const { data: aggregatedUsers, isLoading: usersLoading } = useQuery({
    queryKey: ['users', 'all-tenants'],
    queryFn: async () => {
      if (!allTenantsForAggregation || allTenantsForAggregation.length === 0) return [];
      const allUsers = await Promise.all(
        allTenantsForAggregation.map((tenant) => userApi.list(tenant.id))
      );
      return allUsers.flat();
    },
    enabled: shouldAggregateAllTenants && !!allTenantsForAggregation,
  });

  // Aggregate roles from all tenants when "All Tenants" is selected
  const { data: aggregatedRoles, isLoading: rolesLoading } = useQuery({
    queryKey: ['roles', 'all-tenants'],
    queryFn: async () => {
      if (!allTenantsForAggregation || allTenantsForAggregation.length === 0) return [];
      const allRoles = await Promise.all(
        allTenantsForAggregation.map((tenant) => roleApi.list(tenant.id))
      );
      return allRoles.flat();
    },
    enabled: shouldAggregateAllTenants && !!allTenantsForAggregation,
  });

  // Aggregate permissions from all tenants when "All Tenants" is selected
  const { data: aggregatedPermissions, isLoading: permissionsLoading } = useQuery({
    queryKey: ['permissions', 'all-tenants'],
    queryFn: async () => {
      if (!allTenantsForAggregation || allTenantsForAggregation.length === 0) return [];
      const allPermissions = await Promise.all(
        allTenantsForAggregation.map((tenant) => permissionApi.list(tenant.id))
      );
      return allPermissions.flat();
    },
    enabled: shouldAggregateAllTenants && !!allTenantsForAggregation,
  });

  // Fetch single tenant data when a specific tenant is selected
  const { data: users, isLoading: usersLoadingSingle } = useQuery({
    queryKey: ['users', currentTenantId],
    queryFn: () => userApi.list(currentTenantId || undefined),
    enabled: shouldFetchTenantData && !shouldAggregateAllTenants,
  });

  const { data: roles, isLoading: rolesLoadingSingle } = useQuery({
    queryKey: ['roles', currentTenantId],
    queryFn: () => roleApi.list(currentTenantId || undefined),
    enabled: shouldFetchTenantData && !shouldAggregateAllTenants,
  });

  const { data: permissions, isLoading: permissionsLoadingSingle } = useQuery({
    queryKey: ['permissions', currentTenantId],
    queryFn: () => permissionApi.list(currentTenantId || undefined),
    enabled: shouldFetchTenantData && !shouldAggregateAllTenants,
  });

  // Capability metrics
  const { data: systemCapabilities } = useQuery({
    queryKey: ['system', 'capabilities'],
    queryFn: () => systemCapabilityApi.list(),
    enabled: isSystemUser(),
  });

  const { data: tenantCapabilities } = useQuery({
    queryKey: ['tenant', 'capabilities', currentTenantId],
    queryFn: () => tenantCapabilityApi.list(currentTenantId!),
    enabled: !!currentTenantId,
  });

  const { data: tenantFeatures } = useQuery({
    queryKey: ['tenant', 'features'],
    queryFn: () => tenantFeatureApi.list(),
    enabled: !isSystemUser() && !!tenantId,
  });

  // Get user capability enrollments count (for tenant users)
  const { data: allUsers } = useQuery({
    queryKey: ['users', currentTenantId],
    queryFn: () => userApi.list(currentTenantId || undefined),
    enabled: !!currentTenantId && !isSystemUser(),
  });

  // Count enrolled users (simplified - would need to fetch all user capabilities)
  const enrolledUsersCount = 0; // TODO: Calculate from user capabilities

  // Use aggregated data when "All Tenants" is selected, otherwise use single tenant data
  const finalUsers = shouldAggregateAllTenants ? aggregatedUsers : users;
  const finalRoles = shouldAggregateAllTenants ? aggregatedRoles : roles;
  const finalPermissions = shouldAggregateAllTenants ? aggregatedPermissions : permissions;
  const finalUsersLoading = shouldAggregateAllTenants ? usersLoading : usersLoadingSingle;
  const finalRolesLoading = shouldAggregateAllTenants ? rolesLoading : rolesLoadingSingle;
  const finalPermissionsLoading = shouldAggregateAllTenants ? permissionsLoading : permissionsLoadingSingle;

  const isLoading = tenantsLoading || finalUsersLoading || finalRolesLoading || finalPermissionsLoading;

  // Calculate statistics - ensure data is always an array
  const stats = {
    tenants: {
      total: Array.isArray(tenants) ? tenants.length : 0,
      active: Array.isArray(tenants) ? tenants.filter((t) => t.status === 'active').length : 0,
    },
    users: {
      total: Array.isArray(finalUsers) ? finalUsers.length : 0,
      active: Array.isArray(finalUsers) ? finalUsers.filter((u) => u.status === 'active').length : 0,
    },
    roles: {
      total: Array.isArray(finalRoles) ? finalRoles.length : 0,
    },
    permissions: {
      total: Array.isArray(finalPermissions) ? finalPermissions.length : 0,
    },
    capabilities: {
      systemTotal: Array.isArray(systemCapabilities) ? systemCapabilities.length : 0,
      systemEnabled: Array.isArray(systemCapabilities) ? systemCapabilities.filter((c) => c.enabled).length : 0,
      tenantTotal: Array.isArray(tenantCapabilities) ? tenantCapabilities.length : 0,
      tenantEnabled: Array.isArray(tenantCapabilities) ? tenantCapabilities.filter((c) => c.enabled).length : 0,
      featuresEnabled: Array.isArray(tenantFeatures) ? tenantFeatures.filter((f) => f.enabled).length : 0,
      usersEnrolled: enrolledUsersCount,
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
  const dashboardTitle = isSystemUser() 
    ? (shouldAggregateAllTenants ? 'System Dashboard - All Tenants' : 'System Dashboard')
    : 'Tenant Dashboard';
  const dashboardDescription = isSystemUser()
    ? (shouldAggregateAllTenants 
        ? 'Aggregated overview of all tenants and their resources'
        : 'Overview of all tenants and system-wide statistics')
    : 'Overview of your tenant and resources';

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">{dashboardTitle}</h1>
          <p className="text-gray-600 mt-1">{dashboardDescription}</p>
          {isSystemUser() && !selectedTenantId && (
            <Alert className="mt-4 bg-blue-50 border-blue-200 text-blue-800">
              <Info className="h-4 w-4 mr-2" />
              <AlertTitle>All Tenants View</AlertTitle>
              <AlertDescription>
                Showing aggregated statistics across all tenants. Select a specific tenant from the header dropdown to view tenant-specific data.
              </AlertDescription>
            </Alert>
          )}
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
          <StatCard
            title="Tenants"
            value={stats.tenants.total}
            icon={Building2}
            variant="primary"
            trend={{
              value: stats.tenants.active > 0 ? Math.round((stats.tenants.active / stats.tenants.total) * 100) : 0,
              label: 'active'
            }}
            onClick={() => navigate('/tenants')}
          />
        )}

        {/* Current Tenant Info - Only show for TENANT users */}
        {!isSystemUser() && tenants && tenants.length > 0 && (
          <StatCard
            title="Tenant"
            value={tenants[0]?.name || 'N/A'}
            icon={Building2}
            variant="primary"
          />
        )}

        <StatCard
          title="Users"
          value={stats.users.total}
          icon={Users}
          variant="success"
          trend={{
            value: stats.users.active > 0 ? Math.round((stats.users.active / stats.users.total) * 100) : 0,
            label: 'active'
          }}
          onClick={() => navigate('/users')}
        />

        <StatCard
          title="Roles"
          value={stats.roles.total}
          icon={Shield}
          variant="default"
          onClick={() => navigate('/roles')}
        />

        <StatCard
          title="Permissions"
          value={stats.permissions.total}
          icon={Key}
          variant="default"
          onClick={() => navigate('/permissions')}
        />

        {/* Capability Metrics */}
        {isSystemUser() && (
          <StatCard
            title="System Capabilities"
            value={stats.capabilities.systemTotal}
            icon={Settings}
            variant="warning"
            trend={{
              value: stats.capabilities.systemEnabled > 0 ? Math.round((stats.capabilities.systemEnabled / stats.capabilities.systemTotal) * 100) : 0,
              label: 'enabled'
            }}
            onClick={() => navigate('/capabilities/system')}
          />
        )}

        {currentTenantId && (
          <StatCard
            title="Tenant Capabilities"
            value={stats.capabilities.tenantTotal}
            icon={Zap}
            variant="warning"
            trend={{
              value: stats.capabilities.tenantEnabled > 0 ? Math.round((stats.capabilities.tenantEnabled / stats.capabilities.tenantTotal) * 100) : 0,
              label: 'enabled'
            }}
            onClick={isSystemUser() ? () => navigate('/capabilities/tenant-assignment') : undefined}
          />
        )}

        {!isSystemUser() && (
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Enabled Features</CardTitle>
              <Settings className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.capabilities.featuresEnabled}</div>
              <p className="text-xs text-muted-foreground">Active features</p>
              <Button
                variant="link"
                className="p-0 h-auto mt-2"
                onClick={() => navigate('/capabilities/features')}
              >
                View all →
              </Button>
            </CardContent>
          </Card>
        )}
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

      {/* Capability Visualization */}
      {systemCapabilities && systemCapabilities.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Capability Inheritance</CardTitle>
            <CardDescription>
              Visualize how capabilities flow from System → Tenant → User
            </CardDescription>
          </CardHeader>
          <CardContent>
            <CapabilityInheritanceVisualization
              capabilityKey={systemCapabilities[0].capability_key}
              tenantId={currentTenantId || undefined}
            />
          </CardContent>
        </Card>
      )}

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

