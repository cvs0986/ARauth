/**
 * Header Component
 * 
 * GUARDRAIL #1: Backend Is Law
 * - User info from PrincipalContext (JWT claims)
 * 
 * GUARDRAIL #6: UI Quality Bar
 * - Clear authority context (mode badge, tenant selector)
 * - Professional, calm design
 */

import { useState } from 'react';
import { usePrincipalContext } from '@/contexts/PrincipalContext';
import { useNavigate } from 'react-router-dom';
import { TenantSelector } from '@/components/TenantSelector';
import { UserTypeBadge } from '@/components/UserTypeBadge';
import { Shield, LogOut, ChevronDown } from 'lucide-react';
import { useAuthStore } from '@/store/authStore';

export function Header() {
  const { username, email, principalType, isAuthenticated } = usePrincipalContext();
  const { clearAuth } = useAuthStore();
  const navigate = useNavigate();
  const [showUserMenu, setShowUserMenu] = useState(false);

  const handleLogout = () => {
    clearAuth();
    navigate('/login');
  };

  if (!isAuthenticated) {
    return null;
  }

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
          <UserTypeBadge />
        </div>
        <div className="flex items-center gap-4">
          {principalType === 'SYSTEM' && <TenantSelector />}
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
                        <UserTypeBadge />
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
