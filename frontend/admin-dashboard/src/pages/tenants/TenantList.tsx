/**
 * Tenant List Page
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { tenantApi } from '@/services/api';
import { Button } from '@/components/ui/button';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { useState } from 'react';
import { CreateTenantDialog } from './CreateTenantDialog';
import { EditTenantDialog } from './EditTenantDialog';
import { DeleteTenantDialog } from './DeleteTenantDialog';
import type { Tenant } from '@shared/types/api';

export function TenantList() {
  const queryClient = useQueryClient();
  const [createOpen, setCreateOpen] = useState(false);
  const [editTenant, setEditTenant] = useState<Tenant | null>(null);
  const [deleteTenant, setDeleteTenant] = useState<Tenant | null>(null);

  const { data: tenants, isLoading, error } = useQuery({
    queryKey: ['tenants'],
    queryFn: () => tenantApi.list(),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => tenantApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenants'] });
      setDeleteTenant(null);
    },
  });

  if (isLoading) {
    return <div className="p-4">Loading tenants...</div>;
  }

  if (error) {
    return (
      <div className="p-4 text-red-600">
        Error loading tenants: {error instanceof Error ? error.message : 'Unknown error'}
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Tenants</h1>
        <Button onClick={() => setCreateOpen(true)}>Create Tenant</Button>
      </div>

      <div className="border rounded-lg">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Domain</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Created</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {tenants?.map((tenant) => (
              <TableRow key={tenant.id}>
                <TableCell className="font-medium">{tenant.name}</TableCell>
                <TableCell>{tenant.domain}</TableCell>
                <TableCell>
                  <span
                    className={`px-2 py-1 rounded text-xs ${
                      tenant.status === 'active'
                        ? 'bg-green-100 text-green-800'
                        : 'bg-gray-100 text-gray-800'
                    }`}
                  >
                    {tenant.status}
                  </span>
                </TableCell>
                <TableCell>
                  {new Date(tenant.created_at).toLocaleDateString()}
                </TableCell>
                <TableCell className="text-right">
                  <div className="flex justify-end gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setEditTenant(tenant)}
                    >
                      Edit
                    </Button>
                    <Button
                      variant="destructive"
                      size="sm"
                      onClick={() => setDeleteTenant(tenant)}
                    >
                      Delete
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            ))}
            {tenants?.length === 0 && (
              <TableRow>
                <TableCell colSpan={5} className="text-center text-gray-500">
                  No tenants found. Create your first tenant to get started.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      <CreateTenantDialog open={createOpen} onOpenChange={setCreateOpen} />
      {editTenant && (
        <EditTenantDialog
          tenant={editTenant}
          open={!!editTenant}
          onOpenChange={(open) => !open && setEditTenant(null)}
        />
      )}
      {deleteTenant && (
        <DeleteTenantDialog
          tenant={deleteTenant}
          open={!!deleteTenant}
          onOpenChange={(open) => !open && setDeleteTenant(null)}
          onConfirm={() => {
            deleteMutation.mutate(deleteTenant.id);
          }}
          isLoading={deleteMutation.isPending}
        />
      )}
    </div>
  );
}

