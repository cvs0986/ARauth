/**
 * Delete Permission Dialog
 */

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import type { Permission } from '@shared/types/api';

interface DeletePermissionDialogProps {
  permission: Permission;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onConfirm: () => void;
  isLoading?: boolean;
}

export function DeletePermissionDialog({
  permission,
  open,
  onOpenChange,
  onConfirm,
  isLoading = false,
}: DeletePermissionDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Delete Permission</DialogTitle>
          <DialogDescription>
            Are you sure you want to delete permission <strong>{permission.resource}:{permission.action}</strong>?
            This action cannot be undone. All roles using this permission will lose access to it.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={isLoading}
          >
            Cancel
          </Button>
          <Button
            type="button"
            variant="destructive"
            onClick={onConfirm}
            disabled={isLoading}
          >
            {isLoading ? 'Deleting...' : 'Delete Permission'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

