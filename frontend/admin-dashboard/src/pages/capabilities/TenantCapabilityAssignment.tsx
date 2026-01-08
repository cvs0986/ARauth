/**
 * Tenant Capability Assignment Page
 * For SYSTEM users to assign capabilities to tenants
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { tenantCapabilityApi, systemCapabilityApi } from '@/services/api';
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
import { AssignTenantCapabilityDialog } from './AssignTenantCapabilityDialog';
import { SearchInput } from '@/components/SearchInput';
import { Badge } from '@/components/ui/badge';
import { Building2, Shield, Settings } from 'lucide-react';
import type { Tenant, SystemCapability, TenantCapability } from '@shared/types/api';

export function TenantCapabilityAssignment() {
  const queryClient = useQueryClient();
  const [selectedTenant, setSelectedTenant] = useState<Tenant | null>(null);
  const [assignDialogOpen, setAssignDialogOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');

  const { data: tenants } = useQuery({
    queryKey: ['tenants'],
    queryFn: () => tenantApi.list(),
  });

  const { data: systemCapabilities } = useQuery({
    queryKey: ['system', 'capabilities'],
    queryFn: () => systemCapabilityApi.list(),
  });

  const { data: tenantCapabilities, isLoading } = useQuery({
    queryKey: ['tenant', 'capabilities', selectedTenant?.id],
    queryFn: () => tenantCapabilityApi.list(selectedTenant!.id),
    enabled: !!selectedTenant,
  });

  const deleteMutation = useMutation({
    mutationFn: ({ tenantId, key }: { tenantId: string; key: string }) =>
      tenantCapabilityApi.delete(tenantId, key),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenant', 'capabilities'] });
    },
  });

  const filteredTenants = tenants?.filter((tenant) =>
    tenant.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    tenant.domain.toLowerCase().includes(searchQuery.toLowerCase())
  ) || [];

  if (!selectedTenant && tenants && tenants.length > 0) {
    setSelectedTenant(tenants[0]);
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Building2 className="h-8 w-8" />
          <div>
            <h1 className="text-3xl font-bold">Tenant Capability Assignment</h1>
            <p className="text-gray-500 mt-1">
              Assign system capabilities to tenants
            </p>
          </div>
        </div>
      </div>

      <div className="flex items-center gap-4">
        <SearchInput
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          placeholder="Search tenants..."
          className="max-w-md"
        />
        <select
          value={selectedTenant?.id || ''}
          onChange={(e) => {
            const tenant = tenants?.find((t) => t.id === e.target.value);
            setSelectedTenant(tenant || null);
          }}
          className="px-4 py-2 border rounded-md"
        >
          {filteredTenants.map((tenant) => (
            <option key={tenant.id} value={tenant.id}>
              {tenant.name} ({tenant.domain})
            </option>
          ))}
        </select>
      </div>

      {selectedTenant && (
        <>
          <div className="bg-white rounded-lg shadow p-4">
            <div className="flex items-center justify-between mb-4">
              <div>
                <h2 className="text-xl font-semibold">{selectedTenant.name}</h2>
                <p className="text-sm text-gray-500">{selectedTenant.domain}</p>
              </div>
              <Button onClick={() => setAssignDialogOpen(true)}>
                <Shield className="h-4 w-4 mr-2" />
                Assign Capability
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
                    <TableHead>Value</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {tenantCapabilities && tenantCapabilities.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={4} className="text-center py-8 text-gray-500">
                        No capabilities assigned
                      </TableCell>
                    </TableRow>
                  ) : (
                    tenantCapabilities?.map((cap) => (
                      <TableRow key={cap.capability_key}>
                        <TableCell className="font-mono font-medium">
                          {cap.capability_key}
                        </TableCell>
                        <TableCell>
                          <Badge variant={cap.enabled ? 'default' : 'secondary'}>
                            {cap.enabled ? 'Enabled' : 'Disabled'}
                          </Badge>
                        </TableCell>
                        <TableCell className="font-mono text-xs">
                          {cap.value
                            ? JSON.stringify(cap.value, null, 2).substring(0, 50) + '...'
                            : '-'}
                        </TableCell>
                        <TableCell>
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => {
                              if (confirm('Revoke this capability?')) {
                                deleteMutation.mutate({
                                  tenantId: selectedTenant.id,
                                  key: cap.capability_key,
                                });
                              }
                            }}
                          >
                            Revoke
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            )}
          </div>

          {assignDialogOpen && selectedTenant && systemCapabilities && (
            <AssignTenantCapabilityDialog
              tenant={selectedTenant}
              systemCapabilities={systemCapabilities}
              existingCapabilities={tenantCapabilities || []}
              open={assignDialogOpen}
              onClose={() => setAssignDialogOpen(false)}
              onSave={(key, data) => {
                tenantCapabilityApi.set(selectedTenant.id, key, data).then(() => {
                  queryClient.invalidateQueries({ queryKey: ['tenant', 'capabilities'] });
                  setAssignDialogOpen(false);
                });
              }}
            />
          )}
        </>
      )}
    </div>
  );
}

