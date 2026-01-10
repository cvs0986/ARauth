/**
 * Audit Log List Page
 * 
 * AUTHORITY MODEL:
 * Who: SYSTEM users (all events) OR TENANT users (tenant events)
 * Scope: SYSTEM mode (cross-tenant) OR TENANT mode (tenant-scoped)
 * Permission: audit:read
 * 
 * OBSERVABILITY:
 * - Advanced filtering (actor, action, result, time range)
 * - Export capability (CSV, JSON)
 * - Operator-grade compliance tool
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
import { FileText, Download, Filter } from 'lucide-react';
import { APINotConnectedError, isAPINotConnected } from '@/lib/errors';

interface AuditLog {
    id: string;
    timestamp: string;
    actor: string;
    action: string;
    target: string;
    result: 'success' | 'failure';
    ip_address: string;
    user_agent: string;
}

export function AuditLogList() {
    const { principalType, homeTenantId, selectedTenantId, consoleMode } = usePrincipalContext();
    const [exportOpen, setExportOpen] = useState(false);

    // Determine effective tenant ID
    const effectiveTenantId = principalType === 'SYSTEM' ? selectedTenantId : homeTenantId;

    // Fetch audit logs
    const { data: logs, isLoading, error } = useQuery({
        queryKey: consoleMode === 'SYSTEM' && !effectiveTenantId
            ? ['system-audit-logs']
            : ['audit-logs', effectiveTenantId],
        queryFn: async () => {
            // UI CONTRACT MODE: Throw APINotConnectedError
            throw new APINotConnectedError('audit.logs.list');
        },
        enabled: consoleMode === 'SYSTEM' ? true : !!effectiveTenantId,
        retry: false,
    });

    return (
        <PermissionGate permission="audit:read" systemPermission={principalType === 'SYSTEM'}>
            <div className="space-y-6">
                {/* Header */}
                <div className="flex items-center justify-between">
                    <div>
                        <h1 className="text-3xl font-bold">Audit Logs</h1>
                        <p className="text-sm text-gray-600 mt-1">
                            {consoleMode === 'SYSTEM' && !effectiveTenantId
                                ? 'View all system and tenant audit events'
                                : 'View audit events for this tenant'}
                        </p>
                    </div>
                    <div className="flex gap-2">
                        <Button variant="outline" size="sm" disabled>
                            <Filter className="h-4 w-4 mr-2" />
                            Filters
                        </Button>
                        <PermissionGate permission="audit:export" systemPermission={principalType === 'SYSTEM'}>
                            <Button variant="outline" size="sm" onClick={() => setExportOpen(true)}>
                                <Download className="h-4 w-4 mr-2" />
                                Export
                            </Button>
                        </PermissionGate>
                    </div>
                </div>

                {/* Scope Indicator */}
                <Alert className="bg-blue-50 border-blue-200">
                    <FileText className="h-4 w-4 text-blue-600" />
                    <AlertDescription className="text-blue-800 text-sm">
                        {consoleMode === 'SYSTEM' && !effectiveTenantId ? (
                            <><strong>Scope:</strong> Viewing all system and tenant audit events</>
                        ) : (
                            <><strong>Scope:</strong> Viewing audit events for this tenant only</>
                        )}
                    </AlertDescription>
                </Alert>

                {/* API Not Connected Notice */}
                {error && isAPINotConnected(error) && (
                    <Alert className="bg-blue-50 border-blue-200">
                        <FileText className="h-4 w-4 text-blue-600" />
                        <AlertDescription className="text-blue-800">
                            <strong>Backend Integration Pending:</strong> The audit logs API is not yet connected.
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
                ) : !logs || logs.length === 0 ? (
                    <EmptyState
                        icon={FileText}
                        title="No Audit Logs"
                        description="Audit logs will appear here as actions are performed in the system."
                    />
                ) : (
                    <div className="border rounded-lg">
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Timestamp</TableHead>
                                    <TableHead>Actor</TableHead>
                                    <TableHead>Action</TableHead>
                                    <TableHead>Target</TableHead>
                                    <TableHead>Result</TableHead>
                                    <TableHead>IP Address</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {logs.map((log) => (
                                    <TableRow key={log.id}>
                                        <TableCell className="font-mono text-sm">
                                            {new Date(log.timestamp).toLocaleString()}
                                        </TableCell>
                                        <TableCell className="font-medium">{log.actor}</TableCell>
                                        <TableCell>
                                            <Badge variant="outline" className="font-mono text-xs">
                                                {log.action}
                                            </Badge>
                                        </TableCell>
                                        <TableCell className="font-mono text-sm">{log.target}</TableCell>
                                        <TableCell>
                                            <Badge variant={log.result === 'success' ? 'default' : 'destructive'}>
                                                {log.result}
                                            </Badge>
                                        </TableCell>
                                        <TableCell className="font-mono text-sm">{log.ip_address}</TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </div>
                )}
            </div>
        </PermissionGate>
    );
}
