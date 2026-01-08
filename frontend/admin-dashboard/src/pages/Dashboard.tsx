/**
 * Dashboard Home Page
 */

import { useQuery } from '@tanstack/react-query';
import { tenantApi, userApi, roleApi, permissionApi } from '@/services/api';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { useNavigate } from 'react-router-dom';
import { Users, Building2, Shield, Key, TrendingUp, Activity } from 'lucide-react';

export function Dashboard() {
  const navigate = useNavigate();

  // Fetch statistics
  const { data: tenants, isLoading: tenantsLoading } = useQuery({
    queryKey: ['tenants'],
    queryFn: () => tenantApi.list(),
  });

  const { data: users, isLoading: usersLoading } = useQuery({
    queryKey: ['users'],
    queryFn: () => userApi.list(),
  });

  const { data: roles, isLoading: rolesLoading } = useQuery({
    queryKey: ['roles'],
    queryFn: () => roleApi.list(),
  });

  const { data: permissions, isLoading: permissionsLoading } = useQuery({
    queryKey: ['permissions'],
    queryFn: () => permissionApi.list(),
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

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Dashboard</h1>
          <p className="text-gray-600 mt-1">Overview of your ARauth Identity system</p>
        </div>
      </div>

      {/* Statistics Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Tenants</CardTitle>
            <Building2 className="h-4 w-4 text-muted-foreground" />
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
            <Button
              variant="outline"
              className="w-full justify-start"
              onClick={() => navigate('/tenants')}
            >
              <Building2 className="mr-2 h-4 w-4" />
              Manage Tenants
            </Button>
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
          </CardContent>
        </Card>

        {/* System Overview */}
        <Card>
          <CardHeader>
            <CardTitle>System Overview</CardTitle>
            <CardDescription>Current system status</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Activity className="h-4 w-4 text-green-500" />
                <span className="text-sm">System Status</span>
              </div>
              <span className="px-2 py-1 bg-green-100 text-green-800 rounded text-xs font-medium">
                Operational
              </span>
            </div>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <TrendingUp className="h-4 w-4 text-blue-500" />
                <span className="text-sm">Total Entities</span>
              </div>
              <span className="text-sm font-medium">
                {stats.tenants.total + stats.users.total + stats.roles.total + stats.permissions.total}
              </span>
            </div>
            <div className="pt-2 border-t">
              <p className="text-xs text-gray-600">
                Active Tenants: {stats.tenants.active} / {stats.tenants.total}
              </p>
              <p className="text-xs text-gray-600">
                Active Users: {stats.users.active} / {stats.users.total}
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

