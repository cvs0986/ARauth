/**
 * Settings Page
 * Shows System Settings for SYSTEM users, Tenant Settings for all users
 */

import { useState, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Alert } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Shield, Key, Building2, Server, Settings as SettingsIcon } from 'lucide-react';
import { useAuthStore } from '@/store/authStore';
import { systemApi, tenantApi, systemCapabilityApi, tenantCapabilityApi, tenantFeatureApi } from '@/services/api';
import { Link } from 'react-router-dom';
import { formatCapabilityName, getCapabilityDescription } from '@/utils/capabilityNames';
import { Switch } from '@/components/ui/switch';

// Settings schemas
const securitySettingsSchema = z.object({
  minPasswordLength: z.number().min(8).max(128),
  requireUppercase: z.boolean(),
  requireLowercase: z.boolean(),
  requireNumbers: z.boolean(),
  requireSpecialChars: z.boolean(),
  passwordExpiryDays: z.number().min(0).optional(),
  mfaRequired: z.boolean(),
  rateLimitRequests: z.number().min(1),
  rateLimitWindow: z.number().min(1),
});

const oauthSettingsSchema = z.object({
  hydraAdminUrl: z.string().url(),
  hydraPublicUrl: z.string().url(),
  accessTokenTTL: z.number().min(60),
  refreshTokenTTL: z.number().min(3600),
  idTokenTTL: z.number().min(60),
});

const systemSettingsSchema = z.object({
  jwtSecret: z.string().min(32).optional(),
  jwtIssuer: z.string().optional(),
  jwtAudience: z.string().optional(),
  sessionTimeout: z.number().min(300),
  maxLoginAttempts: z.number().min(1),
  lockoutDuration: z.number().min(60),
});

const tokenSettingsSchema = z.object({
  accessTokenTTLMinutes: z.number().min(1).max(1440), // 1 minute to 24 hours
  refreshTokenTTLDays: z.number().min(1).max(365), // 1 day to 1 year
  idTokenTTLMinutes: z.number().min(1).max(1440),
  rememberMeEnabled: z.boolean(),
  rememberMeRefreshTokenTTLDays: z.number().min(1).max(365),
  rememberMeAccessTokenTTLMinutes: z.number().min(1).max(1440),
  tokenRotationEnabled: z.boolean(),
  requireMFAForExtendedSessions: z.boolean(),
});

type SecuritySettings = z.infer<typeof securitySettingsSchema>;
type OAuthSettings = z.infer<typeof oauthSettingsSchema>;
type SystemSettings = z.infer<typeof systemSettingsSchema>;
type TokenSettings = z.infer<typeof tokenSettingsSchema>;

