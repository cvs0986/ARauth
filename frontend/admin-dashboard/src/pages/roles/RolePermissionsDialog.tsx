/**
 * Role Permissions Dialog
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { roleApi, permissionApi } from '@/services/api';
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
import { useState } from 'react';
import type { Role, Permission } from '@shared/types/api';

interface RolePermissionsDialogProps {
  role: Role;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function RolePermissionsDialog({
  role,
  open,
  onOpenChange,
}: RolePermissionsDialogProps) {
  const queryClient = useQueryClient();
  const [error, setError] = useState<string | null>(null);

  const { data: allPermissions } = useQuery({
    queryKey: ['permissions'],
    queryFn: () => permissionApi.list(),
  });

  const { data: rolePermissions } = useQuery({
    queryKey: ['roles', role.id, 'permissions'],
    queryFn: () => roleApi.getPermissions(role.id),
    enabled: open,
  });

  const assignMutation = useMutation({
    mutationFn: ({ roleId, permissionId }: { roleId: string; permissionId: string }) =>
      roleApi.assignPermission(roleId, permissionId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles', role.id, 'permissions'] });
      setError(null);
    },
    onError: (err) => {
      setError(handleApiError(err));
    },
  });

  const removeMutation = useMutation({
    mutationFn: ({ roleId, permissionId }: { roleId: string; permissionId: string }) =>
      roleApi.removePermission(roleId, permissionId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles', role.id, 'permissions'] });
      setError(null);
    },
    onError: (err) => {
      setError(handleApiError(err));
    },
  });

  const hasPermission = (permissionId: string) => {
    return rolePermissions?.some((p) => p.id === permissionId) || false;
  };

  const togglePermission = (permission: Permission) => {
    if (hasPermission(permission.id)) {
      removeMutation.mutate({ roleId: role.id, permissionId: permission.id });
    } else {
      assignMutation.mutate({ roleId: role.id, permissionId: permission.id });
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Manage Permissions: {role.name}</DialogTitle>
          <DialogDescription>
            Assign or remove permissions for this role. Users with this role will inherit these permissions.
          </DialogDescription>
        </DialogHeader>

        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
            {error}
          </div>
        )}

        <div className="space-y-2 max-h-96 overflow-y-auto">
          {allPermissions?.map((permission) => (
            <div
              key={permission.id}
              className="flex items-center justify-between p-3 border rounded-lg"
            >
              <div>
                <div className="font-medium">
                  {permission.resource}:{permission.action}
                </div>
                {permission.description && (
                  <div className="text-sm text-gray-500">{permission.description}</div>
                )}
              </div>
              <Button
                variant={hasPermission(permission.id) ? 'default' : 'outline'}
                size="sm"
                onClick={() => togglePermission(permission)}
                disabled={assignMutation.isPending || removeMutation.isPending}
              >
                {hasPermission(permission.id) ? 'Remove' : 'Assign'}
              </Button>
            </div>
          ))}
          {allPermissions?.length === 0 && (
            <div className="text-center text-gray-500 py-8">
              No permissions available. Create permissions first.
            </div>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

