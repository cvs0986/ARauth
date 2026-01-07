/**
 * Delete Tenant Dialog
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
import type { Tenant } from '@shared/types/api';

interface DeleteTenantDialogProps {
  tenant: Tenant;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onConfirm: () => void;
  isLoading?: boolean;
}

export function DeleteTenantDialog({
  tenant,
  open,
  onOpenChange,
  onConfirm,
  isLoading = false,
}: DeleteTenantDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Delete Tenant</DialogTitle>
          <DialogDescription>
            Are you sure you want to delete <strong>{tenant.name}</strong>? This action cannot be undone.
            All users, roles, and permissions associated with this tenant will also be deleted.
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
            {isLoading ? 'Deleting...' : 'Delete Tenant'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

