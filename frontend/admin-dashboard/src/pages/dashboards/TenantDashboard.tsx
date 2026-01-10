/**
 * Tenant Dashboard
 * 
 * GUARDRAIL #1: Backend Is Law
 * - All metrics from backend APIs
 * 
 * GUARDRAIL #4: Data Gaps Explicit
 * - User activity timeline not available (show "Coming Soon")
 * - MFA enrollment rate calculation not available (show "Coming Soon")
 * 
 * GUARDRAIL #6: UI Quality Bar
 * - Informational, not functional
 * - Control plane aesthetic
 */

import { useQuery } from '@tanstack/react-query';
import { userApi, roleApi, permissionApi } from '@/services/api';
import { usePrincipalContext } from '@/contexts/PrincipalContext';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { MetricCard } from '@/components/dashboard/MetricCard';
import { useNavigate } from 'react-router-dom';
import {
    Users,
    Shield,
    Key,
    Activity,
    Lock
} from 'lucide-react';

export function TenantDashboard() {
    const navigate = useNavigate();
    const { homeTenantId, selectedTenantId, principalType } = usePrincipalContext();

    // Determine which tenant to show
    // SYSTEM users with selected tenant: show that tenant
    // TENANT users: show their home tenant
    const effectiveTenantId = principalType === 'SYSTEM' ? selectedTenantId : homeTenantId;

    // Fetch tenant data
    const { data: users, isLoading: usersLoading } = useQuery({
        queryKey: ['users', effectiveTenantId],
        queryFn: () => userApi.list(effectiveTenantId!),
        enabled: !!effectiveTenantId,
    });

    const { data: roles, isLoading: rolesLoading } = useQuery({
        queryKey: ['roles', effectiveTenantId],
        queryFn: () => roleApi.list(effectiveTenantId!),
        enabled: !!effectiveTenantId,
    });

    const { data: permissions, isLoading: permissionsLoading } = useQuery({
        queryKey: ['permissions', effectiveTenantId],
        queryFn: () => permissionApi.list(effectiveTenantId!),
        enabled: !!effectiveTenantId,
    });

    const isLoading = usersLoading || rolesLoading || permissionsLoading;

    if (!effectiveTenantId) {
        return (
            <div className="space-y-4">
                <h1 className="text-3xl font-bold">Tenant Dashboard</h1>
                <Card>
                    <CardContent className="py-8 text-center text-gray-500">
                        <p>Select a tenant from the header to view tenant dashboard</p>
                    </CardContent>
                </Card>
            </div>
        );
    }

    if (isLoading) {
        return (
            <div className="space-y-4">
                <h1 className="text-3xl font-bold">Tenant Dashboard</h1>
                <div className="text-center py-8">Loading tenant metrics...</div>
            </div>
        );
    }

    // Calculate stats from available data
    const stats = {
        totalUsers: users?.length || 0,
        activeUsers: users?.filter(u => u.status === 'active').length || 0,
        totalRoles: roles?.length || 0,
        totalPermissions: permissions?.length || 0,
    };

    return (
        <div className="space-y-6">
            {/* Header */}
            <div>
                <h1 className="text-3xl font-bold">Tenant Dashboard</h1>
                <p className="text-gray-600 mt-1">Operational overview for this tenant</p>
            </div>

            {/* Metrics Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                <MetricCard
                    title="Total Users"
                    value={stats.totalUsers}
                    subtitle={`${stats.activeUsers} active`}
                    icon={Users}
                    variant="primary"
                    trend={{
                        value: stats.totalUsers > 0
                            ? Math.round((stats.activeUsers / stats.totalUsers) * 100)
                            : 0,
                        label: 'active',
                        isPositive: true,
                    }}
                    onClick={() => navigate('/users')}
                />

                {/* GUARDRAIL #4: Data gap - MFA enrollment rate */}
                <MetricCard
                    title="MFA Enrolled"
                    value="—"
                    icon={Lock}
                    variant="success"
                    comingSoon
                />

                <MetricCard
                    title="Roles"
                    value={stats.totalRoles}
                    icon={Shield}
                    variant="default"
                    onClick={() => navigate('/roles')}
                />

                <MetricCard
                    title="Permissions"
                    value={stats.totalPermissions}
                    icon={Key}
                    variant="default"
                    onClick={() => navigate('/permissions')}
                />
            </div>

            {/* Security Overview */}
            <Card>
                <CardHeader>
                    <CardTitle>Security Overview</CardTitle>
                    <CardDescription>Current security posture</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="flex items-center justify-between py-2">
                        <div className="flex items-center gap-2">
                            <Activity className="h-5 w-5 text-green-500" />
                            <span className="font-medium">Tenant Status</span>
                        </div>
                        <span className="px-3 py-1 bg-green-100 text-green-800 rounded-md text-sm font-medium">
                            Active
                        </span>
                    </div>
                    <div className="border-t pt-4 space-y-2">
                        <div className="flex justify-between text-sm">
                            <span className="text-gray-600">Total Users</span>
                            <span className="font-medium">{stats.totalUsers}</span>
                        </div>
                        <div className="flex justify-between text-sm">
                            <span className="text-gray-600">Active Users</span>
                            <span className="font-medium">{stats.activeUsers}</span>
                        </div>
                        <div className="flex justify-between text-sm">
                            <span className="text-gray-600">Total Roles</span>
                            <span className="font-medium">{stats.totalRoles}</span>
                        </div>
                        <div className="flex justify-between text-sm">
                            <span className="text-gray-600">Total Permissions</span>
                            <span className="font-medium">{stats.totalPermissions}</span>
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* User Activity - GUARDRAIL #4: Data gap */}
            <Card>
                <CardHeader>
                    <CardTitle>User Activity (7 days)</CardTitle>
                    <CardDescription>Login and API activity trends</CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="text-center py-8 text-gray-500">
                        <Activity className="h-8 w-8 mx-auto mb-2 opacity-50" />
                        <p className="text-sm font-medium">Coming Soon</p>
                        <p className="text-xs mt-1">User activity metrics API not yet implemented</p>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}
