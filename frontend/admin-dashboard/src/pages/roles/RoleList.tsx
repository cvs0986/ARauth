/**
 * Role List Page
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { roleApi } from '@/services/api';
import { useAuthStore } from '@/store/authStore';
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
import { CreateRoleDialog } from './CreateRoleDialog';
import { EditRoleDialog } from './EditRoleDialog';
import { DeleteRoleDialog } from './DeleteRoleDialog';
import { RolePermissionsDialog } from './RolePermissionsDialog';
import { SearchInput } from '@/components/SearchInput';
import { Pagination } from '@/components/Pagination';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Info } from 'lucide-react';
import type { Role } from '@shared/types/api';

export function RoleList() {
  const queryClient = useQueryClient();
  const { isSystemUser, selectedTenantId, tenantId, getCurrentTenantId } = useAuthStore();
  const [createOpen, setCreateOpen] = useState(false);
  const [editRole, setEditRole] = useState<Role | null>(null);
  const [deleteRole, setDeleteRole] = useState<Role | null>(null);
  const [permissionsRole, setPermissionsRole] = useState<Role | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  // Get current tenant context (selected tenant for SYSTEM, own tenant for TENANT)
  const currentTenantId = getCurrentTenantId();

  const { data: roles, isLoading, error } = useQuery({
    queryKey: ['roles', currentTenantId],
    queryFn: () => roleApi.list(currentTenantId || undefined),
    enabled: !!currentTenantId || !isSystemUser(), // For SYSTEM users, require tenant selection
  });

  // Filter roles based on search
  const filteredRoles = useMemo(() => {
    if (!roles) return [];
    
    return roles.filter((role) => {
      return (
        role.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        (role.description || '').toLowerCase().includes(searchQuery.toLowerCase())
      );
    });
  }, [roles, searchQuery]);

  // Paginate filtered roles
  const paginatedRoles = useMemo(() => {
    const start = (currentPage - 1) * pageSize;
    const end = start + pageSize;
    return filteredRoles.slice(start, end);
  }, [filteredRoles, currentPage, pageSize]);

  const totalPages = Math.ceil(filteredRoles.length / pageSize);

  // Reset to page 1 when filters change
  useMemo(() => {
    if (currentPage > totalPages && totalPages > 0) {
      setCurrentPage(1);
    }
  }, [totalPages, currentPage]);

  const deleteMutation = useMutation({
    mutationFn: (id: string) => roleApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] });
      setDeleteRole(null);
    },
  });

  if (isLoading) {
    return <div className="p-4">Loading roles...</div>;
  }

  if (error) {
    return (
      <div className="p-4 text-red-600">
        Error loading roles: {error instanceof Error ? error.message : 'Unknown error'}
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Roles</h1>
        <Button onClick={() => setCreateOpen(true)}>Create Role</Button>
      </div>

      <div className="flex items-center gap-4">
        <SearchInput
          value={searchQuery}
          onChange={setSearchQuery}
          placeholder="Search by name or description..."
        />
      </div>

      <div className="border rounded-lg">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Created</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {paginatedRoles.map((role) => (
              <TableRow key={role.id}>
                <TableCell className="font-medium">{role.name}</TableCell>
                <TableCell>{role.description || '-'}</TableCell>
                <TableCell>
                  {new Date(role.created_at).toLocaleDateString()}
                </TableCell>
                <TableCell className="text-right">
                  <div className="flex justify-end gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setPermissionsRole(role)}
                    >
                      Permissions
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setEditRole(role)}
                    >
                      Edit
                    </Button>
                    <Button
                      variant="destructive"
                      size="sm"
                      onClick={() => setDeleteRole(role)}
                    >
                      Delete
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            ))}
            {filteredRoles.length === 0 && roles && roles.length > 0 && (
              <TableRow>
                <TableCell colSpan={4} className="text-center text-gray-500">
                  No roles match your search criteria.
                </TableCell>
              </TableRow>
            )}
            {roles?.length === 0 && (
              <TableRow>
                <TableCell colSpan={4} className="text-center text-gray-500">
                  No roles found. Create your first role to get started.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>

        {filteredRoles.length > 0 && (
          <Pagination
            currentPage={currentPage}
            totalPages={totalPages}
            pageSize={pageSize}
            totalItems={filteredRoles.length}
            onPageChange={setCurrentPage}
            onPageSizeChange={(size) => {
              setPageSize(size);
              setCurrentPage(1);
            }}
          />
        )}
      </div>

      <CreateRoleDialog open={createOpen} onOpenChange={setCreateOpen} />
      {editRole && (
        <EditRoleDialog
          role={editRole}
          open={!!editRole}
          onOpenChange={(open) => !open && setEditRole(null)}
        />
      )}
      {deleteRole && (
        <DeleteRoleDialog
          role={deleteRole}
          open={!!deleteRole}
          onOpenChange={(open) => !open && setDeleteRole(null)}
          onConfirm={() => {
            deleteMutation.mutate(deleteRole.id);
          }}
          isLoading={deleteMutation.isPending}
        />
      )}
      {permissionsRole && (
        <RolePermissionsDialog
          role={permissionsRole}
          open={!!permissionsRole}
          onOpenChange={(open) => !open && setPermissionsRole(null)}
        />
      )}
    </div>
  );
}

