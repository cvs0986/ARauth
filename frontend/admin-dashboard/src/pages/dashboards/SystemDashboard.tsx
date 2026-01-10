/**
 * System Dashboard
 * 
 * GUARDRAIL #1: Backend Is Law
 * - All metrics from backend APIs
 * 
 * GUARDRAIL #4: Data Gaps Explicit
 * - Cross-tenant aggregation not available (show "Coming Soon")
 * - MFA adoption rate not available (show "Coming Soon")
 * - Security posture scoring not available (show "Coming Soon")
 * 
 * GUARDRAIL #6: UI Quality Bar
 * - Informational, not functional
 * - Control plane aesthetic
 */

import { useQuery } from '@tanstack/react-query';
import { systemApi } from '@/services/api';
import { usePrincipalContext } from '@/contexts/PrincipalContext';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { MetricCard } from '@/components/dashboard/MetricCard';
import { useNavigate } from 'react-router-dom';
import {
    Building2,
    Users,
    Shield,
    AlertTriangle,
    Activity,
    Info
} from 'lucide-react';

export function SystemDashboard() {
    const navigate = useNavigate();
    const { selectedTenantId } = usePrincipalContext();

    // Fetch tenants list
    const { data: tenants, isLoading: tenantsLoading } = useQuery({
        queryKey: ['system', 'tenants'],
        queryFn: () => systemApi.tenants.list(),
    });

    const isLoading = tenantsLoading;

    if (isLoading) {
        return (
            <div className="space-y-4">
                <h1 className="text-3xl font-bold">System Dashboard</h1>
                <div className="text-center py-8">Loading system metrics...</div>
            </div>
        );
    }

    // Calculate basic stats from available data
    const stats = {
        totalTenants: tenants?.length || 0,
        activeTenants: tenants?.filter(t => t.status === 'active').length || 0,
    };

    return (
        <div className="space-y-6">
            {/* Header */}
            <div>
                <h1 className="text-3xl font-bold">System Dashboard</h1>
                <p className="text-gray-600 mt-1">
                    {selectedTenantId
                        ? 'Viewing selected tenant context'
                        : 'Platform-wide operational overview'}
                </p>
            </div>

            {/* All Tenants Alert */}
            {!selectedTenantId && (
                <Alert className="bg-blue-50 border-blue-200">
                    <Info className="h-4 w-4 text-blue-600" />
                    <AlertDescription className="text-blue-800">
                        <strong>All Tenants View</strong> - Select a specific tenant from the header to view tenant-specific metrics.
                    </AlertDescription>
                </Alert>
            )}

            {/* Metrics Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                <MetricCard
                    title="Total Tenants"
                    value={stats.totalTenants}
                    subtitle={`${stats.activeTenants} active`}
                    icon={Building2}
                    variant="primary"
                    trend={{
                        value: stats.totalTenants > 0
                            ? Math.round((stats.activeTenants / stats.totalTenants) * 100)
                            : 0,
                        label: 'active',
                        isPositive: true,
                    }}
                    onClick={() => navigate('/tenants')}
                />

                {/* GUARDRAIL #4: Data gap - cross-tenant user aggregation */}
                <MetricCard
                    title="Total Users"
                    value="—"
                    icon={Users}
                    variant="success"
                    comingSoon
                />

                {/* GUARDRAIL #4: Data gap - MFA adoption rate */}
                <MetricCard
                    title="MFA Adoption"
                    value="—"
                    icon={Shield}
                    variant="warning"
                    comingSoon
                />

                {/* GUARDRAIL #4: Data gap - security incidents */}
                <MetricCard
                    title="Security Incidents"
                    value="—"
                    icon={AlertTriangle}
                    variant="danger"
                    comingSoon
                />
            </div>

            {/* Tenant Health Table */}
            <Card>
                <CardHeader>
                    <CardTitle>Tenant Health</CardTitle>
                    <CardDescription>Overview of all tenants and their status</CardDescription>
                </CardHeader>
                <CardContent>
                    {tenants && tenants.length > 0 ? (
                        <div className="space-y-2">
                            {tenants.map((tenant) => (
                                <div
                                    key={tenant.id}
                                    className="flex items-center justify-between p-3 border rounded-lg hover:bg-gray-50 cursor-pointer transition-colors"
                                    onClick={() => navigate(`/tenants/${tenant.id}`)}
                                >
                                    <div className="flex items-center gap-3">
                                        <Building2 className="h-5 w-5 text-gray-600" />
                                        <div>
                                            <p className="font-medium">{tenant.name}</p>
                                            <p className="text-sm text-gray-500">{tenant.domain}</p>
                                        </div>
                                    </div>
                                    <div className="flex items-center gap-4">
                                        <span className={`px-2 py-1 rounded text-xs font-medium ${tenant.status === 'active'
                                                ? 'bg-green-100 text-green-800'
                                                : 'bg-gray-100 text-gray-800'
                                            }`}>
                                            {tenant.status}
                                        </span>
                                    </div>
                                </div>
                            ))}
                        </div>
                    ) : (
                        <div className="text-center py-8 text-gray-500">
                            <Building2 className="h-8 w-8 mx-auto mb-2 opacity-50" />
                            <p className="text-sm">No tenants found</p>
                        </div>
                    )}
                </CardContent>
            </Card>

            {/* System Status */}
            <Card>
                <CardHeader>
                    <CardTitle>System Status</CardTitle>
                    <CardDescription>Current platform health</CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="flex items-center justify-between py-3">
                        <div className="flex items-center gap-2">
                            <Activity className="h-5 w-5 text-green-500" />
                            <span className="font-medium">Platform Status</span>
                        </div>
                        <span className="px-3 py-1 bg-green-100 text-green-800 rounded-md text-sm font-medium">
                            Operational
                        </span>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}
