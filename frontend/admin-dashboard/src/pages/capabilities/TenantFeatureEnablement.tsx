/**
 * Tenant Feature Enablement Page
 * For TENANT users to enable/disable features
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { tenantFeatureApi, tenantCapabilityApi } from '@/services/api';
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
import { EnableTenantFeatureDialog } from './EnableTenantFeatureDialog';
import { SearchInput } from '@/components/SearchInput';
import { Badge } from '@/components/ui/badge';
import { ToggleLeft, ToggleRight, Settings } from 'lucide-react';
import { useAuthStore } from '@/store/authStore';
import type { TenantFeature, CapabilityEvaluation } from '@shared/types/api';

export function TenantFeatureEnablement() {
  const queryClient = useQueryClient();
  const { tenantId } = useAuthStore();
  const [enableDialogOpen, setEnableDialogOpen] = useState(false);
  const [selectedFeature, setSelectedFeature] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');

  const { data: enabledFeatures, isLoading } = useQuery({
    queryKey: ['tenant', 'features'],
    queryFn: () => tenantFeatureApi.list(),
  });

  const { data: evaluations } = useQuery({
    queryKey: ['tenant', 'capabilities', 'evaluation', tenantId],
    queryFn: () => tenantCapabilityApi.evaluate(tenantId || ''),
    enabled: !!tenantId,
  });

  const disableMutation = useMutation({
    mutationFn: (key: string) => tenantFeatureApi.disable(key),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenant', 'features'] });
    },
  });

  // Get available capabilities (allowed but not enabled)
  const availableCapabilities = evaluations?.filter(
    (evaluation) => evaluation.tenant_allowed && !evaluation.tenant_enabled
  ) || [];

  // Filter enabled features
  const filteredFeatures = enabledFeatures?.filter((feature) =>
    feature.capability_key.toLowerCase().includes(searchQuery.toLowerCase())
  ) || [];

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <ToggleRight className="h-8 w-8" />
          <div>
            <h1 className="text-3xl font-bold">Feature Enablement</h1>
            <p className="text-gray-500 mt-1">
              Enable or disable features for your tenant
            </p>
          </div>
        </div>
        <Button onClick={() => setEnableDialogOpen(true)}>
          <ToggleLeft className="h-4 w-4 mr-2" />
          Enable Feature
        </Button>
      </div>

      <div className="flex items-center gap-4">
        <SearchInput
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          placeholder="Search features..."
          className="max-w-md"
        />
      </div>

      <div className="bg-white rounded-lg shadow">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Feature Key</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Configuration</TableHead>
              <TableHead>Enabled At</TableHead>
              <TableHead>Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell colSpan={5} className="text-center py-8">
                  Loading features...
                </TableCell>
              </TableRow>
            ) : filteredFeatures.length === 0 ? (
              <TableRow>
                <TableCell colSpan={5} className="text-center py-8 text-gray-500">
                  No features enabled
                </TableCell>
              </TableRow>
            ) : (
              filteredFeatures.map((feature) => (
                <TableRow key={feature.capability_key}>
                  <TableCell className="font-mono font-medium">
                    {feature.capability_key}
                  </TableCell>
                  <TableCell>
                    <Badge variant={feature.enabled ? 'default' : 'secondary'}>
                      {feature.enabled ? 'Enabled' : 'Disabled'}
                    </Badge>
                  </TableCell>
                  <TableCell className="font-mono text-xs">
                    {feature.configuration
                      ? JSON.stringify(feature.configuration, null, 2).substring(0, 50) + '...'
                      : '-'}
                  </TableCell>
                  <TableCell>
                    {feature.enabled_at
                      ? new Date(feature.enabled_at).toLocaleDateString()
                      : '-'}
                  </TableCell>
                  <TableCell>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => {
                        if (confirm('Disable this feature?')) {
                          disableMutation.mutate(feature.capability_key);
                        }
                      }}
                    >
                      Disable
                    </Button>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {enableDialogOpen && evaluations && (
        <EnableTenantFeatureDialog
          availableCapabilities={availableCapabilities}
          open={enableDialogOpen}
          onClose={() => setEnableDialogOpen(false)}
          onSave={(key, data) => {
            tenantFeatureApi.enable(key, data).then(() => {
              queryClient.invalidateQueries({ queryKey: ['tenant', 'features'] });
              setEnableDialogOpen(false);
            });
          }}
        />
      )}
    </div>
  );
}

