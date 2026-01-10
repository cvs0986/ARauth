/**
 * Tenant Selector Component
 * 
 * GUARDRAIL #1: Backend Is Law
 * - Tenant list fetched from backend API
 * 
 * GUARDRAIL #6: UI Quality Bar
 * - Only shown for SYSTEM users
 * - Clean, professional dropdown
 */

import { useQuery } from '@tanstack/react-query';
import { systemApi } from '@/services/api';
import { usePrincipalContext } from '@/contexts/PrincipalContext';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Building2 } from 'lucide-react';

export function TenantSelector() {
  const { principalType, selectedTenantId, selectTenant } = usePrincipalContext();

  // Only show for SYSTEM users
  if (principalType !== 'SYSTEM') {
    return null;
  }

  // Fetch all tenants for SYSTEM users
  const { data: tenants, isLoading } = useQuery({
    queryKey: ['system', 'tenants'],
    queryFn: () => systemApi.tenants.list(),
    enabled: principalType === 'SYSTEM',
  });

  const handleTenantChange = (tenantId: string) => {
    if (tenantId === 'all') {
      selectTenant(null);
    } else {
      selectTenant(tenantId);
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
