/**
 * External OIDC Identity Providers List
 * 
 * AUTHORITY MODEL:
 * Who: TENANT users OR SYSTEM users (tenant selected)
 * Scope: Tenant-scoped external IdPs
 * Permission: federation:idp:read
 * 
 * SECURITY:
 * - Identity-safe design
 * - No automatic linking
 * - Clear login impact warnings
 * 
 * UI CONTRACT MODE:
 * - All backend calls throw APINotConnectedError
 */

import { useState } from 'react';
import { useQuery, useMutation } from '@tanstack/react-query';
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
import { Alert, AlertDescription } from '@/components/ui/alert';
import { EmptyState } from '@/components/EmptyState';
import { Plus, AlertTriangle, Link2, CheckCircle2, Loader2 } from 'lucide-react';
import { isAPINotConnected } from '@/lib/errors';
import { CreateOIDCIdPDialog } from './CreateOIDCIdPDialog';
import { federationApi } from '../../services/api';
// @ts-ignore
import { IdentityProvider } from '../../../../shared/types/api';

export function OIDCIdPList() {
    const { principalType, homeTenantId, selectedTenantId } = usePrincipalContext();
    const [createOpen, setCreateOpen] = useState(false);
    const [verifyingId, setVerifyingId] = useState<string | null>(null);

    // Determine effective tenant ID
    const effectiveTenantId = principalType === 'SYSTEM' ? (selectedTenantId || undefined) : (homeTenantId || undefined);

    // Fetch OIDC IdPs
    const { data: idps, isLoading, error } = useQuery({
        queryKey: ['oidc-idps', effectiveTenantId],
        queryFn: async () => {
            return await federationApi.list(effectiveTenantId);
        },
        enabled: !!effectiveTenantId,
        retry: false,
    });

    const verifyMutation = useMutation({
        mutationFn: async (id: string) => {
            setVerifyingId(id);
            try {
                const result = await federationApi.verify(id);
                if (result.success) {
                    alert(`Connection Successful: ${result.message}`);
                } else {
                    alert(`Connection Failed: ${result.message}\nError: ${result.error || 'Unknown error'}`);
                }
            } catch (err: any) {
                alert(`Connection Error: ${err.message || 'Failed to connect'}`);
            } finally {
                setVerifyingId(null);
            }
        },
    });

    const handleVerify = (id: string) => {
        verifyMutation.mutate(id);
    };

    if (!effectiveTenantId) {
        return (
            <div className="space-y-4">
                <h1 className="text-3xl font-bold">External OIDC Providers</h1>
                <Alert className="bg-yellow-50 border-yellow-200">
                    <AlertTriangle className="h-4 w-4 text-yellow-600" />
                    <AlertDescription className="text-yellow-800">
                        Select a tenant to manage external identity providers
                    </AlertDescription>
                </Alert>
            </div>
        );
    }

    // Filter only OIDC providers (assuming backend returns all types)
    const oidcProviders = idps?.filter((idp: IdentityProvider) => idp.type === 'oidc') || [];

    return (
        <PermissionGate permission="federation:idp:read" systemPermission={principalType === 'SYSTEM'}>
            <div className="space-y-6">
                {/* Header */}
                <div className="flex items-center justify-between">
                    <div>
                        <h1 className="text-3xl font-bold">External OIDC Providers</h1>
                        <p className="text-sm text-gray-600 mt-1">
                            Configure external OpenID Connect identity providers for federated authentication
                        </p>
                    </div>
                    <PermissionGate permission="federation:idp:create" systemPermission={principalType === 'SYSTEM'}>
                        <Button onClick={() => setCreateOpen(true)}>
                            <Plus className="h-4 w-4 mr-2" />
                            Add OIDC Provider
                        </Button>
                    </PermissionGate>
                </div>

                {/* Security Notice */}
                <Alert className="bg-orange-50 border-orange-200">
                    <AlertTriangle className="h-4 w-4 text-orange-600" />
                    <AlertDescription className="text-orange-800 text-sm">
                        <strong>Security:</strong> External identity providers allow users to authenticate using external accounts.
                        Misconfiguration can impact user login. Always test connections before enabling.
                    </AlertDescription>
                </Alert>

                {/* API Not Connected Notice */}
                {error && isAPINotConnected(error) && (
                    <Alert className="bg-blue-50 border-blue-200">
                        <Link2 className="h-4 w-4 text-blue-600" />
                        <AlertDescription className="text-blue-800">
                            <strong>Backend Integration Pending:</strong> The federation API is not yet connected.
                            This UI serves as the contract for implementation.
                        </AlertDescription>
                    </Alert>
                )}

                {/* Table */}
                {isLoading ? (
                    <div className="p-4">Loading OIDC providers...</div>
                ) : !oidcProviders || oidcProviders.length === 0 ? (
                    <EmptyState
                        icon={Link2}
                        title="No OIDC Providers Configured"
                        description="Add an external OpenID Connect provider to enable federated authentication for your users."
                        action={{
                            label: 'Add OIDC Provider',
                            onClick: () => setCreateOpen(true),
                        }}
                    />
                ) : (
                    <div className="border rounded-lg">
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Provider Name</TableHead>
                                    <TableHead>Issuer URL</TableHead>
                                    <TableHead>Client ID</TableHead>
                                    <TableHead>Status</TableHead>
                                    <TableHead>Created</TableHead>
                                    <TableHead className="text-right">Actions</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {oidcProviders.map((idp: IdentityProvider) => (
                                    <TableRow key={idp.id}>
                                        <TableCell className="font-medium">{idp.name}</TableCell>
                                        <TableCell className="font-mono text-sm">
                                            {String(idp.configuration?.issuer_url || '-')}
                                        </TableCell>
                                        <TableCell className="font-mono text-sm">
                                            {String(idp.configuration?.client_id || '-')}
                                        </TableCell>
                                        <TableCell>
                                            <Badge variant={idp.enabled ? 'default' : 'secondary'}>
                                                {idp.enabled ? 'Active' : 'Disabled'}
                                            </Badge>
                                        </TableCell>
                                        <TableCell>{new Date(idp.created_at).toLocaleDateString()}</TableCell>
                                        <TableCell className="text-right">
                                            <div className="flex justify-end gap-2">
                                                <Button
                                                    variant="outline"
                                                    size="sm"
                                                    onClick={() => handleVerify(idp.id)}
                                                    disabled={verifyingId === idp.id}
                                                >
                                                    {verifyingId === idp.id ? (
                                                        <Loader2 className="h-4 w-4 animate-spin mr-2" />
                                                    ) : (
                                                        <CheckCircle2 className="h-4 w-4 mr-2" />
                                                    )}
                                                    Test Connection
                                                </Button>
                                                <Button variant="outline" size="sm">
                                                    Configure
                                                </Button>
                                            </div>
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </div>
                )}

                {/* Create Dialog */}
                <CreateOIDCIdPDialog
                    open={createOpen}
                    onOpenChange={setCreateOpen}
                    tenantId={effectiveTenantId || ''}
                />
            </div>
        </PermissionGate>
    );
}
