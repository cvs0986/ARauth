/**
 * OAuth2 Client List Page
 * 
 * AUTHORITY MODEL:
 * Who: TENANT users OR SYSTEM users (tenant selected)
 * Scope: Tenant-scoped OAuth2 clients
 * Permission: oauth:clients:read
 * 
 * GUARDRAILS:
 * - No token minting
 * - No assuming grant behavior
 * - No default scopes
 * - Secret rotation requires confirmation + audit
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
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Key, Plus, RefreshCw, Trash2 } from 'lucide-react';
import { isAPINotConnected } from '@/lib/errors';
import { CreateOAuth2ClientDialog } from './CreateOAuth2ClientDialog';
import { oauthClientApi } from '../../services/api';
// @ts-ignore
import { OAuth2Client } from '../../../../shared/types/api';

export function OAuth2ClientList() {
    const { principalType, homeTenantId, selectedTenantId, consoleMode } = usePrincipalContext();
    const queryClient = useQueryClient();
    const [createOpen, setCreateOpen] = useState(false);
    const [rotatingId, setRotatingId] = useState<string | null>(null);

    // Determine effective tenant ID
    const effectiveTenantId = principalType === 'SYSTEM' ? (selectedTenantId || undefined) : (homeTenantId || undefined);

    // Fetch OAuth2 clients for tenant
    const { data: clients, isLoading, error } = useQuery({
        queryKey: ['oauth-clients', effectiveTenantId],
        queryFn: async () => {
            return await oauthClientApi.list(effectiveTenantId);
        },
        enabled: consoleMode === 'SYSTEM' ? true : !!effectiveTenantId,
        retry: false,
    });

    const deleteMutation = useMutation({
        mutationFn: async (id: string) => {
            await oauthClientApi.delete(id);
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['oauth-clients', effectiveTenantId] });
        },
    });

    const rotateMutation = useMutation({
        mutationFn: async (id: string) => {
            setRotatingId(id);
            try {
                const response = await oauthClientApi.rotateSecret(id);
                // Show new secret
                // Ideally this would be a nice dialog, but alert for now as per plan
                prompt('Secret Rotated Successfully. Copy the new secret now:', response.client_secret);
            } catch (err: any) {
                alert(`Failed to rotate secret: ${err.message}`);
            } finally {
                setRotatingId(null);
            }
        },
    });

    const handleDelete = (id: string) => {
        if (confirm('Are you sure you want to delete this client application?')) {
            deleteMutation.mutate(id);
        }
    };

    const handleRotateSecret = (id: string) => {
        if (confirm('Are you sure you want to rotate the client secret? The old secret will stop working immediately.')) {
            rotateMutation.mutate(id);
        }
    };

    if (isLoading) {
        return <div className="p-4">Loading OAuth2 clients...</div>;
    }

    return (
        <PermissionGate permission="oauth:clients:read" systemPermission={principalType === 'SYSTEM'}>
            <div className="space-y-4">
                {/* Header */}
                <div className="flex items-center justify-between">
                    <div>
                        <h1 className="text-3xl font-bold">OAuth2 Clients</h1>
                        <p className="text-sm text-gray-600 mt-1">
                            Manage OAuth2 and OpenID Connect client applications
                        </p>
                    </div>
                    <PermissionGate permission="oauth:clients:create" systemPermission={principalType === 'SYSTEM'}>
                        <Button onClick={() => setCreateOpen(true)}>
                            <Plus className="h-4 w-4 mr-2" />
                            Create Client
                        </Button>
                    </PermissionGate>
                </div>

                {/* API Not Connected Notice */}
                {error && isAPINotConnected(error) && (
                    <Alert className="bg-blue-50 border-blue-200">
                        <Key className="h-4 w-4 text-blue-600" />
                        <AlertDescription className="text-blue-800">
                            <strong>Backend Integration Pending:</strong> The OAuth client API is not yet connected.
                            This UI serves as the contract for implementation.
                        </AlertDescription>
                    </Alert>
                )}

                {/* Table */}
                <div className="border rounded-lg">
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Client Name</TableHead>
                                <TableHead>Client ID</TableHead>
                                <TableHead>Grant Types</TableHead>
                                <TableHead>Redirect URIs</TableHead>
                                <TableHead>Created</TableHead>
                                <TableHead className="text-right">Actions</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {!clients || clients.length === 0 ? (
                                <TableRow>
                                    <TableCell colSpan={6} className="text-center text-gray-500 py-8">
                                        No OAuth2 clients configured. Create your first client to get started.
                                    </TableCell>
                                </TableRow>
                            ) : (
                                clients.map((client: OAuth2Client) => (
                                    <TableRow key={client.id}>
                                        <TableCell className="font-medium">{client.name}</TableCell>
                                        <TableCell className="font-mono text-sm">{client.client_id}</TableCell>
                                        <TableCell>
                                            <div className="flex flex-wrap gap-1">
                                                {client.grant_types.map((g: string) => (
                                                    <Badge key={g} variant="outline" className="text-xs">
                                                        {g}
                                                    </Badge>
                                                ))}
                                            </div>
                                        </TableCell>
                                        <TableCell>
                                            <div className="flex flex-col gap-1">
                                                {client.redirect_uris.map((uri: string) => (
                                                    <span key={uri} className="font-mono text-xs text-gray-600 truncate max-w-[200px]" title={uri}>
                                                        {uri}
                                                    </span>
                                                ))}
                                            </div>
                                        </TableCell>
                                        <TableCell>{new Date(client.created_at).toLocaleDateString()}</TableCell>
                                        <TableCell className="text-right">
                                            <div className="flex justify-end gap-2">
                                                <PermissionGate permission="oauth:clients:rotate_secret" systemPermission={principalType === 'SYSTEM'}>
                                                    <Button
                                                        variant="ghost"
                                                        size="sm"
                                                        onClick={() => handleRotateSecret(client.id)}
                                                        disabled={rotatingId === client.id}
                                                        title="Rotate Secret"
                                                    >
                                                        <RefreshCw className={`h-4 w-4 ${rotatingId === client.id ? 'animate-spin' : ''}`} />
                                                    </Button>
                                                </PermissionGate>
                                                <PermissionGate permission="oauth:clients:delete" systemPermission={principalType === 'SYSTEM'}>
                                                    <Button
                                                        variant="ghost"
                                                        size="sm"
                                                        onClick={() => handleDelete(client.id)}
                                                        className="text-red-600 hover:text-red-700 hover:bg-red-50"
                                                    >
                                                        <Trash2 className="h-4 w-4" />
                                                    </Button>
                                                </PermissionGate>
                                            </div>
                                        </TableCell>
                                    </TableRow>
                                ))
                            )}
                        </TableBody>
                    </Table>
                </div>

                {/* Create Dialog */}
                <CreateOAuth2ClientDialog
                    open={createOpen}
                    onOpenChange={setCreateOpen}
                    tenantId={effectiveTenantId || ''}
                />
            </div>
        </PermissionGate>
    );
}
