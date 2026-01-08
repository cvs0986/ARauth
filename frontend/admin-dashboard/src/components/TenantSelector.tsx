/**
 * Tenant Selector Component
 * Allows SYSTEM users to select a tenant context for viewing/managing tenant-specific data
 */

import { useQuery } from '@tanstack/react-query';
import { systemApi } from '@/services/api';
import { useAuthStore } from '@/store/authStore';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Building2 } from 'lucide-react';

export function TenantSelector() {
  const { isSystemUser, selectedTenantId, setSelectedTenantId } = useAuthStore();

  // Only show for SYSTEM users
  if (!isSystemUser()) {
    return null;
  }

  // Fetch all tenants for SYSTEM users
  const { data: tenants, isLoading } = useQuery({
    queryKey: ['system', 'tenants'],
    queryFn: () => systemApi.tenants.list(),
    enabled: isSystemUser(),
  });

  const handleTenantChange = (tenantId: string) => {
    if (tenantId === 'all') {
      setSelectedTenantId(null);
    } else {
      setSelectedTenantId(tenantId);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 text-sm text-gray-600">
        <Building2 className="h-4 w-4" />
        <span>Loading tenants...</span>
      </div>
    );
  }

  return (
    <div className="flex items-center gap-2">
      <Building2 className="h-4 w-4 text-gray-600" />
      <Select
        value={selectedTenantId || 'all'}
        onValueChange={handleTenantChange}
      >
        <SelectTrigger className="w-[200px]">
          <SelectValue placeholder="Select tenant" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Tenants</SelectItem>
          {tenants?.map((tenant) => (
            <SelectItem key={tenant.id} value={tenant.id}>
              {tenant.name} ({tenant.domain})
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
}

