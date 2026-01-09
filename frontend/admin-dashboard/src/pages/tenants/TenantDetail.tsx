/**
 * Tenant Detail Page with Tabbed Navigation
 */

import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useState, useEffect } from 'react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { systemApi, userApi, roleApi, permissionApi, tenantCapabilityApi, tenantFeatureApi } from '@/services/api';
import { useAuthStore } from '@/store/authStore';
import { ArrowLeft, Building2, Users, Key, Shield, Settings, Activity, Zap } from 'lucide-react';
import { StatCard } from '@/components/dashboard/StatCard';
import { UserList } from '../users/UserList';
import { RoleList } from '../roles/RoleList';
import { PermissionList } from '../permissions/PermissionList';
import { TenantCapabilityAssignment } from '../capabilities/TenantCapabilityAssignment';
import { TenantFeatureEnablement } from '../capabilities/TenantFeatureEnablement';
import { CreateUserDialog } from '../users/CreateUserDialog';
import { CreateRoleDialog } from '../roles/CreateRoleDialog';
import { EditTenantDialog } from './EditTenantDialog';
import { TenantSettingsForm } from './TenantSettingsForm';

export function TenantDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { isSystemUser, setSelectedTenantId } = useAuthStore();
  const [createUserOpen, setCreateUserOpen] = useState(false);
  const [createRoleOpen, setCreateRoleOpen] = useState(false);
  const [editTenantOpen, setEditTenantOpen] = useState(false);
  const [activeTab, setActiveTab] = useState('overview');

  // Set selected tenant ID for SYSTEM users when component mounts
  useEffect(() => {
    if (isSystemUser() && id) {
      setSelectedTenantId(id);
    }
  }, [id, isSystemUser, setSelectedTenantId]);

  // Fetch tenant details
  const { data: tenant, isLoading: tenantLoading } = useQuery({
    queryKey: ['tenant', id],
    queryFn: () => systemApi.tenants.getById(id!),
    enabled: !!id && isSystemUser(),
  });

  // Fetch tenant statistics
  const { data: users, isLoading: usersLoading } = useQuery({
    queryKey: ['users', id],
    queryFn: () => userApi.list(id!),
    enabled: !!id,
  });

  const { data: roles, isLoading: rolesLoading } = useQuery({
    queryKey: ['roles', id],
    queryFn: () => roleApi.list(id!),
    enabled: !!id,
  });

  const { data: permissions, isLoading: permissionsLoading } = useQuery({
    queryKey: ['permissions', id],
    queryFn: () => permissionApi.list(id!),
    enabled: !!id,
  });

  const { data: capabilities } = useQuery({
    queryKey: ['tenant', 'capabilities', id],
    queryFn: () => tenantCapabilityApi.list(id!),
    enabled: !!id && isSystemUser(),
  });

  const { data: features } = useQuery({
    queryKey: ['tenant', 'features', id],
    queryFn: () => tenantFeatureApi.list(id!),
    enabled: !!id,
  });

  if (tenantLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-gray-500">Loading tenant details...</div>
      </div>
    );
  }

  if (!tenant) {
    return (
      <div className="flex flex-col items-center justify-center h-64">
        <p className="text-gray-500 mb-4">Tenant not found</p>
        <Button onClick={() => navigate('/tenants')}>Back to Tenants</Button>
      </div>
    );
  }

  const stats = [
    {
      title: 'Total Users',
      value: users?.length || 0,
      icon: Users,
      variant: 'primary' as const,
      loading: usersLoading,
    },
    {
      title: 'Total Roles',
      value: roles?.length || 0,
      icon: Key,
      variant: 'success' as const,
      loading: rolesLoading,
    },
    {
      title: 'Total Permissions',
      value: permissions?.length || 0,
      icon: Shield,
      variant: 'default' as const,
      loading: permissionsLoading,
    },
    {
      title: 'Enabled Features',
      value: features?.filter(f => f.enabled).length || 0,
      icon: Zap,
      variant: 'warning' as const,
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
            onClick={() => navigate('/tenants')}
          >
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <div>
            <h1 className="text-3xl font-bold text-gray-900 flex items-center gap-3">
              <Building2 className="h-8 w-8 text-primary-600" />
              {tenant.name}
            </h1>
            <p className="text-gray-600 mt-1">{tenant.domain || tenant.id}</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Badge variant={tenant.status === 'active' ? 'default' : 'secondary'}>
            {tenant.status}
          </Badge>
          <Button variant="outline" onClick={() => setEditTenantOpen(true)}>Edit Tenant</Button>
        </div>
      </div>

      {/* Statistics Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {stats.map((stat, index) => (
          <StatCard key={index} {...stat} />
        ))}
      </div>

      {/* Tabbed Content */}
      <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-4">
        <TabsList className="bg-gray-100">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="users">Users</TabsTrigger>
          <TabsTrigger value="roles">Roles</TabsTrigger>
          <TabsTrigger value="permissions">Permissions</TabsTrigger>
          {isSystemUser() && (
            <>
              <TabsTrigger value="capabilities">Capabilities</TabsTrigger>
              <TabsTrigger value="features">Features</TabsTrigger>
            </>
          )}
          <TabsTrigger value="settings">Settings</TabsTrigger>
          <TabsTrigger value="audit">Audit Logs</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Card>
              <CardHeader>
                <CardTitle>Tenant Information</CardTitle>
                <CardDescription>Basic tenant details</CardDescription>
              </CardHeader>
              <CardContent className="space-y-3">
                <div>
                  <p className="text-sm text-gray-600">Tenant ID</p>
                  <p className="font-mono text-sm">{tenant.id}</p>
                </div>
                <div>
                  <p className="text-sm text-gray-600">Domain</p>
                  <p className="text-sm">{tenant.domain || 'N/A'}</p>
                </div>
                <div>
                  <p className="text-sm text-gray-600">Status</p>
                  <Badge>{tenant.status}</Badge>
                </div>
                <div>
                  <p className="text-sm text-gray-600">Created</p>
                  <p className="text-sm">{new Date(tenant.created_at).toLocaleDateString()}</p>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Quick Actions</CardTitle>
                <CardDescription>Common operations for this tenant</CardDescription>
              </CardHeader>
              <CardContent className="space-y-2">
                <Button 
                  className="w-full justify-start" 
                  variant="outline"
                  onClick={() => {
                    setCreateUserOpen(true);
                  }}
                >
                  <Users className="h-4 w-4 mr-2" />
                  Create User
                </Button>
                <Button 
                  className="w-full justify-start" 
                  variant="outline"
                  onClick={() => {
                    setCreateRoleOpen(true);
                  }}
                >
                  <Key className="h-4 w-4 mr-2" />
                  Create Role
                </Button>
                <Button 
                  className="w-full justify-start" 
                  variant="outline"
                  onClick={() => {
                    setActiveTab('settings');
                  }}
                >
                  <Settings className="h-4 w-4 mr-2" />
                  Configure Settings
                </Button>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="users">
          <UserList tenantId={id} />
        </TabsContent>

        <TabsContent value="roles">
          <RoleList tenantId={id} />
        </TabsContent>

        <TabsContent value="permissions">
          <PermissionList tenantId={id} />
        </TabsContent>

        {isSystemUser() && (
          <>
            <TabsContent value="capabilities">
              <TenantCapabilityAssignment tenantId={id!} />
            </TabsContent>

            <TabsContent value="features">
              <TenantFeatureEnablement tenantId={id!} />
            </TabsContent>
          </>
        )}

        <TabsContent value="settings">
          <Card>
            <CardHeader>
              <CardTitle>Tenant Settings</CardTitle>
              <CardDescription>Configure tenant-specific settings</CardDescription>
            </CardHeader>
            <CardContent>
              {id && <TenantSettingsForm tenantId={id} />}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="audit">
          <Card>
            <CardHeader>
              <CardTitle>Audit Logs</CardTitle>
              <CardDescription>Activity history for this tenant</CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-gray-500">Audit logs coming soon...</p>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* Dialogs */}
      <CreateUserDialog
        open={createUserOpen}
        onOpenChange={(open) => {
          setCreateUserOpen(open);
          if (!open) {
            // Refresh user list when dialog closes
            queryClient.invalidateQueries({ queryKey: ['users', id] });
          }
        }}
        tenantId={id || undefined}
      />
      <CreateRoleDialog
        open={createRoleOpen}
        onOpenChange={(open) => {
          setCreateRoleOpen(open);
          if (!open) {
            // Refresh role list when dialog closes
            queryClient.invalidateQueries({ queryKey: ['roles', id] });
          }
        }}
      />
      {tenant && (
        <EditTenantDialog
          tenant={tenant}
          open={editTenantOpen}
          onOpenChange={setEditTenantOpen}
        />
      )}
    </div>
  );
}

