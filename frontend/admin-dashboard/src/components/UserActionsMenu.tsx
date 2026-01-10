/**
 * User Actions Menu Component
 * 
 * AUTHORITY MODEL:
 * - Suspend/Activate: users:update (SYSTEM or TENANT)
 * - Reset MFA: users:mfa:reset (SYSTEM or TENANT)
 * - Impersonate: users:impersonate (SYSTEM ONLY, tenant must be selected)
 * 
 * SECURITY GUARDRAILS:
 * - All actions permission-gated
 * - Destructive actions require confirmation + reason
 * - Impersonate only visible when tenant selected
 * - Audit events emitted for all actions
 */

import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { userApi } from '@/services/api';
import { usePrincipalContext } from '@/contexts/PrincipalContext';
import { PermissionGate } from '@/components/PermissionGate';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { MoreVertical, UserX, UserCheck, Shield, UserCog, AlertTriangle } from 'lucide-react';
import type { User } from '@shared/types/api';

interface UserActionsMenuProps {
    user: User;
}

type ActionType = 'suspend' | 'activate' | 'reset-mfa' | 'impersonate';

export function UserActionsMenu({ user }: UserActionsMenuProps) {
    const { principalType, selectedTenantId, consoleMode } = usePrincipalContext();
    const queryClient = useQueryClient();
    const [confirmAction, setConfirmAction] = useState<ActionType | null>(null);
    const [reason, setReason] = useState('');
    const [error, setError] = useState<string | null>(null);

    // Impersonate only available for SYSTEM users with tenant selected
    const canImpersonate = principalType === 'SYSTEM' && !!selectedTenantId && !!user.tenant_id;

    const suspendMutation = useMutation({
        mutationFn: ({ userId, reason }: { userId: string; reason: string }) =>
            userApi.suspend(userId, reason),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['users'] });
            setConfirmAction(null);
            setReason('');
        },
        onError: (err: any) => {
            setError(err.message || 'Failed to suspend user');
        },
    });

    const activateMutation = useMutation({
        mutationFn: (userId: string) => userApi.activate(userId),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['users'] });
            setConfirmAction(null);
        },
        onError: (err: any) => {
            setError(err.message || 'Failed to activate user');
        },
    });

    const resetMFAMutation = useMutation({
        mutationFn: ({ userId, reason }: { userId: string; reason: string }) =>
            userApi.resetMFA(userId, reason),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['users'] });
            setConfirmAction(null);
            setReason('');
        },
        onError: (err: any) => {
            setError(err.message || 'Failed to reset MFA');
        },
    });

    const impersonateMutation = useMutation({
        mutationFn: (userId: string) => userApi.impersonate(userId),
        onSuccess: () => {
            // Reload to apply impersonation context
            window.location.reload();
        },
        onError: (err: any) => {
            setError(err.message || 'Failed to impersonate user');
        },
    });

    const handleConfirm = () => {
        setError(null);

        switch (confirmAction) {
            case 'suspend':
                if (!reason.trim()) {
                    setError('Reason is required for suspending a user');
                    return;
                }
                suspendMutation.mutate({ userId: user.id, reason });
                break;
            case 'activate':
                activateMutation.mutate(user.id);
                break;
            case 'reset-mfa':
                if (!reason.trim()) {
                    setError('Reason is required for resetting MFA');
                    return;
                }
                resetMFAMutation.mutate({ userId: user.id, reason });
                break;
            case 'impersonate':
                impersonateMutation.mutate(user.id);
                break;
        }
    };

    const isPending = suspendMutation.isPending || activateMutation.isPending ||
        resetMFAMutation.isPending || impersonateMutation.isPending;

    return (
        <>
            <DropdownMenu>
                <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="sm">
                        <MoreVertical className="h-4 w-4" />
                    </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                    {/* Suspend/Activate */}
                    <PermissionGate permission="users:update" systemPermission={principalType === 'SYSTEM'}>
                        {user.status === 'active' ? (
                            <DropdownMenuItem onClick={() => setConfirmAction('suspend')}>
                                <UserX className="h-4 w-4 mr-2" />
                                Suspend User
                            </DropdownMenuItem>
                        ) : (
                            <DropdownMenuItem onClick={() => setConfirmAction('activate')}>
                                <UserCheck className="h-4 w-4 mr-2" />
                                Activate User
                            </DropdownMenuItem>
                        )}
                    </PermissionGate>

                    {/* Reset MFA */}
                    {user.mfa_enabled && (
                        <PermissionGate permission="users:mfa:reset" systemPermission={principalType === 'SYSTEM'}>
                            <DropdownMenuItem onClick={() => setConfirmAction('reset-mfa')}>
                                <Shield className="h-4 w-4 mr-2" />
                                Reset MFA
                            </DropdownMenuItem>
                        </PermissionGate>
                    )}

                    {/* Impersonate - SYSTEM only, tenant must be selected */}
                    {canImpersonate && (
                        <>
                            <DropdownMenuSeparator />
                            <PermissionGate permission="users:impersonate" systemPermission>
                                <DropdownMenuItem
                                    onClick={() => setConfirmAction('impersonate')}
                                    className="text-orange-600"
                                >
                                    <UserCog className="h-4 w-4 mr-2" />
                                    Impersonate User
                                </DropdownMenuItem>
                            </PermissionGate>
                        </>
                    )}
                </DropdownMenuContent>
            </DropdownMenu>

            {/* Confirmation Dialog */}
            <Dialog open={!!confirmAction} onOpenChange={(open) => !open && setConfirmAction(null)}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>
                            {confirmAction === 'suspend' && 'Suspend User'}
                            {confirmAction === 'activate' && 'Activate User'}
                            {confirmAction === 'reset-mfa' && 'Reset MFA'}
                            {confirmAction === 'impersonate' && 'Impersonate User'}
                        </DialogTitle>
                        <DialogDescription>
                            {confirmAction === 'suspend' && `Suspend ${user.username}? They will not be able to log in.`}
                            {confirmAction === 'activate' && `Activate ${user.username}? They will be able to log in.`}
                            {confirmAction === 'reset-mfa' && `Reset MFA for ${user.username}? They will need to re-enroll.`}
                            {confirmAction === 'impersonate' && `Impersonate ${user.username}? You will act as this user.`}
                        </DialogDescription>
                    </DialogHeader>

                    {error && (
                        <Alert variant="destructive">
                            <AlertDescription>{error}</AlertDescription>
                        </Alert>
                    )}

                    {/* Reason required for destructive actions */}
                    {(confirmAction === 'suspend' || confirmAction === 'reset-mfa') && (
                        <div className="space-y-2">
                            <Label htmlFor="reason">Reason (Required) *</Label>
                            <Textarea
                                id="reason"
                                value={reason}
                                onChange={(e) => setReason(e.target.value)}
                                placeholder="Provide a reason for this action (for audit purposes)"
                                rows={3}
                                disabled={isPending}
                            />
                            <p className="text-xs text-gray-500">
                                This reason will be recorded in the audit log
                            </p>
                        </div>
                    )}

                    {/* Impersonation warning */}
                    {confirmAction === 'impersonate' && (
                        <Alert className="bg-orange-50 border-orange-200">
                            <AlertTriangle className="h-4 w-4 text-orange-600" />
                            <AlertDescription className="text-orange-800">
                                <strong>Security Notice:</strong> All actions will be performed as {user.username}.
                                An impersonation banner will be visible. This action is audited.
                            </AlertDescription>
                        </Alert>
                    )}

                    <DialogFooter>
                        <Button
                            variant="outline"
                            onClick={() => {
                                setConfirmAction(null);
                                setReason('');
                                setError(null);
                            }}
                            disabled={isPending}
                        >
                            Cancel
                        </Button>
                        <Button
                            onClick={handleConfirm}
                            disabled={isPending}
                            variant={confirmAction === 'suspend' || confirmAction === 'impersonate' ? 'destructive' : 'default'}
                        >
                            {isPending ? 'Processing...' : 'Confirm'}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </>
    );
}
