/**
 * Create User Dialog
 * 
 * AUTHORITY MODEL:
 * Who: SYSTEM users OR TENANT users
 * Scope: Platform-wide (SYSTEM) OR Tenant-scoped (TENANT)
 * Permission: users:create
 * 
 * SYSTEM Users:
 * - Must select tenant (required)
 * - Can create system users OR tenant users
 * - Roles filtered by tenant scope
 * 
 * TENANT Users:
 * - No tenant selector
 * - Roles limited to tenant roles only
 * - Cannot create system users
 */

import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useMutation, useQueryClient, useQuery } from '@tanstack/react-query';
import { userApi, roleApi, systemApi } from '@/services/api';
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
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { useState } from 'react';
import { usePrincipalContext } from '@/contexts/PrincipalContext';
import { Building2, Shield, Info } from 'lucide-react';

const createUserSchema = z.object({
  username: z.string().min(1, 'Username is required'),
  email: z.string().email('Invalid email address'),
  password: z.string().min(12, 'Password must be at least 12 characters'),
  first_name: z.string().optional(),
  last_name: z.string().optional(),
  role_id: z.string().optional(),
  tenant_id: z.string().optional(), // For SYSTEM users
});

type CreateUserFormData = z.infer<typeof createUserSchema>;

interface CreateUserDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  tenantId?: string; // Pre-selected tenant ID
}

export function CreateUserDialog({ open, onOpenChange, tenantId: propTenantId }: CreateUserDialogProps) {
  const queryClient = useQueryClient();
  const { principalType, homeTenantId, selectedTenantId } = usePrincipalContext();
  const [error, setError] = useState<string | null>(null);
  const [selectedRoleId, setSelectedRoleId] = useState<string>('');
  const [selectedTenantId, setSelectedTenantId] = useState<string>(propTenantId || '');

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
    setValue,
  } = useForm<CreateUserFormData>({
    resolver: zodResolver(createUserSchema),
  });

  // Determine effective tenant ID
  const effectiveTenantId = propTenantId || selectedTenantId ||
    (principalType === 'TENANT' ? homeTenantId : null);

  // Fetch tenants for SYSTEM users
  const { data: tenants } = useQuery({
    queryKey: ['system', 'tenants'],
    queryFn: () => systemApi.tenants.list(),
    enabled: open && principalType === 'SYSTEM',
  });

  // Fetch roles based on selected tenant
  const { data: roles = [], isLoading: rolesLoading } = useQuery({
    queryKey: effectiveTenantId ? ['roles', effectiveTenantId] : ['systemRoles'],
    queryFn: () => (effectiveTenantId ? roleApi.list(effectiveTenantId) : roleApi.listSystem()),
    enabled: open,
  });

  const createUserMutation = useMutation({
    mutationFn: async (data: CreateUserFormData) => {
      const { role_id, tenant_id, ...userData } = data;

      // Determine target tenant
      const targetTenantId = tenant_id || effectiveTenantId;

      // Create user
      const user = targetTenantId
        ? await userApi.create({ ...userData, tenant_id: targetTenantId })
        : await userApi.createSystem(userData);

      // Assign role if selected
      if (role_id && user.id) {
        try {
          await roleApi.assignRoleToUser(user.id, role_id, targetTenantId || undefined);
        } catch (err) {
          console.error('Failed to assign role:', err);
        }
      }

      return user;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      queryClient.invalidateQueries({ queryKey: ['system', 'users'] });
      reset();
      setSelectedRoleId('');
      setSelectedTenantId('');
      setError(null);
      onOpenChange(false);
    },
    onError: (err) => {
      setError(handleApiError(err));
    },
  });

  const onSubmit = (data: CreateUserFormData) => {
    setError(null);

    // Validate tenant selection for SYSTEM users
    if (principalType === 'SYSTEM' && !effectiveTenantId && !data.tenant_id) {
      setError('Please select a tenant');
      return;
    }

    createUserMutation.mutate({
      ...data,
      tenant_id: effectiveTenantId || undefined,
    });
  };

  const isSystemUser = principalType === 'SYSTEM';
  const showTenantSelector = isSystemUser && !propTenantId;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Create User</DialogTitle>
          <DialogDescription>
            {isSystemUser && effectiveTenantId
              ? `Create a new user for the selected tenant`
              : isSystemUser
                ? 'Create a new system user or tenant user'
                : 'Create a new user account'}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {/* Authority Context Badge */}
          {effectiveTenantId ? (
            <Badge className="bg-green-100 text-green-800">
              <Building2 className="h-3 w-3 mr-1" />
              Creating Tenant User
            </Badge>
          ) : (
            <Badge className="bg-blue-100 text-blue-800">
              <Shield className="h-3 w-3 mr-1" />
              Creating System User
            </Badge>
          )}

          {/* Tenant Selector (SYSTEM users only) */}
          {showTenantSelector && (
            <div className="space-y-2">
              <Label htmlFor="tenant_id">Tenant *</Label>
              <Select
                value={selectedTenantId}
                onValueChange={(value) => {
                  setSelectedTenantId(value);
                  setValue('tenant_id', value);
                  setSelectedRoleId(''); // Reset role when tenant changes
                }}
                disabled={createUserMutation.isPending}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select tenant (or leave empty for system user)" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="">System User (No Tenant)</SelectItem>
                  {tenants?.map((tenant) => (
                    <SelectItem key={tenant.id} value={tenant.id}>
                      {tenant.name} ({tenant.domain})
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <p className="text-xs text-gray-500">
                Select a tenant to create a tenant user, or leave empty to create a system user
              </p>
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="username">Username *</Label>
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
            <Label htmlFor="email">Email *</Label>
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
            <Label htmlFor="password">Password *</Label>
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
              {effectiveTenantId
                ? 'Roles are filtered by selected tenant'
                : 'System roles only'}
            </p>
          </div>

          {/* Authority Notice */}
          <Alert className="bg-blue-50 border-blue-200">
            <Info className="h-4 w-4 text-blue-600" />
            <AlertDescription className="text-blue-800 text-xs">
              {effectiveTenantId
                ? 'This user will be created in the selected tenant scope'
                : 'This user will be created as a system user with platform-wide access'}
            </AlertDescription>
          </Alert>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => {
                onOpenChange(false);
                reset();
                setSelectedRoleId('');
                setSelectedTenantId('');
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
