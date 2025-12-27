'use client';

import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { AlertCircle, CheckCircle, Warehouse, Users, Package, ShoppingCart, CreditCard } from 'lucide-react';

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
    if (max === -1) return <CheckCircle className="h-5 w-5 text-primary" />;
    if (isAtLimit(current, max)) return <AlertCircle className="h-5 w-5 text-destructive" />;
    if (isNearLimit(current, max)) return <AlertCircle className="h-5 w-5 text-yellow-500" />;
    return <CheckCircle className="h-5 w-5 text-primary" />;
  };

  // Progress component handles basic color, but we can assume default is primary.
  // We can use indicatorClassName prop if Shadcn supports it or custom class logic if needed.
  // For standard Shadcn, indicator is typically primary. We might need a custom class wrapper for red/yellow states if strict color matching is required,
  // but for now, let's stick to simple layout updates. To change color, we often use `[&>div]:bg-red-500` utility on Progress.

  const getProgressClass = (current: number, max: number) => {
    if (max === -1) return '[&>div]:bg-primary';
    if (isAtLimit(current, max)) return '[&>div]:bg-destructive';
    if (isNearLimit(current, max)) return '[&>div]:bg-yellow-500';
    return '[&>div]:bg-primary';
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
    <Card className="shadow-sm">
      <CardHeader className="pb-4">
        <CardTitle className="flex items-center justify-between text-lg font-semibold">
          <div className="flex items-center gap-2">
            <CreditCard className="h-5 w-5 text-primary" />
            <span>Subscription Usage</span>
          </div>
          <Badge variant="secondary" className="text-sm font-normal">
            {limits.plan_name} Plan
          </Badge>
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-6">
        {usageItems.map((item, index) => {
          const percentage = calculatePercentage(item.current, item.max);
          const Icon = item.icon;

          return (
            <div key={index} className="space-y-3">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <div className="p-1.5 rounded-md bg-muted">
                    <Icon className="h-4 w-4 text-foreground" />
                  </div>
                  <span className="text-sm font-medium text-foreground">{item.label}</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium text-muted-foreground">
                    {item.current} <span className="text-muted-foreground/50">/</span> {formatLimit(item.max)}
                  </span>
                  {getStatusIcon(item.current, item.max)}
                </div>
              </div>

              {item.max !== -1 && (
                <>
                  <Progress
                    value={percentage}
                    className={`h-2 bg-muted ${getProgressClass(item.current, item.max)}`}
                  />
                  {isAtLimit(item.current, item.max) && (
                    <p className="text-xs text-destructive flex items-center gap-1 font-medium">
                      <AlertCircle className="h-3 w-3" />
                      Limit reached. Upgrade to add more.
                    </p>
                  )}
                  {isNearLimit(item.current, item.max) && !isAtLimit(item.current, item.max) && (
                    <p className="text-xs text-yellow-600 flex items-center gap-1">
                      <AlertCircle className="h-3 w-3" />
                      Approaching limit ({Math.round(percentage)}%)
                    </p>
                  )}
                </>
              )}
            </div>
          );
        })}

        {/* Upgrade CTA if any limit is reached */}
        {usageItems.some(item => isAtLimit(item.current, item.max)) && onUpgrade && (
          <div className="pt-4 mt-2">
            <Button
              onClick={onUpgrade}
              className="w-full"
              variant="default"
            >
              Upgrade Plan
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
