/**
 * System Capability Management Page
 * For SYSTEM users to manage global system capabilities
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { systemCapabilityApi } from '@/services/api';
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
import { Shield, Settings, CheckCircle2, XCircle, Info, Edit } from 'lucide-react';
import type { SystemCapability } from '@shared/types/api';
import { useState } from 'react';
import { EditSystemCapabilityDialog } from './EditSystemCapabilityDialog';
import { formatCapabilityName, getCapabilityDescription } from '@/utils/capabilityNames';
import {
  Alert,
  AlertDescription,
} from '@/components/ui/alert';

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
  const filteredCapabilities = (capabilities || []).filter((cap) =>
    cap.capability_key.toLowerCase().includes(searchQuery.toLowerCase()) ||
    cap.description?.toLowerCase().includes(searchQuery.toLowerCase())
  );

  // Group capabilities by status
  const enabledCapabilities = filteredCapabilities.filter((cap) => cap.enabled);
  const disabledCapabilities = filteredCapabilities.filter((cap) => !cap.enabled);

  const handleToggle = (capability: SystemCapability, enabled: boolean) => {
    updateMutation.mutate({
      key: capability.capability_key,
      data: {
        enabled,
        default_value: capability.default_value,
      },
    });
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Shield className="h-8 w-8 text-primary" />
          <div>
            <h1 className="text-3xl font-bold">System Capabilities</h1>
            <p className="text-gray-500 mt-1">
              Manage global system capabilities that define what features are available to tenants
            </p>
          </div>
        </div>
      </div>

      {/* Search and Stats */}
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
            <span>{enabledCapabilities.length} Enabled</span>
          </div>
          <div className="flex items-center gap-2">
            <XCircle className="h-4 w-4 text-gray-400" />
            <span>{disabledCapabilities.length} Disabled</span>
          </div>
          <div className="flex items-center gap-2">
            <Info className="h-4 w-4 text-blue-600" />
            <span>{filteredCapabilities.length} Total</span>
          </div>
        </div>
      </div>

      {/* Error Alert */}
      {error && (
        <Alert variant="destructive">
          <XCircle className="h-4 w-4" />
          <AlertDescription>
            Failed to load capabilities: {error instanceof Error ? error.message : 'Unknown error'}
          </AlertDescription>
        </Alert>
      )}

      {/* Loading State */}
      {isLoading && (
        <div className="text-center py-12">
          <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
          <p className="mt-4 text-gray-500">Loading capabilities...</p>
        </div>
      )}

      {/* Enabled Capabilities */}
      {!isLoading && enabledCapabilities.length > 0 && (
        <div className="space-y-4">
          <h2 className="text-xl font-semibold flex items-center gap-2">
            <CheckCircle2 className="h-5 w-5 text-green-600" />
            Enabled Capabilities ({enabledCapabilities.length})
          </h2>
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {enabledCapabilities.map((capability) => (
              <SystemCapabilityCard
                key={capability.capability_key}
                capability={capability}
                onToggle={(enabled) => handleToggle(capability, enabled)}
                onEdit={() => setEditCapability(capability)}
                isToggling={updateMutation.isPending}
              />
            ))}
          </div>
        </div>
      )}

      {/* Disabled Capabilities */}
      {!isLoading && disabledCapabilities.length > 0 && (
        <div className="space-y-4">
          <h2 className="text-xl font-semibold flex items-center gap-2">
            <XCircle className="h-5 w-5 text-gray-400" />
            Disabled Capabilities ({disabledCapabilities.length})
          </h2>
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {disabledCapabilities.map((capability) => (
              <SystemCapabilityCard
                key={capability.capability_key}
                capability={capability}
                onToggle={(enabled) => handleToggle(capability, enabled)}
                onEdit={() => setEditCapability(capability)}
                isToggling={updateMutation.isPending}
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

      {/* Edit Dialog */}
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

interface SystemCapabilityCardProps {
  capability: SystemCapability;
  onToggle: (enabled: boolean) => void;
  onEdit: () => void;
  isToggling: boolean;
}

function SystemCapabilityCard({
  capability,
  onToggle,
  onEdit,
  isToggling,
}: SystemCapabilityCardProps) {
  return (
    <Card className={`transition-all ${capability.enabled ? 'border-green-200 bg-green-50/50' : 'border-gray-200'}`}>
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <CardTitle className="text-base font-semibold">{formatCapabilityName(capability.capability_key)}</CardTitle>
            <CardDescription className="mt-1 text-sm text-gray-600">
              {getCapabilityDescription(capability.capability_key, capability.description)}
            </CardDescription>
          </div>
          <div className="ml-4">
            {capability.enabled ? (
              <Badge className="bg-green-100 text-green-800 hover:bg-green-100">Enabled</Badge>
            ) : (
              <Badge variant="secondary">Disabled</Badge>
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
                <span className={capability.enabled ? 'text-green-700 font-bold' : 'text-gray-700'}>
                  {capability.enabled ? '✓ Enabled' : '○ Disabled'}
                </span>
              </Label>
            </div>
            <Switch
              id={`toggle-${capability.capability_key}`}
              checked={capability.enabled}
              disabled={isToggling}
              onCheckedChange={onToggle}
              className="h-7 w-14 data-[state=checked]:bg-green-600 data-[state=unchecked]:bg-gray-300 shadow-sm [&>span]:h-6 [&>span]:w-6 [&>span]:bg-white [&>span]:shadow-lg [&>span]:border [&>span]:border-gray-200 [&>span]:data-[state=checked]:translate-x-7 [&>span]:data-[state=unchecked]:translate-x-0.5"
            />
          </div>
          {capability.default_value && Object.keys(capability.default_value).length > 0 && (
            <div className="pt-2 border-t">
              <p className="text-xs font-medium text-gray-700 mb-1">System default available:</p>
              <pre className="text-xs bg-blue-50 p-2 rounded overflow-x-auto max-h-24 border border-blue-200">
                {JSON.stringify(capability.default_value, null, 2)}
              </pre>
            </div>
          )}
          <div className="pt-2 border-t">
            <Button
              variant="outline"
              size="sm"
              className="w-full"
              onClick={onEdit}
            >
              <Edit className="h-4 w-4 mr-2" />
              Edit Details
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
