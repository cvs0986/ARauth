/**
 * Login Page Component - Redesigned with MFA Enrollment Flow
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
import { Alert, AlertDescription } from '@/components/ui/alert';
import { ARauthLogo } from '@/components/ARauthLogo';
import { Shield, Lock, KeyRound, QrCode, CheckCircle2, AlertCircle } from 'lucide-react';

const loginSchema = z.object({
  username: z.string().min(1, 'Username is required'),
  password: z.string().min(1, 'Password is required'),
  tenantId: z.string().optional(),
  rememberMe: z.boolean().optional(),
});

type LoginFormData = z.infer<typeof loginSchema>;

type MFAState = 'none' | 'enrollment' | 'verification';

export function Login() {
  const navigate = useNavigate();
  const { setAuth } = useAuthStore();
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [mfaState, setMfaState] = useState<MFAState>('none');
  const [mfaCode, setMfaCode] = useState('');
  const [mfaChallengeId, setMfaChallengeId] = useState<string | null>(null);
  const [mfaSessionId, setMfaSessionId] = useState<string | null>(null);
  const [mfaVerifying, setMfaVerifying] = useState(false);
  const [enrollData, setEnrollData] = useState<{
    secret: string;
    qr_code: string;
    recovery_codes: string[];
  } | null>(null);
  const [enrollmentStep, setEnrollmentStep] = useState<'qr' | 'verify'>('qr');
  const [userId, setUserId] = useState<string | null>(null);
  const [tenantId, setTenantId] = useState<string | null>(null);

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
        if (!response.user_id) {
          setError('MFA is required but user ID is missing');
          setIsLoading(false);
          return;
        }

        setUserId(response.user_id);
        setTenantId(response.tenant_id || data.tenantId || null);

          // Check if enrollment is required
          if (response.mfa_enrollment_required) {
            // User needs to enroll in MFA
            setMfaState('enrollment');
            try {
              const challengeResponse = await mfaApi.challenge({
                user_id: response.user_id,
                tenant_id: response.tenant_id || data.tenantId || '',
              });
              const sessionId = challengeResponse.session_id || challengeResponse.challenge_id;
              if (!sessionId) {
                throw new Error('No session ID received from challenge');
              }
              setMfaSessionId(sessionId);
              
              // Start enrollment
              const enrollResponse = await mfaApi.enrollForLogin({
                session_id: sessionId,
              });
              setEnrollData(enrollResponse);
              setEnrollmentStep('qr');
            } catch (err) {
              setError('Failed to start MFA enrollment: ' + handleApiError(err));
              setMfaState('none');
            }
          } else {
            // User already enrolled, just needs to verify
            setMfaState('verification');
            try {
              const challengeResponse = await mfaApi.challenge({
                user_id: response.user_id,
                tenant_id: response.tenant_id || data.tenantId || '',
              });
              const sessionId = challengeResponse.session_id || challengeResponse.challenge_id;
              if (!sessionId) {
                throw new Error('No session ID received from challenge');
              }
              setMfaChallengeId(sessionId);
            } catch (err) {
              setError('Failed to initiate MFA challenge: ' + handleApiError(err));
              setMfaState('none');
            }
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
        systemRoles: userInfo.systemRoles,
        username: userInfo.username,
        email: userInfo.email,
        userId: userInfo.userId,
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
        tenantId: userInfo.tenantId || tenantId || null,
        principalType: validPrincipalType,
        systemPermissions: userInfo.systemPermissions,
        permissions: userInfo.permissions,
        systemRoles: userInfo.systemRoles,
        username: userInfo.username,
        email: userInfo.email,
        userId: userInfo.userId,
      });

      // Redirect to dashboard
      navigate('/');
    } catch (err) {
      setError(handleApiError(err));
    } finally {
      setMfaVerifying(false);
    }
  };

  const onEnrollmentVerify = async () => {
    if (!mfaCode || !mfaSessionId) {
      setError('Please enter the verification code');
      return;
    }

    setMfaVerifying(true);
    setError(null);

    try {
      // Verify the enrollment code
      const verifyResponse = await mfaApi.verifyChallenge({
        challenge_id: mfaSessionId,
        code: mfaCode,
      });

      // After successful verification, MFA is enabled and we can complete login
      const { extractUserInfo } = await import('@/../../shared/utils/jwt-decoder');
      const userInfo = extractUserInfo(verifyResponse.access_token);

      const validPrincipalType: 'SYSTEM' | 'TENANT' = 
        (userInfo.principalType === 'SYSTEM' || userInfo.principalType === 'TENANT')
          ? userInfo.principalType
          : 'TENANT';

      setAuth({
        accessToken: verifyResponse.access_token,
        refreshToken: undefined,
        tenantId: userInfo.tenantId || tenantId || null,
        principalType: validPrincipalType,
        systemPermissions: userInfo.systemPermissions,
        permissions: userInfo.permissions,
        systemRoles: userInfo.systemRoles,
        username: userInfo.username,
        email: userInfo.email,
        userId: userInfo.userId,
      });

      navigate('/');
    } catch (err) {
      setError(handleApiError(err));
    } finally {
      setMfaVerifying(false);
    }
  };

  const resetMfa = () => {
    setMfaState('none');
    setMfaCode('');
    setMfaChallengeId(null);
    setMfaSessionId(null);
    setEnrollData(null);
    setEnrollmentStep('qr');
    setError(null);
    setUserId(null);
    setTenantId(null);
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-primary-50 via-white to-accent-50 p-4">
      {/* Background decoration */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute top-0 right-0 w-96 h-96 bg-primary-200 rounded-full mix-blend-multiply filter blur-3xl opacity-20 animate-pulse"></div>
        <div className="absolute bottom-0 left-0 w-96 h-96 bg-accent-200 rounded-full mix-blend-multiply filter blur-3xl opacity-20 animate-pulse delay-1000"></div>
      </div>

      <Card className="w-full max-w-md shadow-2xl border-0 bg-white/90 backdrop-blur-sm relative z-10">
        <CardHeader className="space-y-4 pb-6">
          <div className="flex justify-center mb-4">
            <ARauthLogo size="lg" />
          </div>
          <CardTitle className="text-2xl font-bold text-center text-gray-900">
            {mfaState === 'none' ? 'Welcome Back' : mfaState === 'enrollment' ? 'Enable Multi-Factor Authentication' : 'Verify Your Identity'}
          </CardTitle>
          <CardDescription className="text-center text-gray-600">
            {mfaState === 'none' 
              ? 'Sign in to your ARauth Identity account'
              : mfaState === 'enrollment'
              ? 'Your organization requires MFA. Let\'s set it up.'
              : 'Enter the code from your authenticator app'}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {error && (
            <Alert variant="destructive" className="border-red-200 bg-red-50">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription className="text-red-700">{error}</AlertDescription>
            </Alert>
          )}

          {mfaState === 'none' ? (
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
              <div className="space-y-2">
                <Label htmlFor="username" className="text-gray-700 font-medium flex items-center gap-2">
                  <KeyRound className="h-4 w-4 text-primary-600" />
                  Username
                </Label>
                <Input
                  id="username"
                  type="text"
                  {...register('username')}
                  placeholder="Enter your username"
                  disabled={isLoading}
                  className="h-11 border-gray-300 focus:border-primary-500 focus:ring-primary-500"
                />
                {errors.username && (
                  <p className="text-sm text-red-600 flex items-center gap-1">
                    <AlertCircle className="h-3 w-3" />
                    {errors.username.message}
                  </p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="password" className="text-gray-700 font-medium flex items-center gap-2">
                  <Lock className="h-4 w-4 text-primary-600" />
                  Password
                </Label>
                <Input
                  id="password"
                  type="password"
                  {...register('password')}
                  placeholder="Enter your password"
                  disabled={isLoading}
                  className="h-11 border-gray-300 focus:border-primary-500 focus:ring-primary-500"
                />
                {errors.password && (
                  <p className="text-sm text-red-600 flex items-center gap-1">
                    <AlertCircle className="h-3 w-3" />
                    {errors.password.message}
                  </p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="tenantId" className="text-gray-700 font-medium flex items-center gap-2">
                  <Shield className="h-4 w-4 text-primary-600" />
                  Tenant ID <span className="text-gray-400 font-normal text-xs">(Optional)</span>
                </Label>
                <Input
                  id="tenantId"
                  type="text"
                  {...register('tenantId')}
                  placeholder="Enter tenant ID"
                  disabled={isLoading}
                  className="h-11 border-gray-300 focus:border-primary-500 focus:ring-primary-500"
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
                  className="text-sm font-normal cursor-pointer text-gray-600"
                >
                  Remember me
                </Label>
              </div>

              <Button 
                type="submit" 
                className="w-full h-11 bg-gradient-to-r from-primary-600 to-primary-700 hover:from-primary-700 hover:to-primary-800 text-white font-semibold shadow-lg hover:shadow-xl transition-all duration-200" 
                disabled={isLoading}
              >
                {isLoading ? (
                  <span className="flex items-center gap-2">
                    <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                    Signing in...
                  </span>
                ) : (
                  'Sign In'
                )}
              </Button>
            </form>
          ) : mfaState === 'enrollment' ? (
            <div className="space-y-6">
              {enrollmentStep === 'qr' && enrollData ? (
                <>
                  <Alert className="bg-blue-50 border-blue-200">
                    <QrCode className="h-4 w-4 text-blue-600" />
                    <AlertDescription className="text-blue-800">
                      Scan this QR code with your authenticator app (Google Authenticator, Authy, etc.)
                    </AlertDescription>
                  </Alert>

                  <div className="flex flex-col items-center space-y-4">
                    <div className="bg-white p-4 rounded-lg border-2 border-gray-200 shadow-sm">
                      <img 
                        src={enrollData.qr_code} 
                        alt="MFA QR Code" 
                        className="w-48 h-48"
                      />
                    </div>
                    <div className="text-center space-y-2">
                      <p className="text-sm text-gray-600">Or enter this code manually:</p>
                      <code className="block px-4 py-2 bg-gray-100 rounded-md font-mono text-sm text-gray-800">
                        {enrollData.secret}
                      </code>
                    </div>
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="enrollmentCode" className="text-gray-700 font-medium">
                      Enter verification code
                    </Label>
                    <Input
                      id="enrollmentCode"
                      type="text"
                      placeholder="000000"
                      value={mfaCode}
                      onChange={(e) => setMfaCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
                      maxLength={6}
                      disabled={mfaVerifying}
                      className="h-11 font-mono text-lg tracking-widest text-center border-gray-300 focus:border-primary-500 focus:ring-primary-500"
                    />
                  </div>

                  <div className="flex space-x-3">
                    <Button
                      variant="outline"
                      className="flex-1 border-gray-300 hover:bg-gray-50"
                      onClick={resetMfa}
                      disabled={mfaVerifying}
                    >
                      Cancel
                    </Button>
                    <Button
                      className="flex-1 bg-gradient-to-r from-primary-600 to-primary-700 hover:from-primary-700 hover:to-primary-800 text-white font-semibold shadow-lg"
                      onClick={onEnrollmentVerify}
                      disabled={mfaCode.length !== 6 || mfaVerifying}
                    >
                      {mfaVerifying ? (
                        <span className="flex items-center gap-2">
                          <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                          Verifying...
                        </span>
                      ) : (
                        'Verify & Continue'
                      )}
                    </Button>
                  </div>
                </>
              ) : null}
            </div>
          ) : (
            <div className="space-y-4">
              <Alert className="bg-primary-50 border-primary-200">
                <Shield className="h-4 w-4 text-primary-600" />
                <AlertDescription className="text-primary-800">
                  Multi-factor authentication is required. Please enter the verification code from your authenticator app.
                </AlertDescription>
              </Alert>

              <div className="space-y-2">
                <Label htmlFor="mfaCode" className="text-gray-700 font-medium">
                  Verification Code
                </Label>
                <Input
                  id="mfaCode"
                  type="text"
                  placeholder="000000"
                  value={mfaCode}
                  onChange={(e) => setMfaCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
                  maxLength={6}
                  disabled={mfaVerifying}
                  className="h-11 font-mono text-lg tracking-widest text-center border-gray-300 focus:border-primary-500 focus:ring-primary-500"
                />
              </div>

              <div className="flex space-x-3">
                <Button
                  variant="outline"
                  className="flex-1 border-gray-300 hover:bg-gray-50"
                  onClick={resetMfa}
                  disabled={mfaVerifying}
                >
                  Back
                </Button>
                <Button
                  className="flex-1 bg-gradient-to-r from-primary-600 to-primary-700 hover:from-primary-700 hover:to-primary-800 text-white font-semibold shadow-lg"
                  onClick={onMfaVerify}
                  disabled={mfaCode.length !== 6 || mfaVerifying}
                >
                  {mfaVerifying ? (
                    <span className="flex items-center gap-2">
                      <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                      Verifying...
                    </span>
                  ) : (
                    'Verify'
                  )}
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
