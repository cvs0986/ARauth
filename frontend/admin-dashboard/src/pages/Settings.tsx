/**
 * System Settings Page
 */

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Alert } from '@/components/ui/alert';
import { Shield, Key, Settings as SettingsIcon, Mail, Clock } from 'lucide-react';

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

export function Settings() {
  const [activeTab, setActiveTab] = useState('security');
  const [success, setSuccess] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

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

  // Token Settings Form
  const {
    register: registerToken,
    handleSubmit: handleSubmitToken,
    watch,
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

  const rememberMeEnabled = watch('rememberMeEnabled');

  const onSecuritySubmit = async (data: SecuritySettings) => {
    setIsLoading(true);
    setError(null);
    setSuccess(null);

    try {
      // TODO: Implement API call to save security settings
      console.log('Security settings:', data);
      await new Promise((resolve) => setTimeout(resolve, 500)); // Simulate API call
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

  const onTokenSubmit = async (data: TokenSettings) => {
    setIsLoading(true);
    setError(null);
    setSuccess(null);

    try {
      // TODO: Implement API call to save token settings
      // await tokenSettingsApi.update(data);
      console.log('Token settings:', data);
      setSuccess('Token settings saved successfully');
    } catch (err) {
      setError('Failed to save token settings');
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

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">System Settings</h1>
        <p className="text-gray-600 mt-1">Configure system-wide settings and policies</p>
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
          <TabsTrigger value="security">
            <Shield className="mr-2 h-4 w-4" />
            Security
          </TabsTrigger>
          <TabsTrigger value="oauth">
            <Key className="mr-2 h-4 w-4" />
            OAuth2/OIDC
          </TabsTrigger>
          <TabsTrigger value="system">
            <SettingsIcon className="mr-2 h-4 w-4" />
            System
          </TabsTrigger>
          <TabsTrigger value="tokens">
            <Clock className="mr-2 h-4 w-4" />
            Token Settings
          </TabsTrigger>
        </TabsList>

        {/* Security Settings */}
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
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="rateLimitWindow">Window (seconds)</Label>
                      <Input
                        id="rateLimitWindow"
                        type="number"
                        {...registerSecurity('rateLimitWindow', { valueAsNumber: true })}
                        disabled={isLoading}
                      />
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

        {/* OAuth2/OIDC Settings */}
        <TabsContent value="oauth">
          <Card>
            <CardHeader>
              <CardTitle>OAuth2/OIDC Settings</CardTitle>
              <CardDescription>
                Configure ORY Hydra integration and token settings
              </CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleSubmitOAuth(onOAuthSubmit)} className="space-y-6">
                <div>
                  <h3 className="text-lg font-semibold mb-4">Hydra Configuration</h3>
                  <div className="space-y-4">
                    <div className="space-y-2">
                      <Label htmlFor="hydraAdminUrl">Hydra Admin URL</Label>
                      <Input
                        id="hydraAdminUrl"
                        type="url"
                        {...registerOAuth('hydraAdminUrl')}
                        disabled={isLoading}
                        placeholder="http://localhost:4445"
                      />
                      {oauthErrors.hydraAdminUrl && (
                        <p className="text-sm text-red-600">{oauthErrors.hydraAdminUrl.message}</p>
                      )}
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="hydraPublicUrl">Hydra Public URL</Label>
                      <Input
                        id="hydraPublicUrl"
                        type="url"
                        {...registerOAuth('hydraPublicUrl')}
                        disabled={isLoading}
                        placeholder="http://localhost:4444"
                      />
                      {oauthErrors.hydraPublicUrl && (
                        <p className="text-sm text-red-600">{oauthErrors.hydraPublicUrl.message}</p>
                      )}
                    </div>
                  </div>
                </div>

                <div>
                  <h3 className="text-lg font-semibold mb-4">Token TTL Settings</h3>
                  <div className="grid grid-cols-3 gap-4">
                    <div className="space-y-2">
                      <Label htmlFor="accessTokenTTL">Access Token TTL (seconds)</Label>
                      <Input
                        id="accessTokenTTL"
                        type="number"
                        {...registerOAuth('accessTokenTTL', { valueAsNumber: true })}
                        disabled={isLoading}
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="refreshTokenTTL">Refresh Token TTL (seconds)</Label>
                      <Input
                        id="refreshTokenTTL"
                        type="number"
                        {...registerOAuth('refreshTokenTTL', { valueAsNumber: true })}
                        disabled={isLoading}
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="idTokenTTL">ID Token TTL (seconds)</Label>
                      <Input
                        id="idTokenTTL"
                        type="number"
                        {...registerOAuth('idTokenTTL', { valueAsNumber: true })}
                        disabled={isLoading}
                      />
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

        {/* System Settings */}
        <TabsContent value="system">
          <Card>
            <CardHeader>
              <CardTitle>System Configuration</CardTitle>
              <CardDescription>
                Configure JWT settings, session management, and security policies
              </CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleSubmitSystem(onSystemSubmit)} className="space-y-6">
                <div>
                  <h3 className="text-lg font-semibold mb-4">JWT Configuration</h3>
                  <div className="space-y-4">
                    <div className="space-y-2">
                      <Label htmlFor="jwtSecret">JWT Secret (leave empty to keep current)</Label>
                      <Input
                        id="jwtSecret"
                        type="password"
                        {...registerSystem('jwtSecret')}
                        disabled={isLoading}
                        placeholder="Enter new secret (min 32 characters)"
                      />
                      {systemErrors.jwtSecret && (
                        <p className="text-sm text-red-600">{systemErrors.jwtSecret.message}</p>
                      )}
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="jwtIssuer">JWT Issuer</Label>
                      <Input
                        id="jwtIssuer"
                        {...registerSystem('jwtIssuer')}
                        disabled={isLoading}
                        placeholder="arauth-identity"
                      />
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="jwtAudience">JWT Audience</Label>
                      <Input
                        id="jwtAudience"
                        {...registerSystem('jwtAudience')}
                        disabled={isLoading}
                        placeholder="Optional"
                      />
                    </div>
                  </div>
                </div>

                <div>
                  <h3 className="text-lg font-semibold mb-4">Session Management</h3>
                  <div className="space-y-4">
                    <div className="space-y-2">
                      <Label htmlFor="sessionTimeout">Session Timeout (seconds)</Label>
                      <Input
                        id="sessionTimeout"
                        type="number"
                        {...registerSystem('sessionTimeout', { valueAsNumber: true })}
                        disabled={isLoading}
                      />
                    </div>
                  </div>
                </div>

                <div>
                  <h3 className="text-lg font-semibold mb-4">Account Lockout</h3>
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label htmlFor="maxLoginAttempts">Max Login Attempts</Label>
                      <Input
                        id="maxLoginAttempts"
                        type="number"
                        {...registerSystem('maxLoginAttempts', { valueAsNumber: true })}
                        disabled={isLoading}
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="lockoutDuration">Lockout Duration (seconds)</Label>
                      <Input
                        id="lockoutDuration"
                        type="number"
                        {...registerSystem('lockoutDuration', { valueAsNumber: true })}
                        disabled={isLoading}
                      />
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

        {/* Token Settings */}
        <TabsContent value="tokens">
          <Card>
            <CardHeader>
              <CardTitle>Token Lifetime Settings</CardTitle>
              <CardDescription>
                Configure JWT token lifetimes and Remember Me settings
              </CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleSubmitToken(onTokenSubmit)} className="space-y-6">
                <div>
                  <h3 className="text-lg font-semibold mb-4">Default Token Lifetimes</h3>
                  <div className="grid grid-cols-3 gap-4">
                    <div className="space-y-2">
                      <Label htmlFor="accessTokenTTLMinutes">Access Token TTL (minutes)</Label>
                      <Input
                        id="accessTokenTTLMinutes"
                        type="number"
                        {...registerToken('accessTokenTTLMinutes', { valueAsNumber: true })}
                        disabled={isLoading}
                        placeholder="15"
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
                        disabled={isLoading}
                        placeholder="30"
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
                        disabled={isLoading}
                        placeholder="60"
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
                  <h3 className="text-lg font-semibold mb-4">Remember Me Settings</h3>
                  <div className="space-y-4">
                    <div className="flex items-center space-x-2">
                      <input
                        type="checkbox"
                        id="rememberMeEnabled"
                        {...registerToken('rememberMeEnabled')}
                        disabled={isLoading}
                        className="rounded"
                      />
                      <Label htmlFor="rememberMeEnabled">Enable Remember Me</Label>
                    </div>

                    {rememberMeEnabled && (
                      <div className="ml-6 space-y-4 border-l-2 pl-4">
                        <div className="grid grid-cols-2 gap-4">
                          <div className="space-y-2">
                            <Label htmlFor="rememberMeRefreshTokenTTLDays">
                              Remember Me Refresh Token TTL (days)
                            </Label>
                            <Input
                              id="rememberMeRefreshTokenTTLDays"
                              type="number"
                              {...registerToken('rememberMeRefreshTokenTTLDays', { valueAsNumber: true })}
                              disabled={isLoading}
                              placeholder="90"
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
                              disabled={isLoading}
                              placeholder="60"
                            />
                            {tokenErrors.rememberMeAccessTokenTTLMinutes && (
                              <p className="text-sm text-red-600">
                                {tokenErrors.rememberMeAccessTokenTTLMinutes.message}
                              </p>
                            )}
                          </div>
                        </div>
                      </div>
                    )}
                  </div>
                </div>

                <div>
                  <h3 className="text-lg font-semibold mb-4">Security Options</h3>
                  <div className="space-y-4">
                    <div className="flex items-center space-x-2">
                      <input
                        type="checkbox"
                        id="tokenRotationEnabled"
                        {...registerToken('tokenRotationEnabled')}
                        disabled={isLoading}
                        className="rounded"
                      />
                      <Label htmlFor="tokenRotationEnabled">
                        Enable Token Rotation (recommended)
                      </Label>
                    </div>
                    <div className="flex items-center space-x-2">
                      <input
                        type="checkbox"
                        id="requireMFAForExtendedSessions"
                        {...registerToken('requireMFAForExtendedSessions')}
                        disabled={isLoading}
                        className="rounded"
                      />
                      <Label htmlFor="requireMFAForExtendedSessions">
                        Require MFA for extended sessions (Remember Me)
                      </Label>
                    </div>
                  </div>
                </div>

                <Button type="submit" disabled={isLoading}>
                  {isLoading ? 'Saving...' : 'Save Token Settings'}
                </Button>
              </form>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}

