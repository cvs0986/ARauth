/**
 * No Access Page
 * Shown when user doesn't have admin:access permission
 */

import { useAuthStore } from '@/store/authStore';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { AlertCircle, Home, LogOut } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

export function NoAccess() {
  const navigate = useNavigate();
  const { clearAuth, username, email } = useAuthStore();

  const handleLogout = () => {
    clearAuth();
    navigate('/login');
  };

  return (
    <div className="flex items-center justify-center min-h-screen bg-gray-50">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-red-100">
            <AlertCircle className="h-8 w-8 text-red-600" />
          </div>
          <CardTitle className="text-2xl">No Admin Access</CardTitle>
          <CardDescription>
            You don't have permission to access the admin dashboard.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="rounded-lg bg-gray-50 p-4">
            <p className="text-sm text-gray-600 mb-2">
              <strong>Logged in as:</strong>
            </p>
            {username && (
              <p className="text-sm font-medium text-gray-900">{username}</p>
            )}
            {email && (
              <p className="text-sm text-gray-600">{email}</p>
            )}
          </div>

          <div className="space-y-2">
            <p className="text-sm text-gray-600">
              To access the admin dashboard, you need the <code className="px-1.5 py-0.5 bg-gray-100 rounded text-xs">tenant.admin.access</code> permission.
            </p>
            <p className="text-sm text-gray-600">
              Please contact your administrator to request access or assign the appropriate role.
            </p>
          </div>

          <div className="flex gap-2 pt-4">
            <Button
              variant="outline"
              className="flex-1"
              onClick={() => navigate('/')}
            >
              <Home className="h-4 w-4 mr-2" />
              Go Home
            </Button>
            <Button
              variant="outline"
              className="flex-1"
              onClick={handleLogout}
            >
              <LogOut className="h-4 w-4 mr-2" />
              Logout
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

