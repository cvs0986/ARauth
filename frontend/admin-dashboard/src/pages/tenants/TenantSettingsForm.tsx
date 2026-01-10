/**
 * Tenant Settings Form Component
 */

import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { systemApi } from '@/services/api';
import { handleApiError } from '@/services/api';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useState, useEffect } from 'react';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Loader2, Save, AlertCircle } from 'lucide-react';
import type { TenantSettings, UpdateTenantSettingsRequest } from '@shared/types/api';

const tenantSettingsSchema = z.object({
  // Token lifetimes
  access_token_ttl_minutes: z.number().min(1).max(1440),
  refresh_token_ttl_days: z.number().min(1).max(365),
  id_token_ttl_minutes: z.number().min(1).max(1440),
  // Remember Me
  remember_me_enabled: z.boolean(),
  remember_me_refresh_token_ttl_days: z.number().min(1).max(365).optional(),
  remember_me_access_token_ttl_minutes: z.number().min(1).max(1440).optional(),
  // Token rotation
  token_rotation_enabled: z.boolean(),
  require_mfa_for_extended_sessions: z.boolean(),
  // Password policy
  min_password_length: z.number().min(8).max(128),
  require_uppercase: z.boolean(),
  require_lowercase: z.boolean(),
  require_numbers: z.boolean(),
  require_special_chars: z.boolean(),
  password_expiry_days: z.number().min(0).max(365).nullable().optional(),
  // MFA
  mfa_required: z.boolean(),
  // Rate limiting
  rate_limit_requests: z.number().min(1).max(10000),
  rate_limit_window_seconds: z.number().min(1).max(3600),
});

type TenantSettingsFormData = z.infer<typeof tenantSettingsSchema>;

interface TenantSettingsFormProps {
  tenantId: string;
}

