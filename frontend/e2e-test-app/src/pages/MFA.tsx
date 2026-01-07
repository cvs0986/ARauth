/**
 * MFA Management Page for E2E Testing App
 */

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { mfaApi } from '../services/mfaApi';
import { handleApiError } from '../services/mfaApi';
import { useAuthStore } from '../store/authStore';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert } from '@/components/ui/alert';

export function MFA() {
  const { setTokens } = useAuthStore();
  const queryClient = useQueryClient();
  const [enrollStep, setEnrollStep] = useState<'start' | 'qr' | 'verify'>('start');
  const [verifyCode, setVerifyCode] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [enrollData, setEnrollData] = useState<{
    secret: string;
    qr_code: string;
    recovery_codes: string[];
  } | null>(null);

  const enrollMutation = useMutation({
    mutationFn: () => mfaApi.enroll(),
    onSuccess: (data) => {
      setEnrollData(data);
      setEnrollStep('qr');
      setError(null);
    },
    onError: (err) => {
      setError(handleApiError(err));
    },
  });

  const verifyMutation = useMutation({
    mutationFn: (code: string) => mfaApi.verify({ code }),
    onSuccess: () => {
      setEnrollStep('start');
      setEnrollData(null);
      setVerifyCode('');
      setError(null);
      queryClient.invalidateQueries({ queryKey: ['mfa'] });
    },
    onError: (err) => {
      setError(handleApiError(err));
    },
  });

  const handleEnroll = () => {
    enrollMutation.mutate();
  };

  const handleVerify = () => {
    if (!verifyCode) {
      setError('Please enter a verification code');
      return;
    }
    verifyMutation.mutate(verifyCode);
  };

  const copySecret = () => {
    if (enrollData?.secret) {
      navigator.clipboard.writeText(enrollData.secret);
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
    }
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b">
        <div className="container mx-auto px-4 py-4">
          <h1 className="text-xl font-bold">MFA Management</h1>
        </div>
      </header>

      <main className="container mx-auto px-4 py-8 max-w-2xl">
        <Card>
          <CardHeader>
            <CardTitle>Multi-Factor Authentication</CardTitle>
            <CardDescription>
              Secure your account with two-factor authentication
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            {error && (
              <Alert variant="destructive">
                {error}
              </Alert>
            )}

            {enrollStep === 'start' && (
              <div className="space-y-4">
                <p className="text-gray-600">
                  Enable MFA to add an extra layer of security to your account. You'll need an
                  authenticator app like Google Authenticator or Authy.
                </p>
                <Button onClick={handleEnroll} disabled={enrollMutation.isPending}>
                  {enrollMutation.isPending ? 'Enrolling...' : 'Enable MFA'}
                </Button>
              </div>
            )}

            {enrollStep === 'qr' && enrollData && (
              <div className="space-y-4">
                <div>
                  <h3 className="font-semibold mb-2">Step 1: Scan QR Code</h3>
                  <p className="text-sm text-gray-600 mb-4">
                    Scan this QR code with your authenticator app:
                  </p>
                  <div className="flex justify-center mb-4">
                    <img
                      src={enrollData.qr_code}
                      alt="MFA QR Code"
                      className="border rounded-lg"
                    />
                  </div>
                </div>

                <div>
                  <h3 className="font-semibold mb-2">Or Enter Secret Manually</h3>
                  <div className="flex items-center gap-2">
                    <Input value={enrollData.secret} readOnly className="font-mono" />
                    <Button variant="outline" size="sm" onClick={copySecret}>
                      Copy
                    </Button>
                  </div>
                </div>

                <div>
                  <h3 className="font-semibold mb-2">Recovery Codes</h3>
                  <p className="text-sm text-gray-600 mb-2">
                    Save these recovery codes in a safe place. You can use them if you lose
                    access to your authenticator app.
                  </p>
                  <div className="bg-gray-50 p-4 rounded-lg mb-2">
                    <ul className="list-disc list-inside space-y-1 font-mono text-sm">
                      {enrollData.recovery_codes.map((code, index) => (
                        <li key={index}>{code}</li>
                      ))}
                    </ul>
                  </div>
                  <Button variant="outline" size="sm" onClick={downloadRecoveryCodes}>
                    Download Recovery Codes
                  </Button>
                </div>

                <div>
                  <h3 className="font-semibold mb-2">Step 2: Verify Setup</h3>
                  <p className="text-sm text-gray-600 mb-4">
                    Enter the 6-digit code from your authenticator app to complete setup:
                  </p>
                  <div className="flex gap-2">
                    <Input
                      type="text"
                      placeholder="000000"
                      value={verifyCode}
                      onChange={(e) => setVerifyCode(e.target.value)}
                      maxLength={6}
                      className="max-w-xs"
                    />
                    <Button onClick={handleVerify} disabled={verifyMutation.isPending}>
                      {verifyMutation.isPending ? 'Verifying...' : 'Verify'}
                    </Button>
                  </div>
                </div>

                <Button
                  variant="outline"
                  onClick={() => {
                    setEnrollStep('start');
                    setEnrollData(null);
                  }}
                >
                  Cancel
                </Button>
              </div>
            )}
          </CardContent>
        </Card>
      </main>
    </div>
  );
}

