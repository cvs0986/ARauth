/**
 * Metric Card Component
 * 
 * GUARDRAIL #6: UI Quality Bar
 * - Clean, professional metric display
 * - No clutter, clear hierarchy
 */

import { Card, CardContent } from '@/components/ui/card';
import type { LucideIcon } from 'lucide-react';
import { cn } from '@/lib/utils';

interface MetricCardProps {
    title: string;
    value: string | number;
    subtitle?: string;
    icon: LucideIcon;
    variant?: 'default' | 'primary' | 'success' | 'warning' | 'danger';
    trend?: {
        value: number;
        label: string;
        isPositive?: boolean;
    };
    onClick?: () => void;
    comingSoon?: boolean;
}

const variantStyles = {
    default: 'text-gray-600',
    primary: 'text-blue-600',
    success: 'text-green-600',
    warning: 'text-yellow-600',
    danger: 'text-red-600',
};

export function MetricCard({
    title,
    value,
    subtitle,
    icon: Icon,
    variant = 'default',
    trend,
    onClick,
    comingSoon = false,
}: MetricCardProps) {
    return (
        <Card
            className={cn(
                "transition-all",
                onClick && "cursor-pointer hover:shadow-md hover:border-blue-300",
                comingSoon && "opacity-60"
            )}
            onClick={comingSoon ? undefined : onClick}
        >
            <CardContent className="p-6">
                <div className="flex items-start justify-between">
                    <div className="flex-1">
                        <p className="text-sm font-medium text-gray-600">{title}</p>
                        {comingSoon ? (
                            <div className="mt-2">
                                <p className="text-2xl font-bold text-gray-400">Coming Soon</p>
                                <p className="text-xs text-gray-400 mt-1">API not yet implemented</p>
                            </div>
                        ) : (
                            <>
                                <p className="text-3xl font-bold mt-2">{value}</p>
                                {subtitle && (
                                    <p className="text-xs text-gray-500 mt-1">{subtitle}</p>
                                )}
                                {trend && (
                                    <div className="flex items-center gap-1 mt-2">
                                        <span className={cn(
                                            "text-sm font-medium",
                                            trend.isPositive === false ? "text-red-600" : "text-green-600"
                                        )}>
                                            {trend.value}%
                                        </span>
                                        <span className="text-xs text-gray-500">{trend.label}</span>
                                    </div>
                                )}
                            </>
                        )}
                    </div>
                    <div className={cn(
                        "p-3 rounded-lg",
                        variant === 'primary' && "bg-blue-100",
                        variant === 'success' && "bg-green-100",
                        variant === 'warning' && "bg-yellow-100",
                        variant === 'danger' && "bg-red-100",
                        variant === 'default' && "bg-gray-100"
                    )}>
                        <Icon className={cn("h-6 w-6", variantStyles[variant])} />
                    </div>
                </div>
            </CardContent>
        </Card>
    );
}
