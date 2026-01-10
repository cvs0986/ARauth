/**
 * SCIM Token List Component
 * 
 * AUTHORITY MODEL:
 * Who: TENANT users OR SYSTEM users (tenant selected)
 * Scope: Tenant-scoped SCIM tokens
 * Permission: scim:tokens:read, scim:tokens:create, scim:tokens:revoke
 * 
 * SECURITY GUARDRAILS:
 * - SCIM tokens are root credentials
 * - One-time secret display only
 * - Revoke requires audit reason
 * - No token re-display
 * - No edit token (rotate = revoke + create)
 */

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { usePrincipalContext } from '@/contexts/PrincipalContext';
import { PermissionGate } from '@/components/PermissionGate';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
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
import { Plus, Trash2, AlertTriangle, Shield } from 'lucide-react';
import { CreateSCIMTokenDialog } from './CreateSCIMTokenDialog';

interface SCIMTokenListProps {
    tenantId: string;
}

interface SCIMToken {
    id: string;
    name: string;
    created_at: string;
    last_used_at: string | null;
    status: 'active' | 'revoked';
}

export function SCIMTokenList({ tenantId }: SCIMTokenListProps) {
    const { principalType } = usePrincipalContext();
    const queryClient = useQueryClient();
    const [createOpen, setCreateOpen] = useState(false);
    const [revokeToken, setRevokeToken] = useState<SCIMToken | null>(null);
    const [revokeReason, setRevokeReason] = useState('');
    const [error, setError] = useState<string | null>(null);

    // Fetch SCIM tokens
    const { data: tokens, isLoading } = useQuery({
        queryKey: ['scim-tokens', tenantId],
        queryFn: async () => {
            // TODO: Implement API call
            // return scimApi.listTokens(tenantId);
            return [] as SCIMToken[];
        },
        enabled: !!tenantId,
    });

    // Revoke token mutation
    const revokeMutation = useMutation({
        mutationFn: async ({ tokenId, reason }: { tokenId: string; reason: string }) => {
            // TODO: Implement API call
            // return scimApi.revokeToken(tenantId, tokenId, reason);
            console.log('Revoking token:', tokenId, 'Reason:', reason);
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['scim-tokens', tenantId] });
            setRevokeToken(null);
            setRevokeReason('');
            setError(null);
        },
        onError: (err: any) => {
            setError(err.message || 'Failed to revoke token');
        },
    });

    const handleRevoke = () => {
        if (!revokeReason.trim()) {
            setError('Audit reason is required');
            return;
        }
        if (revokeToken) {
            revokeMutation.mutate({ tokenId: revokeToken.id, reason: revokeReason });
        }
    };

    if (isLoading) {
        return <div className="p-4">Loading SCIM tokens...</div>;
    }

    return (
        <>
            <Card>
                <CardHeader>
                    <div className="flex items-center justify-between">
                        <div>
                            <CardTitle className="flex items-center gap-2">
                                <Shield className="h-5 w-5" />
                                SCIM Tokens
                            </CardTitle>
                            <CardDescription>
                                Manage authentication tokens for SCIM provisioning
                            </CardDescription>
                        </div>
                        <PermissionGate permission="scim:tokens:create" systemPermission={principalType === 'SYSTEM'}>
                            <Button onClick={() => setCreateOpen(true)} size="sm">
                                <Plus className="h-4 w-4 mr-2" />
                                Create Token
                            </Button>
                        </PermissionGate>
                    </div>
                </CardHeader>
                <CardContent>
                    {/* Security Warning */}
                    <Alert className="bg-orange-50 border-orange-200 mb-4">
                        <AlertTriangle className="h-4 w-4 text-orange-600" />
                        <AlertDescription className="text-orange-800 text-sm">
                            <strong>Security:</strong> SCIM tokens grant full provisioning access.
                            Treat them as root credentials and rotate regularly.
                        </AlertDescription>
                    </Alert>

                    {/* Token Table */}
                    <div className="border rounded-lg">
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Name</TableHead>
                                    <TableHead>Status</TableHead>
                                    <TableHead>Created</TableHead>
                                    <TableHead>Last Used</TableHead>
                                    <TableHead className="text-right">Actions</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {tokens && tokens.length === 0 && (
                                    <TableRow>
                                        <TableCell colSpan={5} className="text-center text-gray-500">
                                            No SCIM tokens configured. Create a token to enable provisioning.
                                        </TableCell>
                                    </TableRow>
                                )}
                                {tokens?.map((token) => (
                                    <TableRow key={token.id}>
                                        <TableCell className="font-medium">{token.name}</TableCell>
                                        <TableCell>
                                            <Badge variant={token.status === 'active' ? 'default' : 'secondary'}>
                                                {token.status}
                                            </Badge>
                                        </TableCell>
                                        <TableCell>{new Date(token.created_at).toLocaleDateString()}</TableCell>
                                        <TableCell>
                                            {token.last_used_at ? (
                                                new Date(token.last_used_at).toLocaleDateString()
                                            ) : (
                                                <span className="text-gray-400">Never</span>
                                            )}
                                        </TableCell>
                                        <TableCell className="text-right">
                                            {token.status === 'active' && (
                                                <PermissionGate permission="scim:tokens:revoke" systemPermission={principalType === 'SYSTEM'}>
                                                    <Button
                                                        variant="destructive"
                                                        size="sm"
                                                        onClick={() => setRevokeToken(token)}
                                                    >
                                                        <Trash2 className="h-4 w-4 mr-2" />
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
                </CardContent>
            </Card>

            {/* Create Token Dialog */}
            <CreateSCIMTokenDialog
                open={createOpen}
                onOpenChange={setCreateOpen}
                tenantId={tenantId}
            />

            {/* Revoke Token Dialog */}
            <Dialog open={!!revokeToken} onOpenChange={(open) => !open && setRevokeToken(null)}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>Revoke SCIM Token</DialogTitle>
                        <DialogDescription>
                            Revoke "{revokeToken?.name}"? This action cannot be undone.
                        </DialogDescription>
                    </DialogHeader>

                    {error && (
                        <Alert variant="destructive">
                            <AlertDescription>{error}</AlertDescription>
                        </Alert>
                    )}

                    <div className="space-y-2">
                        <Label htmlFor="revoke-reason">Audit Reason (Required) *</Label>
                        <Textarea
                            id="revoke-reason"
                            value={revokeReason}
                            onChange={(e) => setRevokeReason(e.target.value)}
                            placeholder="Provide a reason for revoking this token (for audit purposes)"
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
                            <strong>Warning:</strong> Revoking this token will immediately stop all SCIM provisioning using it.
                        </AlertDescription>
                    </Alert>

                    <DialogFooter>
                        <Button
                            variant="outline"
                            onClick={() => {
                                setRevokeToken(null);
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
                            {revokeMutation.isPending ? 'Revoking...' : 'Revoke Token'}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </>
    );
}
