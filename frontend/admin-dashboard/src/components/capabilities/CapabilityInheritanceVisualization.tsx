/**
 * Capability Inheritance Visualization Component
 * Shows the three-layer capability model: System → Tenant → User
 */

import { useQuery } from '@tanstack/react-query';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { systemCapabilityApi, tenantCapabilityApi, tenantFeatureApi, userCapabilityApi } from '@/services/api';
import { useAuthStore } from '@/store/authStore';
import { ArrowDown, CheckCircle, XCircle, AlertCircle } from 'lucide-react';
import type { SystemCapability, TenantCapability, TenantFeature, UserCapabilityState, CapabilityEvaluation } from '@shared/types/api';

interface CapabilityInheritanceVisualizationProps {
  capabilityKey: string;
  tenantId?: string;
  userId?: string;
}

export function CapabilityInheritanceVisualization({
  capabilityKey,
  tenantId,
  userId,
}: CapabilityInheritanceVisualizationProps) {
  const { tenantId: currentTenantId } = useAuthStore();

  const effectiveTenantId = tenantId || currentTenantId;

  // Fetch system capability
  const { data: systemCapability } = useQuery({
    queryKey: ['system', 'capabilities', capabilityKey],
    queryFn: () => systemCapabilityApi.getByKey(capabilityKey),
  });

  // Fetch tenant capability (if tenant is selected)
  const { data: tenantCapability } = useQuery({
    queryKey: ['tenant', 'capabilities', effectiveTenantId, capabilityKey],
    queryFn: () => tenantCapabilityApi.list(effectiveTenantId!),
    enabled: !!effectiveTenantId,
  });

  // Fetch tenant feature (if tenant is selected)
  const { data: tenantFeature } = useQuery({
    queryKey: ['tenant', 'features', capabilityKey],
    queryFn: () => tenantFeatureApi.list(),
    enabled: !!effectiveTenantId,
  });

  // Fetch user capability state (if user is selected)
  const { data: userCapabilityState } = useQuery({
    queryKey: ['user', 'capabilities', userId, capabilityKey],
    queryFn: async () => {
      try {
        return await userCapabilityApi.getByKey(userId!, capabilityKey);
      } catch (error: any) {
        // If user capability doesn't exist (404/400), return null instead of throwing
        if (error?.response?.status === 404 || error?.response?.status === 400) {
          return null;
        }
        throw error;
      }
    },
    enabled: !!userId,
  });

  // Get evaluation
  const { data: evaluation } = useQuery({
    queryKey: ['capability', 'evaluation', effectiveTenantId, userId, capabilityKey],
    queryFn: async () => {
      if (!effectiveTenantId) return null;
      const evaluations = await tenantCapabilityApi.evaluate(effectiveTenantId, userId);
      return evaluations.find((e) => e.capability_key === capabilityKey);
    },
    enabled: !!effectiveTenantId,
  });

  const tenantCap = Array.isArray(tenantCapability) 
    ? tenantCapability.find((tc) => tc.capability_key === capabilityKey)
    : undefined;
  const feature = Array.isArray(tenantFeature)
    ? tenantFeature.find((tf) => tf.capability_key === capabilityKey)
    : undefined;

  return (
    <Card>
      <CardHeader>
        <CardTitle>Capability Inheritance</CardTitle>
        <CardDescription>
          Visual representation of capability flow: System → Tenant → User
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* System Level */}
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <h3 className="font-semibold">System Level</h3>
              {systemCapability ? (
                systemCapability.enabled ? (
                  <CheckCircle className="h-5 w-5 text-green-500" />
                ) : (
                  <XCircle className="h-5 w-5 text-red-500" />
                )
              ) : (
                <AlertCircle className="h-5 w-5 text-gray-400" />
              )}
            </div>
            <Badge variant={systemCapability?.enabled ? 'default' : 'secondary'}>
              {systemCapability?.enabled ? 'Supported' : 'Not Supported'}
            </Badge>
          </div>
          <div className="pl-4 border-l-2 border-gray-300">
            <div className="font-mono text-sm">{capabilityKey}</div>
            {systemCapability?.description && (
              <div className="text-sm text-gray-500 mt-1">{systemCapability.description}</div>
            )}
          </div>
        </div>

        <div className="flex justify-center">
          <ArrowDown className="h-6 w-6 text-gray-400" />
        </div>

        {/* Tenant Level (System → Tenant Assignment) */}
        {effectiveTenantId && (
          <>
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <h3 className="font-semibold">Tenant Assignment</h3>
                  {tenantCap ? (
                    tenantCap.enabled ? (
                      <CheckCircle className="h-5 w-5 text-green-500" />
                    ) : (
                      <XCircle className="h-5 w-5 text-red-500" />
                    )
                  ) : (
                    <AlertCircle className="h-5 w-5 text-gray-400" />
                  )}
                </div>
                <Badge variant={tenantCap?.enabled ? 'default' : 'secondary'}>
                  {tenantCap?.enabled ? 'Allowed' : 'Not Allowed'}
                </Badge>
              </div>
              <div className="pl-4 border-l-2 border-gray-300">
                {tenantCap ? (
                  <>
                    <div className="font-mono text-sm">{capabilityKey}</div>
                    {tenantCap.value && (
                      <div className="text-xs text-gray-500 mt-1 font-mono">
                        Value: {JSON.stringify(tenantCap.value).substring(0, 50)}...
                      </div>
                    )}
                  </>
                ) : (
                  <div className="text-sm text-gray-400">Not assigned to tenant</div>
                )}
              </div>
            </div>

            <div className="flex justify-center">
              <ArrowDown className="h-6 w-6 text-gray-400" />
            </div>

            {/* Tenant Level (Feature Enablement) */}
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <h3 className="font-semibold">Feature Enablement</h3>
                  {feature ? (
                    feature.enabled ? (
                      <CheckCircle className="h-5 w-5 text-green-500" />
                    ) : (
                      <XCircle className="h-5 w-5 text-red-500" />
                    )
                  ) : (
                    <AlertCircle className="h-5 w-5 text-gray-400" />
                  )}
                </div>
                <Badge variant={feature?.enabled ? 'default' : 'secondary'}>
                  {feature?.enabled ? 'Enabled' : 'Not Enabled'}
                </Badge>
              </div>
              <div className="pl-4 border-l-2 border-gray-300">
                {feature ? (
                  <>
                    <div className="font-mono text-sm">{capabilityKey}</div>
                    {feature.configuration && (
                      <div className="text-xs text-gray-500 mt-1 font-mono">
                        Config: {JSON.stringify(feature.configuration).substring(0, 50)}...
                      </div>
                    )}
                  </>
                ) : (
                  <div className="text-sm text-gray-400">Feature not enabled by tenant</div>
                )}
              </div>
            </div>

            {userId && (
              <>
                <div className="flex justify-center">
                  <ArrowDown className="h-6 w-6 text-gray-400" />
                </div>

                {/* User Level */}
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <h3 className="font-semibold">User Enrollment</h3>
                      {userCapabilityState ? (
                        userCapabilityState.enrolled ? (
                          <CheckCircle className="h-5 w-5 text-green-500" />
                        ) : (
                          <XCircle className="h-5 w-5 text-red-500" />
                        )
                      ) : (
                        <AlertCircle className="h-5 w-5 text-gray-400" />
                      )}
                    </div>
                    <Badge variant={userCapabilityState?.enrolled ? 'default' : 'secondary'}>
                      {userCapabilityState?.enrolled ? 'Enrolled' : 'Not Enrolled'}
                    </Badge>
                  </div>
                  <div className="pl-4 border-l-2 border-gray-300">
                    {userCapabilityState ? (
                      <>
                        <div className="font-mono text-sm">{capabilityKey}</div>
                        {userCapabilityState.state_data && (
                          <div className="text-xs text-gray-500 mt-1 font-mono">
                            State: {JSON.stringify(userCapabilityState.state_data).substring(0, 50)}...
                          </div>
                        )}
                      </>
                    ) : (
                      <div className="text-sm text-gray-400">User not enrolled</div>
                    )}
                  </div>
                </div>
              </>
            )}

            {/* Final Evaluation */}
            {evaluation && (
              <div className="mt-6 p-4 bg-gray-50 rounded-lg border-2 border-dashed">
                <div className="flex items-center justify-between mb-2">
                  <h3 className="font-semibold">Final Evaluation</h3>
                  <Badge variant={evaluation.can_use ? 'default' : 'destructive'}>
                    {evaluation.can_use ? 'Can Use' : 'Cannot Use'}
                  </Badge>
                </div>
                <div className="text-sm space-y-1">
                  <div className="flex items-center gap-2">
                    <span>System Supported:</span>
                    <Badge variant={evaluation.system_supported ? 'default' : 'secondary'} className="text-xs">
                      {evaluation.system_supported ? 'Yes' : 'No'}
                    </Badge>
                  </div>
                  <div className="flex items-center gap-2">
                    <span>Tenant Allowed:</span>
                    <Badge variant={evaluation.tenant_allowed ? 'default' : 'secondary'} className="text-xs">
                      {evaluation.tenant_allowed ? 'Yes' : 'No'}
                    </Badge>
                  </div>
                  <div className="flex items-center gap-2">
                    <span>Tenant Enabled:</span>
                    <Badge variant={evaluation.tenant_enabled ? 'default' : 'secondary'} className="text-xs">
                      {evaluation.tenant_enabled ? 'Yes' : 'No'}
                    </Badge>
                  </div>
                  {userId && (
                    <div className="flex items-center gap-2">
                      <span>User Enrolled:</span>
                      <Badge variant={evaluation.user_enrolled ? 'default' : 'secondary'} className="text-xs">
                        {evaluation.user_enrolled ? 'Yes' : 'No'}
                      </Badge>
                    </div>
                  )}
                  {evaluation.reason && (
                    <div className="mt-2 text-xs text-gray-600">
                      <strong>Reason:</strong> {evaluation.reason}
                    </div>
                  )}
                </div>
              </div>
            )}
          </>
        )}

        {!effectiveTenantId && (
          <div className="text-center text-gray-500 py-4">
            Select a tenant to view capability inheritance
          </div>
        )}
      </CardContent>
    </Card>
  );
}

