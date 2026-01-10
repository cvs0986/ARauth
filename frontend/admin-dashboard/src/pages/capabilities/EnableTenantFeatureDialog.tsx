/**
 * Enable Tenant Feature Dialog
 */

import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { useState } from 'react';
import type { CapabilityEvaluation, EnableTenantFeatureRequest } from '@shared/types/api';

interface EnableTenantFeatureDialogProps {
  availableCapabilities: CapabilityEvaluation[];
  open: boolean;
  onClose: () => void;
  onSave: (key: string, data: EnableTenantFeatureRequest) => void;
}

export function EnableTenantFeatureDialog({
  availableCapabilities,
  open,
  onClose,
  onSave,
}: EnableTenantFeatureDialogProps) {
  const [selectedKey, setSelectedKey] = useState('');
  const [configuration, setConfiguration] = useState('');

  const handleSave = () => {
    if (!selectedKey) {
      alert('Please select a feature');
      return;
    }

    let parsedConfiguration: Record<string, unknown> | undefined;
    if (configuration.trim()) {
      try {
        parsedConfiguration = JSON.parse(configuration);
      } catch (e) {
        alert('Invalid JSON in configuration');
        return;
      }
    }

    onSave(selectedKey, {
      configuration: parsedConfiguration,
    });
  };

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Enable Feature</DialogTitle>
          <DialogDescription>
            Enable a feature for your tenant
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-4">
          <div className="space-y-2">
            <Label htmlFor="feature">Feature</Label>
            <select
              id="feature"
              value={selectedKey}
              onChange={(e) => setSelectedKey(e.target.value)}
              className="w-full px-4 py-2 border rounded-md"
            >
              <option value="">Select a feature...</option>
              {availableCapabilities.map((cap) => (
                <option key={cap.capability_key} value={cap.capability_key}>
                  {cap.capability_key}
                </option>
              ))}
            </select>
            {availableCapabilities.length === 0 && (
              <p className="text-sm text-gray-500">
                No available features to enable. This could mean:
                <br />• No capabilities have been assigned to your tenant (contact system administrator)
                <br />• All assigned capabilities are already enabled as features
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="configuration">Configuration (JSON, optional)</Label>
            <Textarea
              id="configuration"
              value={configuration}
              onChange={(e) => setConfiguration(e.target.value)}
              placeholder='{"key": "value"}'
              rows={6}
              className="font-mono text-sm"
            />
            <p className="text-xs text-gray-500">
              Enter valid JSON for the feature configuration
            </p>
          </div>

          <div className="flex justify-end gap-2 pt-4">
            <Button variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button onClick={handleSave} disabled={!selectedKey || availableCapabilities.length === 0}>
              Enable Feature
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}

