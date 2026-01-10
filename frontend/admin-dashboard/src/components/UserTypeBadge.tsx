/**
 * User Type Badge Component
 * 
 * GUARDRAIL #1: Backend Is Law
 * - Principal type comes from JWT claims
 * 
 * GUARDRAIL #6: UI Quality Bar
 * - Clear visual indicator of user authority
 */

import { usePrincipalContext } from '@/contexts/PrincipalContext';
import { Shield, Building2 } from 'lucide-react';
import { cn } from '@/lib/utils';

export function UserTypeBadge() {
    const { principalType, systemRoles } = usePrincipalContext();

    if (!principalType) {
        return null;
    }

    const isSystem = principalType === 'SYSTEM';

    // Get role display name
    const getRoleDisplayName = () => {
        if (!isSystem) {
            return 'Tenant Admin';
        }
        if (systemRoles && systemRoles.length > 0) {
            // Format role name: system_owner -> System Owner
            const role = systemRoles[0];
            return role.split('_').map(word =>
                word.charAt(0).toUpperCase() + word.slice(1)
            ).join(' ');
        }
        return 'System Admin';
    };

    const roleDisplayName = getRoleDisplayName();
    const Icon = isSystem ? Shield : Building2;

    return (
        <div className={cn(
            "flex items-center gap-2 px-3 py-1.5 rounded-md text-sm font-medium",
            isSystem
                ? "bg-blue-100 text-blue-800"
                : "bg-gray-100 text-gray-800"
        )}>
            <Icon className="h-4 w-4" />
            <span>{roleDisplayName}</span>
        </div>
    );
}
