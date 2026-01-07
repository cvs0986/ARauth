/**
 * Dashboard Page for E2E Testing App
 */

import { useAuthStore } from '../store/authStore';
import { useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

export function Dashboard() {
  const { clearAuth, tenantId } = useAuthStore();
  const navigate = useNavigate();

  const handleLogout = () => {
    clearAuth();
    navigate('/login');
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b">
        <div className="container mx-auto px-4 py-4 flex items-center justify-between">
          <h1 className="text-xl font-bold">ARauth Identity - E2E Testing</h1>
          <Button variant="outline" onClick={handleLogout}>
            Logout
          </Button>
        </div>
      </header>

      <main className="container mx-auto px-4 py-8">
        <div className="space-y-6">
          <div>
            <h2 className="text-2xl font-bold mb-4">Welcome to E2E Testing Dashboard</h2>
            <p className="text-gray-600">
              This app is designed for end-to-end testing of all ARauth Identity features.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <Card>
              <CardHeader>
                <CardTitle>Profile</CardTitle>
                <CardDescription>View and edit your profile</CardDescription>
              </CardHeader>
              <CardContent>
                <Button onClick={() => navigate('/profile')} className="w-full">
                  Go to Profile
                </Button>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>MFA</CardTitle>
                <CardDescription>Manage multi-factor authentication</CardDescription>
              </CardHeader>
              <CardContent>
                <Button onClick={() => navigate('/mfa')} className="w-full">
                  Manage MFA
                </Button>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Roles & Permissions</CardTitle>
                <CardDescription>View your assigned roles and permissions</CardDescription>
              </CardHeader>
              <CardContent>
                <Button onClick={() => navigate('/roles')} className="w-full">
                  View Roles & Permissions
                </Button>
              </CardContent>
            </Card>
          </div>

          {tenantId && (
            <Card>
              <CardHeader>
                <CardTitle>Current Tenant</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-gray-600">Tenant ID: {tenantId}</p>
              </CardContent>
            </Card>
          )}
        </div>
      </main>
    </div>
  );
}

