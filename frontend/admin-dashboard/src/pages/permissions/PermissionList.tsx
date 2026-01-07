/**
 * Permission List Page
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { permissionApi } from '@/services/api';
import { Button } from '@/components/ui/button';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { useState, useMemo } from 'react';
import { CreatePermissionDialog } from './CreatePermissionDialog';
import { EditPermissionDialog } from './EditPermissionDialog';
import { DeletePermissionDialog } from './DeletePermissionDialog';
import { SearchInput } from '@/components/SearchInput';
import { Pagination } from '@/components/Pagination';
import type { Permission } from '@shared/types/api';

export function PermissionList() {
  const queryClient = useQueryClient();
  const [createOpen, setCreateOpen] = useState(false);
  const [editPermission, setEditPermission] = useState<Permission | null>(null);
  const [deletePermission, setDeletePermission] = useState<Permission | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  const { data: permissions, isLoading, error } = useQuery({
    queryKey: ['permissions'],
    queryFn: () => permissionApi.list(),
  });

  // Filter permissions based on search
  const filteredPermissions = useMemo(() => {
    if (!permissions) return [];
    
    return permissions.filter((permission) => {
      const searchLower = searchQuery.toLowerCase();
      return (
        permission.resource.toLowerCase().includes(searchLower) ||
        permission.action.toLowerCase().includes(searchLower) ||
        (permission.description || '').toLowerCase().includes(searchLower)
      );
    });
  }, [permissions, searchQuery]);

  const deleteMutation = useMutation({
    mutationFn: (id: string) => permissionApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['permissions'] });
      setDeletePermission(null);
    },
  });

  if (isLoading) {
    return <div className="p-4">Loading permissions...</div>;
  }

  if (error) {
    return (
      <div className="p-4 text-red-600">
        Error loading permissions: {error instanceof Error ? error.message : 'Unknown error'}
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Permissions</h1>
        <Button onClick={() => setCreateOpen(true)}>Create Permission</Button>
      </div>

      <div className="flex items-center gap-4">
        <SearchInput
          value={searchQuery}
          onChange={setSearchQuery}
          placeholder="Search by resource, action, or description..."
        />
      </div>

      <div className="border rounded-lg">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Resource</TableHead>
              <TableHead>Action</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Created</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {paginatedPermissions.map((permission) => (
              <TableRow key={permission.id}>
                <TableCell className="font-medium">{permission.resource}</TableCell>
                <TableCell>{permission.action}</TableCell>
                <TableCell>{permission.description || '-'}</TableCell>
                <TableCell>
                  {new Date(permission.created_at).toLocaleDateString()}
                </TableCell>
                <TableCell className="text-right">
                  <div className="flex justify-end gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setEditPermission(permission)}
                    >
                      Edit
                    </Button>
                    <Button
                      variant="destructive"
                      size="sm"
                      onClick={() => setDeletePermission(permission)}
                    >
                      Delete
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            ))}
            {filteredPermissions.length === 0 && permissions && permissions.length > 0 && (
              <TableRow>
                <TableCell colSpan={5} className="text-center text-gray-500">
                  No permissions match your search criteria.
                </TableCell>
              </TableRow>
            )}
            {permissions?.length === 0 && (
              <TableRow>
                <TableCell colSpan={5} className="text-center text-gray-500">
                  No permissions found. Create your first permission to get started.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>

        {filteredPermissions.length > 0 && (
          <Pagination
            currentPage={currentPage}
            totalPages={totalPages}
            pageSize={pageSize}
            totalItems={filteredPermissions.length}
            onPageChange={setCurrentPage}
            onPageSizeChange={(size) => {
              setPageSize(size);
              setCurrentPage(1);
            }}
          />
        )}
      </div>

      <CreatePermissionDialog open={createOpen} onOpenChange={setCreateOpen} />
      {editPermission && (
        <EditPermissionDialog
          permission={editPermission}
          open={!!editPermission}
          onOpenChange={(open) => !open && setEditPermission(null)}
        />
      )}
      {deletePermission && (
        <DeletePermissionDialog
          permission={deletePermission}
          open={!!deletePermission}
          onOpenChange={(open) => !open && setDeletePermission(null)}
          onConfirm={() => {
            deleteMutation.mutate(deletePermission.id);
          }}
          isLoading={deleteMutation.isPending}
        />
      )}
    </div>
  );
}

