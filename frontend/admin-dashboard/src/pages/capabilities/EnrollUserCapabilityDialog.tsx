/**
 * Enroll User Capability Dialog
 */

import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { useState } from 'react';
import type { User, CapabilityEvaluation, EnrollUserCapabilityRequest } from '@shared/types/api';

interface EnrollUserCapabilityDialogProps {
  user: User;
  availableCapabilities: CapabilityEvaluation[];
  open: boolean;
  onClose: () => void;
  onSave: (key: string, data: EnrollUserCapabilityRequest) => void;
}

export function EnrollUserCapabilityDialog({
  user,
  availableCapabilities,
  open,
  onClose,
  onSave,
}: EnrollUserCapabilityDialogProps) {
  const [selectedKey, setSelectedKey] = useState('');
  const [stateData, setStateData] = useState('');

  const handleSave = () => {
    if (!selectedKey) {
      alert('Please select a capability');
      return;
    }

    let parsedStateData: Record<string, unknown> | undefined;
    if (stateData.trim()) {
      try {
        parsedStateData = JSON.parse(stateData);
      } catch (e) {
        alert('Invalid JSON in state data');
        return;
      }
    }

    onSave(selectedKey, {
      state_data: parsedStateData,
    });
  };

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Enroll User in Capability</DialogTitle>
          <DialogDescription>
            Enroll {user.username} in a capability
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
                  {cap.capability_key}
                </option>
              ))}
            </select>
            {availableCapabilities.length === 0 && (
              <p className="text-sm text-gray-500">
                No available capabilities. All enabled capabilities are already enrolled.
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="stateData">State Data (JSON, optional)</Label>
            <Textarea
              id="stateData"
              value={stateData}
              onChange={(e) => setStateData(e.target.value)}
              placeholder='{"key": "value"}'
              rows={6}
              className="font-mono text-sm"
            />
            <p className="text-xs text-gray-500">
              Enter valid JSON for the capability state data
            </p>
          </div>

          <div className="flex justify-end gap-2 pt-4">
            <Button variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button onClick={handleSave} disabled={!selectedKey || availableCapabilities.length === 0}>
              Enroll User
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}

