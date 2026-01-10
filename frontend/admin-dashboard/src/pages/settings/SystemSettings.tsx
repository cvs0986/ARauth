/**
 * System Settings Page
 * 
 * AUTHORITY MODEL:
 * Who: SYSTEM users only
 * Scope: Platform-wide configuration
 * Permission: system:configure
 * Why: Platform infrastructure affects all tenants
 * 
 * GUARDRAIL #1: Backend Is Law
 * - All configuration from backend APIs
 * - No invented limits or defaults
 */

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { PermissionGate } from '@/components/PermissionGate';
import { Server, Key, Zap, Info } from 'lucide-react';

export function SystemSettings() {
    return (
        <PermissionGate permission="system:configure" systemPermission fallback={
            <div className="space-y-4">
                <h1 className="text-3xl font-bold">System Settings</h1>
                <Alert variant="destructive">
                    <AlertDescription>
                        You do not have permission to view system settings. Required permission: system:configure
                    </AlertDescription>
                </Alert>
            </div>
        }>
            <div className="space-y-6">
                {/* Header */}
                <div>
                    <h1 className="text-3xl font-bold">System Settings</h1>
                    <p className="text-gray-600 mt-1">Platform-wide configuration</p>
                    <Badge className="mt-2 bg-blue-100 text-blue-800">
                        <Server className="h-3 w-3 mr-1" />
                        Applies to all tenants
                    </Badge>
                </div>

                {/* Authority Notice */}
                <Alert className="bg-blue-50 border-blue-200">
                    <Info className="h-4 w-4 text-blue-600" />
                    <AlertDescription className="text-blue-800">
                        <strong>Platform-Level Settings</strong> - Changes here affect the entire ARauth platform and all tenants.
                    </AlertDescription>
                </Alert>

                {/* OAuth2/OIDC Configuration */}
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <Key className="h-5 w-5" />
                            OAuth2 / OIDC Provider
                        </CardTitle>
                        <CardDescription>
                            Platform-wide OAuth2 and OpenID Connect configuration
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="text-center py-8 text-gray-500">
                            <Key className="h-8 w-8 mx-auto mb-2 opacity-50" />
                            <p className="text-sm font-medium">Coming Soon</p>
                            <p className="text-xs mt-1">OAuth2 configuration API not yet exposed</p>
                        </div>
                    </CardContent>
                </Card>

                {/* JWT Configuration */}
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <Server className="h-5 w-5" />
                            JWT Configuration
                        </CardTitle>
                        <CardDescription>
                            Platform-wide JWT signing and validation settings
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="text-center py-8 text-gray-500">
                            <Server className="h-8 w-8 mx-auto mb-2 opacity-50" />
                            <p className="text-sm font-medium">Coming Soon</p>
                            <p className="text-xs mt-1">JWT configuration API not yet exposed</p>
                        </div>
                    </CardContent>
                </Card>

                {/* System Capabilities */}
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <Zap className="h-5 w-5" />
                            System Capabilities
                        </CardTitle>
                        <CardDescription>
                            Enable or disable platform features available to tenants
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="text-center py-8 text-gray-500">
                            <Zap className="h-8 w-8 mx-auto mb-2 opacity-50" />
                            <p className="text-sm font-medium">Manage via Capabilities Page</p>
                            <p className="text-xs mt-1">
                                <a href="/capabilities/system" className="text-blue-600 hover:underline">
                                    Go to System Capabilities →
                                </a>
                            </p>
                        </div>
                    </CardContent>
                </Card>
            </div>
        </PermissionGate>
    );
}
