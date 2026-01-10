/**
 * Dashboard Router
 * 
 * GUARDRAIL #1: Backend Is Law
 * - Routes based on console mode from PrincipalContext
 * 
 * GUARDRAIL #6: UI Quality Bar
 * - Different dashboards for different authority levels
 */

import { usePrincipalContext } from '@/contexts/PrincipalContext';
import { SystemDashboard } from './dashboards/SystemDashboard';
import { TenantDashboard } from './dashboards/TenantDashboard';

export function Dashboard() {
  const { consoleMode } = usePrincipalContext();

  // Route to appropriate dashboard based on console mode
  if (consoleMode === 'SYSTEM') {
    return <SystemDashboard />;
  }

  return <TenantDashboard />;
}