export function TenantSettingsForm({ tenantId }: TenantSettingsFormProps) {
  const queryClient = useQueryClient();
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  // Fetch tenant settings
  const { data: settings, isLoading, error: fetchError } = useQuery({
    queryKey: ['tenant', 'settings', tenantId],
    queryFn: () => systemApi.tenants.getSettings(tenantId),
    enabled: !!tenantId,
  });

  const {
    register,
    handleSubmit,
    formState: { errors, isDirty },
    reset,
    watch,
  } = useForm<TenantSettingsFormData>({
    resolver: zodResolver(tenantSettingsSchema),
    defaultValues: {
      access_token_ttl_minutes: 15,
      refresh_token_ttl_days: 30,
      id_token_ttl_minutes: 60,
      remember_me_enabled: true,
      remember_me_refresh_token_ttl_days: 90,
      remember_me_access_token_ttl_minutes: 60,
      token_rotation_enabled: true,
      require_mfa_for_extended_sessions: false,
      min_password_length: 12,
      require_uppercase: true,
      require_lowercase: true,
      require_numbers: true,
      require_special_chars: true,
      password_expiry_days: null,
      mfa_required: false,
      rate_limit_requests: 100,
      rate_limit_window_seconds: 60,
    },
  });

  const rememberMeEnabled = watch('remember_me_enabled');

  // Reset form when settings are loaded
  useEffect(() => {
    if (settings) {
      // If settings don't exist (message field present), use defaults
      if ((settings as any).message) {
        // Settings don't exist, keep default values
        return;
      }
      
      // Settings exist, populate form
      const tenantSettings = settings as TenantSettings;
      reset({
        access_token_ttl_minutes: tenantSettings.access_token_ttl_minutes || 15,
        refresh_token_ttl_days: tenantSettings.refresh_token_ttl_days || 30,
        id_token_ttl_minutes: tenantSettings.id_token_ttl_minutes || 60,
        remember_me_enabled: tenantSettings.remember_me_enabled ?? true,
        remember_me_refresh_token_ttl_days: tenantSettings.remember_me_refresh_token_ttl_days || 90,
        remember_me_access_token_ttl_minutes: tenantSettings.remember_me_access_token_ttl_minutes || 60,
        token_rotation_enabled: tenantSettings.token_rotation_enabled ?? true,
        require_mfa_for_extended_sessions: tenantSettings.require_mfa_for_extended_sessions ?? false,
        min_password_length: tenantSettings.min_password_length || 12,
        require_uppercase: tenantSettings.require_uppercase ?? true,
        require_lowercase: tenantSettings.require_lowercase ?? true,
        require_numbers: tenantSettings.require_numbers ?? true,
        require_special_chars: tenantSettings.require_special_chars ?? true,
        password_expiry_days: tenantSettings.password_expiry_days ?? null,
        mfa_required: tenantSettings.mfa_required ?? false,
        rate_limit_requests: tenantSettings.rate_limit_requests || 100,
        rate_limit_window_seconds: tenantSettings.rate_limit_window_seconds || 60,
      });
    }
  }, [settings, reset]);

  const mutation = useMutation({
    mutationFn: (data: UpdateTenantSettingsRequest) =>
      systemApi.tenants.updateSettings(tenantId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenant', 'settings', tenantId] });
      setError(null);
      setSuccess(true);
      setTimeout(() => setSuccess(false), 3000);
    },
    onError: (err) => {
      setError(handleApiError(err));
      setSuccess(false);
    },
  });

  const onSubmit = (data: TenantSettingsFormData) => {
    setError(null);
    setSuccess(false);
    mutation.mutate(data);
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="h-6 w-6 animate-spin text-gray-500" />
        <span className="ml-2 text-gray-500">Loading settings...</span>
      </div>
    );
  }

  if (fetchError) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          Error loading settings: {fetchError instanceof Error ? fetchError.message : 'Unknown error'}
        </AlertDescription>
      </Alert>
    );
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {success && (
        <Alert className="bg-green-50 border-green-200 text-green-800">
          <AlertDescription>Settings updated successfully!</AlertDescription>
        </Alert>
      )}

      {/* Token Lifetime Settings */}
      <Card>
        <CardHeader>
          <CardTitle>Token Lifetime Settings</CardTitle>
          <CardDescription>Configure token expiration times</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="space-y-2">
              <Label htmlFor="access_token_ttl_minutes">
                Access Token TTL (minutes)
              </Label>
              <Input
                id="access_token_ttl_minutes"
                type="number"
                {...register('access_token_ttl_minutes', { valueAsNumber: true })}
                min={1}
                max={1440}
                disabled={mutation.isPending}
              />
              {errors.access_token_ttl_minutes && (
                <p className="text-sm text-red-600">
                  {errors.access_token_ttl_minutes.message}
                </p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="refresh_token_ttl_days">
                Refresh Token TTL (days)
              </Label>
              <Input
                id="refresh_token_ttl_days"
                type="number"
                {...register('refresh_token_ttl_days', { valueAsNumber: true })}
                min={1}
                max={365}
                disabled={mutation.isPending}
              />
              {errors.refresh_token_ttl_days && (
                <p className="text-sm text-red-600">
                  {errors.refresh_token_ttl_days.message}
                </p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="id_token_ttl_minutes">
                ID Token TTL (minutes)
              </Label>
              <Input
                id="id_token_ttl_minutes"
                type="number"
                {...register('id_token_ttl_minutes', { valueAsNumber: true })}
                min={1}
                max={1440}
                disabled={mutation.isPending}
              />
              {errors.id_token_ttl_minutes && (
                <p className="text-sm text-red-600">
                  {errors.id_token_ttl_minutes.message}
                </p>
              )}
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Remember Me Settings */}
      <Card>
        <CardHeader>
          <CardTitle>Remember Me Settings</CardTitle>
          <CardDescription>Configure extended session settings</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center space-x-2">
            <input
              type="checkbox"
              id="remember_me_enabled"
              {...register('remember_me_enabled')}
              disabled={mutation.isPending}
              className="h-4 w-4 rounded border-gray-300"
            />
            <Label htmlFor="remember_me_enabled" className="font-normal">
              Enable Remember Me
            </Label>
          </div>

          {rememberMeEnabled && (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 pl-6 border-l-2 border-gray-200">
              <div className="space-y-2">
                <Label htmlFor="remember_me_refresh_token_ttl_days">
                  Remember Me Refresh Token TTL (days)
                </Label>
                <Input
                  id="remember_me_refresh_token_ttl_days"
                  type="number"
                  {...register('remember_me_refresh_token_ttl_days', {
                    valueAsNumber: true,
                  })}
                  min={1}
                  max={365}
                  disabled={mutation.isPending}
                />
                {errors.remember_me_refresh_token_ttl_days && (
                  <p className="text-sm text-red-600">
                    {errors.remember_me_refresh_token_ttl_days.message}
                  </p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="remember_me_access_token_ttl_minutes">
                  Remember Me Access Token TTL (minutes)
                </Label>
                <Input
                  id="remember_me_access_token_ttl_minutes"
                  type="number"
                  {...register('remember_me_access_token_ttl_minutes', {
                    valueAsNumber: true,
                  })}
                  min={1}
                  max={1440}
                  disabled={mutation.isPending}
                />
                {errors.remember_me_access_token_ttl_minutes && (
                  <p className="text-sm text-red-600">
                    {errors.remember_me_access_token_ttl_minutes.message}
                  </p>
                )}
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Token Security Settings */}
      <Card>
        <CardHeader>
          <CardTitle>Token Security</CardTitle>
          <CardDescription>Configure token security features</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center space-x-2">
            <input
              type="checkbox"
              id="token_rotation_enabled"
              {...register('token_rotation_enabled')}
              disabled={mutation.isPending}
              className="h-4 w-4 rounded border-gray-300"
            />
            <Label htmlFor="token_rotation_enabled" className="font-normal">
              Enable Token Rotation
            </Label>
          </div>

          <div className="flex items-center space-x-2">
            <input
              type="checkbox"
              id="require_mfa_for_extended_sessions"
              {...register('require_mfa_for_extended_sessions')}
              disabled={mutation.isPending}
              className="h-4 w-4 rounded border-gray-300"
            />
            <Label htmlFor="require_mfa_for_extended_sessions" className="font-normal">
              Require MFA for Extended Sessions
            </Label>
          </div>
        </CardContent>
      </Card>

      {/* Password Policy */}
      <Card>
        <CardHeader>
          <CardTitle>Password Policy</CardTitle>
          <CardDescription>Configure password requirements</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="min_password_length">
                Minimum Password Length
              </Label>
              <Input
                id="min_password_length"
                type="number"
                {...register('min_password_length', { valueAsNumber: true })}
                min={8}
                max={128}
                disabled={mutation.isPending}
              />
              {errors.min_password_length && (
                <p className="text-sm text-red-600">
                  {errors.min_password_length.message}
                </p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="password_expiry_days">
                Password Expiry (days, 0 = never)
              </Label>
              <Input
                id="password_expiry_days"
                type="number"
                {...register('password_expiry_days', {
                  valueAsNumber: true,
                  setValueAs: (v) => (v === '' || v === 0 ? null : Number(v)),
                })}
                min={0}
                max={365}
                disabled={mutation.isPending}
              />
              {errors.password_expiry_days && (
                <p className="text-sm text-red-600">
                  {errors.password_expiry_days.message}
                </p>
              )}
            </div>
          </div>

          <div className="border-t border-gray-200 my-4" />

          <div className="space-y-3">
            <div className="flex items-center space-x-2">
              <input
                type="checkbox"
                id="require_uppercase"
                {...register('require_uppercase')}
                disabled={mutation.isPending}
                className="h-4 w-4 rounded border-gray-300"
              />
              <Label htmlFor="require_uppercase" className="font-normal">
                Require Uppercase Letters
              </Label>
            </div>

            <div className="flex items-center space-x-2">
              <input
                type="checkbox"
                id="require_lowercase"
                {...register('require_lowercase')}
                disabled={mutation.isPending}
                className="h-4 w-4 rounded border-gray-300"
              />
              <Label htmlFor="require_lowercase" className="font-normal">
                Require Lowercase Letters
              </Label>
            </div>

            <div className="flex items-center space-x-2">
              <input
                type="checkbox"
                id="require_numbers"
                {...register('require_numbers')}
                disabled={mutation.isPending}
                className="h-4 w-4 rounded border-gray-300"
              />
              <Label htmlFor="require_numbers" className="font-normal">
                Require Numbers
              </Label>
            </div>

            <div className="flex items-center space-x-2">
              <input
                type="checkbox"
                id="require_special_chars"
                {...register('require_special_chars')}
                disabled={mutation.isPending}
                className="h-4 w-4 rounded border-gray-300"
              />
              <Label htmlFor="require_special_chars" className="font-normal">
                Require Special Characters
              </Label>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* MFA Settings */}
      <Card>
        <CardHeader>
          <CardTitle>Multi-Factor Authentication</CardTitle>
          <CardDescription>Configure MFA requirements</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center space-x-2">
            <input
              type="checkbox"
              id="mfa_required"
              {...register('mfa_required')}
              disabled={mutation.isPending}
              className="h-4 w-4 rounded border-gray-300"
            />
            <Label htmlFor="mfa_required" className="font-normal">
              Require MFA for All Users
            </Label>
          </div>
        </CardContent>
      </Card>

      {/* Rate Limiting */}
      <Card>
        <CardHeader>
          <CardTitle>Rate Limiting</CardTitle>
          <CardDescription>Configure API rate limits</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="rate_limit_requests">
                Max Requests per Window
              </Label>
              <Input
                id="rate_limit_requests"
                type="number"
                {...register('rate_limit_requests', { valueAsNumber: true })}
                min={1}
                max={10000}
                disabled={mutation.isPending}
              />
              {errors.rate_limit_requests && (
                <p className="text-sm text-red-600">
                  {errors.rate_limit_requests.message}
                </p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="rate_limit_window_seconds">
                Rate Limit Window (seconds)
              </Label>
              <Input
                id="rate_limit_window_seconds"
                type="number"
                {...register('rate_limit_window_seconds', { valueAsNumber: true })}
                min={1}
                max={3600}
                disabled={mutation.isPending}
              />
              {errors.rate_limit_window_seconds && (
                <p className="text-sm text-red-600">
                  {errors.rate_limit_window_seconds.message}
                </p>
              )}
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Submit Button */}
      <div className="flex justify-end">
        <Button
          type="submit"
          disabled={mutation.isPending || !isDirty}
          className="min-w-[120px]"
        >
          {mutation.isPending ? (
            <>
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              Saving...
            </>
          ) : (
            <>
              <Save className="h-4 w-4 mr-2" />
              Save Settings
            </>
          )}
        </Button>
      </div>
    </form>
  );
}

