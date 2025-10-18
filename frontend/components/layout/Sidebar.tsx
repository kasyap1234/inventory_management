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
  CreditCard,
  User as UserIcon
} from 'lucide-react';
import { useAuth } from '@/hooks/useAuth';
import { cn } from '@/lib/utils';

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
  { name: 'My Profile', href: '/dashboard/profile', icon: UserIcon },
  { name: 'Settings', href: '/dashboard/settings', icon: Settings },
];

export function Sidebar() {
  const pathname = usePathname();
  const { logout, user } = useAuth();

  return (
    <div className="flex flex-col w-64 bg-white border-r border-gray-200 h-screen fixed left-0 top-0">
      {/* Logo section with gradient */}
      <div className="flex items-center gap-3 h-16 px-6 border-b border-gray-200">
        <div className="flex items-center justify-center w-10 h-10 gradient-primary rounded-xl shadow-colored">
          <Package className="w-5 h-5 text-white" />
        </div>
        <h1 className="text-xl font-bold gradient-text">AgroMart</h1>
      </div>
      
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
                  'group flex items-center justify-between px-3 py-2.5 text-sm font-medium rounded-lg transition-all duration-200',
                  isActive
                    ? 'bg-gray-900 text-white'
                    : 'text-gray-700 hover:bg-gray-100'
                )}
              >
                <div className="flex items-center">
                  <item.icon className={cn(
                    "mr-3 h-5 w-5 transition-colors duration-200",
                    isActive ? "text-white" : "text-gray-500 group-hover:text-gray-900"
                  )} />
                  <span>{item.name}</span>
                </div>
                {isActive && (
                  <div className="w-1 h-6 bg-blue-500 rounded-full" />
                )}
              </Link>
            );
          })}
        </nav>
      </div>
      
      {/* User section with modern card */}
      <div className="border-t border-gray-200 p-4">
        <Link href="/dashboard/profile">
          <div className="flex items-center gap-3 mb-3 px-3 py-2.5 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors cursor-pointer">
            <div className="flex items-center justify-center w-9 h-9 gradient-primary rounded-full text-white font-semibold text-xs">
              {user?.first_name?.[0]}{user?.last_name?.[0]}
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-semibold text-gray-900 truncate">
                {user?.first_name} {user?.last_name}
              </p>
              <p className="text-xs text-gray-500 truncate">{user?.email}</p>
            </div>
          </div>
        </Link>
        <button
          onClick={() => logout.mutate()}
          className="flex items-center justify-center w-full px-4 py-2 text-sm font-medium text-gray-700 hover:text-red-600 hover:bg-red-50 rounded-lg transition-all duration-200 border border-gray-200 hover:border-red-200"
        >
          <LogOut className="mr-2 h-4 w-4" />
          Logout
        </button>
      </div>
    </div>
  );
}
