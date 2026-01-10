/**
 * Tenant Settings Page
 * 
 * AUTHORITY MODEL:
 * Who: TENANT users OR SYSTEM users (when tenant selected)
 * Scope: Single tenant configuration
 * Permission: settings:update (tenant) OR tenant:update (system)
 * Why: Tenant controls their own security posture and behavior
 * 
 * GUARDRAIL #1: Backend Is Law
 * - All settings from backend APIs
 * - Respect backend min/max constraints
 */

import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { PermissionGate } from '@/components/PermissionGate';
import { usePrincipalContext } from '@/contexts/PrincipalContext';
import { systemApi, tenantApi } from '@/services/api';
import { Building2, Shield, Clock, AlertTriangle, Info, CheckCircle2 } from 'lucide-react';
import { useState } from 'react';

// Token settings schema with backend constraints
const tokenSettingsSchema = z.object({
    accessTokenTTLMinutes: z.number().min(1).max(1440),
    refreshTokenTTLDays: z.number().min(1).max(365),
    idTokenTTLMinutes: z.number().min(1).max(1440),
});

// Password policy schema
const passwordPolicySchema = z.object({
    minPasswordLength: z.number().min(8).max(128),
    requireUppercase: z.boolean(),
    requireLowercase: z.boolean(),
    requireNumbers: z.boolean(),
    requireSpecialChars: z.boolean(),
});

// MFA settings schema
const mfaSettingsSchema = z.object({
    mfaRequired: z.boolean(),
});

// Rate limiting schema
const rateLimitSchema = z.object({
    rateLimitRequests: z.number().min(1),
    rateLimitWindow: z.number().min(1),
});

type TokenSettings = z.infer<typeof tokenSettingsSchema>;
type PasswordPolicy = z.infer<typeof passwordPolicySchema>;
type MFASettings = z.infer<typeof mfaSettingsSchema>;
type RateLimit = z.infer<typeof rateLimitSchema>;

