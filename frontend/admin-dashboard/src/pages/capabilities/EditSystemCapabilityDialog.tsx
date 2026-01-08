/**
 * Edit System Capability Dialog
 */

import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Switch } from '@/components/ui/switch';
import { useState } from 'react';
import type { SystemCapability, UpdateSystemCapabilityRequest } from '@shared/types/api';

interface EditSystemCapabilityDialogProps {
  capability: SystemCapability;
  open: boolean;
  onClose: () => void;
  onSave: (data: UpdateSystemCapabilityRequest) => void;
}

export function EditSystemCapabilityDialog({
  capability,
  open,
  onClose,
  onSave,
}: EditSystemCapabilityDialogProps) {
  const [enabled, setEnabled] = useState(capability.enabled);
  const [description, setDescription] = useState(capability.description || '');
  const [defaultValue, setDefaultValue] = useState(
    capability.default_value ? JSON.stringify(capability.default_value, null, 2) : ''
  );

  const handleSave = () => {
    let parsedDefaultValue: Record<string, unknown> | undefined;
    if (defaultValue.trim()) {
      try {
        parsedDefaultValue = JSON.parse(defaultValue);
      } catch (e) {
        alert('Invalid JSON in default value');
        return;
      }
    }

    onSave({
      enabled,
      description: description || undefined,
      default_value: parsedDefaultValue,
    });
  };

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Edit System Capability</DialogTitle>
          <DialogDescription>
            Update the system capability: {capability.capability_key}
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-4">
          <div className="space-y-2">
            <Label>Capability Key</Label>
            <Input value={capability.capability_key} disabled className="font-mono" />
          </div>

          <div className="flex items-center justify-between">
            <Label htmlFor="enabled">Enabled</Label>
            <Switch id="enabled" checked={enabled} onCheckedChange={setEnabled} />
          </div>

          <div className="space-y-2">
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Enter capability description..."
              rows={3}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="defaultValue">Default Value (JSON)</Label>
            <Textarea
              id="defaultValue"
              value={defaultValue}
              onChange={(e) => setDefaultValue(e.target.value)}
              placeholder='{"key": "value"}'
              rows={6}
              className="font-mono text-sm"
            />
            <p className="text-xs text-gray-500">
              Enter valid JSON for the default value
            </p>
          </div>

          <div className="flex justify-end gap-2 pt-4">
            <Button variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button onClick={handleSave}>Save Changes</Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}

