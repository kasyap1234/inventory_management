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
  Settings,
  LogOut,
  Sprout,
  ChevronRight,
  BarChart3,
  Bell,
  Shield,
  CreditCard
} from 'lucide-react';
import { useAuth } from '@/hooks/useAuth';
import { cn } from '@/lib/utils';
import { ThemeToggle } from '@/components/theme-toggle';

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
  { name: 'Subscriptions', href: '/dashboard/subscriptions', icon: CreditCard },
  { name: 'Notifications', href: '/dashboard/notifications', icon: Bell },
  { name: 'Audit Logs', href: '/dashboard/audit-logs', icon: Shield },
  { name: 'Settings', href: '/dashboard/settings', icon: Settings },
];

export function Sidebar() {
  const pathname = usePathname();
  const { logout, user } = useAuth();

  return (
    <div className="flex flex-col w-64 bg-card border-r border-border h-screen fixed left-0 top-0">
      {/* Logo section */}
      <div className="flex items-center justify-between h-16 px-6 border-b border-border">
        <div className="flex items-center gap-3">
          <div className="flex items-center justify-center w-9 h-9 bg-primary rounded-lg">
            <Sprout className="w-5 h-5 text-primary-foreground" />
          </div>
          <h1 className="text-xl font-bold text-foreground">Agromart</h1>
        </div>
        <ThemeToggle />
      </div>
      
      {/* Navigation */}
      <div className="flex-1 overflow-y-auto py-4">
        <nav className="px-3 space-y-1">
          {navigation.map((item) => {
            const isActive = pathname === item.href;
            return (
              <Link
                key={item.name}
                href={item.href}
                className={cn(
                  'group flex items-center justify-between px-3 py-2.5 text-sm font-medium rounded-lg transition-all duration-200',
                  isActive
                    ? 'bg-primary text-primary-foreground'
                    : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
                )}
              >
                <div className="flex items-center gap-3">
                  <item.icon className="h-4 w-4" />
                  <span>{item.name}</span>
                </div>
                {isActive && (
                  <ChevronRight className="h-4 w-4" />
                )}
              </Link>
            );
          })}
        </nav>
      </div>
      
      {/* User section */}
      <div className="border-t border-border p-4">
        <div className="flex items-center gap-3 mb-3 px-3 py-2 bg-accent/50 rounded-lg">
          <div className="flex items-center justify-center w-9 h-9 bg-primary rounded-full text-primary-foreground font-semibold text-sm">
            {user?.first_name?.[0]}{user?.last_name?.[0]}
          </div>
          <div className="flex-1 min-w-0">
            <p className="text-sm font-medium text-foreground truncate">
              {user?.first_name} {user?.last_name}
            </p>
            <p className="text-xs text-muted-foreground truncate">{user?.email}</p>
          </div>
        </div>
        <button
          onClick={() => logout.mutate()}
          className="flex items-center gap-3 w-full px-3 py-2.5 text-sm font-medium text-destructive hover:bg-destructive/10 rounded-lg transition-all duration-200"
        >
          <LogOut className="h-4 w-4" />
          <span>Logout</span>
        </button>
      </div>
    </div>
  );
}
