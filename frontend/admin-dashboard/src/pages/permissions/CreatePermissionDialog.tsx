/**
 * Create Permission Dialog
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
import { useState } from 'react';

const createPermissionSchema = z.object({
  resource: z.string().min(1, 'Resource is required'),
  action: z.string().min(1, 'Action is required'),
  description: z.string().optional(),
});

type CreatePermissionFormData = z.infer<typeof createPermissionSchema>;

interface CreatePermissionDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function CreatePermissionDialog({ open, onOpenChange }: CreatePermissionDialogProps) {
  const queryClient = useQueryClient();
  const [error, setError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
  } = useForm<CreatePermissionFormData>({
    resolver: zodResolver(createPermissionSchema),
  });

  const mutation = useMutation({
    mutationFn: permissionApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['permissions'] });
      reset();
      setError(null);
      onOpenChange(false);
    },
    onError: (err) => {
      setError(handleApiError(err));
    },
  });

  const onSubmit = (data: CreatePermissionFormData) => {
    setError(null);
    mutation.mutate(data);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create Permission</DialogTitle>
          <DialogDescription>
            Create a new permission. Permissions define what actions can be performed on resources.
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
              {mutation.isPending ? 'Creating...' : 'Create Permission'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

