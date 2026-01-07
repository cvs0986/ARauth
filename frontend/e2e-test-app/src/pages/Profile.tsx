/**
 * Profile Page for E2E Testing App
 */

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { userApi } from '../services/api';
import { handleApiError } from '../services/api';
import { useAuthStore } from '../store/authStore';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert } from '@/components/ui/alert';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';

const profileSchema = z.object({
  email: z.string().email('Invalid email address'),
  first_name: z.string().optional(),
  last_name: z.string().optional(),
});

const passwordSchema = z.object({
  current_password: z.string().min(1, 'Current password is required'),
  new_password: z.string().min(12, 'Password must be at least 12 characters'),
  confirm_password: z.string().min(12, 'Password must be at least 12 characters'),
}).refine((data) => data.new_password === data.confirm_password, {
  message: "Passwords don't match",
  path: ['confirm_password'],
});

type ProfileFormData = z.infer<typeof profileSchema>;
type PasswordFormData = z.infer<typeof passwordSchema>;

export function Profile() {
  const { tenantId } = useAuthStore();
  const queryClient = useQueryClient();
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'profile' | 'password'>('profile');

  // Get current user (would need user ID from auth token or separate endpoint)
  // For now, we'll use a placeholder
  const userId = 'current-user-id'; // This should come from auth token

  const {
    register: registerProfile,
    handleSubmit: handleSubmitProfile,
    formState: { errors: profileErrors },
    reset: resetProfile,
  } = useForm<ProfileFormData>({
    resolver: zodResolver(profileSchema),
  });

  const {
    register: registerPassword,
    handleSubmit: handleSubmitPassword,
    formState: { errors: passwordErrors },
    reset: resetPassword,
  } = useForm<PasswordFormData>({
    resolver: zodResolver(passwordSchema),
  });

  const updateProfileMutation = useMutation({
    mutationFn: (data: ProfileFormData) => userApi.update(userId, data),
    onSuccess: () => {
      setSuccess('Profile updated successfully');
      setError(null);
      queryClient.invalidateQueries({ queryKey: ['user', userId] });
      setTimeout(() => setSuccess(null), 3000);
    },
    onError: (err) => {
      setError(handleApiError(err));
      setSuccess(null);
    },
  });

  const changePasswordMutation = useMutation({
    mutationFn: async (data: PasswordFormData) => {
      // Note: This would need a separate password change endpoint
      // For now, this is a placeholder
      throw new Error('Password change endpoint not yet implemented');
    },
    onSuccess: () => {
      setSuccess('Password changed successfully');
      setError(null);
      resetPassword();
      setTimeout(() => setSuccess(null), 3000);
    },
    onError: (err) => {
      setError(handleApiError(err));
      setSuccess(null);
    },
  });

  const onProfileSubmit = (data: ProfileFormData) => {
    setError(null);
    updateProfileMutation.mutate(data);
  };

  const onPasswordSubmit = (data: PasswordFormData) => {
    setError(null);
    changePasswordMutation.mutate(data);
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b">
        <div className="container mx-auto px-4 py-4">
          <h1 className="text-xl font-bold">Profile</h1>
        </div>
      </header>

      <main className="container mx-auto px-4 py-8 max-w-4xl">
        <div className="flex gap-4 mb-6">
          <Button
            variant={activeTab === 'profile' ? 'default' : 'outline'}
            onClick={() => setActiveTab('profile')}
          >
            Profile Information
          </Button>
          <Button
            variant={activeTab === 'password' ? 'default' : 'outline'}
            onClick={() => setActiveTab('password')}
          >
            Change Password
          </Button>
        </div>

        {error && (
          <Alert variant="destructive" className="mb-4">
            {error}
          </Alert>
        )}

        {success && (
          <Alert className="mb-4 bg-green-50 border-green-200 text-green-700">
            {success}
          </Alert>
        )}

        {activeTab === 'profile' && (
          <Card>
            <CardHeader>
              <CardTitle>Profile Information</CardTitle>
              <CardDescription>Update your profile details</CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleSubmitProfile(onProfileSubmit)} className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="email">Email</Label>
                  <Input
                    id="email"
                    type="email"
                    {...registerProfile('email')}
                    disabled={updateProfileMutation.isPending}
                  />
                  {profileErrors.email && (
                    <p className="text-sm text-red-600">{profileErrors.email.message}</p>
                  )}
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label htmlFor="first_name">First Name</Label>
                    <Input
                      id="first_name"
                      {...registerProfile('first_name')}
                      disabled={updateProfileMutation.isPending}
                    />
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="last_name">Last Name</Label>
                    <Input
                      id="last_name"
                      {...registerProfile('last_name')}
                      disabled={updateProfileMutation.isPending}
                    />
                  </div>
                </div>

                {tenantId && (
                  <div className="space-y-2">
                    <Label>Tenant ID</Label>
                    <Input value={tenantId} disabled />
                  </div>
                )}

                <Button type="submit" disabled={updateProfileMutation.isPending}>
                  {updateProfileMutation.isPending ? 'Updating...' : 'Update Profile'}
                </Button>
              </form>
            </CardContent>
          </Card>
        )}

        {activeTab === 'password' && (
          <Card>
            <CardHeader>
              <CardTitle>Change Password</CardTitle>
              <CardDescription>Update your account password</CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleSubmitPassword(onPasswordSubmit)} className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="current_password">Current Password</Label>
                  <Input
                    id="current_password"
                    type="password"
                    {...registerPassword('current_password')}
                    disabled={changePasswordMutation.isPending}
                  />
                  {passwordErrors.current_password && (
                    <p className="text-sm text-red-600">
                      {passwordErrors.current_password.message}
                    </p>
                  )}
                </div>

                <div className="space-y-2">
                  <Label htmlFor="new_password">New Password</Label>
                  <Input
                    id="new_password"
                    type="password"
                    {...registerPassword('new_password')}
                    placeholder="Minimum 12 characters"
                    disabled={changePasswordMutation.isPending}
                  />
                  {passwordErrors.new_password && (
                    <p className="text-sm text-red-600">{passwordErrors.new_password.message}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <Label htmlFor="confirm_password">Confirm New Password</Label>
                  <Input
                    id="confirm_password"
                    type="password"
                    {...registerPassword('confirm_password')}
                    placeholder="Confirm your new password"
                    disabled={changePasswordMutation.isPending}
                  />
                  {passwordErrors.confirm_password && (
                    <p className="text-sm text-red-600">
                      {passwordErrors.confirm_password.message}
                    </p>
                  )}
                </div>

                <Button type="submit" disabled={changePasswordMutation.isPending}>
                  {changePasswordMutation.isPending ? 'Changing...' : 'Change Password'}
                </Button>
              </form>
            </CardContent>
          </Card>
        )}
      </main>
    </div>
  );
}

