/**
 * Create Tenant Dialog
 */

import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { tenantApi } from '@/services/api';
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
import { useState } from 'react';

const createTenantSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  domain: z.string().min(1, 'Domain is required').regex(/^[a-z0-9.-]+$/, 'Invalid domain format'),
  status: z.enum(['active', 'inactive']).default('active'),
});

type CreateTenantFormData = z.infer<typeof createTenantSchema>;

interface CreateTenantDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function CreateTenantDialog({ open, onOpenChange }: CreateTenantDialogProps) {
  const queryClient = useQueryClient();
  const [error, setError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
  } = useForm<CreateTenantFormData>({
    resolver: zodResolver(createTenantSchema),
    defaultValues: {
      status: 'active',
    },
  });

  const mutation = useMutation({
    mutationFn: tenantApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenants'] });
      reset();
      setError(null);
      onOpenChange(false);
    },
    onError: (err) => {
      setError(handleApiError(err));
    },
  });

  const onSubmit = (data: CreateTenantFormData) => {
    setError(null);
    mutation.mutate(data);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create Tenant</DialogTitle>
          <DialogDescription>
            Create a new tenant. Each tenant is isolated and has its own users, roles, and permissions.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          {error && (
            <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
              {error}
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="name">Name</Label>
            <Input
              id="name"
              {...register('name')}
              placeholder="Acme Corporation"
              disabled={mutation.isPending}
            />
            {errors.name && (
              <p className="text-sm text-red-600">{errors.name.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="domain">Domain</Label>
            <Input
              id="domain"
              {...register('domain')}
              placeholder="acme.com"
              disabled={mutation.isPending}
            />
            {errors.domain && (
              <p className="text-sm text-red-600">{errors.domain.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="status">Status</Label>
            <select
              id="status"
              {...register('status')}
              className="w-full px-3 py-2 border border-gray-300 rounded-md"
              disabled={mutation.isPending}
            >
              <option value="active">Active</option>
              <option value="inactive">Inactive</option>
            </select>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={mutation.isPending}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={mutation.isPending}>
              {mutation.isPending ? 'Creating...' : 'Create Tenant'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

