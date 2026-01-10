/**
 * Active Sessions Management
 * 
 * AUTHORITY MODEL:
 * Who: SYSTEM users (cross-tenant) OR TENANT users (tenant sessions)
 * Scope: SYSTEM mode (all sessions) OR TENANT mode (tenant sessions)
 * Permission: sessions:read, sessions:revoke
 * 
 * SECURITY:
 * - Revoke session requires audit reason
 * - Clear warnings about immediate logout
 * - Revoke all sessions for user capability
 * 
 * UI CONTRACT MODE:
 * - All backend calls throw APINotConnectedError
 */

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { usePrincipalContext } from '@/contexts/PrincipalContext';
import { PermissionGate } from '@/components/PermissionGate';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from '@/components/ui/table';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { EmptyState } from '@/components/EmptyState';
import { Skeleton } from '@/components/ui/skeleton';
import { Monitor, AlertTriangle, Clock } from 'lucide-react';
import { APINotConnectedError, getAPINotConnectedMessage, isAPINotConnected } from '@/lib/errors';

interface Session {
    id: string;
    user_id: string;
    user_email: string;
    ip_address: string;
    user_agent: string;
    started_at: string;
    last_activity_at: string;
    status: 'active' | 'expired';
}

export function ActiveSessions() {
    const { principalType, homeTenantId, selectedTenantId, consoleMode } = usePrincipalContext();
    const queryClient = useQueryClient();
    const [revokeSession, setRevokeSession] = useState<Session | null>(null);
    const [revokeReason, setRevokeReason] = useState('');
    const [error, setError] = useState<string | null>(null);

    // Determine effective tenant ID
    const effectiveTenantId = principalType === 'SYSTEM' ? selectedTenantId : homeTenantId;

    // Fetch active sessions
    const { data: sessions, isLoading } = useQuery({
        queryKey: consoleMode === 'SYSTEM' && !effectiveTenantId
            ? ['system-sessions']
            : ['sessions', effectiveTenantId],
        queryFn: async () => {
            // UI CONTRACT MODE: Throw APINotConnectedError
            throw new APINotConnectedError('sessions.list');
        },
        enabled: consoleMode === 'SYSTEM' ? true : !!effectiveTenantId,
        retry: false,
    });

    // Revoke session mutation
    const revokeMutation = useMutation({
        mutationFn: async ({ sessionId, reason }: { sessionId: string; reason: string }) => {
            // UI CONTRACT MODE: Throw APINotConnectedError
            throw new APINotConnectedError('sessions.revoke');
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: consoleMode === 'SYSTEM' && !effectiveTenantId ? ['system-sessions'] : ['sessions', effectiveTenantId] });
            setRevokeSession(null);
            setRevokeReason('');
            setError(null);
        },
        onError: (err: any) => {
            setError(getAPINotConnectedMessage(err));
        },
    });

    const handleRevoke = () => {
        if (!revokeReason.trim()) {
            setError('Audit reason is required');
            return;
        }
        if (revokeSession) {
            revokeMutation.mutate({ sessionId: revokeSession.id, reason: revokeReason });
        }
    };

    return (
        <PermissionGate permission="sessions:read" systemPermission={principalType === 'SYSTEM'}>
            <div className="space-y-6">
                {/* Header */}
                <div>
                    <h1 className="text-3xl font-bold">Active Sessions</h1>
                    <p className="text-sm text-gray-600 mt-1">
                        {consoleMode === 'SYSTEM' && !effectiveTenantId
                            ? 'View and manage all active user sessions'
                            : 'View and manage active sessions for this tenant'}
                    </p>
                </div>

                {/* Scope Indicator */}
                <Alert className="bg-blue-50 border-blue-200">
                    <Monitor className="h-4 w-4 text-blue-600" />
                    <AlertDescription className="text-blue-800 text-sm">
                        {consoleMode === 'SYSTEM' && !effectiveTenantId ? (
                            <><strong>Scope:</strong> Viewing all active sessions across all tenants</>
                        ) : (
                            <><strong>Scope:</strong> Viewing active sessions for this tenant only</>
                        )}
                    </AlertDescription>
                </Alert>

                {/* API Not Connected Notice */}
                {isLoading && (
                    <Alert className="bg-blue-50 border-blue-200">
                        <Monitor className="h-4 w-4 text-blue-600" />
                        <AlertDescription className="text-blue-800">
                            <strong>Backend Integration Pending:</strong> The sessions API is not yet connected.
                        </AlertDescription>
                    </Alert>
                )}

                {/* Table */}
                {isLoading ? (
                    <div className="space-y-3">
                        <Skeleton className="h-12 w-full" />
                        <Skeleton className="h-12 w-full" />
                        <Skeleton className="h-12 w-full" />
                    </div>
                ) : !sessions || sessions.length === 0 ? (
                    <EmptyState
                        icon={Monitor}
                        title="No Active Sessions"
                        description="No active user sessions found."
                    />
                ) : (
                    <div className="border rounded-lg">
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>User</TableHead>
                                    <TableHead>IP Address</TableHead>
                                    <TableHead>User Agent</TableHead>
                                    <TableHead>Started</TableHead>
                                    <TableHead>Last Activity</TableHead>
                                    <TableHead>Status</TableHead>
                                    <TableHead className="text-right">Actions</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {sessions.map((session) => (
                                    <TableRow key={session.id}>
                                        <TableCell className="font-medium">{session.user_email}</TableCell>
                                        <TableCell className="font-mono text-sm">{session.ip_address}</TableCell>
                                        <TableCell className="text-sm truncate max-w-xs">{session.user_agent}</TableCell>
                                        <TableCell>
                                            <div className="flex items-center gap-1">
                                                <Clock className="h-4 w-4 text-gray-400" />
                                                <span className="text-sm">{new Date(session.started_at).toLocaleString()}</span>
                                            </div>
                                        </TableCell>
                                        <TableCell className="text-sm">{new Date(session.last_activity_at).toLocaleString()}</TableCell>
                                        <TableCell>
                                            <Badge variant={session.status === 'active' ? 'default' : 'secondary'}>
                                                {session.status}
                                            </Badge>
                                        </TableCell>
                                        <TableCell className="text-right">
                                            {session.status === 'active' && (
                                                <PermissionGate permission="sessions:revoke" systemPermission={principalType === 'SYSTEM'}>
                                                    <Button
                                                        variant="destructive"
                                                        size="sm"
                                                        onClick={() => setRevokeSession(session)}
                                                    >
                                                        Revoke
                                                    </Button>
                                                </PermissionGate>
                                            )}
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </div>
                )}

                {/* Revoke Dialog */}
                <Dialog open={!!revokeSession} onOpenChange={(open) => !open && setRevokeSession(null)}>
                    <DialogContent>
                        <DialogHeader>
                            <DialogTitle>Revoke Session</DialogTitle>
                            <DialogDescription>
                                Revoke session for {revokeSession?.user_email}? This will log the user out immediately.
                            </DialogDescription>
                        </DialogHeader>

                        {error && (
                            <Alert variant={isAPINotConnected(error) ? 'default' : 'destructive'} className={isAPINotConnected(error) ? 'bg-blue-50 border-blue-200' : ''}>
                                <AlertDescription className={isAPINotConnected(error) ? 'text-blue-800' : ''}>
                                    {error}
                                </AlertDescription>
                            </Alert>
                        )}

                        <div className="space-y-2">
                            <Label htmlFor="revoke-reason">Audit Reason (Required) *</Label>
                            <Textarea
                                id="revoke-reason"
                                value={revokeReason}
                                onChange={(e) => setRevokeReason(e.target.value)}
                                placeholder="Provide a reason for revoking this session (for audit purposes)"
                                rows={3}
                                disabled={revokeMutation.isPending}
                            />
                            <p className="text-xs text-gray-500">
                                This reason will be recorded in the audit log
                            </p>
                        </div>

                        <Alert className="bg-orange-50 border-orange-200">
                            <AlertTriangle className="h-4 w-4 text-orange-600" />
                            <AlertDescription className="text-orange-800 text-sm">
                                <strong>Warning:</strong> Revoking this session will log the user out immediately.
                            </AlertDescription>
                        </Alert>

                        <DialogFooter>
                            <Button
                                variant="outline"
                                onClick={() => {
                                    setRevokeSession(null);
                                    setRevokeReason('');
                                    setError(null);
                                }}
                                disabled={revokeMutation.isPending}
                            >
                                Cancel
                            </Button>
                            <Button
                                variant="destructive"
                                onClick={handleRevoke}
                                disabled={revokeMutation.isPending}
                            >
                                {revokeMutation.isPending ? 'Revoking...' : 'Revoke Session'}
                            </Button>
                        </DialogFooter>
                    </DialogContent>
                </Dialog>
            </div>
        </PermissionGate>
    );
}
