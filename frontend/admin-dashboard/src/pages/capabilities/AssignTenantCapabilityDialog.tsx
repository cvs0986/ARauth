/**
 * Assign Tenant Capability Dialog
 */

import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Switch } from '@/components/ui/switch';
import { useState } from 'react';
import type { Tenant, SystemCapability, TenantCapability, SetTenantCapabilityRequest } from '@shared/types/api';

interface AssignTenantCapabilityDialogProps {
  tenant: Tenant;
  systemCapabilities: SystemCapability[];
  existingCapabilities: TenantCapability[];
  open: boolean;
  onClose: () => void;
  onSave: (key: string, data: SetTenantCapabilityRequest) => void;
}

export function AssignTenantCapabilityDialog({
  tenant,
  systemCapabilities,
  existingCapabilities,
  open,
  onClose,
  onSave,
}: AssignTenantCapabilityDialogProps) {
  const [selectedKey, setSelectedKey] = useState('');
  const [enabled, setEnabled] = useState(true);
  const [value, setValue] = useState('');

  const availableCapabilities = systemCapabilities.filter(
    (cap) => !existingCapabilities.some((ec) => ec.capability_key === cap.capability_key)
  );

  const handleSave = () => {
    if (!selectedKey) {
      alert('Please select a capability');
      return;
    }

    let parsedValue: Record<string, unknown> | undefined;
    if (value.trim()) {
      try {
        parsedValue = JSON.parse(value);
      } catch (e) {
        alert('Invalid JSON in value');
        return;
      }
    }

    onSave(selectedKey, {
      enabled,
      value: parsedValue,
    });
  };

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Assign Capability to Tenant</DialogTitle>
          <DialogDescription>
            Assign a system capability to {tenant.name}
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-4">
          <div className="space-y-2">
            <Label htmlFor="capability">Capability</Label>
            <select
              id="capability"
              value={selectedKey}
              onChange={(e) => setSelectedKey(e.target.value)}
              className="w-full px-4 py-2 border rounded-md"
            >
              <option value="">Select a capability...</option>
              {availableCapabilities.map((cap) => (
                <option key={cap.capability_key} value={cap.capability_key}>
                  {cap.capability_key} - {cap.description || 'No description'}
                </option>
              ))}
            </select>
          </div>

          <div className="flex items-center justify-between">
            <Label htmlFor="enabled">Enabled</Label>
            <Switch id="enabled" checked={enabled} onCheckedChange={setEnabled} />
          </div>

          <div className="space-y-2">
            <Label htmlFor="value">Value (JSON, optional)</Label>
            <Textarea
              id="value"
              value={value}
              onChange={(e) => setValue(e.target.value)}
              placeholder='{"key": "value"}'
              rows={6}
              className="font-mono text-sm"
            />
            <p className="text-xs text-gray-500">
              Enter valid JSON for the capability value
            </p>
          </div>

          <div className="flex justify-end gap-2 pt-4">
            <Button variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button onClick={handleSave}>Assign Capability</Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}

