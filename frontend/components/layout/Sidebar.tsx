'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
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
  { name: 'Orders', href: '/dashboard/orders', icon: ShoppingCart },
  { name: 'Invoices', href: '/dashboard/invoices', icon: FileText },
  { name: 'Warehouses', href: '/dashboard/warehouses', icon: Building2 },
  { name: 'Suppliers', href: '/dashboard/suppliers', icon: Truck },
  { name: 'Distributors', href: '/dashboard/distributors', icon: Users },
  { name: 'Users & Roles', href: '/dashboard/users', icon: Users },
  { name: 'RBAC Management', href: '/dashboard/rbac', icon: Shield },
  { name: 'Tenants', href: '/dashboard/tenants', icon: Building2 },
  { name: 'Subscriptions', href: '/dashboard/subscriptions', icon: CreditCard },
  { name: 'Notifications', href: '/dashboard/notifications', icon: Bell },
  { name: 'Audit Logs', href: '/dashboard/audit-logs', icon: Shield },
];

export function Sidebar() {
  const pathname = usePathname();
  const { isAuthenticated } = useAuth();

  return (
    <div className="flex flex-col w-64 bg-background/80 backdrop-blur-xl border-r border-border h-screen fixed left-0 top-0 z-30 shadow-lg">
      {/* Logo section */}
      <Link href={isAuthenticated ? "/dashboard" : "/"} className="flex items-center gap-3 h-16 px-6 border-b border-border hover:bg-muted/50 transition-colors">
        <div className="flex items-center justify-center w-9 h-9 rounded-xl bg-primary text-primary-foreground shadow-sm">
          <Package className="w-5 h-5" />
        </div>
        <h1 className="text-xl font-bold text-foreground">AgroMart</h1>
      </Link>

      {/* Navigation */}
      <div className="flex-1 overflow-y-auto py-6">
        <nav className="px-3 space-y-1.5">
          {navigation.map((item) => {
            const isActive = pathname === item.href;
            return (
              <Link
                key={item.name}
                href={item.href}
                className={cn(
                  'group flex items-center justify-between px-3 py-2.5 text-sm font-medium rounded-xl transition-all duration-200 mx-2',
                  isActive
                    ? 'bg-primary/10 text-primary shadow-sm border border-primary/20'
                    : 'text-muted-foreground hover:bg-muted hover:text-foreground'
                )}
              >
                <div className="flex items-center">
                  <item.icon className={cn(
                    "mr-3 h-5 w-5 transition-all duration-200",
                    isActive
                      ? "text-primary"
                      : "text-muted-foreground group-hover:text-foreground group-hover:scale-110"
                  )} />
                  <span className={cn(
                    "font-medium",
                    isActive ? "text-foreground" : "text-muted-foreground group-hover:text-foreground"
                  )}>{item.name}</span>
                </div>
                {isActive && (
                  <div className="w-1.5 h-1.5 bg-primary rounded-full shadow-sm" />
                )}
              </Link>
            );
          })}
        </nav>
      </div>

      {/* Footer info */}
      <div className="border-t border-border p-4">
        <div className="text-center">
          <p className="text-xs text-muted-foreground">AgroMart Inventory</p>
          <p className="text-xs text-muted-foreground/70 mt-1">v1.0.0</p>
        </div>
      </div>
    </div>
  );
}
