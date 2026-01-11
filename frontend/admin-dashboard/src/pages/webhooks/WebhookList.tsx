/**
 * Webhook List Page
 * 
 * AUTHORITY MODEL:
 * Who: SYSTEM users (system webhooks) OR TENANT users (tenant webhooks)
 * Scope: SYSTEM mode (platform-wide) OR TENANT mode (tenant-scoped)
 * Permission: webhooks:read
 * 
 * SECURITY:
 * - Webhooks receive sensitive data
 * - One-time signing secret display
 * - Delete requires audit reason
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
import { Alert, AlertDescription } from '@/components/ui/alert';
import { EmptyState } from '@/components/EmptyState';
import { Skeleton } from '@/components/ui/skeleton';
import { Plus, Webhook as WebhookIcon, AlertTriangle, Trash2 } from 'lucide-react';
import { isAPINotConnected } from '@/lib/errors';
import { CreateWebhookDialog } from './CreateWebhookDialog';
import { webhookApi } from '../../services/api';
// @ts-ignore
import { Webhook } from '../../../../shared/types/api';

export function WebhookList() {
    const { principalType, homeTenantId, selectedTenantId, consoleMode } = usePrincipalContext();
    const [createOpen, setCreateOpen] = useState(false);
    const queryClient = useQueryClient();

    // Determine effective tenant ID
    const effectiveTenantId = principalType === 'SYSTEM' ? (selectedTenantId || undefined) : (homeTenantId || undefined);

    // Fetch webhooks
    const { data: webhooks, isLoading, error } = useQuery({
        queryKey: consoleMode === 'SYSTEM' && !effectiveTenantId
            ? ['system-webhooks']
            : ['webhooks', effectiveTenantId],
        queryFn: async () => {
            return await webhookApi.list(effectiveTenantId);
        },
        enabled: consoleMode === 'SYSTEM' ? true : !!effectiveTenantId,
        retry: false,
    });

    const deleteMutation = useMutation({
        mutationFn: async (id: string) => {
            await webhookApi.delete(id);
        },
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: consoleMode === 'SYSTEM' && !effectiveTenantId
                    ? ['system-webhooks']
                    : ['webhooks', effectiveTenantId]
            });
        },
    });

    const handleDelete = (id: string) => {
        if (confirm('Are you sure you want to delete this webhook?')) {
            deleteMutation.mutate(id);
        }
    };

    return (
        <PermissionGate permission="webhooks:read" systemPermission={principalType === 'SYSTEM'}>
            <div className="space-y-6">
                {/* Header */}
                <div className="flex items-center justify-between">
                    <div>
                        <h1 className="text-3xl font-bold">Webhooks</h1>
                        <p className="text-sm text-gray-600 mt-1">
                            {consoleMode === 'SYSTEM' && !effectiveTenantId
                                ? 'Manage system-wide webhook subscriptions'
                                : 'Manage webhook subscriptions for this tenant'}
                        </p>
                    </div>
                    <PermissionGate permission="webhooks:create" systemPermission={principalType === 'SYSTEM'}>
                        <Button onClick={() => setCreateOpen(true)}>
                            <Plus className="h-4 w-4 mr-2" />
                            Create Webhook
                        </Button>
                    </PermissionGate>
                </div>

                {/* Scope Indicator */}
                <Alert className="bg-blue-50 border-blue-200">
                    <WebhookIcon className="h-4 w-4 text-blue-600" />
                    <AlertDescription className="text-blue-800 text-sm">
                        {consoleMode === 'SYSTEM' && !effectiveTenantId ? (
                            <><strong>Scope:</strong> System webhooks apply to all tenants</>
                        ) : (
                            <><strong>Scope:</strong> Webhooks apply only to this tenant</>
                        )}
                    </AlertDescription>
                </Alert>

                {/* Security Warning */}
                <Alert className="bg-orange-50 border-orange-200">
                    <AlertTriangle className="h-4 w-4 text-orange-600" />
                    <AlertDescription className="text-orange-800 text-sm">
                        <strong>Security:</strong> Webhooks receive sensitive event data. Ensure endpoints are secure and trusted.
                    </AlertDescription>
                </Alert>

                {/* API Not Connected Notice */}
                {error && isAPINotConnected(error) && (
                    <Alert className="bg-blue-50 border-blue-200">
                        <WebhookIcon className="h-4 w-4 text-blue-600" />
                        <AlertDescription className="text-blue-800">
                            <strong>Backend Integration Pending:</strong> The webhooks API is not yet connected.
                            This UI serves as the contract for implementation.
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
                ) : !webhooks || webhooks.length === 0 ? (
                    <EmptyState
                        icon={WebhookIcon}
                        title="No Webhooks Configured"
                        description="Create a webhook to receive real-time notifications for events in your system."
                        action={{
                            label: 'Create Webhook',
                            onClick: () => setCreateOpen(true),
                        }}
                    />
                ) : (
                    <div className="border rounded-lg">
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Name</TableHead>
                                    <TableHead>URL</TableHead>
                                    <TableHead>Events</TableHead>
                                    <TableHead>Status</TableHead>
                                    <TableHead>Created</TableHead>
                                    <TableHead className="text-right">Actions</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {webhooks.map((webhook: Webhook) => (
                                    <TableRow key={webhook.id}>
                                        <TableCell className="font-medium">{webhook.name}</TableCell>
                                        <TableCell className="font-mono text-sm">{webhook.url}</TableCell>
                                        <TableCell>
                                            <div className="flex flex-wrap gap-1">
                                                {webhook.events.slice(0, 2).map((event) => (
                                                    <Badge key={event} variant="outline" className="text-xs">
                                                        {event}
                                                    </Badge>
                                                ))}
                                                {webhook.events.length > 2 && (
                                                    <Badge variant="outline" className="text-xs">
                                                        +{webhook.events.length - 2}
                                                    </Badge>
                                                )}
                                            </div>
                                        </TableCell>
                                        <TableCell>
                                            <Badge variant={webhook.enabled ? 'default' : 'secondary'}>
                                                {webhook.enabled ? 'Active' : 'Disabled'}
                                            </Badge>
                                        </TableCell>
                                        <TableCell>{new Date(webhook.created_at).toLocaleDateString()}</TableCell>
                                        <TableCell className="text-right">
                                            <PermissionGate permission="webhooks:delete" systemPermission={principalType === 'SYSTEM'}>
                                                <Button
                                                    variant="ghost"
                                                    size="sm"
                                                    onClick={() => handleDelete(webhook.id)}
                                                    className="text-red-600 hover:text-red-700 hover:bg-red-50"
                                                >
                                                    <Trash2 className="h-4 w-4" />
                                                </Button>
                                            </PermissionGate>
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </div>
                )}

                {/* Create Dialog */}
                <CreateWebhookDialog
                    open={createOpen}
                    onOpenChange={setCreateOpen}
                    scope={consoleMode === 'SYSTEM' && !effectiveTenantId ? 'system' : 'tenant'}
                    tenantId={effectiveTenantId || undefined}
                />
            </div>
        </PermissionGate>
    );
}
