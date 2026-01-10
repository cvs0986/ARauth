/**
 * Identity Linking Component
 * 
 * AUTHORITY MODEL:
 * Who: TENANT users OR SYSTEM users (tenant selected)
 * Scope: Per-user external identity links
 * Permission: federation:link, federation:unlink
 * 
 * SECURITY:
 * - No automatic linking
 * - Link requires confirmation
 * - Unlink requires audit reason
 * - Clear provider visibility
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
import { Link2, Unlink, AlertTriangle, Plus } from 'lucide-react';
import { APINotConnectedError, getAPINotConnectedMessage, isAPINotConnected } from '@/lib/errors';

interface ExternalIdentity {
    id: string;
    provider_name: string;
    provider_type: 'oidc' | 'saml';
    external_id: string;
    linked_at: string;
}

interface IdentityLinkingProps {
    userId: string;
}

export function IdentityLinking({ userId }: IdentityLinkingProps) {
    const { principalType } = usePrincipalContext();
    const queryClient = useQueryClient();
    const [unlinkIdentity, setUnlinkIdentity] = useState<ExternalIdentity | null>(null);
    const [unlinkReason, setUnlinkReason] = useState('');
    const [error, setError] = useState<string | null>(null);

    // Fetch linked identities
    const { data: identities, isLoading } = useQuery({
        queryKey: ['user-identities', userId],
        queryFn: async () => {
            // UI CONTRACT MODE: Throw APINotConnectedError
            throw new APINotConnectedError('federation.identities.list');
        },
        enabled: !!userId,
        retry: false,
    });

    // Unlink identity mutation
    const unlinkMutation = useMutation({
        mutationFn: async ({ identityId, reason }: { identityId: string; reason: string }) => {
            // UI CONTRACT MODE: Throw APINotConnectedError
            throw new APINotConnectedError('federation.identities.unlink');
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['user-identities', userId] });
            setUnlinkIdentity(null);
            setUnlinkReason('');
            setError(null);
        },
        onError: (err: any) => {
            setError(getAPINotConnectedMessage(err));
        },
    });

    const handleUnlink = () => {
        if (!unlinkReason.trim()) {
            setError('Audit reason is required');
            return;
        }
        if (unlinkIdentity) {
            unlinkMutation.mutate({ identityId: unlinkIdentity.id, reason: unlinkReason });
        }
    };

    return (
        <PermissionGate permission="federation:link" systemPermission={principalType === 'SYSTEM'}>
            <div className="space-y-4">
                <div className="flex items-center justify-between">
                    <div>
                        <h3 className="text-lg font-semibold">Linked External Identities</h3>
                        <p className="text-sm text-gray-600">
                            External identity providers linked to this user account
                        </p>
                    </div>
                    <PermissionGate permission="federation:link" systemPermission={principalType === 'SYSTEM'}>
                        <Button size="sm" disabled>
                            <Plus className="h-4 w-4 mr-2" />
                            Link Identity
                        </Button>
                    </PermissionGate>
                </div>

                {/* API Not Connected Notice */}
                {isLoading && (
                    <Alert className="bg-blue-50 border-blue-200">
                        <Link2 className="h-4 w-4 text-blue-600" />
                        <AlertDescription className="text-blue-800">
                            <strong>Backend Integration Pending:</strong> Identity linking API is not yet connected.
                        </AlertDescription>
                    </Alert>
                )}

                {/* Table */}
                {!identities || identities.length === 0 ? (
                    <EmptyState
                        icon={Link2}
                        title="No External Identities Linked"
                        description="This user has not linked any external identity providers. Linked identities allow federated authentication."
                    />
                ) : (
                    <div className="border rounded-lg">
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Provider</TableHead>
                                    <TableHead>Type</TableHead>
                                    <TableHead>External ID</TableHead>
                                    <TableHead>Linked At</TableHead>
                                    <TableHead className="text-right">Actions</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {identities.map((identity) => (
                                    <TableRow key={identity.id}>
                                        <TableCell className="font-medium">{identity.provider_name}</TableCell>
                                        <TableCell>
                                            <Badge variant="outline" className="uppercase text-xs">
                                                {identity.provider_type}
                                            </Badge>
                                        </TableCell>
                                        <TableCell className="font-mono text-sm">{identity.external_id}</TableCell>
                                        <TableCell>{new Date(identity.linked_at).toLocaleDateString()}</TableCell>
                                        <TableCell className="text-right">
                                            <PermissionGate permission="federation:unlink" systemPermission={principalType === 'SYSTEM'}>
                                                <Button
                                                    variant="destructive"
                                                    size="sm"
                                                    onClick={() => setUnlinkIdentity(identity)}
                                                >
                                                    <Unlink className="h-4 w-4 mr-2" />
                                                    Unlink
                                                </Button>
                                            </PermissionGate>
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </div>
                )}

                {/* Unlink Dialog */}
                <Dialog open={!!unlinkIdentity} onOpenChange={(open) => !open && setUnlinkIdentity(null)}>
                    <DialogContent>
                        <DialogHeader>
                            <DialogTitle>Unlink External Identity</DialogTitle>
                            <DialogDescription>
                                Unlink "{unlinkIdentity?.provider_name}" from this user? This action cannot be undone.
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
                            <Label htmlFor="unlink-reason">Audit Reason (Required) *</Label>
                            <Textarea
                                id="unlink-reason"
                                value={unlinkReason}
                                onChange={(e) => setUnlinkReason(e.target.value)}
                                placeholder="Provide a reason for unlinking this identity (for audit purposes)"
                                rows={3}
                                disabled={unlinkMutation.isPending}
                            />
                            <p className="text-xs text-gray-500">
                                This reason will be recorded in the audit log
                            </p>
                        </div>

                        <Alert className="bg-orange-50 border-orange-200">
                            <AlertTriangle className="h-4 w-4 text-orange-600" />
                            <AlertDescription className="text-orange-800 text-sm">
                                <strong>Warning:</strong> Unlinking this identity will prevent the user from authenticating via this provider.
                            </AlertDescription>
                        </Alert>

                        <DialogFooter>
                            <Button
                                variant="outline"
                                onClick={() => {
                                    setUnlinkIdentity(null);
                                    setUnlinkReason('');
                                    setError(null);
                                }}
                                disabled={unlinkMutation.isPending}
                            >
                                Cancel
                            </Button>
                            <Button
                                variant="destructive"
                                onClick={handleUnlink}
                                disabled={unlinkMutation.isPending}
                            >
                                {unlinkMutation.isPending ? 'Unlinking...' : 'Unlink Identity'}
                            </Button>
                        </DialogFooter>
                    </DialogContent>
                </Dialog>
            </div>
        </PermissionGate>
    );
}
