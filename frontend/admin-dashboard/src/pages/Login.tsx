/**
 * Login Page Component
 */

import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useAuthStore } from '../store/authStore';
import { authApi, mfaApi } from '../services/api';
import { handleApiError } from '../services/api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Checkbox } from '@/components/ui/checkbox';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert } from '@/components/ui/alert';

const loginSchema = z.object({
  username: z.string().min(1, 'Username is required'),
  password: z.string().min(1, 'Password is required'),
  tenantId: z.string().optional(),
  rememberMe: z.boolean().optional(),
});

type LoginFormData = z.infer<typeof loginSchema>;

export function Login() {
  const navigate = useNavigate();
  const { setAuth, userId, tenantId } = useAuthStore();
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [mfaRequired, setMfaRequired] = useState(false);
  const [mfaCode, setMfaCode] = useState('');
  const [mfaChallengeId, setMfaChallengeId] = useState<string | null>(null);
  const [mfaVerifying, setMfaVerifying] = useState(false);

  const {
    register,
    handleSubmit,
    watch,
    setValue,
    formState: { errors },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      rememberMe: false,
    },
  });

  const rememberMe = watch('rememberMe');

  const onSubmit = async (data: LoginFormData) => {
    setIsLoading(true);
    setError(null);

    try {
      const response = await authApi.login({
        username: data.username,
        password: data.password,
        tenant_id: data.tenantId,
        remember_me: data.rememberMe || false,
      });

      // Check if MFA is required
      if (response.mfa_required) {
        setMfaRequired(true);
        // Initiate MFA challenge
        try {
          const challengeResponse = await mfaApi.challenge({
            user_id: '', // Will be set by backend from login context
            tenant_id: data.tenantId,
          });
          setMfaChallengeId(challengeResponse.challenge_id);
        } catch (err) {
          setError('Failed to initiate MFA challenge: ' + handleApiError(err));
          setMfaRequired(false);
        }
        setIsLoading(false);
        return;
      }

      // If no MFA required, proceed with normal login
      if (!response.access_token) {
        throw new Error('No access token received');
      }

      // Extract user info from JWT token
      const { extractUserInfo } = await import('@/../../shared/utils/jwt-decoder');
      const userInfo = extractUserInfo(response.access_token);

      // Ensure principalType is valid (SYSTEM or TENANT only, not SERVICE)
      const validPrincipalType: 'SYSTEM' | 'TENANT' = 
        (userInfo.principalType === 'SYSTEM' || userInfo.principalType === 'TENANT')
          ? userInfo.principalType
          : 'TENANT'; // Default to TENANT if invalid or SERVICE

      // Store auth data (including principal_type and permissions)
      setAuth({
        accessToken: response.access_token,
        refreshToken: response.refresh_token,
        tenantId: userInfo.tenantId || data.tenantId || null,
        principalType: validPrincipalType,
        systemPermissions: userInfo.systemPermissions,
        permissions: userInfo.permissions,
      });

      // Redirect to dashboard
      navigate('/');
    } catch (err) {
      setError(handleApiError(err));
    } finally {
      setIsLoading(false);
    }
  };

  const onMfaVerify = async () => {
    if (!mfaCode || !mfaChallengeId) {
      setError('Please enter the verification code');
      return;
    }

    setMfaVerifying(true);
    setError(null);

    try {
      const response = await mfaApi.verifyChallenge({
        challenge_id: mfaChallengeId,
        code: mfaCode,
      });

      // Extract user info from JWT token
      const { extractUserInfo } = await import('@/../../shared/utils/jwt-decoder');
      const userInfo = extractUserInfo(response.access_token);

      // Ensure principalType is valid
      const validPrincipalType: 'SYSTEM' | 'TENANT' = 
        (userInfo.principalType === 'SYSTEM' || userInfo.principalType === 'TENANT')
          ? userInfo.principalType
          : 'TENANT';

      // Store auth data
      setAuth({
        accessToken: response.access_token,
        refreshToken: undefined,
        tenantId: userInfo.tenantId || null,
        principalType: validPrincipalType,
        systemPermissions: userInfo.systemPermissions,
        permissions: userInfo.permissions,
      });

      // Redirect to dashboard
      navigate('/');
    } catch (err) {
      setError(handleApiError(err));
    } finally {
      setMfaVerifying(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>ARauth Identity Admin</CardTitle>
          <CardDescription>Sign in to your account</CardDescription>
        </CardHeader>
        <CardContent>
          {!mfaRequired ? (
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
              {error && (
                <Alert className="bg-red-50 border-red-200 text-red-700">
                  {error}
                </Alert>
              )}

              <div className="space-y-2">
                <Label htmlFor="username">Username</Label>
                <Input
                  id="username"
                  type="text"
                  {...register('username')}
                  placeholder="Enter your username"
                  disabled={isLoading}
                />
                {errors.username && (
                  <p className="text-sm text-red-600">{errors.username.message}</p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="password">Password</Label>
                <Input
                  id="password"
                  type="password"
                  {...register('password')}
                  placeholder="Enter your password"
                  disabled={isLoading}
                />
                {errors.password && (
                  <p className="text-sm text-red-600">{errors.password.message}</p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="tenantId">Tenant ID (Optional)</Label>
                <Input
                  id="tenantId"
                  type="text"
                  {...register('tenantId')}
                  placeholder="Enter tenant ID"
                  disabled={isLoading}
                />
              </div>

              <div className="flex items-center space-x-2">
                <Checkbox
                  id="rememberMe"
                  checked={rememberMe}
                  onCheckedChange={(checked) => {
                    setValue('rememberMe', checked === true);
                  }}
                  disabled={isLoading}
                />
                <Label
                  htmlFor="rememberMe"
                  className="text-sm font-normal cursor-pointer"
                >
                  Remember me
                </Label>
              </div>

              <Button type="submit" className="w-full" disabled={isLoading}>
                {isLoading ? 'Signing in...' : 'Sign In'}
              </Button>
            </form>
          ) : (
            <div className="space-y-4">
              <Alert>
                Multi-factor authentication is required. Please enter the verification code from your authenticator app.
              </Alert>

              {error && (
                <Alert className="bg-red-50 border-red-200 text-red-700">
                  {error}
                </Alert>
              )}

              <div className="space-y-2">
                <Label htmlFor="mfaCode">Verification Code</Label>
                <Input
                  id="mfaCode"
                  type="text"
                  placeholder="Enter 6-digit code"
                  value={mfaCode}
                  onChange={(e) => setMfaCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
                  maxLength={6}
                  disabled={mfaVerifying}
                  className="font-mono text-lg tracking-widest"
                />
              </div>

              <div className="flex space-x-2">
                <Button
                  variant="outline"
                  className="flex-1"
                  onClick={() => {
                    setMfaRequired(false);
                    setMfaCode('');
                    setMfaChallengeId(null);
                    setError(null);
                  }}
                  disabled={mfaVerifying}
                >
                  Back
                </Button>
                <Button
                  className="flex-1"
                  onClick={onMfaVerify}
                  disabled={mfaCode.length !== 6 || mfaVerifying}
                >
                  {mfaVerifying ? 'Verifying...' : 'Verify'}
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

