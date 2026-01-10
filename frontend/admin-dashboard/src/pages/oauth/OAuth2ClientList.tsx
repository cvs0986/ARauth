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
import { Key, Plus, AlertTriangle } from 'lucide-react';

export function OAuth2ClientList() {
    const { principalType, homeTenantId, selectedTenantId } = usePrincipalContext();
    const queryClient = useQueryClient();
    const [createOpen, setCreateOpen] = useState(false);

    // Determine effective tenant ID
    const effectiveTenantId = principalType === 'SYSTEM' ? selectedTenantId : homeTenantId;

    // Fetch OAuth2 clients for tenant
    const { data: clients, isLoading, error } = useQuery({
        queryKey: ['oauth-clients', effectiveTenantId],
        queryFn: async () => {
            // TODO: Implement API call
            // return oauthApi.listClients(effectiveTenantId);
            return [];
        },
        enabled: !!effectiveTenantId,
    });

    if (!effectiveTenantId) {
        return (
            <div className="space-y-4">
                <h1 className="text-3xl font-bold">OAuth2 Clients</h1>
                <Alert className="bg-yellow-50 border-yellow-200">
                    <AlertTriangle className="h-4 w-4 text-yellow-600" />
                    <AlertDescription className="text-yellow-800">
                        Select a tenant to manage OAuth2 clients
                    </AlertDescription>
                </Alert>
            </div>
        );
    }

    if (isLoading) {
        return <div className="p-4">Loading OAuth2 clients...</div>;
    }

    if (error) {
        return (
            <div className="p-4 text-red-600">
                Error loading OAuth2 clients: {error instanceof Error ? error.message : 'Unknown error'}
            </div>
        );
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

                {/* Coming Soon Notice */}
                <Alert className="bg-blue-50 border-blue-200">
                    <Key className="h-4 w-4 text-blue-600" />
                    <AlertDescription className="text-blue-800">
                        <strong>Coming Soon:</strong> OAuth2 client management API is not yet exposed.
                        This interface will allow you to create and manage OAuth2/OIDC clients for your tenant.
                    </AlertDescription>
                </Alert>

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
                            {clients && clients.length === 0 && (
                                <TableRow>
                                    <TableCell colSpan={6} className="text-center text-gray-500">
                                        No OAuth2 clients configured. Create your first client to get started.
                                    </TableCell>
                                </TableRow>
                            )}
                        </TableBody>
                    </Table>
                </div>
            </div>
        </PermissionGate>
    );
}
