/**
 * Header Component
 * Shows tenant selector for SYSTEM users and user info
 */

import { useAuthStore } from '@/store/authStore';
import { Button } from '@/components/ui/button';
import { useNavigate } from 'react-router-dom';
import { TenantSelector } from '@/components/TenantSelector';
import { Shield, Building2 } from 'lucide-react';

export function Header() {
  const { clearAuth, isAuthenticated, isSystemUser, tenantId } = useAuthStore();
  const navigate = useNavigate();

  const handleLogout = () => {
    clearAuth();
    navigate('/login');
  };

  if (!isAuthenticated) {
    return null;
  }

  return (
    <header className="border-b bg-white">
      <div className="flex items-center justify-between px-6 py-4">
        <div className="flex items-center gap-6">
          <div className="flex items-center gap-2">
            <Shield className="h-6 w-6 text-blue-600" />
            <h1 className="text-xl font-bold text-gray-900">ARauth Identity</h1>
          </div>
          {isSystemUser() && (
            <div className="flex items-center gap-2 text-sm text-gray-600">
              <span className="px-2 py-1 bg-blue-100 text-blue-800 rounded">System Admin</span>
            </div>
          )}
          {!isSystemUser() && tenantId && (
            <div className="flex items-center gap-2 text-sm text-gray-600">
              <Building2 className="h-4 w-4" />
              <span>Tenant Admin</span>
            </div>
          )}
        </div>
        <div className="flex items-center gap-4">
          {isSystemUser() && <TenantSelector />}
          <Button variant="outline" onClick={handleLogout}>
            Logout
          </Button>
        </div>
      </div>
    </header>
  );
}

