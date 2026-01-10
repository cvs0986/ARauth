/**
 * Impersonation Banner Component
 * 
 * SECURITY REQUIREMENTS (NON-NEGOTIABLE):
 * - Always visible when impersonating
 * - Cannot be dismissed
 * - Shows impersonated user and tenant
 * - One-click "End Impersonation"
 * - Appears before any tenant UI
 * - Emits audit event on exit
 * 
 * AUTHORITY MODEL:
 * - Only SYSTEM users can impersonate
 * - Must have selected a tenant
 * - Impersonation context stored in auth state
 */

import { useMutation } from '@tanstack/react-query';
import { userApi } from '@/services/api';
import { useAuthStore } from '@/store/authStore';
import { Button } from '@/components/ui/button';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { AlertTriangle, UserCog, LogOut } from 'lucide-react';

export function ImpersonationBanner() {
    const { impersonatedUser, impersonatedTenant, clearImpersonation } = useAuthStore();

    const endImpersonationMutation = useMutation({
        mutationFn: () => userApi.endImpersonation(),
        onSuccess: () => {
            clearImpersonation();
            // Reload to clear impersonation context
            window.location.reload();
        },
    });

    // Only show if impersonating
    if (!impersonatedUser) {
        return null;
    }

    return (
        <Alert className="rounded-none border-x-0 border-t-0 bg-orange-100 border-orange-300">
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                    <AlertTriangle className="h-5 w-5 text-orange-600" />
                    <div>
                        <AlertDescription className="text-orange-900 font-semibold">
                            <UserCog className="h-4 w-4 inline mr-1" />
                            Impersonating: {impersonatedUser.email}
                        </AlertDescription>
                        {impersonatedTenant && (
                            <p className="text-xs text-orange-700 mt-0.5">
                                Tenant: {impersonatedTenant.name}
                            </p>
                        )}
                    </div>
                </div>
                <Button
                    variant="destructive"
                    size="sm"
                    onClick={() => endImpersonationMutation.mutate()}
                    disabled={endImpersonationMutation.isPending}
                    className="bg-orange-600 hover:bg-orange-700"
                >
                    <LogOut className="h-4 w-4 mr-2" />
                    {endImpersonationMutation.isPending ? 'Ending...' : 'End Impersonation'}
                </Button>
            </div>
        </Alert>
    );
}
