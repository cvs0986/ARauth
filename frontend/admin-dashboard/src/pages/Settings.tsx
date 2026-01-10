/**
 * Settings Router
 * 
 * AUTHORITY MODEL:
 * - SYSTEM mode → SystemSettings (platform configuration)
 * - TENANT mode → TenantSettings (tenant configuration)
 * 
 * GUARDRAIL #6: UI Quality Bar
 * - No shared components
 * - No conditionals inside pages
 * - Clean plane separation
 */

import { usePrincipalContext } from '@/contexts/PrincipalContext';
import { SystemSettings } from './settings/SystemSettings';
import { TenantSettings } from './settings/TenantSettings';

export function Settings() {
  const { consoleMode } = usePrincipalContext();

  // Route based on console mode
  if (consoleMode === 'SYSTEM') {
    return <SystemSettings />;
  }

  return <TenantSettings />;
}
