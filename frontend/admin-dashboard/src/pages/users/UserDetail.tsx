/**
 * User Detail Page with Tabbed Navigation
 */

import { useParams, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { userApi, roleApi, permissionApi, userCapabilityApi } from '@/services/api';
import { useAuthStore } from '@/store/authStore';
import { ArrowLeft, User, Shield, Key, Settings, Activity, Mail, Calendar, CheckCircle2, XCircle } from 'lucide-react';
import { StatCard } from '@/components/dashboard/StatCard';
import { EditUserDialog } from './EditUserDialog';
import { useState } from 'react';

export function UserDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { isSystemUser, selectedTenantId, tenantId, getCurrentTenantId } = useAuthStore();
  const currentTenantId = getCurrentTenantId();
  const [editUserOpen, setEditUserOpen] = useState(false);
  const [activeTab, setActiveTab] = useState('overview');

  // Fetch user details
  const { data: user, isLoading: userLoading } = useQuery({
    queryKey: ['user', id, currentTenantId],
    queryFn: () => userApi.getById(id!, currentTenantId || undefined),
    enabled: !!id,
  });

  // Fetch user roles
  const { data: userRoles, isLoading: rolesLoading } = useQuery({
    queryKey: ['user', id, 'roles', currentTenantId],
    queryFn: () => roleApi.getUserRoles(id!),
    enabled: !!id, // Enable for both system and tenant users
  });

  // Fetch all roles for assignment
  const { data: allRoles } = useQuery({
    queryKey: ['roles', currentTenantId],
    queryFn: () => roleApi.list(currentTenantId || undefined),
    enabled: !!id && !!currentTenantId, // Only for tenant users
  });

  // Fetch user permissions
  const { data: userPermissions } = useQuery({
    queryKey: ['user', id, 'permissions', currentTenantId],
    queryFn: () => userApi.getUserPermissions(id!),
    enabled: !!id, // Enable for both system and tenant users
  });

  // Fetch user capabilities
  const { data: userCapabilities } = useQuery({
    queryKey: ['user', id, 'capabilities', currentTenantId],
    queryFn: () => userCapabilityApi.list(id!),
    enabled: !!id, // Enable for both system and tenant users
  });

  if (userLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-gray-500">Loading user details...</div>
      </div>
    );
  }

  if (!user) {
    return (
      <div className="flex flex-col items-center justify-center h-64">
        <p className="text-gray-500 mb-4">User not found</p>
        <Button onClick={() => navigate('/users')}>Back to Users</Button>
      </div>
    );
  }

  const stats = [
    {
      title: 'Assigned Roles',
      value: userRoles?.length || 0,
      icon: Shield,
      variant: 'primary' as const,
      loading: rolesLoading,
    },
    {
      title: 'Enrolled Capabilities',
      value: userCapabilities?.filter(c => c.enrolled).length || 0,
      icon: Key,
      variant: 'success' as const,
    },
    {
      title: 'MFA Status',
      value: user.mfa_enabled ? 'Enabled' : 'Disabled',
      icon: user.mfa_enabled ? CheckCircle2 : XCircle,
      variant: user.mfa_enabled ? 'success' as const : 'default' as const,
    },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Button
            variant="outline"
            size="icon"
            onClick={() => navigate('/users')}
          >
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <div>
            <h1 className="text-3xl font-bold text-gray-900 flex items-center gap-3">
              <User className="h-8 w-8 text-primary-600" />
              {user.first_name || user.last_name 
                ? `${user.first_name || ''} ${user.last_name || ''}`.trim()
                : user.username}
            </h1>
            <p className="text-gray-600 mt-1">{user.email}</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Badge variant={user.status === 'active' ? 'default' : 'secondary'}>
            {user.status}
          </Badge>
          <Button variant="outline" onClick={() => setEditUserOpen(true)}>Edit User</Button>
        </div>
      </div>

      {/* Statistics Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {stats.map((stat, index) => (
          <StatCard key={index} {...stat} />
        ))}
      </div>

      {/* Tabbed Content */}
      <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-4">
        <TabsList className="bg-gray-100">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="roles">Roles</TabsTrigger>
          <TabsTrigger value="permissions">Permissions</TabsTrigger>
          <TabsTrigger value="capabilities">Capabilities</TabsTrigger>
          <TabsTrigger value="activity">Activity</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Card>
              <CardHeader>
                <CardTitle>User Information</CardTitle>
                <CardDescription>Basic user details</CardDescription>
              </CardHeader>
              <CardContent className="space-y-3">
                <div>
                  <p className="text-sm text-gray-600 flex items-center gap-2">
                    <Mail className="h-4 w-4" />
                    Email
                  </p>
                  <p className="text-sm">{user.email}</p>
                </div>
                <div>
                  <p className="text-sm text-gray-600 flex items-center gap-2">
                    <User className="h-4 w-4" />
                    Username
                  </p>
                  <p className="text-sm">{user.username}</p>
                </div>
                <div>
                  <p className="text-sm text-gray-600 flex items-center gap-2">
                    <Shield className="h-4 w-4" />
                    Principal Type
                  </p>
                  <Badge>{user.principal_type}</Badge>
                </div>
                <div>
                  <p className="text-sm text-gray-600 flex items-center gap-2">
                    <CheckCircle2 className="h-4 w-4" />
                    MFA Status
                  </p>
                  <Badge variant={user.mfa_enabled ? 'default' : 'secondary'}>
                    {user.mfa_enabled ? 'Enabled' : 'Disabled'}
                  </Badge>
                </div>
                <div>
                  <p className="text-sm text-gray-600 flex items-center gap-2">
                    <Calendar className="h-4 w-4" />
                    Created
                  </p>
                  <p className="text-sm">{new Date(user.created_at).toLocaleDateString()}</p>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Quick Actions</CardTitle>
                <CardDescription>Common operations for this user</CardDescription>
              </CardHeader>
              <CardContent className="space-y-2">
                <Button 
                  className="w-full justify-start" 
                  variant="outline"
                  onClick={() => setActiveTab('roles')}
                >
                  <Shield className="h-4 w-4 mr-2" />
                  Assign Role
                </Button>
                <Button 
                  className="w-full justify-start" 
                  variant="outline"
                  onClick={() => setActiveTab('capabilities')}
                >
                  <Key className="h-4 w-4 mr-2" />
                  Enroll in Capability
                </Button>
                <Button 
                  className="w-full justify-start" 
                  variant="outline"
                  onClick={() => {
                    // TODO: Implement MFA configuration
                    alert('MFA configuration coming soon');
                  }}
                >
                  <Settings className="h-4 w-4 mr-2" />
                  Configure MFA
                </Button>
                <Button 
                  className="w-full justify-start" 
                  variant="outline"
                  onClick={() => setActiveTab('activity')}
                >
                  <Activity className="h-4 w-4 mr-2" />
                  View Activity
                </Button>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="roles">
          <Card>
            <CardHeader>
              <CardTitle>Assigned Roles</CardTitle>
              <CardDescription>Roles assigned to this user</CardDescription>
            </CardHeader>
            <CardContent>
              {rolesLoading ? (
                <p className="text-gray-500">Loading roles...</p>
              ) : userRoles && userRoles.length > 0 ? (
                <div className="space-y-2">
                  {userRoles.map((role) => (
                    <div key={role.id} className="flex items-center justify-between p-3 border rounded-lg">
                      <div>
                        <p className="font-medium">{role.name}</p>
                        <p className="text-sm text-gray-600">{role.description || 'No description'}</p>
                      </div>
                      <Button variant="outline" size="sm">Remove</Button>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-gray-500">No roles assigned</p>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="permissions">
          <Card>
            <CardHeader>
              <CardTitle>Effective Permissions</CardTitle>
              <CardDescription>Permissions granted through roles</CardDescription>
            </CardHeader>
            <CardContent>
              {userPermissions && userPermissions.length > 0 ? (
                <div className="space-y-2">
                  {userPermissions.map((perm: any, index: number) => (
                    <div key={index} className="flex items-center justify-between p-3 border rounded-lg">
                      <div>
                        <p className="font-medium">{perm.permission || `${perm.resource}:${perm.action}`}</p>
                        <p className="text-sm text-gray-600">
                          {perm.description || `${perm.resource} ${perm.action}`}
                        </p>
                      </div>
                      <Badge variant="default">Active</Badge>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-gray-500">No permissions assigned</p>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="capabilities">
          <Card>
            <CardHeader>
              <CardTitle>Capability Enrollment</CardTitle>
              <CardDescription>Capabilities enrolled by this user</CardDescription>
            </CardHeader>
            <CardContent>
              {userCapabilities && userCapabilities.length > 0 ? (
                <div className="space-y-2">
                  {userCapabilities.map((cap) => (
                    <div key={cap.capability_key} className="flex items-center justify-between p-3 border rounded-lg">
                      <div>
                        <p className="font-medium">{cap.capability_key}</p>
                        <p className="text-sm text-gray-600">
                          {cap.enrolled ? 'Enrolled' : 'Not Enrolled'}
                        </p>
                      </div>
                      <Badge variant={cap.enrolled ? 'default' : 'secondary'}>
                        {cap.enrolled ? 'Active' : 'Inactive'}
                      </Badge>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-gray-500">No capabilities enrolled</p>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="activity">
          <Card>
            <CardHeader>
              <CardTitle>Activity Log</CardTitle>
              <CardDescription>Recent activity for this user</CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-gray-500">Activity logs coming soon...</p>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {user && (
        <EditUserDialog
          user={user}
          open={editUserOpen}
          onOpenChange={setEditUserOpen}
        />
      )}
    </div>
  );
}

