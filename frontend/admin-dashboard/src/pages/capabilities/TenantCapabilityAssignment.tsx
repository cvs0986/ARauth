/**
 * Tenant Capability Assignment Page
 * For SYSTEM users to assign capabilities to tenants
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { tenantCapabilityApi, systemCapabilityApi, systemApi } from '@/services/api';
import { tenantApi } from '@/services/api';
import { Switch } from '@/components/ui/switch';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { SearchInput } from '@/components/SearchInput';
import { Building2, Shield, CheckCircle2, XCircle, AlertCircle, Info } from 'lucide-react';
import type { Tenant, SystemCapability, TenantCapability } from '@shared/types/api';
import { useState, useEffect } from 'react';
import { formatCapabilityName, getCapabilityDescription } from '@/utils/capabilityNames';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

export function TenantCapabilityAssignment({ tenantId: propTenantId }: { tenantId?: string } = {}) {
  const queryClient = useQueryClient();
  const [selectedTenant, setSelectedTenant] = useState<Tenant | null>(null);
  const [searchQuery, setSearchQuery] = useState('');

  const { data: tenants, isLoading: tenantsLoading } = useQuery({
    queryKey: ['tenants'],
    queryFn: () => tenantApi.list(),
  });

  // If propTenantId is provided, find and set that tenant
  const { data: propTenant } = useQuery({
    queryKey: ['tenant', propTenantId],
    queryFn: () => systemApi.tenants.getById(propTenantId!),
    enabled: !!propTenantId,
    onSuccess: (tenant) => {
      if (tenant && !selectedTenant) {
        setSelectedTenant(tenant);
      }
    },
  });

  // Use propTenantId if provided, otherwise use selectedTenant
  const effectiveTenantId = propTenantId || selectedTenant?.id;

  const { data: systemCapabilities, isLoading: systemCapabilitiesLoading } = useQuery({
    queryKey: ['system', 'capabilities'],
    queryFn: () => systemCapabilityApi.list(),
  });

  const { data: tenantCapabilities, isLoading: tenantCapabilitiesLoading } = useQuery({
    queryKey: ['tenant', 'capabilities', effectiveTenantId],
    queryFn: () => tenantCapabilityApi.list(effectiveTenantId!),
    enabled: !!effectiveTenantId,
  });

  // Auto-select first tenant only if propTenantId is not provided
  useEffect(() => {
    if (!propTenantId && !selectedTenant && tenants && tenants.length > 0) {
      setSelectedTenant(tenants[0]);
    }
  }, [tenants, selectedTenant, propTenantId]);

  const assignMutation = useMutation({
    mutationFn: ({ tenantId, key, data }: { tenantId: string; key: string; data: any }) =>
      tenantCapabilityApi.set(tenantId, key, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenant', 'capabilities'] });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: ({ tenantId, key }: { tenantId: string; key: string }) =>
      tenantCapabilityApi.delete(tenantId, key),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenant', 'capabilities'] });
    },
  });

  const handleToggle = (capability: SystemCapability, assigned: boolean) => {
    if (!effectiveTenantId) return;

    if (assigned) {
      // Assign capability
      assignMutation.mutate({
        tenantId: effectiveTenantId,
        key: capability.capability_key,
        data: {
          enabled: true,
          value: capability.default_value,
        },
      });
    } else {
      // Unassign capability
      deleteMutation.mutate({
        tenantId: effectiveTenantId,
        key: capability.capability_key,
      });
    }
  };

  // Create a map of assigned capabilities
  const assignedCapabilitiesMap = new Map(
    (Array.isArray(tenantCapabilities) ? tenantCapabilities : []).map((cap) => [
      cap.capability_key,
      cap,
    ])
  );

  // Filter system capabilities based on search
  const filteredCapabilities = (systemCapabilities || []).filter((cap) =>
    cap.capability_key.toLowerCase().includes(searchQuery.toLowerCase()) ||
    cap.description?.toLowerCase().includes(searchQuery.toLowerCase())
  );

  // Group capabilities
  const assignedCapabilities = filteredCapabilities.filter((cap) =>
    assignedCapabilitiesMap.has(cap.capability_key)
  );
  const availableCapabilities = filteredCapabilities.filter(
    (cap) => !assignedCapabilitiesMap.has(cap.capability_key)
  );

  const isLoading = tenantsLoading || systemCapabilitiesLoading || tenantCapabilitiesLoading;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Building2 className="h-8 w-8 text-primary" />
          <div>
            <h1 className="text-3xl font-bold">Tenant Capability Assignment</h1>
            <p className="text-gray-500 mt-1">
              Assign system capabilities to tenants. Toggle switches to assign or unassign capabilities.
            </p>
          </div>
        </div>
      </div>

      {/* Tenant Selector - Only show if propTenantId is not provided */}
      {!propTenantId && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Select Tenant</CardTitle>
            <CardDescription>Choose a tenant to manage its capabilities</CardDescription>
          </CardHeader>
          <CardContent>
            {tenantsLoading ? (
              <div className="text-center py-4">
                <div className="inline-block animate-spin rounded-full h-6 w-6 border-b-2 border-primary"></div>
              </div>
            ) : (
              <Select
                value={selectedTenant?.id || ''}
                onValueChange={(value) => {
                  const tenant = tenants?.find((t) => t.id === value);
                  setSelectedTenant(tenant || null);
                }}
              >
                <SelectTrigger className="w-full max-w-md">
                  <SelectValue placeholder="Select a tenant..." />
                </SelectTrigger>
                <SelectContent>
                  {tenants?.map((tenant) => (
                    <SelectItem key={tenant.id} value={tenant.id}>
                      {tenant.name} ({tenant.domain})
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          </CardContent>
        </Card>
      )}

      {effectiveTenantId && (propTenant || selectedTenant) && (
        <>
          {/* Tenant Info */}
          <Card className="border-primary/20 bg-primary/5">
            <CardContent className="pt-6">
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-xl font-semibold">{(propTenant || selectedTenant)?.name}</h2>
                  <p className="text-sm text-gray-500 mt-1">{(propTenant || selectedTenant)?.domain}</p>
                </div>
                <Badge variant={(propTenant || selectedTenant)?.status === 'active' ? 'default' : 'secondary'}>
                  {(propTenant || selectedTenant)?.status}
                </Badge>
              </div>
            </CardContent>
          </Card>

          {/* Search */}
          <div className="flex items-center gap-4">
            <SearchInput
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search capabilities by name or description..."
              className="max-w-md"
            />
            <div className="flex items-center gap-4 text-sm text-gray-600">
              <div className="flex items-center gap-2">
                <CheckCircle2 className="h-4 w-4 text-green-600" />
                <span>{assignedCapabilities.length} Assigned</span>
              </div>
              <div className="flex items-center gap-2">
                <AlertCircle className="h-4 w-4 text-blue-600" />
                <span>{availableCapabilities.length} Available</span>
              </div>
            </div>
          </div>

          {/* Loading State */}
          {isLoading && (
            <div className="text-center py-12">
              <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
              <p className="mt-4 text-gray-500">Loading capabilities...</p>
            </div>
          )}

          {/* Assigned Capabilities */}
          {!isLoading && assignedCapabilities.length > 0 && (
            <div className="space-y-4">
              <h2 className="text-xl font-semibold flex items-center gap-2">
                <CheckCircle2 className="h-5 w-5 text-green-600" />
                Assigned Capabilities ({assignedCapabilities.length})
              </h2>
              <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                {assignedCapabilities.map((capability) => {
                  const assigned = assignedCapabilitiesMap.get(capability.capability_key);
                  return (
                    <CapabilityCard
                      key={capability.capability_key}
                      capability={capability}
                      assigned={true}
                      tenantCapability={assigned}
                      onToggle={(assigned) => handleToggle(capability, assigned)}
                      isToggling={deleteMutation.isPending}
                    />
                  );
                })}
              </div>
            </div>
          )}

          {/* Available Capabilities */}
          {!isLoading && availableCapabilities.length > 0 && (
            <div className="space-y-4">
              <h2 className="text-xl font-semibold flex items-center gap-2">
                <AlertCircle className="h-5 w-5 text-blue-600" />
                Available Capabilities ({availableCapabilities.length})
              </h2>
              <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                {availableCapabilities.map((capability) => (
                  <CapabilityCard
                    key={capability.capability_key}
                    capability={capability}
                    assigned={false}
                    onToggle={(assigned) => handleToggle(capability, assigned)}
                    isToggling={assignMutation.isPending}
                  />
                ))}
              </div>
            </div>
          )}

          {/* Empty State */}
          {!isLoading && filteredCapabilities.length === 0 && (
            <Card>
              <CardContent className="py-12 text-center">
                <Info className="h-12 w-12 text-gray-400 mx-auto mb-4" />
                <p className="text-lg font-medium text-gray-900 mb-2">No capabilities found</p>
                <p className="text-gray-500">
                  {searchQuery
                    ? 'Try adjusting your search query'
                    : 'No system capabilities available'}
                </p>
              </CardContent>
            </Card>
          )}
        </>
      )}

      {!effectiveTenantId && !tenantsLoading && (
        <Card>
          <CardContent className="py-12 text-center">
            <Info className="h-12 w-12 text-gray-400 mx-auto mb-4" />
            <p className="text-lg font-medium text-gray-900 mb-2">No tenant selected</p>
            <p className="text-gray-500">Please select a tenant to manage its capabilities</p>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

interface CapabilityCardProps {
  capability: SystemCapability;
  assigned: boolean;
  tenantCapability?: TenantCapability;
  onToggle: (assigned: boolean) => void;
  isToggling: boolean;
}

function CapabilityCard({
  capability,
  assigned,
  tenantCapability,
  onToggle,
  isToggling,
}: CapabilityCardProps) {
  return (
    <Card className={`transition-all ${assigned ? 'border-green-200 bg-green-50/50' : ''}`}>
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <CardTitle className="text-base font-semibold">{formatCapabilityName(capability.capability_key)}</CardTitle>
            <CardDescription className="mt-1 text-sm text-gray-600">
              {getCapabilityDescription(capability.capability_key, capability.description)}
            </CardDescription>
          </div>
          <div className="ml-4">
            {assigned ? (
              <Badge className="bg-green-100 text-green-800 hover:bg-green-100">Assigned</Badge>
            ) : (
              <Badge variant="outline">Available</Badge>
            )}
          </div>
        </div>
      </CardHeader>
      <CardContent className="pt-0">
        <div className="space-y-3">
          {/* Prominent Toggle Section */}
          <div className="flex items-center justify-between p-4 bg-white rounded-lg border-2 border-gray-200 shadow-sm">
            <div className="flex items-center gap-3">
              <Label 
                htmlFor={`toggle-${capability.capability_key}`} 
                className="text-base font-semibold cursor-pointer flex items-center gap-2"
              >
                <span className={assigned ? 'text-green-700 font-bold' : 'text-gray-700'}>
                  {assigned ? '✓ Assigned' : '○ Not Assigned'}
                </span>
              </Label>
            </div>
            <Switch
              id={`toggle-${capability.capability_key}`}
              checked={assigned}
              disabled={isToggling}
              onCheckedChange={onToggle}
              className="h-7 w-14 data-[state=checked]:bg-green-600 data-[state=unchecked]:bg-gray-300 shadow-sm [&>span]:h-6 [&>span]:w-6 [&>span]:bg-white [&>span]:shadow-lg [&>span]:border [&>span]:border-gray-200 [&>span]:data-[state=checked]:translate-x-7 [&>span]:data-[state=unchecked]:translate-x-0.5"
            />
          </div>
          
          {/* Value Information */}
          <div className="space-y-2 pt-2 border-t">
            {assigned && tenantCapability && (
              <div>
                <p className="text-xs text-gray-500 mb-1">
                  Status: <Badge variant={tenantCapability.enabled ? 'default' : 'secondary'} className="ml-1">
                    {tenantCapability.enabled ? 'Enabled' : 'Disabled'}
                  </Badge>
                </p>
              </div>
            )}
            {tenantCapability?.value && Object.keys(tenantCapability.value).length > 0 && (
              <div>
                <p className="text-xs font-medium text-gray-700 mb-1">Custom value configured:</p>
                <pre className="text-xs bg-gray-100 p-2 rounded overflow-x-auto max-h-24">
                  {JSON.stringify(tenantCapability.value, null, 2)}
                </pre>
              </div>
            )}
            {capability.default_value && Object.keys(capability.default_value).length > 0 && (
              <div>
                <p className="text-xs font-medium text-gray-700 mb-1">System default available:</p>
                <pre className="text-xs bg-blue-50 p-2 rounded overflow-x-auto max-h-24 border border-blue-200">
                  {JSON.stringify(capability.default_value, null, 2)}
                </pre>
              </div>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
