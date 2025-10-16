'use client';

import { Badge } from '@/components/ui/badge';
import { CheckCircle, Clock, Package, Truck, Home, XCircle } from 'lucide-react';

interface OrderStatusBadgeProps {
  status: string;
  showIcon?: boolean;
}

const statusConfig = {
  pending: {
    label: 'Pending',
    color: 'bg-yellow-100 text-yellow-800',
    icon: Clock,
  },
  confirmed: {
    label: 'Confirmed',
    color: 'bg-blue-100 text-blue-800',
    icon: CheckCircle,
  },
  processing: {
    label: 'Processing',
    color: 'bg-purple-100 text-purple-800',
    icon: Package,
  },
  shipped: {
    label: 'Shipped',
    color: 'bg-indigo-100 text-indigo-800',
    icon: Truck,
  },
  delivered: {
    label: 'Delivered',
    color: 'bg-green-100 text-green-800',
    icon: Home,
  },
  cancelled: {
    label: 'Cancelled',
    color: 'bg-red-100 text-red-800',
    icon: XCircle,
  },
};

export function OrderStatusBadge({ status, showIcon = true }: OrderStatusBadgeProps) {
  const config = statusConfig[status as keyof typeof statusConfig] || {
    label: status,
    color: 'bg-gray-100 text-gray-800',
    icon: Clock,
  };

  const Icon = config.icon;

  return (
    <Badge className={`${config.color} flex items-center gap-1`}>
      {showIcon && <Icon className="h-3 w-3" />}
      {config.label}
    </Badge>
  );
}