export function TenantSettings() {
    const { principalType, homeTenantId, selectedTenantId } = usePrincipalContext();
    const queryClient = useQueryClient();
    const [success, setSuccess] = useState<string | null>(null);
    const [error, setError] = useState<string | null>(null);

    // Determine effective tenant ID
    const effectiveTenantId = principalType === 'SYSTEM' ? selectedTenantId : homeTenantId;

    // Fetch tenant settings
    const { data: tenantSettings, isLoading } = useQuery({
        queryKey: ['tenant-settings', effectiveTenantId],
        queryFn: async () => {
            if (!effectiveTenantId) return null;
            if (principalType === 'SYSTEM') {
                return systemApi.tenants.getSettings(effectiveTenantId);
            } else {
                return tenantApi.getSettings();
            }
        },
        enabled: !!effectiveTenantId,
    });

    // Token settings form
    const tokenForm = useForm<TokenSettings>({
        resolver: zodResolver(tokenSettingsSchema),
        defaultValues: {
            accessTokenTTLMinutes: 15,
            refreshTokenTTLDays: 30,
            idTokenTTLMinutes: 60,
        },
    });

    // Password policy form
    const passwordForm = useForm<PasswordPolicy>({
        resolver: zodResolver(passwordPolicySchema),
        defaultValues: {
            minPasswordLength: 12,
            requireUppercase: true,
            requireLowercase: true,
            requireNumbers: true,
            requireSpecialChars: true,
        },
    });

    // MFA settings form
    const mfaForm = useForm<MFASettings>({
        resolver: zodResolver(mfaSettingsSchema),
        defaultValues: {
            mfaRequired: false,
        },
    });

    // Rate limit form
    const rateLimitForm = useForm<RateLimit>({
        resolver: zodResolver(rateLimitSchema),
        defaultValues: {
            rateLimitRequests: 100,
            rateLimitWindow: 60,
        },
    });

    // Update forms when settings load
    useEffect(() => {
        if (tenantSettings && !isLoading) {
            tokenForm.reset({
                accessTokenTTLMinutes: tenantSettings.access_token_ttl_minutes || 15,
                refreshTokenTTLDays: tenantSettings.refresh_token_ttl_days || 30,
                idTokenTTLMinutes: tenantSettings.id_token_ttl_minutes || 60,
            });

            passwordForm.reset({
                minPasswordLength: tenantSettings.min_password_length || 12,
                requireUppercase: tenantSettings.require_uppercase ?? true,
                requireLowercase: tenantSettings.require_lowercase ?? true,
                requireNumbers: tenantSettings.require_numbers ?? true,
                requireSpecialChars: tenantSettings.require_special_chars ?? true,
            });

            mfaForm.reset({
                mfaRequired: tenantSettings.mfa_required ?? false,
            });

            rateLimitForm.reset({
                rateLimitRequests: tenantSettings.rate_limit_requests || 100,
                rateLimitWindow: tenantSettings.rate_limit_window_seconds || 60,
            });
        }
    }, [tenantSettings, isLoading]);

    // Save settings mutation
    const saveSettings = useMutation({
        mutationFn: async (data: any) => {
            if (!effectiveTenantId) throw new Error('No tenant selected');
            if (principalType === 'SYSTEM') {
                return systemApi.tenants.updateSettings(effectiveTenantId, data);
            } else {
                return tenantApi.updateSettings(data);
            }
        },
        onSuccess: () => {
            setSuccess('Settings saved successfully');
            queryClient.invalidateQueries({ queryKey: ['tenant-settings'] });
            setTimeout(() => setSuccess(null), 3000);
        },
        onError: (err: any) => {
            setError(err.message || 'Failed to save settings');
            setTimeout(() => setError(null), 5000);
        },
    });

    const onTokenSubmit = (data: TokenSettings) => {
        saveSettings.mutate({
            access_token_ttl_minutes: data.accessTokenTTLMinutes,
            refresh_token_ttl_days: data.refreshTokenTTLDays,
            id_token_ttl_minutes: data.idTokenTTLMinutes,
        });
    };

    const onPasswordSubmit = (data: PasswordPolicy) => {
        saveSettings.mutate({
            min_password_length: data.minPasswordLength,
            require_uppercase: data.requireUppercase,
            require_lowercase: data.requireLowercase,
            require_numbers: data.requireNumbers,
            require_special_chars: data.requireSpecialChars,
        });
    };

    const onMFASubmit = (data: MFASettings) => {
        saveSettings.mutate({
            mfa_required: data.mfaRequired,
        });
    };

    const onRateLimitSubmit = (data: RateLimit) => {
        saveSettings.mutate({
            rate_limit_requests: data.rateLimitRequests,
            rate_limit_window_seconds: data.rateLimitWindow,
        });
    };

    if (!effectiveTenantId) {
        return (
            <div className="space-y-4">
                <h1 className="text-3xl font-bold">Tenant Settings</h1>
                <Alert className="bg-yellow-50 border-yellow-200">
                    <AlertTriangle className="h-4 w-4 text-yellow-600" />
                    <AlertDescription className="text-yellow-800">
                        Select a tenant from the header to configure tenant settings
                    </AlertDescription>
                </Alert>
            </div>
        );
    }

    return (
        <PermissionGate
            permission={principalType === 'SYSTEM' ? 'tenant:update' : 'settings:update'}
            systemPermission={principalType === 'SYSTEM'}
            fallback={
                <div className="space-y-4">
                    <h1 className="text-3xl font-bold">Tenant Settings</h1>
                    <Alert variant="destructive">
                        <AlertDescription>
                            You do not have permission to modify tenant settings
                        </AlertDescription>
                    </Alert>
                </div>
            }
        >
            <div className="space-y-6">
                {/* Header */}
                <div>
                    <h1 className="text-3xl font-bold">Tenant Settings</h1>
                    <p className="text-gray-600 mt-1">Configure settings for this tenant</p>
                    <Badge className="mt-2 bg-green-100 text-green-800">
                        <Building2 className="h-3 w-3 mr-1" />
                        Applies only to this tenant
                    </Badge>
                </div>

                {/* Success/Error Messages */}
                {success && (
                    <Alert className="bg-green-50 border-green-200">
                        <CheckCircle2 className="h-4 w-4 text-green-600" />
                        <AlertDescription className="text-green-800">{success}</AlertDescription>
                    </Alert>
                )}
                {error && (
                    <Alert variant="destructive">
                        <AlertDescription>{error}</AlertDescription>
                    </Alert>
                )}

                {/* Authority Notice */}
                <Alert className="bg-green-50 border-green-200">
                    <Info className="h-4 w-4 text-green-600" />
                    <AlertDescription className="text-green-800">
                        <strong>Tenant-Level Settings</strong> - Changes here affect only this tenant
                    </AlertDescription>
                </Alert>

                {/* Token Lifetimes */}
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <Clock className="h-5 w-5" />
                            Token Lifetimes
                        </CardTitle>
                        <CardDescription>
                            Configure how long tokens remain valid
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <form onSubmit={tokenForm.handleSubmit(onTokenSubmit)} className="space-y-4">
                            <div className="grid grid-cols-3 gap-4">
                                <div className="space-y-2">
                                    <Label htmlFor="accessTokenTTL">Access Token (minutes)</Label>
                                    <Input
                                        id="accessTokenTTL"
                                        type="number"
                                        {...tokenForm.register('accessTokenTTLMinutes', { valueAsNumber: true })}
                                        disabled={saveSettings.isPending}
                                    />
                                    <p className="text-xs text-gray-500">Range: 1-1440 minutes</p>
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="refreshTokenTTL">Refresh Token (days)</Label>
                                    <Input
                                        id="refreshTokenTTL"
                                        type="number"
                                        {...tokenForm.register('refreshTokenTTLDays', { valueAsNumber: true })}
                                        disabled={saveSettings.isPending}
                                    />
                                    <p className="text-xs text-gray-500">Range: 1-365 days</p>
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="idTokenTTL">ID Token (minutes)</Label>
                                    <Input
                                        id="idTokenTTL"
                                        type="number"
                                        {...tokenForm.register('idTokenTTLMinutes', { valueAsNumber: true })}
                                        disabled={saveSettings.isPending}
                                    />
                                    <p className="text-xs text-gray-500">Range: 1-1440 minutes</p>
                                </div>
                            </div>
                            <Button type="submit" disabled={saveSettings.isPending}>
                                {saveSettings.isPending ? 'Saving...' : 'Save Token Settings'}
                            </Button>
                        </form>
                    </CardContent>
                </Card>

                {/* Password Policy */}
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <Shield className="h-5 w-5" />
                            Password Policy
                        </CardTitle>
                        <CardDescription>
                            Configure password requirements for users
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <form onSubmit={passwordForm.handleSubmit(onPasswordSubmit)} className="space-y-4">
                            <div className="space-y-2">
                                <Label htmlFor="minPasswordLength">Minimum Length</Label>
                                <Input
                                    id="minPasswordLength"
                                    type="number"
                                    {...passwordForm.register('minPasswordLength', { valueAsNumber: true })}
                                    disabled={saveSettings.isPending}
                                />
                                <p className="text-xs text-gray-500">Range: 8-128 characters</p>
                            </div>
                            <div className="space-y-3">
                                <div className="flex items-center gap-2">
                                    <input
                                        type="checkbox"
                                        id="requireUppercase"
                                        {...passwordForm.register('requireUppercase')}
                                        disabled={saveSettings.isPending}
                                        className="rounded"
                                    />
                                    <Label htmlFor="requireUppercase">Require uppercase letters</Label>
                                </div>
                                <div className="flex items-center gap-2">
                                    <input
                                        type="checkbox"
                                        id="requireLowercase"
                                        {...passwordForm.register('requireLowercase')}
                                        disabled={saveSettings.isPending}
                                        className="rounded"
                                    />
                                    <Label htmlFor="requireLowercase">Require lowercase letters</Label>
                                </div>
                                <div className="flex items-center gap-2">
                                    <input
                                        type="checkbox"
                                        id="requireNumbers"
                                        {...passwordForm.register('requireNumbers')}
                                        disabled={saveSettings.isPending}
                                        className="rounded"
                                    />
                                    <Label htmlFor="requireNumbers">Require numbers</Label>
                                </div>
                                <div className="flex items-center gap-2">
                                    <input
                                        type="checkbox"
                                        id="requireSpecialChars"
                                        {...passwordForm.register('requireSpecialChars')}
                                        disabled={saveSettings.isPending}
                                        className="rounded"
                                    />
                                    <Label htmlFor="requireSpecialChars">Require special characters</Label>
                                </div>
                            </div>
                            <Button type="submit" disabled={saveSettings.isPending}>
                                {saveSettings.isPending ? 'Saving...' : 'Save Password Policy'}
                            </Button>
                        </form>
                    </CardContent>
                </Card>

                {/* MFA Requirements */}
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <Shield className="h-5 w-5" />
                            MFA Requirements
                        </CardTitle>
                        <CardDescription>
                            Configure multi-factor authentication requirements
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <form onSubmit={mfaForm.handleSubmit(onMFASubmit)} className="space-y-4">
                            <div className="flex items-center gap-2">
                                <input
                                    type="checkbox"
                                    id="mfaRequired"
                                    {...mfaForm.register('mfaRequired')}
                                    disabled={saveSettings.isPending}
                                    className="rounded"
                                />
                                <Label htmlFor="mfaRequired">Require MFA for all users</Label>
                            </div>
                            <Button type="submit" disabled={saveSettings.isPending}>
                                {saveSettings.isPending ? 'Saving...' : 'Save MFA Settings'}
                            </Button>
                        </form>
                    </CardContent>
                </Card>

                {/* Rate Limiting */}
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <AlertTriangle className="h-5 w-5" />
                            Rate Limiting
                        </CardTitle>
                        <CardDescription>
                            Configure API rate limits for this tenant
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <form onSubmit={rateLimitForm.handleSubmit(onRateLimitSubmit)} className="space-y-4">
                            <div className="grid grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <Label htmlFor="rateLimitRequests">Requests per Window</Label>
                                    <Input
                                        id="rateLimitRequests"
                                        type="number"
                                        {...rateLimitForm.register('rateLimitRequests', { valueAsNumber: true })}
                                        disabled={saveSettings.isPending}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="rateLimitWindow">Window (seconds)</Label>
                                    <Input
                                        id="rateLimitWindow"
                                        type="number"
                                        {...rateLimitForm.register('rateLimitWindow', { valueAsNumber: true })}
                                        disabled={saveSettings.isPending}
                                    />
                                </div>
                            </div>
                            <Button type="submit" disabled={saveSettings.isPending}>
                                {saveSettings.isPending ? 'Saving...' : 'Save Rate Limits'}
                            </Button>
                        </form>
                    </CardContent>
                </Card>
            </div>
        </PermissionGate>
    );
}
