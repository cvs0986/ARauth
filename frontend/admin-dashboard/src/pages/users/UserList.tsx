/**
 * User List Page
 * 
 * AUTHORITY MODEL:
 * Who: SYSTEM users (all tenants) OR TENANT users (own tenant only)
 * Scope: Platform-wide (SYSTEM) OR Tenant-scoped (TENANT)
 * Permission: users:read
 * 
 * GUARDRAIL #1: Backend Is Law
 * - All user data from backend APIs
 * - Tenant filtering based on console mode
 * 
 * GUARDRAIL #3: Strict Plane Separation
 * - SYSTEM users see tenant column and filter
 * - TENANT users never see cross-tenant data
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { userApi } from '@/services/api';
import { usePrincipalContext } from '@/contexts/PrincipalContext';
import { Button } from '@/components/ui/button';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { useState, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { CreateUserDialog } from './CreateUserDialog';
import { EditUserDialog } from './EditUserDialog';
import { DeleteUserDialog } from './DeleteUserDialog';
import { SearchInput } from '@/components/SearchInput';
import { Pagination } from '@/components/Pagination';
import { PermissionGate } from '@/components/PermissionGate';
import { Shield, CheckCircle2, XCircle } from 'lucide-react';
import type { User } from '@shared/types/api';

export function UserList({ tenantId: propTenantId }: { tenantId?: string | null } = {}) {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const { principalType, homeTenantId, selectedTenantId, consoleMode } = usePrincipalContext();
  const [createOpen, setCreateOpen] = useState(false);
  const [editUser, setEditUser] = useState<User | null>(null);
  const [deleteUser, setDeleteUser] = useState<User | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  // Determine effective tenant ID
  // Priority: prop > selected > home
  const effectiveTenantId = propTenantId ||
    (principalType === 'SYSTEM' ? selectedTenantId : homeTenantId);

  // SYSTEM mode without tenant = system users
  const isSystemView = consoleMode === 'SYSTEM' && !effectiveTenantId;

  // Fetch users based on context
  const { data: users, isLoading, error } = useQuery({
    queryKey: isSystemView ? ['system', 'users'] : ['users', effectiveTenantId],
    queryFn: () => isSystemView ? userApi.listSystem() : userApi.list(effectiveTenantId!),
    enabled: isSystemView || !!effectiveTenantId,
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

  return (
    <PermissionGate permission="users:read" systemPermission={principalType === 'SYSTEM'}>
      <div className="space-y-4">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold">
              {isSystemView ? 'System Users' : 'Users'}
            </h1>
            {consoleMode === 'SYSTEM' && effectiveTenantId && (
              <p className="text-sm text-gray-600 mt-1">
                Managing users for selected tenant
              </p>
            )}
            {isSystemView && (
              <p className="text-sm text-gray-500 mt-1">
                System users (principal_type = SYSTEM). Select a tenant to manage tenant users.
              </p>
            )}
          </div>
          <PermissionGate permission="users:create" systemPermission={principalType === 'SYSTEM'}>
            <Button onClick={() => setCreateOpen(true)} disabled={!isSystemView && !effectiveTenantId}>
              Create User
            </Button>
          </PermissionGate>
        </div>

        {/* Filters */}
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

        {/* Table */}
        <div className="border rounded-lg">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Username</TableHead>
                <TableHead>Email</TableHead>
                <TableHead>Name</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>MFA</TableHead>
                <TableHead>Roles</TableHead>
                {consoleMode === 'SYSTEM' && <TableHead>Tenant</TableHead>}
                <TableHead>Last Login</TableHead>
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
                  <TableCell>
                    <Badge
                      variant={
                        user.status === 'active' ? 'default' :
                          user.status === 'locked' ? 'destructive' :
                            'secondary'
                      }
                    >
                      {user.status}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    {user.mfa_enabled ? (
                      <CheckCircle2 className="h-4 w-4 text-green-600" />
                    ) : (
                      <XCircle className="h-4 w-4 text-gray-400" />
                    )}
                  </TableCell>
                  <TableCell>
                    <div className="flex gap-1 flex-wrap">
                      {user.roles && user.roles.length > 0 ? (
                        user.roles.slice(0, 2).map((role) => (
                          <Badge key={role} variant="outline" className="text-xs">
                            {role}
                          </Badge>
                        ))
                      ) : (
                        <span className="text-xs text-gray-400">No roles</span>
                      )}
                      {user.roles && user.roles.length > 2 && (
                        <Badge variant="outline" className="text-xs">
                          +{user.roles.length - 2}
                        </Badge>
                      )}
                    </div>
                  </TableCell>
                  {consoleMode === 'SYSTEM' && (
                    <TableCell>
                      {user.tenant_id ? (
                        <span className="text-sm text-gray-600">{user.tenant_id}</span>
                      ) : (
                        <Badge variant="secondary" className="text-xs">
                          <Shield className="h-3 w-3 mr-1" />
                          System
                        </Badge>
                      )}
                    </TableCell>
                  )}
                  <TableCell>
                    {user.last_login ? (
                      new Date(user.last_login).toLocaleDateString()
                    ) : (
                      <span className="text-gray-400">Never</span>
                    )}
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex justify-end gap-2" onClick={(e) => e.stopPropagation()}>
                      <PermissionGate permission="users:update" systemPermission={principalType === 'SYSTEM'}>
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
                      </PermissionGate>
                      <PermissionGate permission="users:delete" systemPermission={principalType === 'SYSTEM'}>
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
                      </PermissionGate>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
              {filteredUsers.length === 0 && users && users.length > 0 && (
                <TableRow>
                  <TableCell colSpan={consoleMode === 'SYSTEM' ? 9 : 8} className="text-center text-gray-500">
                    No users match your search criteria.
                  </TableCell>
                </TableRow>
              )}
              {users?.length === 0 && (
                <TableRow>
                  <TableCell colSpan={consoleMode === 'SYSTEM' ? 9 : 8} className="text-center text-gray-500">
                    No users found. Create your first user to get started.
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>

        {/* Pagination */}
        {totalPages > 1 && (
          <Pagination
            currentPage={currentPage}
            totalPages={totalPages}
            onPageChange={setCurrentPage}
            pageSize={pageSize}
            onPageSizeChange={setPageSize}
            totalItems={filteredUsers.length}
          />
        )}

        {/* Dialogs */}
        <CreateUserDialog
          open={createOpen}
          onOpenChange={setCreateOpen}
          tenantId={isSystemView ? undefined : effectiveTenantId || undefined}
        />
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
    </PermissionGate>
  );
}
