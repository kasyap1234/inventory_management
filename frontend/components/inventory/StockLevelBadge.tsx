'use client';

import { Badge } from '@/components/ui/badge';
import { AlertTriangle, CheckCircle, XCircle } from 'lucide-react';

interface StockLevelBadgeProps {
  quantity: number;
  minimumLevel?: number;
  showIcon?: boolean;
}

export function StockLevelBadge({ quantity, minimumLevel = 0, showIcon = true }: StockLevelBadgeProps) {
  const getStatus = () => {
    if (quantity === 0) {
      return {
        label: 'Out of Stock',
        color: 'bg-red-100 text-red-800',
        icon: XCircle,
      };
    }
    if (quantity <= minimumLevel) {
      return {
        label: 'Low Stock',
        color: 'bg-orange-100 text-orange-800',
        icon: AlertTriangle,
      };
    }
    return {
      label: 'In Stock',
      color: 'bg-green-100 text-green-800',
      icon: CheckCircle,
    };
  };

  const status = getStatus();
  const Icon = status.icon;

  return (
    <Badge className={`${status.color} flex items-center gap-1`}>
      {showIcon && <Icon className="h-3 w-3" />}
      {status.label}
    </Badge>
  );
}
