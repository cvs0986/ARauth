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
import { useQuery } from '@tanstack/react-query';
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
import { Plus, Webhook, AlertTriangle, Clock } from 'lucide-react';
import { APINotConnectedError, isAPINotConnected } from '@/lib/errors';
import { CreateWebhookDialog } from './CreateWebhookDialog';

interface Webhook {
    id: string;
    name: string;
    url: string;
    events: string[];
    status: 'active' | 'disabled';
    last_delivery_at: string | null;
    last_delivery_status: 'success' | 'failed' | null;
    created_at: string;
}

export function WebhookList() {
    const { principalType, homeTenantId, selectedTenantId, consoleMode } = usePrincipalContext();
    const [createOpen, setCreateOpen] = useState(false);

    // Determine effective tenant ID
    const effectiveTenantId = principalType === 'SYSTEM' ? selectedTenantId : homeTenantId;

    // Fetch webhooks
    const { data: webhooks, isLoading, error } = useQuery({
        queryKey: consoleMode === 'SYSTEM' && !effectiveTenantId
            ? ['system-webhooks']
            : ['webhooks', effectiveTenantId],
        queryFn: async () => {
            // UI CONTRACT MODE: Throw APINotConnectedError
            throw new APINotConnectedError('webhooks.list');
        },
        enabled: consoleMode === 'SYSTEM' ? true : !!effectiveTenantId,
        retry: false,
    });

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
                    <Webhook className="h-4 w-4 text-blue-600" />
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
                        <Webhook className="h-4 w-4 text-blue-600" />
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
                        icon={Webhook}
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
                                    <TableHead>Last Delivery</TableHead>
                                    <TableHead>Created</TableHead>
                                    <TableHead className="text-right">Actions</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {webhooks.map((webhook) => (
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
                                            <Badge variant={webhook.status === 'active' ? 'default' : 'secondary'}>
                                                {webhook.status}
                                            </Badge>
                                        </TableCell>
                                        <TableCell>
                                            {webhook.last_delivery_at ? (
                                                <div className="flex items-center gap-2">
                                                    <Clock className="h-4 w-4 text-gray-400" />
                                                    <span className="text-sm">
                                                        {new Date(webhook.last_delivery_at).toLocaleDateString()}
                                                    </span>
                                                    {webhook.last_delivery_status && (
                                                        <Badge
                                                            variant={webhook.last_delivery_status === 'success' ? 'default' : 'destructive'}
                                                            className="text-xs"
                                                        >
                                                            {webhook.last_delivery_status}
                                                        </Badge>
                                                    )}
                                                </div>
                                            ) : (
                                                <span className="text-gray-400">Never</span>
                                            )}
                                        </TableCell>
                                        <TableCell>{new Date(webhook.created_at).toLocaleDateString()}</TableCell>
                                        <TableCell className="text-right">
                                            <Button variant="outline" size="sm">
                                                Configure
                                            </Button>
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
