'use client';

import { Badge } from '@/components/ui/badge';
import { CheckCircle, Clock, Package, Truck, Home, XCircle } from 'lucide-react';

interface OrderStatusBadgeProps {
  status: string;
  showIcon?: boolean;
}

const statusConfig = {
  pending: {
    label: 'PENDING',
    color: 'bg-yellow-500/10 text-yellow-500 border-yellow-500/20',
    icon: Clock,
    animate: 'animate-pulse',
  },
  confirmed: {
    label: 'CONFIRMED',
    color: 'bg-blue-500/10 text-blue-500 border-blue-500/20',
    icon: CheckCircle,
    animate: '',
  },
  processing: {
    label: 'PROCESSING',
    color: 'bg-purple-500/10 text-purple-500 border-purple-500/20',
    icon: Package,
    animate: 'animate-pulse',
  },
  shipped: {
    label: 'SHIPPED',
    color: 'bg-indigo-500/10 text-indigo-500 border-indigo-500/20',
    icon: Truck,
    animate: '',
  },
  delivered: {
    label: 'DELIVERED',
    color: 'bg-green-500/10 text-green-500 border-green-500/20',
    icon: Home,
    animate: '',
  },
  cancelled: {
    label: 'CANCELLED',
    color: 'bg-red-500/10 text-red-500 border-red-500/20',
    icon: XCircle,
    animate: '',
  },
};

export function OrderStatusBadge({ status, showIcon = true }: OrderStatusBadgeProps) {
  const config = statusConfig[status as keyof typeof statusConfig] || {
    label: status,
    color: 'bg-gray-500/10 text-gray-500 border-gray-500/20',
    icon: Clock,
    animate: '',
  };

  return (
    <Badge className={`${config.color} border flex items-center gap-2 px-2 py-1 rounded-none font-mono text-xs tracking-wider uppercase shadow-none hover:bg-transparent`}>
      {showIcon && (
        <span className={`relative flex h-2 w-2`}>
          <span className={`${config.animate} absolute inline-flex h-full w-full rounded-full opacity-75 bg-current`}></span>
          <span className="relative inline-flex rounded-full h-2 w-2 bg-current"></span>
        </span>
      )}
      {config.label}
    </Badge>
  );
}
