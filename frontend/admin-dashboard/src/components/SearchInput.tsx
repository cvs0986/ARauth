/**
 * Search Input Component
 */

import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';

interface SearchInputProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  label?: string;
}

export function SearchInput({
  value,
  onChange,
  placeholder = 'Search...',
  label,
}: SearchInputProps) {
  return (
    <div className="space-y-2">
      {label && <Label>{label}</Label>}
      <Input
        type="text"
        placeholder={placeholder}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="max-w-sm"
      />
    </div>
  );
}

