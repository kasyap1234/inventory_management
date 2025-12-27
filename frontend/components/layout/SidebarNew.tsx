'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import React, { useMemo } from 'react';
import {
  LayoutDashboard,
  Package,
  Warehouse,
  ShoppingCart,
  FileText,
  Users,
  Truck,
  Building2,
  FolderTree,
  BarChart3,
  Bell,
  Shield,
  CreditCard
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { useAuth } from '@/hooks/useAuth';

const navigation = [
  { name: 'Dashboard', href: '/dashboard', icon: LayoutDashboard },
  { name: 'Analytics', href: '/dashboard/analytics', icon: BarChart3 },
  { name: 'Products', href: '/dashboard/products', icon: Package },
  { name: 'Categories', href: '/dashboard/categories', icon: FolderTree },
  { name: 'Inventory', href: '/dashboard/inventory', icon: Warehouse },
  { name: 'Stock Adjustments', href: '/dashboard/stock-adjustments', icon: Package },
  { name: 'Reservations', href: '/dashboard/reservations', icon: Package },
  { name: 'Orders', href: '/dashboard/orders', icon: ShoppingCart },
  { name: 'Invoices', href: '/dashboard/invoices', icon: FileText },
  { name: 'Reports', href: '/dashboard/reports', icon: FileText },
  { name: 'Warehouses', href: '/dashboard/warehouses', icon: Building2 },
  { name: 'Suppliers', href: '/dashboard/suppliers', icon: Truck },
  { name: 'Distributors', href: '/dashboard/distributors', icon: Users },
  { name: 'Users & Roles', href: '/dashboard/users', icon: Users },
  { name: 'RBAC Management', href: '/dashboard/rbac', icon: Shield },
  { name: 'Tenants', href: '/dashboard/tenants', icon: Building2 },
  { name: 'Subscriptions', href: '/dashboard/subscriptions', icon: CreditCard },
  { name: 'Notifications', href: '/dashboard/notifications', icon: Bell },
  { name: 'Notification Templates', href: '/dashboard/notification-templates', icon: FileText },
  { name: 'Audit Logs', href: '/dashboard/audit-logs', icon: Shield },
  { name: 'Tally Integration', href: '/dashboard/tally', icon: FileText },
  { name: 'Batches', href: '/dashboard/batches', icon: Package },
  { name: 'Alert Rules', href: '/dashboard/alerts', icon: Bell },
];

interface SidebarProps {
  onNavigate?: () => void;
  className?: string; // Allow overriding visibility classes
}

export const Sidebar = React.memo(function Sidebar({ onNavigate, className }: SidebarProps) {
  const pathname = usePathname();
  const { isAuthenticated, user } = useAuth();

  // Memoize filtered navigation to prevent recalculation on every render
  const filteredNavigation = useMemo(() => {
    return navigation.filter((item) => {
      // Permission checks
      if (user?.role === 'super_admin') {
        return item.name === 'Tenants';
      }
      // Hide Tenants from non-super-admins
      if (item.name === 'Tenants') return false;

      // Hide RBAC/Users from non-admins (support both 'admin' and 'tenant_admin' roles)
      const adminRoles = ['admin', 'tenant_admin'];
      if ((item.name === 'Users & Roles' || item.name === 'RBAC Management') &&
        !adminRoles.includes(user?.role || '')) return false;

      return true;
    });
  }, [user?.role]);

  return (
    <div className={cn("border-r bg-background h-full", className)}>
      <div className="flex h-full max-h-screen flex-col gap-2">
        {/* Logo section */}
        <div className="flex h-12 items-center border-b px-4 lg:h-[60px] lg:px-6">
          <Link href={isAuthenticated ? "/dashboard" : "/"} className="flex items-center gap-2 font-semibold" onClick={onNavigate}>
            <Package className="h-6 w-6 text-primary" />
            <span className="text-lg font-bold tracking-tight">AgroMart</span>
          </Link>
        </div>

        {/* Navigation */}
        <div className="flex-1 overflow-y-auto py-1 md:py-2">
          <nav className="grid items-start px-2 text-sm font-medium lg:px-4">
            {filteredNavigation.map((item) => {
              const isActive = pathname === item.href;
              return (
                <Link
                  key={item.name}
                  href={item.href}
                  className={cn(
                    "flex items-center gap-2 md:gap-3 rounded-lg px-3 py-1.5 md:py-2 transition-all hover:text-primary",
                    isActive
                      ? "bg-muted text-primary"
                      : "text-muted-foreground"
                  )}
                  onClick={onNavigate}
                >
                  <item.icon className="h-4 w-4" />
                  {item.name}
                </Link>
              );
            })}
          </nav>
        </div>

        {/* Footer info */}
        <div className="mt-auto p-4 border-t">
          <div className="text-center">
            <p className="text-xs text-muted-foreground font-mono">AgroMart v1.0.0</p>
          </div>
        </div>
      </div>
    </div>
  );
});

