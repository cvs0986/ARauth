/**
 * User Capability Enrollment Page
 * For TENANT users to manage user capability enrollment
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { userCapabilityApi, userApi } from '@/services/api';
import { tenantCapabilityApi } from '@/services/api';
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
import { EnrollUserCapabilityDialog } from './EnrollUserCapabilityDialog';
import { SearchInput } from '@/components/SearchInput';
import { Badge } from '@/components/ui/badge';
import { User, UserCheck, UserX } from 'lucide-react';
import { useAuthStore } from '@/store/authStore';
import type { User as UserType, UserCapabilityState, CapabilityEvaluation } from '@shared/types/api';

export function UserCapabilityEnrollment() {
  const queryClient = useQueryClient();
  const { tenantId } = useAuthStore();
  const [selectedUser, setSelectedUser] = useState<UserType | null>(null);
  const [enrollDialogOpen, setEnrollDialogOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');

  const { data: users } = useQuery({
    queryKey: ['users'],
    queryFn: () => userApi.list(),
  });

  const { data: userCapabilities, isLoading } = useQuery({
    queryKey: ['user', 'capabilities', selectedUser?.id],
    queryFn: () => userCapabilityApi.list(selectedUser!.id),
    enabled: !!selectedUser,
  });

  const { data: evaluations } = useQuery({
    queryKey: ['tenant', 'capabilities', 'evaluation', tenantId, selectedUser?.id],
    queryFn: () => tenantCapabilityApi.evaluate(tenantId || '', selectedUser?.id || ''),
    enabled: !!tenantId && !!selectedUser,
  });

  const unenrollMutation = useMutation({
    mutationFn: ({ userId, key }: { userId: string; key: string }) =>
      userCapabilityApi.unenroll(userId, key),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['user', 'capabilities'] });
    },
  });

  // Get available capabilities (enabled by tenant but not enrolled by user)
  const availableCapabilities = evaluations?.filter(
    (evaluation) => evaluation.tenant_enabled && !evaluation.user_enrolled
  ) || [];

  const filteredUsers = users?.filter((user) =>
    user.username.toLowerCase().includes(searchQuery.toLowerCase()) ||
    user.email.toLowerCase().includes(searchQuery.toLowerCase())
  ) || [];

  if (!selectedUser && users && users.length > 0) {
    setSelectedUser(users[0]);
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <User className="h-8 w-8" />
          <div>
            <h1 className="text-3xl font-bold">User Capability Enrollment</h1>
            <p className="text-gray-500 mt-1">
              Manage user enrollment in tenant-enabled capabilities
            </p>
          </div>
        </div>
      </div>

      <div className="flex items-center gap-4">
        <SearchInput
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          placeholder="Search users..."
          className="max-w-md"
        />
        <select
          value={selectedUser?.id || ''}
          onChange={(e) => {
            const user = users?.find((u) => u.id === e.target.value);
            setSelectedUser(user || null);
          }}
          className="px-4 py-2 border rounded-md"
        >
          {filteredUsers.map((user) => (
            <option key={user.id} value={user.id}>
              {user.username} ({user.email})
            </option>
          ))}
        </select>
      </div>

      {selectedUser && (
        <>
          <div className="bg-white rounded-lg shadow p-4">
            <div className="flex items-center justify-between mb-4">
              <div>
                <h2 className="text-xl font-semibold">{selectedUser.username}</h2>
                <p className="text-sm text-gray-500">{selectedUser.email}</p>
              </div>
              <Button onClick={() => setEnrollDialogOpen(true)}>
                <UserCheck className="h-4 w-4 mr-2" />
                Enroll in Capability
              </Button>
            </div>

            {isLoading ? (
              <div className="p-4">Loading capabilities...</div>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Capability Key</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>State Data</TableHead>
                    <TableHead>Enrolled At</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {userCapabilities && userCapabilities.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={5} className="text-center py-8 text-gray-500">
                        No capabilities enrolled
                      </TableCell>
                    </TableRow>
                  ) : (
                    userCapabilities?.map((cap) => (
                      <TableRow key={cap.capability_key}>
                        <TableCell className="font-mono font-medium">
                          {cap.capability_key}
                        </TableCell>
                        <TableCell>
                          <Badge variant={cap.enrolled ? 'default' : 'secondary'}>
                            {cap.enrolled ? 'Enrolled' : 'Not Enrolled'}
                          </Badge>
                        </TableCell>
                        <TableCell className="font-mono text-xs">
                          {cap.state_data
                            ? JSON.stringify(cap.state_data, null, 2).substring(0, 50) + '...'
                            : '-'}
                        </TableCell>
                        <TableCell>
                          {cap.enrolled_at
                            ? new Date(cap.enrolled_at).toLocaleDateString()
                            : '-'}
                        </TableCell>
                        <TableCell>
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => {
                              if (confirm('Unenroll user from this capability?')) {
                                unenrollMutation.mutate({
                                  userId: selectedUser.id,
                                  key: cap.capability_key,
                                });
                              }
                            }}
                          >
                            <UserX className="h-4 w-4 mr-2" />
                            Unenroll
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            )}
          </div>

          {enrollDialogOpen && selectedUser && evaluations && (
            <EnrollUserCapabilityDialog
              user={selectedUser}
              availableCapabilities={availableCapabilities}
              open={enrollDialogOpen}
              onClose={() => setEnrollDialogOpen(false)}
              onSave={(key, data) => {
                userCapabilityApi.enroll(selectedUser.id, key, data).then(() => {
                  queryClient.invalidateQueries({ queryKey: ['user', 'capabilities'] });
                  setEnrollDialogOpen(false);
                });
              }}
            />
          )}
        </>
      )}
    </div>
  );
}

