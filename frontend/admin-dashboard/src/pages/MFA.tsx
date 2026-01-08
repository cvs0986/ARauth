/**
 * MFA Management Page for Admin Dashboard
 */

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { mfaApi, userApi } from '@/services/api';
import { useAuthStore } from '@/store/authStore';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert } from '@/components/ui/alert';
import { Shield, CheckCircle, AlertCircle } from 'lucide-react';
import { extractUserInfo } from '@/../../shared/utils/jwt-decoder';

export function MFA() {
  const { tenantId, selectedTenantId, isSystemUser } = useAuthStore();
  const queryClient = useQueryClient();
  const [enrollStep, setEnrollStep] = useState<'start' | 'qr' | 'verify'>('start');
  const [verifyCode, setVerifyCode] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [enrollData, setEnrollData] = useState<{
    secret: string;
    qr_code: string;
    recovery_codes: string[];
  } | null>(null);

  // Get current user ID from JWT token
  const getCurrentUserId = (): string | null => {
    const token = localStorage.getItem('access_token');
    if (!token) return null;
    try {
      const userInfo = extractUserInfo(token);
      return userInfo.userId || null;
    } catch {
      return null;
    }
  };

  const currentUserId = getCurrentUserId();
  const currentTenantId = isSystemUser() ? selectedTenantId : tenantId;

  // Fetch current user to check MFA status
  const { data: currentUser, isLoading: userLoading, refetch: refetchUser } = useQuery({
    queryKey: ['user', currentUserId, currentTenantId],
    queryFn: async () => {
      if (!currentUserId) return null;
      return userApi.getById(currentUserId);
    },
    enabled: !!currentUserId,
  });

  const isMFAEnabled = currentUser?.mfa_enabled || false;

  const enrollMutation = useMutation({
    mutationFn: () => mfaApi.enroll(),
    onSuccess: (data) => {
      setEnrollData(data);
      setEnrollStep('qr');
      setError(null);
    },
    onError: (err: any) => {
      setError(err.response?.data?.message || err.message || 'Failed to enroll in MFA');
    },
  });

  const verifyMutation = useMutation({
    mutationFn: (code: string) => mfaApi.verify({ code }),
    onSuccess: () => {
      setEnrollStep('start');
      setEnrollData(null);
      setVerifyCode('');
      setError(null);
      setSuccess('MFA enabled successfully!');
      setTimeout(() => setSuccess(null), 5000);
      // Refetch user to update MFA status
      refetchUser();
      queryClient.invalidateQueries({ queryKey: ['mfa'] });
      queryClient.invalidateQueries({ queryKey: ['user', currentUserId] });
    },
    onError: (err: any) => {
      setError(err.response?.data?.message || err.message || 'Invalid verification code');
    },
  });

  const handleEnroll = () => {
    setError(null);
    enrollMutation.mutate();
  };

  const handleVerify = () => {
    if (!verifyCode) {
      setError('Please enter a verification code');
      return;
    }
    setError(null);
    verifyMutation.mutate(verifyCode);
  };

  const copySecret = () => {
    if (enrollData?.secret) {
      navigator.clipboard.writeText(enrollData.secret);
      setSuccess('Secret copied to clipboard!');
      setTimeout(() => setSuccess(null), 2000);
    }
  };

  const downloadRecoveryCodes = () => {
    if (enrollData?.recovery_codes) {
      const content = enrollData.recovery_codes.join('\n');
      const blob = new Blob([content], { type: 'text/plain' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = 'mfa-recovery-codes.txt';
      a.click();
      URL.revokeObjectURL(url);
      setSuccess('Recovery codes downloaded!');
      setTimeout(() => setSuccess(null), 2000);
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Multi-Factor Authentication</h1>
        <p className="text-gray-600 mt-1">
          Secure your account with two-factor authentication using TOTP
        </p>
      </div>

      {error && (
        <Alert className="bg-red-50 border-red-200 text-red-700">
          <AlertCircle className="h-4 w-4 mr-2" />
          {error}
        </Alert>
      )}

      {success && (
        <Alert className="bg-green-50 border-green-200 text-green-700">
          <CheckCircle className="h-4 w-4 mr-2" />
          {success}
        </Alert>
      )}

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center">
            <Shield className="mr-2 h-5 w-5" />
            MFA Setup
          </CardTitle>
          <CardDescription>
            Enable MFA to add an extra layer of security to your account. You'll need an
            authenticator app like Google Authenticator or Authy.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {userLoading ? (
            <div className="text-center py-4">Loading MFA status...</div>
          ) : isMFAEnabled ? (
            <div className="space-y-4">
              <Alert className="bg-green-50 border-green-200 text-green-700">
                <CheckCircle className="h-4 w-4 mr-2" />
                <div>
                  <p className="font-semibold">MFA is Enabled</p>
                  <p className="text-sm mt-1">
                    Your account is protected with multi-factor authentication. You'll be required
                    to enter a verification code from your authenticator app when logging in.
                  </p>
                </div>
              </Alert>
              <div className="pt-4 border-t">
                <p className="text-sm text-gray-600 mb-4">
                  If you need to disable MFA or set up a new authenticator device, please contact your administrator.
                </p>
              </div>
            </div>
          ) : enrollStep === 'start' ? (
            <div className="space-y-4">
              <p className="text-gray-600">
                Multi-factor authentication (MFA) adds an extra layer of security by requiring
                a verification code from your authenticator app in addition to your password.
              </p>
              <Button onClick={handleEnroll} disabled={enrollMutation.isPending}>
                {enrollMutation.isPending ? 'Enrolling...' : 'Enable MFA'}
              </Button>
            </div>
          ) : null}

          {enrollStep === 'qr' && enrollData && (
            <div className="space-y-6">
              <Alert>
                <AlertCircle className="h-4 w-4 mr-2" />
                Scan the QR code with your authenticator app, then enter the verification code below.
              </Alert>

              <div className="space-y-4">
                <div>
                  <Label>QR Code</Label>
                  <div className="mt-2 p-4 bg-white border rounded-lg inline-block">
                    <img
                      src={enrollData.qr_code}
                      alt="MFA QR Code"
                      className="w-64 h-64"
                    />
                  </div>
                </div>

                <div>
                  <Label>Manual Entry Secret</Label>
                  <div className="mt-2 flex items-center space-x-2">
                    <Input
                      value={enrollData.secret}
                      readOnly
                      className="font-mono"
                    />
                    <Button variant="outline" onClick={copySecret}>
                      Copy
                    </Button>
                  </div>
                  <p className="text-sm text-gray-500 mt-1">
                    If you can't scan the QR code, enter this secret manually in your authenticator app.
                  </p>
                </div>

                <div>
                  <Label htmlFor="verifyCode">Verification Code</Label>
                  <div className="mt-2 flex items-center space-x-2">
                    <Input
                      id="verifyCode"
                      type="text"
                      placeholder="Enter 6-digit code"
                      value={verifyCode}
                      onChange={(e) => setVerifyCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
                      maxLength={6}
                      className="font-mono text-lg tracking-widest"
                    />
                    <Button
                      onClick={handleVerify}
                      disabled={verifyCode.length !== 6 || verifyMutation.isPending}
                    >
                      {verifyMutation.isPending ? 'Verifying...' : 'Verify'}
                    </Button>
                  </div>
                </div>

                {enrollData.recovery_codes && enrollData.recovery_codes.length > 0 && (
                  <div>
                    <Label>Recovery Codes</Label>
                    <Alert className="mt-2 bg-yellow-50 border-yellow-200 text-yellow-800">
                      <AlertCircle className="h-4 w-4 mr-2" />
                      <div>
                        <p className="font-semibold mb-2">Save these recovery codes!</p>
                        <p className="text-sm mb-2">
                          If you lose access to your authenticator app, you can use these codes to sign in.
                          Each code can only be used once.
                        </p>
                        <div className="grid grid-cols-2 gap-2 font-mono text-sm bg-white p-2 rounded border">
                          {enrollData.recovery_codes.map((code, index) => (
                            <div key={index}>{code}</div>
                          ))}
                        </div>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={downloadRecoveryCodes}
                          className="mt-2"
                        >
                          Download Recovery Codes
                        </Button>
                      </div>
                    </Alert>
                  </div>
                )}
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

