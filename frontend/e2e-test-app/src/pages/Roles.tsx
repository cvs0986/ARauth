/**
 * Roles and Permissions View Page for E2E Testing App
 */

import { useQuery } from '@tanstack/react-query';
import { userApi } from '../services/userApi';
import { handleApiError } from '../services/userApi';
import { useAuthStore } from '../store/authStore';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { useNavigate } from 'react-router-dom';
import { Alert } from '@/components/ui/alert';

export function Roles() {
  const navigate = useNavigate();
  const { user } = useAuthStore();
  const userId = user?.id || 'current-user'; // In real app, get from auth token

  const { data: currentUser, isLoading: userLoading } = useQuery({
    queryKey: ['currentUser'],
    queryFn: () => userApi.getCurrentUser(),
    enabled: !!user,
  });

  const { data: roles, isLoading: rolesLoading, error: rolesError } = useQuery({
    queryKey: ['userRoles', userId],
    queryFn: () => userApi.getUserRoles(userId),
    enabled: !!userId && !!currentUser,
  });

  const { data: permissions, isLoading: permissionsLoading, error: permissionsError } = useQuery({
    queryKey: ['userPermissions', userId],
    queryFn: () => userApi.getUserPermissions(userId),
    enabled: !!userId && !!currentUser,
  });

  const isLoading = userLoading || rolesLoading || permissionsLoading;
  const error = rolesError || permissionsError;

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50">
        <header className="bg-white border-b">
          <div className="container mx-auto px-4 py-4">
            <h1 className="text-xl font-bold">Roles & Permissions</h1>
          </div>
        </header>
        <main className="container mx-auto px-4 py-8">
          <div className="text-center">Loading...</div>
        </main>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50">
        <header className="bg-white border-b">
          <div className="container mx-auto px-4 py-4">
            <h1 className="text-xl font-bold">Roles & Permissions</h1>
          </div>
        </header>
        <main className="container mx-auto px-4 py-8">
          <Alert variant="destructive">
            Error loading roles and permissions: {handleApiError(error)}
          </Alert>
        </main>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b">
        <div className="container mx-auto px-4 py-4 flex items-center justify-between">
          <h1 className="text-xl font-bold">Roles & Permissions</h1>
          <Button variant="outline" onClick={() => navigate('/dashboard')}>
            Back to Dashboard
          </Button>
        </div>
      </header>

      <main className="container mx-auto px-4 py-8 max-w-6xl">
        <div className="space-y-6">
          {currentUser && (
            <Card>
              <CardHeader>
                <CardTitle>User Information</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-sm text-gray-600">Username</p>
                    <p className="font-medium">{currentUser.username}</p>
                  </div>
                  <div>
                    <p className="text-sm text-gray-600">Email</p>
                    <p className="font-medium">{currentUser.email}</p>
                  </div>
                  <div>
                    <p className="text-sm text-gray-600">Status</p>
                    <p className="font-medium">
                      <span
                        className={`px-2 py-1 rounded text-xs ${
                          currentUser.status === 'active'
                            ? 'bg-green-100 text-green-800'
                            : 'bg-gray-100 text-gray-800'
                        }`}
                      >
                        {currentUser.status}
                      </span>
                    </p>
                  </div>
                  <div>
                    <p className="text-sm text-gray-600">Tenant ID</p>
                    <p className="font-medium">{currentUser.tenant_id}</p>
                  </div>
                </div>
              </CardContent>
            </Card>
          )}

          <Card>
            <CardHeader>
              <CardTitle>Assigned Roles</CardTitle>
              <CardDescription>
                {roles && roles.length > 0
                  ? `You have ${roles.length} role(s) assigned`
                  : 'No roles assigned'}
              </CardDescription>
            </CardHeader>
            <CardContent>
              {roles && roles.length > 0 ? (
                <div className="space-y-4">
                  {roles.map((role) => (
                    <div
                      key={role.id}
                      className="border rounded-lg p-4 hover:bg-gray-50 transition-colors"
                    >
                      <div className="flex items-start justify-between">
                        <div className="flex-1">
                          <h3 className="font-semibold text-lg">{role.name}</h3>
                          {role.description && (
                            <p className="text-sm text-gray-600 mt-1">{role.description}</p>
                          )}
                          <div className="mt-3">
                            <p className="text-sm font-medium text-gray-700">
                              Permissions ({role.permissions?.length || 0}):
                            </p>
                            {role.permissions && role.permissions.length > 0 ? (
                              <div className="mt-2 flex flex-wrap gap-2">
                                {role.permissions.map((permission, index) => (
                                  <span
                                    key={index}
                                    className="px-2 py-1 bg-blue-100 text-blue-800 rounded text-xs"
                                  >
                                    {permission}
                                  </span>
                                ))}
                              </div>
                            ) : (
                              <p className="text-sm text-gray-500 mt-1">No permissions</p>
                            )}
                          </div>
                        </div>
                        <div className="text-sm text-gray-500">
                          <p>Created: {new Date(role.created_at).toLocaleDateString()}</p>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-gray-500 text-center py-8">
                  No roles assigned. Contact your administrator to assign roles.
                </p>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>All Permissions</CardTitle>
              <CardDescription>
                {permissions && permissions.length > 0
                  ? `You have access to ${permissions.length} permission(s) through your roles`
                  : 'No permissions available'}
              </CardDescription>
            </CardHeader>
            <CardContent>
              {permissions && permissions.length > 0 ? (
                <div className="space-y-2">
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                    {permissions.map((permission) => (
                      <div
                        key={permission.id}
                        className="border rounded-lg p-3 hover:bg-gray-50 transition-colors"
                      >
                        <div className="flex items-start justify-between">
                          <div className="flex-1">
                            <p className="font-medium text-sm">
                              {permission.resource}:{permission.action}
                            </p>
                            {permission.description && (
                              <p className="text-xs text-gray-600 mt-1">{permission.description}</p>
                            )}
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              ) : (
                <p className="text-gray-500 text-center py-8">
                  No permissions available. Permissions are granted through roles.
                </p>
              )}
            </CardContent>
          </Card>

          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <h3 className="font-semibold text-blue-900 mb-2">Testing Information</h3>
            <p className="text-sm text-blue-800">
              This page displays your assigned roles and permissions. Use this to verify that:
            </p>
            <ul className="list-disc list-inside text-sm text-blue-800 mt-2 space-y-1">
              <li>Roles are correctly assigned to your user account</li>
              <li>Permissions are properly inherited from roles</li>
              <li>Role and permission data is displayed correctly</li>
              <li>API endpoints for user roles and permissions work as expected</li>
            </ul>
          </div>
        </div>
      </main>
    </div>
  );
}

