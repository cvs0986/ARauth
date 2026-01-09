/**
 * StatCard Component for Dashboard Statistics
 */

import * as React from 'react';
import { Card, CardContent } from '@/components/ui/card';
import { cn } from '@/lib/utils';

interface StatCardProps {
  title: string;
  value: string | number;
  icon: React.ComponentType<{ className?: string }>;
  trend?: {
    value: number;
    label: string;
  };
  variant?: 'default' | 'primary' | 'success' | 'warning' | 'danger';
  onClick?: () => void;
  loading?: boolean;
}

const variantStyles = {
  default: 'bg-white border-gray-200 hover:border-primary-300',
  primary: 'bg-gradient-to-br from-primary-50 to-primary-100 border-primary-200 hover:border-primary-400',
  success: 'bg-gradient-to-br from-accent-50 to-accent-100 border-accent-200 hover:border-accent-400',
  warning: 'bg-gradient-to-br from-yellow-50 to-yellow-100 border-yellow-200 hover:border-yellow-400',
  danger: 'bg-gradient-to-br from-destructive-50 to-destructive-100 border-destructive-200 hover:border-destructive-400',
};

const iconStyles = {
  default: 'text-gray-600',
  primary: 'text-primary-600',
  success: 'text-accent-600',
  warning: 'text-yellow-600',
  danger: 'text-destructive-600',
};

export function StatCard({ 
  title, 
  value, 
  icon: Icon, 
  trend, 
  variant = 'default',
  onClick,
  loading = false 
}: StatCardProps) {
  return (
    <Card 
      className={cn(
        'transition-all duration-200 cursor-pointer',
        variantStyles[variant],
        onClick && 'hover:shadow-lg hover:-translate-y-1'
      )}
      onClick={onClick}
    >
      <CardContent className="p-6">
        <div className="flex items-center justify-between">
          <div className="flex-1">
            <p className="text-sm font-medium text-gray-600 mb-1">{title}</p>
            {loading ? (
              <div className="h-8 w-20 bg-gray-200 rounded animate-pulse"></div>
            ) : (
              <p className="text-3xl font-bold text-gray-900">{value}</p>
            )}
            {trend && !loading && (
              <p className={cn(
                "text-xs mt-2",
                trend.value >= 0 ? "text-accent-600" : "text-destructive-600"
              )}>
                {trend.value >= 0 ? '↑' : '↓'} {Math.abs(trend.value)}% {trend.label}
              </p>
            )}
          </div>
          <div className={cn(
            "p-3 rounded-lg",
            variant === 'default' ? 'bg-gray-100' : 
            variant === 'primary' ? 'bg-primary-200' :
            variant === 'success' ? 'bg-accent-200' :
            variant === 'warning' ? 'bg-yellow-200' :
            'bg-destructive-200'
          )}>
            <Icon className={cn("h-6 w-6", iconStyles[variant])} />
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

