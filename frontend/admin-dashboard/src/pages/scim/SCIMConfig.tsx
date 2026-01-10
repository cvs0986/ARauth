/**
 * SCIM Configuration Page
 * 
 * AUTHORITY MODEL:
 * Who: TENANT users OR SYSTEM users (tenant selected)
 * Scope: Tenant-scoped SCIM configuration
 * Permission: scim:read
 * 
 * GUARDRAILS:
 * - SCIM tokens are root credentials
 * - No silent enablement
 * - No cross-tenant SCIM
 * - Read-only configuration display
 */

import { useQuery } from '@tanstack/react-query';
import { usePrincipalContext } from '@/contexts/PrincipalContext';
import { PermissionGate } from '@/components/PermissionGate';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Link2, AlertTriangle, Copy, CheckCircle2, Info } from 'lucide-react';
import { useState } from 'react';
import { SCIMTokenList } from './SCIMTokenList';

export function SCIMConfig() {
    const { principalType, homeTenantId, selectedTenantId } = usePrincipalContext();
    const [urlCopied, setUrlCopied] = useState(false);

    // Determine effective tenant ID
    const effectiveTenantId = principalType === 'SYSTEM' ? selectedTenantId : homeTenantId;

    // Fetch SCIM configuration
    const { data: scimConfig, isLoading } = useQuery({
        queryKey: ['scim-config', effectiveTenantId],
        queryFn: async () => {
            // TODO: Implement API call
            // return scimApi.getConfig(effectiveTenantId);

            // Placeholder
            return {
                enabled: true,
                base_url: `https://api.arauth.example.com/scim/v2/tenants/${effectiveTenantId}`,
                tenant_id: effectiveTenantId,
            };
        },
        enabled: !!effectiveTenantId,
    });

    const copyUrl = () => {
        if (scimConfig?.base_url) {
            navigator.clipboard.writeText(scimConfig.base_url);
            setUrlCopied(true);
            setTimeout(() => setUrlCopied(false), 2000);
        }
    };

    if (!effectiveTenantId) {
        return (
            <div className="space-y-4">
                <h1 className="text-3xl font-bold">SCIM Provisioning</h1>
                <Alert className="bg-yellow-50 border-yellow-200">
                    <AlertTriangle className="h-4 w-4 text-yellow-600" />
                    <AlertDescription className="text-yellow-800">
                        Select a tenant to manage SCIM provisioning
                    </AlertDescription>
                </Alert>
            </div>
        );
    }

    if (isLoading) {
        return <div className="p-4">Loading SCIM configuration...</div>;
    }

    return (
        <PermissionGate permission="scim:read" systemPermission={principalType === 'SYSTEM'}>
            <div className="space-y-6">
                {/* Header */}
                <div>
                    <h1 className="text-3xl font-bold">SCIM Provisioning</h1>
                    <p className="text-sm text-gray-600 mt-1">
                        System for Cross-domain Identity Management (SCIM 2.0)
                    </p>
                </div>

                {/* SCIM Overview */}
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <Link2 className="h-5 w-5" />
                            SCIM Configuration
                        </CardTitle>
                        <CardDescription>
                            Configure SCIM provisioning for automated user and group management
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        {/* Status */}
                        <div className="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
                            <div>
                                <Label className="text-sm font-medium">SCIM Status</Label>
                                <p className="text-xs text-gray-500 mt-0.5">
                                    SCIM provisioning for this tenant
                                </p>
                            </div>
                            <Badge variant={scimConfig?.enabled ? 'default' : 'secondary'} className="text-sm">
                                {scimConfig?.enabled ? (
                                    <>
                                        <CheckCircle2 className="h-3 w-3 mr-1" />
                                        Enabled
                                    </>
                                ) : (
                                    'Disabled'
                                )}
                            </Badge>
                        </div>

                        {/* SCIM Base URL */}
                        <div className="space-y-2">
                            <Label>SCIM Base URL</Label>
                            <div className="flex gap-2">
                                <Input
                                    value={scimConfig?.base_url || ''}
                                    readOnly
                                    className="font-mono text-sm"
                                />
                                <Button
                                    variant="outline"
                                    size="sm"
                                    onClick={copyUrl}
                                    className={urlCopied ? 'bg-green-50' : ''}
                                >
                                    {urlCopied ? (
                                        <CheckCircle2 className="h-4 w-4 text-green-600" />
                                    ) : (
                                        <Copy className="h-4 w-4" />
                                    )}
                                </Button>
                            </div>
                            <p className="text-xs text-gray-500">
                                Use this URL as the SCIM endpoint in your identity provider
                            </p>
                        </div>

                        {/* Tenant ID */}
                        <div className="space-y-2">
                            <Label>Tenant ID</Label>
                            <Input
                                value={scimConfig?.tenant_id || ''}
                                readOnly
                                className="font-mono text-sm"
                            />
                        </div>

                        {/* Info Notice */}
                        <Alert className="bg-blue-50 border-blue-200">
                            <Info className="h-4 w-4 text-blue-600" />
                            <AlertDescription className="text-blue-800 text-sm">
                                <strong>SCIM Tokens:</strong> Create a SCIM token below to authenticate provisioning requests.
                                Tokens are shown only once and should be treated as root credentials.
                            </AlertDescription>
                        </Alert>
                    </CardContent>
                </Card>

                {/* SCIM Token Management */}
                <SCIMTokenList tenantId={effectiveTenantId} />

                {/* Attribute Mapping (Coming Soon) */}
                <Card>
                    <CardHeader>
                        <CardTitle>Attribute Mapping</CardTitle>
                        <CardDescription>
                            Map SCIM attributes to ARauth user fields
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <Alert className="bg-blue-50 border-blue-200">
                            <Info className="h-4 w-4 text-blue-600" />
                            <AlertDescription className="text-blue-800">
                                <strong>Coming Soon:</strong> Attribute mapping configuration will be available when the backend API is exposed.
                            </AlertDescription>
                        </Alert>
                    </CardContent>
                </Card>
            </div>
        </PermissionGate>
    );
}
