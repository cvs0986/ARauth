/**
 * User List Page
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { userApi } from '@/services/api';
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
import { useNavigate } from 'react-router-dom';
import { CreateUserDialog } from './CreateUserDialog';
import { EditUserDialog } from './EditUserDialog';
import { DeleteUserDialog } from './DeleteUserDialog';
import { SearchInput } from '@/components/SearchInput';
import { Pagination } from '@/components/Pagination';
import type { User } from '@shared/types/api';

export function UserList({ tenantId: propTenantId }: { tenantId?: string | null } = {}) {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const { isSystemUser, selectedTenantId, tenantId, getCurrentTenantId } = useAuthStore();
  const [createOpen, setCreateOpen] = useState(false);
  const [editUser, setEditUser] = useState<User | null>(null);
  const [deleteUser, setDeleteUser] = useState<User | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  // Get current tenant context (selected tenant for SYSTEM, own tenant for TENANT)
  // Use prop tenantId if provided (for drill-down views), otherwise use context
  const currentTenantId = propTenantId || getCurrentTenantId();
  const isSystemView = isSystemUser() && !currentTenantId;

  // For SYSTEM users without tenant selected, show system users
  // For SYSTEM users with tenant selected, show tenant users
  // For TENANT users, show their own tenant users
  const { data: users, isLoading, error } = useQuery({
    queryKey: isSystemView ? ['system', 'users'] : ['users', currentTenantId],
    queryFn: () => isSystemView ? userApi.listSystem() : userApi.list(currentTenantId),
    enabled: isSystemView || !!currentTenantId || !isSystemUser(),
  });

  // Filter users based on search and status
  const filteredUsers = useMemo(() => {
    if (!users) return [];
    
    return users.filter((user) => {
      const matchesSearch =
        user.username.toLowerCase().includes(searchQuery.toLowerCase()) ||
        user.email.toLowerCase().includes(searchQuery.toLowerCase()) ||
        `${user.first_name || ''} ${user.last_name || ''}`.toLowerCase().includes(searchQuery.toLowerCase());
      
      const matchesStatus = statusFilter === 'all' || user.status === statusFilter;
      
      return matchesSearch && matchesStatus;
    });
  }, [users, searchQuery, statusFilter]);

  // Paginate filtered users
  const paginatedUsers = useMemo(() => {
    const start = (currentPage - 1) * pageSize;
    const end = start + pageSize;
    return filteredUsers.slice(start, end);
  }, [filteredUsers, currentPage, pageSize]);

  const totalPages = Math.ceil(filteredUsers.length / pageSize);

  // Reset to page 1 when filters change
  useMemo(() => {
    if (currentPage > totalPages && totalPages > 0) {
      setCurrentPage(1);
    }
  }, [totalPages, currentPage]);

  const deleteMutation = useMutation({
    mutationFn: (id: string) => userApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      setDeleteUser(null);
    },
  });

  if (isLoading) {
    return <div className="p-4">Loading users...</div>;
  }

  if (error) {
    return (
      <div className="p-4 text-red-600">
        Error loading users: {error instanceof Error ? error.message : 'Unknown error'}
      </div>
    );
  }

  // Show system users for SYSTEM users when no tenant is selected
  // The query above will handle fetching system users

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">
            {isSystemView ? 'System Users' : 'Users'}
          </h1>
          {isSystemUser() && currentTenantId && (
            <p className="text-sm text-gray-600 mt-1">
              Managing users for selected tenant
            </p>
          )}
          {isSystemView && (
            <p className="text-sm text-gray-500 mt-1">
              System users (principal_type = SYSTEM). Select a tenant from the header to manage tenant users.
            </p>
          )}
        </div>
        <Button onClick={() => setCreateOpen(true)} disabled={isSystemView ? false : !currentTenantId}>
          Create User
        </Button>
      </div>

      <div className="flex items-center gap-4">
        <SearchInput
          value={searchQuery}
          onChange={setSearchQuery}
          placeholder="Search by username, email, or name..."
        />
        <div className="space-y-2">
          <label className="text-sm font-medium">Status</label>
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="px-3 py-2 border border-gray-300 rounded-md"
          >
            <option value="all">All</option>
            <option value="active">Active</option>
            <option value="inactive">Inactive</option>
            <option value="locked">Locked</option>
          </select>
        </div>
      </div>

      <div className="border rounded-lg">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Username</TableHead>
              <TableHead>Email</TableHead>
              <TableHead>Name</TableHead>
              {isSystemUser() && <TableHead>Tenant</TableHead>}
              <TableHead>Status</TableHead>
              <TableHead>Created</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {paginatedUsers.map((user) => (
              <TableRow 
                key={user.id}
                className="cursor-pointer hover:bg-gray-50"
                onClick={() => navigate(`/users/${user.id}`)}
              >
                <TableCell className="font-medium text-primary-600 hover:text-primary-700">
                  {user.username}
                </TableCell>
                <TableCell>{user.email}</TableCell>
                <TableCell>
                  {user.first_name || user.last_name
                    ? `${user.first_name || ''} ${user.last_name || ''}`.trim()
                    : '-'}
                </TableCell>
                {isSystemUser() && (
                  <TableCell>
                    {user.tenant_id ? (
                      <span className="text-sm text-gray-600">{user.tenant_id}</span>
                    ) : (
                      <span className="text-sm text-gray-400 italic">System User</span>
                    )}
                  </TableCell>
                )}
                <TableCell>
                  <span
                    className={`px-2 py-1 rounded text-xs ${
                      user.status === 'active'
                        ? 'bg-green-100 text-green-800'
                        : user.status === 'locked'
                        ? 'bg-red-100 text-red-800'
                        : 'bg-gray-100 text-gray-800'
                    }`}
                  >
                    {user.status}
                  </span>
                </TableCell>
                <TableCell>
                  {new Date(user.created_at).toLocaleDateString()}
                </TableCell>
                <TableCell className="text-right">
                  <div className="flex justify-end gap-2" onClick={(e) => e.stopPropagation()}>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={(e) => {
                        e.stopPropagation();
                        setEditUser(user);
                      }}
                    >
                      Edit
                    </Button>
                    <Button
                      variant="destructive"
                      size="sm"
                      onClick={(e) => {
                        e.stopPropagation();
                        setDeleteUser(user);
                      }}
                    >
                      Delete
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            ))}
            {filteredUsers.length === 0 && users && users.length > 0 && (
              <TableRow>
                <TableCell colSpan={isSystemUser() ? 7 : 6} className="text-center text-gray-500">
                  No users match your search criteria.
                </TableCell>
              </TableRow>
            )}
            {users?.length === 0 && (
              <TableRow>
                <TableCell colSpan={isSystemUser() ? 7 : 6} className="text-center text-gray-500">
                  No users found. Create your first user to get started.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {!isSystemView && (
        <CreateUserDialog 
          open={createOpen} 
          onOpenChange={setCreateOpen}
          tenantId={currentTenantId || undefined}
        />
      )}
      {isSystemView && (
        <CreateUserDialog 
          open={createOpen} 
          onOpenChange={setCreateOpen}
          // For system users, don't pass tenantId - the backend should handle creating SYSTEM users
        />
      )}
      {editUser && (
        <EditUserDialog
          user={editUser}
          open={!!editUser}
          onOpenChange={(open) => !open && setEditUser(null)}
        />
      )}
      {deleteUser && (
        <DeleteUserDialog
          user={deleteUser}
          open={!!deleteUser}
          onOpenChange={(open) => !open && setDeleteUser(null)}
          onConfirm={() => {
            deleteMutation.mutate(deleteUser.id);
          }}
          isLoading={deleteMutation.isPending}
        />
      )}
    </div>
  );
}

