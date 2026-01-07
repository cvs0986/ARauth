/**
 * Edit Permission Dialog
 */

import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { permissionApi } from '@/services/api';
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
import { useState, useEffect } from 'react';
import type { Permission } from '@shared/types/api';

const editPermissionSchema = z.object({
  resource: z.string().min(1, 'Resource is required'),
  action: z.string().min(1, 'Action is required'),
  description: z.string().optional(),
});

type EditPermissionFormData = z.infer<typeof editPermissionSchema>;

interface EditPermissionDialogProps {
  permission: Permission;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function EditPermissionDialog({
  permission,
  open,
  onOpenChange,
}: EditPermissionDialogProps) {
  const queryClient = useQueryClient();
  const [error, setError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
  } = useForm<EditPermissionFormData>({
    resolver: zodResolver(editPermissionSchema),
    defaultValues: {
      resource: permission.resource,
      action: permission.action,
      description: permission.description,
    },
  });

  useEffect(() => {
    if (permission) {
      reset({
        resource: permission.resource,
        action: permission.action,
        description: permission.description,
      });
    }
  }, [permission, reset]);

  const mutation = useMutation({
    mutationFn: (data: EditPermissionFormData) =>
      permissionApi.update(permission.id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['permissions'] });
      setError(null);
      onOpenChange(false);
    },
    onError: (err) => {
      setError(handleApiError(err));
    },
  });

  const onSubmit = (data: EditPermissionFormData) => {
    setError(null);
    mutation.mutate(data);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit Permission</DialogTitle>
          <DialogDescription>
            Update permission information. Changes will affect all roles using this permission.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          {error && (
            <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
              {error}
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="resource">Resource</Label>
            <Input
              id="resource"
              {...register('resource')}
              placeholder="users"
              disabled={mutation.isPending}
            />
            {errors.resource && (
              <p className="text-sm text-red-600">{errors.resource.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="action">Action</Label>
            <Input
              id="action"
              {...register('action')}
              placeholder="read, write, delete"
              disabled={mutation.isPending}
            />
            {errors.action && (
              <p className="text-sm text-red-600">{errors.action.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="description">Description</Label>
            <Input
              id="description"
              {...register('description')}
              placeholder="Read user information"
              disabled={mutation.isPending}
            />
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
              {mutation.isPending ? 'Updating...' : 'Update Permission'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

