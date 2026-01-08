/**
 * System Capability Management Page
 * For SYSTEM users to manage global system capabilities
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { systemCapabilityApi } from '@/services/api';
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
import { EditSystemCapabilityDialog } from './EditSystemCapabilityDialog';
import { SearchInput } from '@/components/SearchInput';
import { Badge } from '@/components/ui/badge';
import { Settings, Shield } from 'lucide-react';
import type { SystemCapability } from '@shared/types/api';

export function SystemCapabilityList() {
  const queryClient = useQueryClient();
  const [editCapability, setEditCapability] = useState<SystemCapability | null>(null);
  const [searchQuery, setSearchQuery] = useState('');

  const { data: capabilities, isLoading, error } = useQuery({
    queryKey: ['system', 'capabilities'],
    queryFn: () => systemCapabilityApi.list(),
  });

  const updateMutation = useMutation({
    mutationFn: ({ key, data }: { key: string; data: any }) =>
      systemCapabilityApi.update(key, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['system', 'capabilities'] });
      setEditCapability(null);
    },
  });

  // Filter capabilities based on search
  const filteredCapabilities = capabilities?.filter((cap) =>
    cap.capability_key.toLowerCase().includes(searchQuery.toLowerCase()) ||
    cap.description?.toLowerCase().includes(searchQuery.toLowerCase())
  ) || [];

  if (isLoading) {
    return <div className="p-4">Loading capabilities...</div>;
  }

  if (error) {
    return (
      <div className="p-4 text-red-600">
        Error loading capabilities: {error instanceof Error ? error.message : 'Unknown error'}
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Shield className="h-8 w-8" />
          <div>
            <h1 className="text-3xl font-bold">System Capabilities</h1>
            <p className="text-gray-500 mt-1">
              Manage global system capabilities that define what features are available
            </p>
          </div>
        </div>
      </div>

      <div className="flex items-center gap-4">
        <SearchInput
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          placeholder="Search capabilities..."
          className="max-w-md"
        />
      </div>

      <div className="bg-white rounded-lg shadow">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Capability Key</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Default Value</TableHead>
              <TableHead>Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredCapabilities.length === 0 ? (
              <TableRow>
                <TableCell colSpan={5} className="text-center py-8 text-gray-500">
                  No capabilities found
                </TableCell>
              </TableRow>
            ) : (
              filteredCapabilities.map((capability) => (
                <TableRow key={capability.capability_key}>
                  <TableCell className="font-mono font-medium">
                    {capability.capability_key}
                  </TableCell>
                  <TableCell>{capability.description || '-'}</TableCell>
                  <TableCell>
                    <Badge variant={capability.enabled ? 'default' : 'secondary'}>
                      {capability.enabled ? 'Enabled' : 'Disabled'}
                    </Badge>
                  </TableCell>
                  <TableCell className="font-mono text-xs">
                    {capability.default_value
                      ? JSON.stringify(capability.default_value, null, 2).substring(0, 50) + '...'
                      : '-'}
                  </TableCell>
                  <TableCell>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setEditCapability(capability)}
                    >
                      <Settings className="h-4 w-4 mr-2" />
                      Edit
                    </Button>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {editCapability && (
        <EditSystemCapabilityDialog
          capability={editCapability}
          open={!!editCapability}
          onClose={() => setEditCapability(null)}
          onSave={(data) => {
            updateMutation.mutate({ key: editCapability.capability_key, data });
          }}
        />
      )}
    </div>
  );
}

