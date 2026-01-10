/**
 * Header Component
 * Shows tenant selector for SYSTEM users and user info
 */

import { useState } from 'react';
import { useAuthStore } from '@/store/authStore';
import { Button } from '@/components/ui/button';
import { useNavigate } from 'react-router-dom';
import { TenantSelector } from '@/components/TenantSelector';
import { Shield, Building2, User, LogOut, ChevronDown } from 'lucide-react';
import { cn } from '@/lib/utils';

export function Header() {
  const { clearAuth, isAuthenticated, isSystemUser, tenantId, systemRoles, username, email } = useAuthStore();
  const navigate = useNavigate();
  const [showUserMenu, setShowUserMenu] = useState(false);

  const handleLogout = () => {
    clearAuth();
    navigate('/login');
  };

  if (!isAuthenticated) {
    return null;
  }

  // Get role display name
  const getRoleDisplayName = () => {
    if (!isSystemUser()) {
      return 'Tenant Admin';
    }
    if (systemRoles && systemRoles.length > 0) {
      // Format role name: system_owner -> System Owner
      const role = systemRoles[0];
      return role.split('_').map(word => 
        word.charAt(0).toUpperCase() + word.slice(1)
      ).join(' ');
    }
    return 'System Admin';
  };

  const roleDisplayName = getRoleDisplayName();
  const userInitials = username 
    ? username.substring(0, 2).toUpperCase()
    : email 
    ? email.substring(0, 2).toUpperCase()
    : 'U';

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
              <span className="px-2 py-1 bg-blue-100 text-blue-800 rounded">{roleDisplayName}</span>
            </div>
          )}
          {!isSystemUser() && tenantId && (
            <div className="flex items-center gap-2 text-sm text-gray-600">
              <Building2 className="h-4 w-4" />
              <span>{roleDisplayName}</span>
            </div>
          )}
        </div>
        <div className="flex items-center gap-4">
          {isSystemUser() && <TenantSelector />}
          <div className="relative">
            <button
              onClick={() => setShowUserMenu(!showUserMenu)}
              className="flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-gray-100 transition-colors"
            >
              <div className="h-8 w-8 rounded-full bg-primary-600 text-white flex items-center justify-center text-sm font-semibold">
                {userInitials}
              </div>
              <ChevronDown className="h-4 w-4 text-gray-600" />
            </button>
            {showUserMenu && (
              <>
                <div 
                  className="fixed inset-0 z-10" 
                  onClick={() => setShowUserMenu(false)}
                />
                <div className="absolute right-0 mt-2 w-64 bg-white rounded-lg shadow-lg border border-gray-200 z-20">
                  <div className="p-4 border-b border-gray-200">
                    <div className="flex items-center gap-3">
                      <div className="h-10 w-10 rounded-full bg-primary-600 text-white flex items-center justify-center text-sm font-semibold">
                        {userInitials}
                      </div>
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-semibold text-gray-900 truncate">
                          {username || email || 'User'}
                        </p>
                        <p className="text-xs text-gray-500 truncate">{email}</p>
                        <p className="text-xs text-gray-400 mt-1">{roleDisplayName}</p>
                      </div>
                    </div>
                  </div>
                  <div className="p-2">
                    <button
                      onClick={handleLogout}
                      className="w-full flex items-center gap-2 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
                    >
                      <LogOut className="h-4 w-4" />
                      Logout
                    </button>
                  </div>
                </div>
              </>
            )}
          </div>
        </div>
      </div>
    </header>
  );
}