// Capabilities Settings Tab Component
function CapabilitiesSettingsTab({ isSystemUser, currentTenantId }: { isSystemUser: boolean; currentTenantId: string | null }) {
  const { tenantId } = useAuthStore();
  const queryClient = useQueryClient();

  // For SYSTEM users: show system capabilities
  const { data: systemCapabilities } = useQuery({
    queryKey: ['system', 'capabilities'],
    queryFn: () => systemCapabilityApi.list(),
    enabled: isSystemUser,
  });

  // For SYSTEM users: show tenant capabilities for selected tenant
  const { data: tenantCapabilities } = useQuery({
    queryKey: ['tenant', 'capabilities', currentTenantId],
    queryFn: () => tenantCapabilityApi.list(currentTenantId!),
    enabled: isSystemUser && !!currentTenantId,
  });

  // For TENANT users: show enabled features
  const { data: tenantFeatures } = useQuery({
    queryKey: ['tenant', 'features'],
    queryFn: () => tenantFeatureApi.list(),
    enabled: !isSystemUser && !!tenantId,
  });

  // Toggle system capability mutation
  const toggleSystemCapability = useMutation({
    mutationFn: ({ key, enabled }: { key: string; enabled: boolean }) =>
      systemCapabilityApi.update(key, { enabled }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['system', 'capabilities'] });
    },
  });

  if (isSystemUser) {
    return (
      <div className="space-y-4">
        <Card>
          <CardHeader>
            <CardTitle>System Capabilities</CardTitle>
            <CardDescription>
              Global system capabilities that define what features are available
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {systemCapabilities?.map((cap) => (
                <Card key={cap.capability_key} className={`transition-all ${cap.enabled ? 'border-green-200 bg-green-50/50' : ''}`}>
                  <CardContent className="pt-4">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <h3 className="text-base font-semibold mb-1">
                          {formatCapabilityName(cap.capability_key)}
                        </h3>
                        <p className="text-sm text-gray-600 mb-3">
                          {getCapabilityDescription(cap.capability_key, cap.description)}
                        </p>
                        {/* Toggle Section */}
                        <div className="flex items-center justify-between p-3 bg-white rounded-lg border-2 border-gray-200 shadow-sm">
                          <div className="flex items-center gap-3">
                            <Label 
                              htmlFor={`toggle-${cap.capability_key}`} 
                              className="text-sm font-semibold cursor-pointer flex items-center gap-2"
                            >
                              <span className={cap.enabled ? 'text-green-700 font-bold' : 'text-gray-700'}>
                                {cap.enabled ? '✓ Enabled' : '○ Disabled'}
                              </span>
                            </Label>
                          </div>
                          <Switch
                            id={`toggle-${cap.capability_key}`}
                            checked={cap.enabled}
                            disabled={toggleSystemCapability.isPending}
                            onCheckedChange={(enabled) => {
                              toggleSystemCapability.mutate({ key: cap.capability_key, enabled });
                            }}
                            className="h-7 w-14 data-[state=checked]:bg-green-600 data-[state=unchecked]:bg-gray-300 shadow-sm [&>span]:h-6 [&>span]:w-6 [&>span]:bg-white [&>span]:shadow-lg [&>span]:border [&>span]:border-gray-200 [&>span]:data-[state=checked]:translate-x-7 [&>span]:data-[state=unchecked]:translate-x-0.5"
                          />
                        </div>
                        {cap.default_value && Object.keys(cap.default_value).length > 0 && (
                          <div className="mt-3 pt-3 border-t">
                            <p className="text-xs font-medium text-gray-700 mb-1">System default available:</p>
                            <pre className="text-xs bg-blue-50 p-2 rounded overflow-x-auto max-h-24 border border-blue-200">
                              {JSON.stringify(cap.default_value, null, 2)}
                            </pre>
                          </div>
                        )}
                      </div>
                    </div>
                  </CardContent>
                </Card>
              ))}
              <div className="pt-2">
                <Button variant="outline" asChild>
                  <Link to="/capabilities/system">Manage System Capabilities</Link>
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>

        {currentTenantId && (
          <Card>
            <CardHeader>
              <CardTitle>Tenant Capabilities</CardTitle>
              <CardDescription>
                Capabilities assigned to the selected tenant
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                {tenantCapabilities && tenantCapabilities.length > 0 ? (
                  tenantCapabilities.map((cap) => (
                    <Card key={cap.capability_key} className={`transition-all ${cap.enabled ? 'border-green-200 bg-green-50/50' : ''}`}>
                      <CardContent className="pt-4">
                        <div className="flex items-start justify-between">
                          <div className="flex-1">
                            <h3 className="text-base font-semibold mb-1">
                              {formatCapabilityName(cap.capability_key)}
                            </h3>
                            <p className="text-sm text-gray-600 mb-3">
                              {getCapabilityDescription(cap.capability_key)}
                            </p>
                            <div className="flex items-center gap-2">
                              <Badge variant={cap.enabled ? 'default' : 'secondary'}>
                                {cap.enabled ? 'Enabled' : 'Disabled'}
                              </Badge>
                            </div>
                            {cap.value && Object.keys(cap.value).length > 0 && (
                              <div className="mt-3 pt-3 border-t">
                                <p className="text-xs font-medium text-gray-700 mb-1">Custom value configured:</p>
                                <pre className="text-xs bg-gray-100 p-2 rounded overflow-x-auto max-h-24">
                                  {JSON.stringify(cap.value, null, 2)}
                                </pre>
                              </div>
                            )}
                          </div>
                        </div>
                      </CardContent>
                    </Card>
                  ))
                ) : (
                  <p className="text-sm text-gray-500">No capabilities assigned</p>
                )}
                <div className="pt-2">
                  <Button variant="outline" asChild>
                    <Link to="/capabilities/tenant-assignment">Manage Tenant Capabilities</Link>
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    );
  }

  // TENANT users: show enabled features
  return (
    <Card>
      <CardHeader>
        <CardTitle>Enabled Features</CardTitle>
        <CardDescription>
          Features that are currently enabled for your tenant
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {tenantFeatures && tenantFeatures.length > 0 ? (
            tenantFeatures.map((feature) => (
              <Card key={feature.capability_key} className={`transition-all ${feature.enabled ? 'border-green-200 bg-green-50/50' : ''}`}>
                <CardContent className="pt-4">
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <h3 className="text-base font-semibold mb-1">
                        {formatCapabilityName(feature.capability_key)}
                      </h3>
                      <p className="text-sm text-gray-600 mb-3">
                        {getCapabilityDescription(feature.capability_key)}
                      </p>
                      <div className="flex items-center gap-2">
                        <Badge variant={feature.enabled ? 'default' : 'secondary'}>
                          {feature.enabled ? 'Enabled' : 'Disabled'}
                        </Badge>
                        {feature.enabled_at && (
                          <span className="text-xs text-gray-500">
                            Enabled {new Date(feature.enabled_at).toLocaleDateString()}
                          </span>
                        )}
                      </div>
                      {feature.configuration && typeof feature.configuration === 'object' && Object.keys(feature.configuration).length > 0 && (
                        <div className="mt-3 pt-3 border-t">
                          <p className="text-xs font-medium text-gray-700 mb-1">Configuration:</p>
                          <pre className="text-xs bg-gray-100 p-2 rounded overflow-x-auto max-h-24">
                            {JSON.stringify(feature.configuration, null, 2)}
                          </pre>
                        </div>
                      )}
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))
          ) : (
            <p className="text-sm text-gray-500">No features enabled</p>
          )}
          <div className="pt-2 space-x-2">
            <Button variant="outline" asChild>
              <Link to="/capabilities/features">Manage Features</Link>
            </Button>
            <Button variant="outline" asChild>
              <Link to="/capabilities/user-enrollment">Manage User Capabilities</Link>
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export function Settings() {
  const { isSystemUser, selectedTenantId, tenantId } = useAuthStore();
  const queryClient = useQueryClient();
  // Initialize activeTab based on user type: 'security' for SYSTEM, 'tenant' for TENANT
  const [activeTab, setActiveTab] = useState(() => {
    // Check localStorage for principal_type to determine initial tab
    const storedPrincipalType = localStorage.getItem('principal_type');
    return storedPrincipalType === 'SYSTEM' ? 'security' : 'tenant';
  });
  const [success, setSuccess] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  // Get current tenant ID (selected tenant for SYSTEM users, own tenant for TENANT users)
  const currentTenantId = isSystemUser() ? selectedTenantId : tenantId;

  // Fetch tenant settings
  // For SYSTEM users: fetch settings for selected tenant
  // For TENANT users: fetch their own tenant settings
  const { data: tenantSettings, isLoading: settingsLoading } = useQuery({
    queryKey: ['tenant-settings', currentTenantId, isSystemUser()],
    queryFn: async () => {
      if (isSystemUser()) {
        // SYSTEM users: fetch settings for selected tenant
        if (!currentTenantId) return null;
        return systemApi.tenants.getSettings(currentTenantId);
      } else {
        // TENANT users: fetch their own tenant settings
        if (!tenantId) return null;
        return tenantApi.getSettings();
      }
    },
    enabled: (isSystemUser() ? !!currentTenantId : !!tenantId), // Enable if tenant context is available
  });

  // Security Settings Form
  const {
    register: registerSecurity,
    handleSubmit: handleSubmitSecurity,
    formState: { errors: securityErrors },
  } = useForm<SecuritySettings>({
    resolver: zodResolver(securitySettingsSchema),
    defaultValues: {
      minPasswordLength: 12,
      requireUppercase: true,
      requireLowercase: true,
      requireNumbers: true,
      requireSpecialChars: true,
      mfaRequired: false,
      rateLimitRequests: 100,
      rateLimitWindow: 60,
    },
  });

  // OAuth Settings Form
  const {
    register: registerOAuth,
    handleSubmit: handleSubmitOAuth,
    formState: { errors: oauthErrors },
  } = useForm<OAuthSettings>({
    resolver: zodResolver(oauthSettingsSchema),
    defaultValues: {
      hydraAdminUrl: 'http://localhost:4445',
      hydraPublicUrl: 'http://localhost:4444',
      accessTokenTTL: 3600,
      refreshTokenTTL: 86400,
      idTokenTTL: 3600,
    },
  });

  // System Settings Form
  const {
    register: registerSystem,
    handleSubmit: handleSubmitSystem,
    formState: { errors: systemErrors },
  } = useForm<SystemSettings>({
    resolver: zodResolver(systemSettingsSchema),
    defaultValues: {
      jwtIssuer: 'arauth-identity',
      sessionTimeout: 3600,
      maxLoginAttempts: 5,
      lockoutDuration: 900,
    },
  });

  // Token Settings Form (for tenant settings)
  const {
    register: registerToken,
    handleSubmit: handleSubmitToken,
    watch: watchToken,
    setValue: setTokenValue,
    formState: { errors: tokenErrors },
  } = useForm<TokenSettings>({
    resolver: zodResolver(tokenSettingsSchema),
    defaultValues: {
      accessTokenTTLMinutes: 15, // 15 minutes
      refreshTokenTTLDays: 30, // 30 days
      idTokenTTLMinutes: 60, // 1 hour
      rememberMeEnabled: true,
      rememberMeRefreshTokenTTLDays: 90, // 90 days
      rememberMeAccessTokenTTLMinutes: 60, // 60 minutes
      tokenRotationEnabled: true,
      requireMFAForExtendedSessions: false,
    },
  });

  const rememberMeEnabled = watchToken('rememberMeEnabled');

  // Update form when tenant settings are loaded
  useEffect(() => {
    if (tenantSettings && !settingsLoading && currentTenantId) {
      setTokenValue('accessTokenTTLMinutes', tenantSettings.access_token_ttl_minutes || 15);
      setTokenValue('refreshTokenTTLDays', tenantSettings.refresh_token_ttl_days || 30);
      setTokenValue('idTokenTTLMinutes', tenantSettings.id_token_ttl_minutes || 60);
      setTokenValue('rememberMeEnabled', tenantSettings.remember_me_enabled ?? true);
      setTokenValue('rememberMeRefreshTokenTTLDays', tenantSettings.remember_me_refresh_token_ttl_days || 90);
      setTokenValue('rememberMeAccessTokenTTLMinutes', tenantSettings.remember_me_access_token_ttl_minutes || 60);
      setTokenValue('tokenRotationEnabled', tenantSettings.token_rotation_enabled ?? true);
      setTokenValue('requireMFAForExtendedSessions', tenantSettings.require_mfa_for_extended_sessions ?? false);
    }
  }, [tenantSettings, settingsLoading, currentTenantId, setTokenValue]);

  const onSecuritySubmit = async (data: SecuritySettings) => {
    setIsLoading(true);
    setError(null);
    setSuccess(null);

    try {
      // Prepare request payload
      const payload = {
        min_password_length: data.minPasswordLength,
        require_uppercase: data.requireUppercase,
        require_lowercase: data.requireLowercase,
        require_numbers: data.requireNumbers,
        require_special_chars: data.requireSpecialChars,
        password_expiry_days: data.passwordExpiryDays ?? null,
        mfa_required: data.mfaRequired,
        rate_limit_requests: data.rateLimitRequests,
        rate_limit_window_seconds: data.rateLimitWindow,
      };

      if (isSystemUser()) {
        // SYSTEM users: update settings for selected tenant
        if (!currentTenantId) {
          throw new Error('No tenant selected');
        }
        await systemApi.tenants.updateSettings(currentTenantId, payload);
      } else {
        // TENANT users: update their own tenant settings
        if (!tenantId) {
          throw new Error('No tenant context available');
        }
        await tenantApi.updateSettings(payload);
      }

      // Invalidate query to refetch updated settings
      queryClient.invalidateQueries({ queryKey: ['tenant-settings'] });
      setSuccess('Security settings saved successfully');
      setTimeout(() => setSuccess(null), 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save security settings');
    } finally {
      setIsLoading(false);
    }
  };

  const onOAuthSubmit = async (data: OAuthSettings) => {
    setIsLoading(true);
    setError(null);
    setSuccess(null);

    try {
      // TODO: Implement API call to save OAuth settings
      console.log('OAuth settings:', data);
      await new Promise((resolve) => setTimeout(resolve, 500)); // Simulate API call
      setSuccess('OAuth settings saved successfully');
      setTimeout(() => setSuccess(null), 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save OAuth settings');
    } finally {
      setIsLoading(false);
    }
  };


  const onSystemSubmit = async (data: SystemSettings) => {
    setIsLoading(true);
    setError(null);
    setSuccess(null);

    try {
      // TODO: Implement API call to save system settings
      console.log('System settings:', data);
      await new Promise((resolve) => setTimeout(resolve, 500)); // Simulate API call
      setSuccess('System settings saved successfully');
      setTimeout(() => setSuccess(null), 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save system settings');
    } finally {
      setIsLoading(false);
    }
  };

  // Ensure correct tab is selected based on user type
  useEffect(() => {
    if (isSystemUser()) {
      // SYSTEM users: default to 'security' if current tab doesn't exist
      if (!['system', 'security', 'oauth', 'tenant', 'capabilities'].includes(activeTab)) {
        setActiveTab('security');
      }
    } else {
      // TENANT users: only 'tenant' and 'capabilities' tabs are available
      if (!['tenant', 'capabilities'].includes(activeTab)) {
        setActiveTab('tenant');
      }
    }
  }, [isSystemUser(), activeTab]);

  // Update token form when tenant settings are loaded
  useEffect(() => {
    if (tenantSettings && !settingsLoading) {
      // Populate form with tenant settings
      // Note: This would require setValue from react-hook-form
    }
  }, [tenantSettings, settingsLoading]);

  // Save tenant settings mutation
  const saveTenantSettingsMutation = useMutation({
    mutationFn: async (data: TokenSettings) => {
      if (isSystemUser()) {
        // SYSTEM users: update settings for selected tenant
        if (!currentTenantId) throw new Error('No tenant selected');
        return systemApi.tenants.updateSettings(currentTenantId, data);
      } else {
        // TENANT users: update their own tenant settings
        if (!tenantId) throw new Error('No tenant context available');
        return tenantApi.updateSettings(data);
      }
    },
    onSuccess: () => {
      setSuccess('Tenant settings saved successfully');
      queryClient.invalidateQueries({ queryKey: ['tenant-settings'] });
      setTimeout(() => setSuccess(null), 3000);
    },
    onError: (err: any) => {
      setError(err.message || 'Failed to save tenant settings');
      setTimeout(() => setError(null), 5000);
    },
  });

  const onTenantTokenSubmit = async (data: TokenSettings) => {
    setIsLoading(true);
    setError(null);
    try {
      await saveTenantSettingsMutation.mutateAsync(data);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">
          {isSystemUser() ? 'System & Tenant Settings' : 'Tenant Settings'}
        </h1>
        <p className="text-gray-600 mt-1">
          {isSystemUser()
            ? 'Configure system-wide settings and tenant-specific configurations'
            : 'Configure settings for your tenant'}
        </p>
        {isSystemUser() && !currentTenantId && (
          <Alert className="mt-4 bg-yellow-50 border-yellow-200 text-yellow-800">
            <Building2 className="h-4 w-4 mr-2" />
            Select a tenant from the header to configure tenant-specific settings.
          </Alert>
        )}
      </div>

      {error && (
        <Alert variant="destructive">
          {error}
        </Alert>
      )}

      {success && (
        <Alert className="bg-green-50 border-green-200 text-green-700">
          {success}
        </Alert>
      )}

      <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-4">
        <TabsList>
          {isSystemUser() && (
            <>
              <TabsTrigger value="system">
                <Server className="mr-2 h-4 w-4" />
                System Settings
              </TabsTrigger>
              <TabsTrigger value="security">
                <Shield className="mr-2 h-4 w-4" />
                Security
              </TabsTrigger>
              <TabsTrigger value="oauth">
                <Key className="mr-2 h-4 w-4" />
                OAuth2/OIDC
              </TabsTrigger>
              <TabsTrigger value="capabilities">
                <SettingsIcon className="mr-2 h-4 w-4" />
                Capabilities
              </TabsTrigger>
            </>
          )}
          <TabsTrigger value="tenant">
            <Building2 className="mr-2 h-4 w-4" />
            {isSystemUser() ? 'Tenant Settings' : 'Token Settings'}
          </TabsTrigger>
          {!isSystemUser() && (
            <TabsTrigger value="capabilities">
              <SettingsIcon className="mr-2 h-4 w-4" />
              Capabilities
            </TabsTrigger>
          )}
        </TabsList>

        {/* Capabilities Tab */}
        <TabsContent value="capabilities">
          <CapabilitiesSettingsTab isSystemUser={isSystemUser()} currentTenantId={currentTenantId} />
        </TabsContent>

        {/* Security Settings Tab - Available for both SYSTEM and TENANT users */}
        <TabsContent value="security">
          <Card>
            <CardHeader>
              <CardTitle>Security Settings</CardTitle>
              <CardDescription>
                Configure password policies, MFA requirements, and rate limiting
              </CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleSubmitSecurity(onSecuritySubmit)} className="space-y-6">
                <div>
                  <h3 className="text-lg font-semibold mb-4">Password Policy</h3>
                  <div className="space-y-4">
                    <div className="space-y-2">
                      <Label htmlFor="minPasswordLength">Minimum Password Length</Label>
                      <Input
                        id="minPasswordLength"
                        type="number"
                        {...registerSecurity('minPasswordLength', { valueAsNumber: true })}
                        disabled={isLoading}
                      />
                      {securityErrors.minPasswordLength && (
                        <p className="text-sm text-red-600">
                          {securityErrors.minPasswordLength.message}
                        </p>
                      )}
                    </div>

                    <div className="space-y-2">
                      <div className="flex items-center space-x-2">
                        <input
                          type="checkbox"
                          id="requireUppercase"
                          {...registerSecurity('requireUppercase')}
                          disabled={isLoading}
                          className="rounded"
                        />
                        <Label htmlFor="requireUppercase">Require uppercase letters</Label>
                      </div>
                    </div>

                    <div className="space-y-2">
                      <div className="flex items-center space-x-2">
                        <input
                          type="checkbox"
                          id="requireLowercase"
                          {...registerSecurity('requireLowercase')}
                          disabled={isLoading}
                          className="rounded"
                        />
                        <Label htmlFor="requireLowercase">Require lowercase letters</Label>
                      </div>
                    </div>

                    <div className="space-y-2">
                      <div className="flex items-center space-x-2">
                        <input
                          type="checkbox"
                          id="requireNumbers"
                          {...registerSecurity('requireNumbers')}
                          disabled={isLoading}
                          className="rounded"
                        />
                        <Label htmlFor="requireNumbers">Require numbers</Label>
                      </div>
                    </div>

                    <div className="space-y-2">
                      <div className="flex items-center space-x-2">
                        <input
                          type="checkbox"
                          id="requireSpecialChars"
                          {...registerSecurity('requireSpecialChars')}
                          disabled={isLoading}
                          className="rounded"
                        />
                        <Label htmlFor="requireSpecialChars">Require special characters</Label>
                      </div>
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="passwordExpiryDays">Password Expiry (days, 0 = never)</Label>
                      <Input
                        id="passwordExpiryDays"
                        type="number"
                        {...registerSecurity('passwordExpiryDays', { valueAsNumber: true })}
                        disabled={isLoading}
                      />
                    </div>
                  </div>
                </div>

                <div>
                  <h3 className="text-lg font-semibold mb-4">Multi-Factor Authentication</h3>
                  <div className="space-y-2">
                    <div className="flex items-center space-x-2">
                      <input
                        type="checkbox"
                        id="mfaRequired"
                        {...registerSecurity('mfaRequired')}
                        disabled={isLoading}
                        className="rounded"
                      />
                      <Label htmlFor="mfaRequired">Require MFA for all users</Label>
                    </div>
                  </div>
                </div>

                <div>
                  <h3 className="text-lg font-semibold mb-4">Rate Limiting</h3>
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label htmlFor="rateLimitRequests">Requests per Window</Label>
                      <Input
                        id="rateLimitRequests"
                        type="number"
                        {...registerSecurity('rateLimitRequests', { valueAsNumber: true })}
                        disabled={isLoading}
                      />
                      {securityErrors.rateLimitRequests && (
                        <p className="text-sm text-red-600">
                          {securityErrors.rateLimitRequests.message}
                        </p>
                      )}
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="rateLimitWindow">Window (seconds)</Label>
                      <Input
                        id="rateLimitWindow"
                        type="number"
                        {...registerSecurity('rateLimitWindow', { valueAsNumber: true })}
                        disabled={isLoading}
                      />
                      {securityErrors.rateLimitWindow && (
                        <p className="text-sm text-red-600">
                          {securityErrors.rateLimitWindow.message}
                        </p>
                      )}
                    </div>
                  </div>
                </div>

                <Button type="submit" disabled={isLoading}>
                  {isLoading ? 'Saving...' : 'Save Security Settings'}
                </Button>
              </form>
            </CardContent>
          </Card>
        </TabsContent>

        {/* OAuth Settings Tab - SYSTEM users only */}
        {isSystemUser() && (
          <TabsContent value="oauth">
            <Card>
              <CardHeader>
                <CardTitle>OAuth2/OIDC Settings</CardTitle>
                <CardDescription>
                  Configure OAuth2 and OpenID Connect settings
                </CardDescription>
              </CardHeader>
              <CardContent>
                <form onSubmit={handleSubmitOAuth(onOAuthSubmit)} className="space-y-6">
                  <div className="space-y-4">
                    <div className="space-y-2">
                      <Label htmlFor="hydraAdminUrl">Hydra Admin URL</Label>
                      <Input
                        id="hydraAdminUrl"
                        type="url"
                        {...registerOAuth('hydraAdminUrl')}
                        disabled={isLoading}
                      />
                      {oauthErrors.hydraAdminUrl && (
                        <p className="text-sm text-red-600">
                          {oauthErrors.hydraAdminUrl.message}
                        </p>
                      )}
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="hydraPublicUrl">Hydra Public URL</Label>
                      <Input
                        id="hydraPublicUrl"
                        type="url"
                        {...registerOAuth('hydraPublicUrl')}
                        disabled={isLoading}
                      />
                      {oauthErrors.hydraPublicUrl && (
                        <p className="text-sm text-red-600">
                          {oauthErrors.hydraPublicUrl.message}
                        </p>
                      )}
                    </div>

                    <div className="grid grid-cols-3 gap-4">
                      <div className="space-y-2">
                        <Label htmlFor="accessTokenTTL">Access Token TTL (seconds)</Label>
                        <Input
                          id="accessTokenTTL"
                          type="number"
                          {...registerOAuth('accessTokenTTL', { valueAsNumber: true })}
                          disabled={isLoading}
                        />
                        {oauthErrors.accessTokenTTL && (
                          <p className="text-sm text-red-600">
                            {oauthErrors.accessTokenTTL.message}
                          </p>
                        )}
                      </div>

                      <div className="space-y-2">
                        <Label htmlFor="refreshTokenTTL">Refresh Token TTL (seconds)</Label>
                        <Input
                          id="refreshTokenTTL"
                          type="number"
                          {...registerOAuth('refreshTokenTTL', { valueAsNumber: true })}
                          disabled={isLoading}
                        />
                        {oauthErrors.refreshTokenTTL && (
                          <p className="text-sm text-red-600">
                            {oauthErrors.refreshTokenTTL.message}
                          </p>
                        )}
                      </div>

                      <div className="space-y-2">
                        <Label htmlFor="idTokenTTL">ID Token TTL (seconds)</Label>
                        <Input
                          id="idTokenTTL"
                          type="number"
                          {...registerOAuth('idTokenTTL', { valueAsNumber: true })}
                          disabled={isLoading}
                        />
                        {oauthErrors.idTokenTTL && (
                          <p className="text-sm text-red-600">
                            {oauthErrors.idTokenTTL.message}
                          </p>
                        )}
                      </div>
                    </div>
                  </div>

                  <Button type="submit" disabled={isLoading}>
                    {isLoading ? 'Saving...' : 'Save OAuth Settings'}
                  </Button>
                </form>
              </CardContent>
            </Card>
          </TabsContent>
        )}

        {/* System Settings Tab - SYSTEM users only */}
        {isSystemUser() && (
          <TabsContent value="system">
            <Card>
              <CardHeader>
                <CardTitle>System Settings</CardTitle>
                <CardDescription>
                  Configure system-wide settings
                </CardDescription>
              </CardHeader>
              <CardContent>
                <form onSubmit={handleSubmitSystem(onSystemSubmit)} className="space-y-6">
                  <div className="space-y-4">
                    <div className="space-y-2">
                      <Label htmlFor="jwtIssuer">JWT Issuer</Label>
                      <Input
                        id="jwtIssuer"
                        {...registerSystem('jwtIssuer')}
                        disabled={isLoading}
                        placeholder="arauth-identity"
                      />
                      {systemErrors.jwtIssuer && (
                        <p className="text-sm text-red-600">
                          {systemErrors.jwtIssuer.message}
                        </p>
                      )}
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="jwtAudience">JWT Audience</Label>
                      <Input
                        id="jwtAudience"
                        {...registerSystem('jwtAudience')}
                        disabled={isLoading}
                      />
                      {systemErrors.jwtAudience && (
                        <p className="text-sm text-red-600">
                          {systemErrors.jwtAudience.message}
                        </p>
                      )}
                    </div>

                    <div className="grid grid-cols-3 gap-4">
                      <div className="space-y-2">
                        <Label htmlFor="sessionTimeout">Session Timeout (seconds)</Label>
                        <Input
                          id="sessionTimeout"
                          type="number"
                          {...registerSystem('sessionTimeout', { valueAsNumber: true })}
                          disabled={isLoading}
                        />
                        {systemErrors.sessionTimeout && (
                          <p className="text-sm text-red-600">
                            {systemErrors.sessionTimeout.message}
                          </p>
                        )}
                      </div>

                      <div className="space-y-2">
                        <Label htmlFor="maxLoginAttempts">Max Login Attempts</Label>
                        <Input
                          id="maxLoginAttempts"
                          type="number"
                          {...registerSystem('maxLoginAttempts', { valueAsNumber: true })}
                          disabled={isLoading}
                        />
                        {systemErrors.maxLoginAttempts && (
                          <p className="text-sm text-red-600">
                            {systemErrors.maxLoginAttempts.message}
                          </p>
                        )}
                      </div>

                      <div className="space-y-2">
                        <Label htmlFor="lockoutDuration">Lockout Duration (seconds)</Label>
                        <Input
                          id="lockoutDuration"
                          type="number"
                          {...registerSystem('lockoutDuration', { valueAsNumber: true })}
                          disabled={isLoading}
                        />
                        {systemErrors.lockoutDuration && (
                          <p className="text-sm text-red-600">
                            {systemErrors.lockoutDuration.message}
                          </p>
                        )}
                      </div>
                    </div>
                  </div>

                  <Button type="submit" disabled={isLoading}>
                    {isLoading ? 'Saving...' : 'Save System Settings'}
                  </Button>
                </form>
              </CardContent>
            </Card>
          </TabsContent>
        )}

        {/* Tenant Settings Tab */}
        <TabsContent value="tenant">
          <Card>
            <CardHeader>
              <CardTitle>{isSystemUser() ? 'Tenant Settings' : 'Token Settings'}</CardTitle>
              <CardDescription>
                {isSystemUser()
                  ? 'Configure token settings for the selected tenant'
                  : 'Configure token settings for your tenant'}
              </CardDescription>
            </CardHeader>
            <CardContent>
              {isSystemUser() && !currentTenantId ? (
                <Alert>
                  <Building2 className="h-4 w-4 mr-2" />
                  Please select a tenant from the header to configure tenant settings.
                </Alert>
              ) : (
                <form onSubmit={handleSubmitToken(onTenantTokenSubmit)} className="space-y-6">
                  <div>
                    <h3 className="text-lg font-semibold mb-4">Token Lifetimes</h3>
                    <div className="grid grid-cols-3 gap-4">
                      <div className="space-y-2">
                        <Label htmlFor="accessTokenTTLMinutes">Access Token TTL (minutes)</Label>
                        <Input
                          id="accessTokenTTLMinutes"
                          type="number"
                          {...registerToken('accessTokenTTLMinutes', { valueAsNumber: true })}
                          disabled={isLoading || settingsLoading}
                        />
                        {tokenErrors.accessTokenTTLMinutes && (
                          <p className="text-sm text-red-600">
                            {tokenErrors.accessTokenTTLMinutes.message}
                          </p>
                        )}
                      </div>

                      <div className="space-y-2">
                        <Label htmlFor="refreshTokenTTLDays">Refresh Token TTL (days)</Label>
                        <Input
                          id="refreshTokenTTLDays"
                          type="number"
                          {...registerToken('refreshTokenTTLDays', { valueAsNumber: true })}
                          disabled={isLoading || settingsLoading}
                        />
                        {tokenErrors.refreshTokenTTLDays && (
                          <p className="text-sm text-red-600">
                            {tokenErrors.refreshTokenTTLDays.message}
                          </p>
                        )}
                      </div>

                      <div className="space-y-2">
                        <Label htmlFor="idTokenTTLMinutes">ID Token TTL (minutes)</Label>
                        <Input
                          id="idTokenTTLMinutes"
                          type="number"
                          {...registerToken('idTokenTTLMinutes', { valueAsNumber: true })}
                          disabled={isLoading || settingsLoading}
                        />
                        {tokenErrors.idTokenTTLMinutes && (
                          <p className="text-sm text-red-600">
                            {tokenErrors.idTokenTTLMinutes.message}
                          </p>
                        )}
                      </div>
                    </div>
                  </div>

                  <div>
                    <h3 className="text-lg font-semibold mb-4">Remember Me</h3>
                    <div className="space-y-4">
                      <div className="flex items-center space-x-2">
                        <input
                          type="checkbox"
                          id="rememberMeEnabled"
                          {...registerToken('rememberMeEnabled')}
                          disabled={isLoading || settingsLoading}
                          className="rounded"
                        />
                        <Label htmlFor="rememberMeEnabled">Enable Remember Me</Label>
                      </div>

                      {rememberMeEnabled && (
                        <div className="grid grid-cols-2 gap-4 pl-6">
                          <div className="space-y-2">
                            <Label htmlFor="rememberMeRefreshTokenTTLDays">
                              Remember Me Refresh Token TTL (days)
                            </Label>
                            <Input
                              id="rememberMeRefreshTokenTTLDays"
                              type="number"
                              {...registerToken('rememberMeRefreshTokenTTLDays', { valueAsNumber: true })}
                              disabled={isLoading || settingsLoading}
                            />
                            {tokenErrors.rememberMeRefreshTokenTTLDays && (
                              <p className="text-sm text-red-600">
                                {tokenErrors.rememberMeRefreshTokenTTLDays.message}
                              </p>
                            )}
                          </div>

                          <div className="space-y-2">
                            <Label htmlFor="rememberMeAccessTokenTTLMinutes">
                              Remember Me Access Token TTL (minutes)
                            </Label>
                            <Input
                              id="rememberMeAccessTokenTTLMinutes"
                              type="number"
                              {...registerToken('rememberMeAccessTokenTTLMinutes', { valueAsNumber: true })}
                              disabled={isLoading || settingsLoading}
                            />
                            {tokenErrors.rememberMeAccessTokenTTLMinutes && (
                              <p className="text-sm text-red-600">
                                {tokenErrors.rememberMeAccessTokenTTLMinutes.message}
                              </p>
                            )}
                          </div>
                        </div>
                      )}
                    </div>
                  </div>

                  <div>
                    <h3 className="text-lg font-semibold mb-4">Token Security</h3>
                    <div className="space-y-4">
                      <div className="flex items-center space-x-2">
                        <input
                          type="checkbox"
                          id="tokenRotationEnabled"
                          {...registerToken('tokenRotationEnabled')}
                          disabled={isLoading || settingsLoading}
                          className="rounded"
                        />
                        <Label htmlFor="tokenRotationEnabled">Enable Token Rotation</Label>
                      </div>

                      <div className="flex items-center space-x-2">
                        <input
                          type="checkbox"
                          id="requireMFAForExtendedSessions"
                          {...registerToken('requireMFAForExtendedSessions')}
                          disabled={isLoading || settingsLoading}
                          className="rounded"
                        />
                        <Label htmlFor="requireMFAForExtendedSessions">
                          Require MFA for Extended Sessions
                        </Label>
                      </div>
                    </div>
                  </div>

                  <Button type="submit" disabled={isLoading || settingsLoading}>
                    {isLoading ? 'Saving...' : 'Save Tenant Settings'}
                  </Button>
                </form>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
