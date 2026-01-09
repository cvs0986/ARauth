/**
 * Create User Dialog
 */

import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useMutation, useQueryClient, useQuery } from '@tanstack/react-query';
import { userApi, roleApi } from '@/services/api';
import { handleApiError } from '@/services/api';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { useState } from 'react';
import { useAuthStore } from '@/store/authStore';

const createUserSchema = z.object({
  username: z.string().min(1, 'Username is required'),
  email: z.string().email('Invalid email address'),
  password: z.string().min(12, 'Password must be at least 12 characters'),
  first_name: z.string().optional(),
  last_name: z.string().optional(),
  role_id: z.string().optional(),
});

type CreateUserFormData = z.infer<typeof createUserSchema>;

interface CreateUserDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  tenantId?: string; // Optional tenant ID for SYSTEM users creating tenant users
}

export function CreateUserDialog({ open, onOpenChange, tenantId }: CreateUserDialogProps) {
  const queryClient = useQueryClient();
  const { isSystemUser } = useAuthStore();
  const [error, setError] = useState<string | null>(null);
  const [selectedRoleId, setSelectedRoleId] = useState<string>('');

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
    setValue,
  } = useForm<CreateUserFormData>({
    resolver: zodResolver(createUserSchema),
  });

  // Fetch roles based on context (system or tenant)
  const { data: roles = [], isLoading: rolesLoading } = useQuery({
    queryKey: tenantId ? ['roles', tenantId] : ['systemRoles'],
    queryFn: () => (tenantId ? roleApi.list(tenantId) : roleApi.listSystem()),
    enabled: open, // Only fetch when dialog is open
  });

  const createUserMutation = useMutation({
    mutationFn: async (data: CreateUserFormData) => {
      const { role_id, ...userData } = data;
      
      // Use createSystem for system users, create for tenant users
      const user = tenantId
        ? await userApi.create({ ...userData, tenant_id: tenantId })
        : await userApi.createSystem(userData);
      
      // Assign role if selected
      if (role_id && user.id) {
        try {
          // For system users, don't pass tenantId (system roles don't need tenant context)
          // For tenant users, pass tenantId (tenant roles need tenant context)
          await roleApi.assignRoleToUser(user.id, role_id, tenantId || undefined);
        } catch (err) {
          // If role assignment fails, we still created the user
          // Log error but don't fail the whole operation
          console.error('Failed to assign role:', err);
        }
      }
      
      return user;
    },
    onSuccess: () => {
      // Invalidate both system and tenant user queries
      queryClient.invalidateQueries({ queryKey: ['users'] });
      queryClient.invalidateQueries({ queryKey: ['system', 'users'] });
      queryClient.invalidateQueries({ queryKey: tenantId ? ['roles', tenantId] : ['systemRoles'] });
      reset();
      setSelectedRoleId('');
      setError(null);
      onOpenChange(false);
    },
    onError: (err) => {
      setError(handleApiError(err));
    },
  });

  const onSubmit = (data: CreateUserFormData) => {
    setError(null);
    createUserMutation.mutate(data);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Create User</DialogTitle>
          <DialogDescription>
            {isSystemUser() && tenantId
              ? `Create a new user account for the selected tenant. The user will be able to log in with these credentials.`
              : 'Create a new user account. The user will be able to log in with these credentials.'}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          {error && (
            <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
              {error}
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="username">Username</Label>
            <Input
              id="username"
              {...register('username')}
              placeholder="johndoe"
              disabled={createUserMutation.isPending}
            />
            {errors.username && (
              <p className="text-sm text-red-600">{errors.username.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="email">Email</Label>
            <Input
              id="email"
              type="email"
              {...register('email')}
              placeholder="john@example.com"
              disabled={createUserMutation.isPending}
            />
            {errors.email && (
              <p className="text-sm text-red-600">{errors.email.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="password">Password</Label>
            <Input
              id="password"
              type="password"
              {...register('password')}
              placeholder="Minimum 12 characters"
              disabled={createUserMutation.isPending}
            />
            {errors.password && (
              <p className="text-sm text-red-600">{errors.password.message}</p>
            )}
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="first_name">First Name</Label>
              <Input
                id="first_name"
                {...register('first_name')}
                placeholder="John"
                disabled={createUserMutation.isPending}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="last_name">Last Name</Label>
              <Input
                id="last_name"
                {...register('last_name')}
                placeholder="Doe"
                disabled={createUserMutation.isPending}
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="role_id">Role (Optional)</Label>
            <Select
              value={selectedRoleId || undefined}
              onValueChange={(value) => {
                setSelectedRoleId(value);
                setValue('role_id', value);
              }}
              disabled={createUserMutation.isPending || rolesLoading}
            >
              <SelectTrigger>
                <SelectValue placeholder={rolesLoading ? "Loading roles..." : "Select a role (optional)"} />
              </SelectTrigger>
              <SelectContent>
                {roles.length === 0 && !rolesLoading ? (
                  <div className="px-2 py-1.5 text-sm text-muted-foreground">
                    No roles available
                  </div>
                ) : (
                  roles.map((role) => (
                    <SelectItem key={role.id} value={role.id}>
                      {role.name}
                    </SelectItem>
                  ))
                )}
              </SelectContent>
            </Select>
            <p className="text-xs text-muted-foreground">
              {tenantId ? 'Select a tenant role to assign to this user' : 'Select a system role to assign to this user'}
            </p>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => {
                onOpenChange(false);
                reset();
                setSelectedRoleId('');
              }}
              disabled={createUserMutation.isPending}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={createUserMutation.isPending}>
              {createUserMutation.isPending ? 'Creating...' : 'Create User'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

