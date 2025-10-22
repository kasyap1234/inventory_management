'use client';

import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { AlertCircle, CheckCircle, Warehouse, Users, Package, ShoppingCart } from 'lucide-react';

interface UsageData {
  warehouses_count: number;
  users_count: number;
  products_count: number;
  orders_count_current_month: number;
  suppliers_count: number;
  distributors_count: number;
}

interface LimitsData {
  plan_name: string;
  max_warehouses: number;
  max_users: number;
  max_products: number;
  max_orders_per_month: number;
  max_suppliers: number;
  max_distributors: number;
}

interface SubscriptionUsageCardProps {
  usage: UsageData;
  limits: LimitsData;
  onUpgrade?: () => void;
}

export default function SubscriptionUsageCard({ usage, limits, onUpgrade }: SubscriptionUsageCardProps) {
  const calculatePercentage = (current: number, max: number) => {
    if (max === -1) return 0; // Unlimited
    return Math.min((current / max) * 100, 100);
  };

  const isNearLimit = (current: number, max: number) => {
    if (max === -1) return false;
    return current >= max * 0.8; // 80% threshold
  };

  const isAtLimit = (current: number, max: number) => {
    if (max === -1) return false;
    return current >= max;
  };

  const formatLimit = (limit: number) => {
    return limit === -1 ? 'Unlimited' : limit.toLocaleString();
  };

  const getStatusIcon = (current: number, max: number) => {
    if (max === -1) return <CheckCircle className="h-5 w-5 text-green-500" />;
    if (isAtLimit(current, max)) return <AlertCircle className="h-5 w-5 text-red-500" />;
    if (isNearLimit(current, max)) return <AlertCircle className="h-5 w-5 text-yellow-500" />;
    return <CheckCircle className="h-5 w-5 text-green-500" />;
  };

  const getProgressColor = (current: number, max: number) => {
    if (max === -1) return 'bg-green-500';
    if (isAtLimit(current, max)) return 'bg-red-500';
    if (isNearLimit(current, max)) return 'bg-yellow-500';
    return 'bg-blue-500';
  };

  const usageItems = [
    {
      label: 'Warehouses',
      icon: Warehouse,
      current: usage.warehouses_count,
      max: limits.max_warehouses,
    },
    {
      label: 'Users',
      icon: Users,
      current: usage.users_count,
      max: limits.max_users,
    },
    {
      label: 'Products',
      icon: Package,
      current: usage.products_count,
      max: limits.max_products,
    },
    {
      label: 'Orders (This Month)',
      icon: ShoppingCart,
      current: usage.orders_count_current_month,
      max: limits.max_orders_per_month,
    },
  ];

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center justify-between">
          <span>Subscription Usage</span>
          <span className="text-sm font-normal text-muted-foreground">{limits.plan_name}</span>
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-6">
        {usageItems.map((item, index) => {
          const percentage = calculatePercentage(item.current, item.max);
          const Icon = item.icon;
          
          return (
            <div key={index} className="space-y-2">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <Icon className="h-4 w-4 text-muted-foreground" />
                  <span className="text-sm font-medium">{item.label}</span>
                </div>
                <div className="flex items-center gap-2">
                  {getStatusIcon(item.current, item.max)}
                  <span className="text-sm font-medium">
                    {item.current} / {formatLimit(item.max)}
                  </span>
                </div>
              </div>
              
              {item.max !== -1 && (
                <>
                  <Progress value={percentage} className="h-2" />
                  {isAtLimit(item.current, item.max) && (
                    <p className="text-xs text-red-600 flex items-center gap-1">
                      <AlertCircle className="h-3 w-3" />
                      Limit reached. Please upgrade to add more {item.label.toLowerCase()}.
                    </p>
                  )}
                  {isNearLimit(item.current, item.max) && !isAtLimit(item.current, item.max) && (
                    <p className="text-xs text-yellow-600 flex items-center gap-1">
                      <AlertCircle className="h-3 w-3" />
                      Approaching limit ({Math.round(percentage)}% used)
                    </p>
                  )}
                </>
              )}
            </div>
          );
        })}

        {/* Upgrade CTA if any limit is reached */}
        {usageItems.some(item => isAtLimit(item.current, item.max)) && onUpgrade && (
          <div className="pt-4 border-t">
            <button
              onClick={onUpgrade}
              className="w-full py-2 px-4 bg-gradient-to-r from-blue-500 to-blue-600 text-white rounded-lg font-medium hover:from-blue-600 hover:to-blue-700 transition-all"
            >
              Upgrade Plan
            </button>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
