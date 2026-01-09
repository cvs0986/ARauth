/**
 * Tenant Feature Enablement Page
 * For TENANT users to enable/disable features
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { tenantFeatureApi, tenantCapabilityApi, systemCapabilityApi } from '@/services/api';
import { Switch } from '@/components/ui/switch';
import { Label } from '@/components/ui/label';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { SearchInput } from '@/components/SearchInput';
import { ToggleRight, Info, CheckCircle2, XCircle, AlertCircle } from 'lucide-react';
import { useAuthStore } from '@/store/authStore';
import type { CapabilityEvaluation } from '@shared/types/api';
import { useState } from 'react';
import {
  Alert,
  AlertDescription,
} from '@/components/ui/alert';
import { formatCapabilityName, getCapabilityDescription } from '@/utils/capabilityNames';

export function TenantFeatureEnablement({ tenantId: propTenantId }: { tenantId?: string } = {}) {
  const queryClient = useQueryClient();
  const { tenantId, selectedTenantId, isSystemUser } = useAuthStore();
  const [searchQuery, setSearchQuery] = useState('');

  // For SYSTEM users: use propTenantId if provided (from TenantDetail), otherwise selectedTenantId
  // For TENANT users: use tenantId from store
  const effectiveTenantId = propTenantId || (isSystemUser() ? selectedTenantId : tenantId);

  // Fetch all evaluations (shows all capabilities with their status)
  const { data: evaluations, isLoading: evaluationsLoading, error: evaluationsError } = useQuery({
    queryKey: ['tenant', 'capabilities', 'evaluation', effectiveTenantId || 'current'],
    queryFn: async () => {
      // For SYSTEM users, must have tenantId; for TENANT users, pass empty string to use tenant-scoped endpoint
      if (isSystemUser() && !effectiveTenantId) {
        throw new Error('Tenant ID is required for SYSTEM users');
      }
      const result = await tenantCapabilityApi.evaluate(isSystemUser() ? effectiveTenantId! : '');
      console.log('[TenantFeatureEnablement] Evaluations result:', result);
      return result;
    },
    enabled: !isSystemUser() || !!effectiveTenantId, // Only enable if we have tenantId for SYSTEM users
    retry: false,
    staleTime: 0,
    gcTime: 0,
    refetchOnMount: 'always',
    refetchOnWindowFocus: false,
  });

  // Fetch enabled features (for additional info)
  const { data: enabledFeatures, isLoading: featuresLoading } = useQuery({
    queryKey: ['tenant', 'features', effectiveTenantId || 'current'],
    queryFn: () => tenantFeatureApi.list(effectiveTenantId || undefined),
    enabled: !isSystemUser() || !!effectiveTenantId, // Only enable if we have tenantId for SYSTEM users
  });

  // Fetch system capabilities for descriptions (works for both SYSTEM and TENANT users)
  const { data: systemCapabilities } = useQuery({
    queryKey: ['system', 'capabilities'],
    queryFn: () => systemCapabilityApi.list(),
    enabled: true, // Always fetch (we have tenant-scoped endpoint now)
  });

  // Create a map of system capabilities for quick lookup
  const systemCapabilitiesMap = new Map(
    (systemCapabilities || []).map((cap) => [cap.capability_key, cap])
  );

  // Enable mutation
  const enableMutation = useMutation({
    mutationFn: ({ key, config }: { key: string; config?: Record<string, unknown> }) =>
      tenantFeatureApi.enable(key, { configuration: config }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenant', 'features'] });
      queryClient.invalidateQueries({ queryKey: ['tenant', 'capabilities', 'evaluation'] });
    },
  });

  // Disable mutation
  const disableMutation = useMutation({
    mutationFn: (key: string) => tenantFeatureApi.disable(key),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenant', 'features'] });
      queryClient.invalidateQueries({ queryKey: ['tenant', 'capabilities', 'evaluation'] });
    },
  });

  // Get enabled feature configs for display
  const enabledFeatureMap = new Map(
    (enabledFeatures || []).map((f) => [f.capability_key, f])
  );

  // Filter evaluations based on search
  const filteredEvaluations = (evaluations || []).filter((evaluation) =>
    evaluation.capability_key.toLowerCase().includes(searchQuery.toLowerCase())
  );

  // Group evaluations by status
  const enabledFeaturesList = filteredEvaluations.filter((e) => e.tenant_enabled);
  const availableFeaturesList = filteredEvaluations.filter(
    (e) => e.tenant_allowed && !e.tenant_enabled && e.system_supported
  );
  const unavailableFeaturesList = filteredEvaluations.filter(
    (e) => !e.tenant_allowed || !e.system_supported
  );

  const handleToggle = (evaluation: CapabilityEvaluation, enabled: boolean) => {
    if (enabled) {
      const feature = enabledFeatureMap.get(evaluation.capability_key);
      enableMutation.mutate({
        key: evaluation.capability_key,
        config: feature?.configuration as Record<string, unknown> | undefined,
      });
    } else {
      disableMutation.mutate(evaluation.capability_key);
    }
  };

  const isLoading = evaluationsLoading || featuresLoading;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <ToggleRight className="h-8 w-8 text-primary" />
          <div>
            <h1 className="text-3xl font-bold">Feature Management</h1>
            <p className="text-gray-500 mt-1">
              Enable or disable features for your tenant. Toggle switches to manage features instantly.
            </p>
          </div>
        </div>
      </div>

      {/* Search */}
      <div className="flex items-center gap-4">
        <SearchInput
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          placeholder="Search features by name..."
          className="max-w-md"
        />
        <div className="flex items-center gap-4 text-sm text-gray-600">
          <div className="flex items-center gap-2">
            <CheckCircle2 className="h-4 w-4 text-green-600" />
            <span>{enabledFeaturesList.length} Enabled</span>
          </div>
          <div className="flex items-center gap-2">
            <AlertCircle className="h-4 w-4 text-blue-600" />
            <span>{availableFeaturesList.length} Available</span>
          </div>
          <div className="flex items-center gap-2">
            <XCircle className="h-4 w-4 text-gray-400" />
            <span>{unavailableFeaturesList.length} Unavailable</span>
          </div>
        </div>
      </div>

      {/* Error Alert */}
      {evaluationsError && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            Failed to load features. Please refresh the page or contact support.
            <br />
            <span className="text-xs mt-2 block">
              Principal Type: {principalType}, Is System: {isSystem ? 'true' : 'false'}, Tenant ID: {tenantId || 'none'}
              <br />
              Error: {evaluationsError instanceof Error ? evaluationsError.message : String(evaluationsError)}
            </span>
          </AlertDescription>
        </Alert>
      )}

      {/* Loading State */}
      {isLoading && (
        <div className="text-center py-12">
          <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
          <p className="mt-4 text-gray-500">Loading features...</p>
        </div>
      )}

      {/* Enabled Features */}
      {!isLoading && enabledFeaturesList.length > 0 && (
        <div className="space-y-4">
          <h2 className="text-xl font-semibold flex items-center gap-2">
            <CheckCircle2 className="h-5 w-5 text-green-600" />
            Enabled Features ({enabledFeaturesList.length})
          </h2>
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {enabledFeaturesList.map((evaluation) => {
              const feature = enabledFeatureMap.get(evaluation.capability_key);
              const systemCap = systemCapabilitiesMap.get(evaluation.capability_key);
              return (
                <FeatureCard
                  key={evaluation.capability_key}
                  evaluation={evaluation}
                  enabled={true}
                  feature={feature}
                  systemCapability={systemCap}
                  onToggle={(enabled) => handleToggle(evaluation, enabled)}
                  isToggling={disableMutation.isPending}
                />
              );
            })}
          </div>
        </div>
      )}

      {/* Available Features */}
      {!isLoading && availableFeaturesList.length > 0 && (
        <div className="space-y-4">
          <h2 className="text-xl font-semibold flex items-center gap-2">
            <AlertCircle className="h-5 w-5 text-blue-600" />
            Available Features ({availableFeaturesList.length})
          </h2>
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {availableFeaturesList.map((evaluation) => {
              const systemCap = systemCapabilitiesMap.get(evaluation.capability_key);
              return (
                <FeatureCard
                  key={evaluation.capability_key}
                  evaluation={evaluation}
                  enabled={false}
                  systemCapability={systemCap}
                  onToggle={(enabled) => handleToggle(evaluation, enabled)}
                  isToggling={enableMutation.isPending}
                />
              );
            })}
          </div>
        </div>
      )}

      {/* Unavailable Features */}
      {!isLoading && unavailableFeaturesList.length > 0 && (
        <div className="space-y-4">
          <h2 className="text-xl font-semibold flex items-center gap-2">
            <XCircle className="h-5 w-5 text-gray-400" />
            Unavailable Features ({unavailableFeaturesList.length})
          </h2>
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {unavailableFeaturesList.map((evaluation) => {
              const systemCap = systemCapabilitiesMap.get(evaluation.capability_key);
              return (
                <FeatureCard
                  key={evaluation.capability_key}
                  evaluation={evaluation}
                  enabled={false}
                  disabled={true}
                  systemCapability={systemCap}
                  onToggle={() => {}}
                  isToggling={false}
                />
              );
            })}
          </div>
        </div>
      )}

      {/* Empty State */}
      {!isLoading && filteredEvaluations.length === 0 && (
        <Card>
          <CardContent className="py-12 text-center">
            <Info className="h-12 w-12 text-gray-400 mx-auto mb-4" />
            <p className="text-lg font-medium text-gray-900 mb-2">No features found</p>
            <p className="text-gray-500">
              {searchQuery
                ? 'Try adjusting your search query'
                : 'No capabilities have been assigned to your tenant. Contact your system administrator.'}
            </p>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

interface FeatureCardProps {
  evaluation: CapabilityEvaluation;
  enabled: boolean;
  disabled?: boolean;
  feature?: { configuration?: unknown; enabled_at?: string };
  systemCapability?: { description?: string };
  onToggle: (enabled: boolean) => void;
  isToggling: boolean;
}

function FeatureCard({
  evaluation,
  enabled,
  disabled = false,
  feature,
  systemCapability,
  onToggle,
  isToggling,
}: FeatureCardProps) {
  const displayName = formatCapabilityName(evaluation.capability_key);
  const description = getCapabilityDescription(evaluation.capability_key, systemCapability?.description);

  const getStatusBadge = () => {
    if (disabled) {
      if (!evaluation.system_supported) {
        return <Badge variant="secondary">Not Supported</Badge>;
      }
      if (!evaluation.tenant_allowed) {
        return <Badge variant="secondary">Not Allowed</Badge>;
      }
      return <Badge variant="secondary">Unavailable</Badge>;
    }
    return enabled ? (
      <Badge className="bg-green-100 text-green-800 hover:bg-green-100">Enabled</Badge>
    ) : (
      <Badge variant="outline">Available</Badge>
    );
  };

  const hasSystemValue = evaluation.system_value && Object.keys(evaluation.system_value).length > 0;
  const hasTenantValue = evaluation.tenant_value && Object.keys(evaluation.tenant_value).length > 0;
  const hasTenantConfig = evaluation.tenant_configuration && Object.keys(evaluation.tenant_configuration).length > 0;
  const hasCustomValue = feature?.configuration && typeof feature.configuration === 'object' && Object.keys(feature.configuration).length > 0;

  return (
    <Card className={`transition-all ${enabled ? 'border-green-200 bg-green-50/50' : ''}`}>
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <CardTitle className="text-base font-semibold">{displayName}</CardTitle>
            <CardDescription className="mt-1 text-sm text-gray-600">{description}</CardDescription>
          </div>
          <div className="ml-4">{getStatusBadge()}</div>
        </div>
      </CardHeader>
      <CardContent className="pt-0">
        <div className="space-y-3">
          {/* Prominent Toggle Section */}
          <div className="flex items-center justify-between p-4 bg-white rounded-lg border-2 border-gray-200 shadow-sm">
            <div className="flex items-center gap-3">
              <Label 
                htmlFor={`toggle-${evaluation.capability_key}`} 
                className="text-base font-semibold cursor-pointer flex items-center gap-2"
              >
                <span className={enabled ? 'text-green-700 font-bold' : 'text-gray-700'}>
                  {enabled ? '✓ Enabled' : '○ Disabled'}
                </span>
              </Label>
            </div>
            <Switch
              id={`toggle-${evaluation.capability_key}`}
              checked={enabled}
              disabled={disabled || isToggling}
              onCheckedChange={onToggle}
              className="h-7 w-14 data-[state=checked]:bg-green-600 data-[state=unchecked]:bg-gray-300 shadow-sm [&>span]:h-6 [&>span]:w-6 [&>span]:bg-white [&>span]:shadow-lg [&>span]:border [&>span]:border-gray-200 [&>span]:data-[state=checked]:translate-x-7 [&>span]:data-[state=unchecked]:translate-x-0.5"
            />
          </div>
          
          {/* Value Information */}
          <div className="space-y-2 pt-2 border-t">
            {hasCustomValue && (
              <div>
                <p className="text-xs font-medium text-gray-700 mb-1">Custom value configured:</p>
                <pre className="text-xs bg-gray-100 p-2 rounded overflow-x-auto max-h-24">
                  {JSON.stringify(feature.configuration, null, 2)}
                </pre>
              </div>
            )}
            {hasTenantConfig && (
              <div>
                <p className="text-xs font-medium text-gray-700 mb-1">Tenant configuration:</p>
                <pre className="text-xs bg-gray-100 p-2 rounded overflow-x-auto max-h-24">
                  {JSON.stringify(evaluation.tenant_configuration, null, 2)}
                </pre>
              </div>
            )}
            {hasTenantValue && (
              <div>
                <p className="text-xs font-medium text-gray-700 mb-1">Tenant value:</p>
                <pre className="text-xs bg-gray-100 p-2 rounded overflow-x-auto max-h-24">
                  {JSON.stringify(evaluation.tenant_value, null, 2)}
                </pre>
              </div>
            )}
            {hasSystemValue && (
              <div>
                <p className="text-xs font-medium text-gray-700 mb-1">System default available:</p>
                <pre className="text-xs bg-blue-50 p-2 rounded overflow-x-auto max-h-24 border border-blue-200">
                  {JSON.stringify(evaluation.system_value, null, 2)}
                </pre>
              </div>
            )}
            {enabled && feature?.enabled_at && (
              <p className="text-xs text-gray-500">
                Enabled: {new Date(feature.enabled_at).toLocaleDateString()}
              </p>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
