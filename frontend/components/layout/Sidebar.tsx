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
    <div className="flex flex-col w-64 bg-gradient-to-b from-slate-50 via-white to-slate-50 border-r border-gray-200/60 h-screen fixed left-0 top-0 shadow-elegant">
      {/* Logo section with gradient */}
      <div className="flex items-center gap-3 h-20 px-6 border-b border-gray-200/60 bg-white/80 backdrop-blur-sm">
        <div className="flex items-center justify-center w-11 h-11 bg-gradient-to-br from-indigo-600 via-purple-600 to-pink-600 rounded-xl shadow-colored pulse-glow">
          <Sprout className="w-6 h-6 text-white" />
        </div>
        <h1 className="text-2xl font-bold gradient-text">Agromart</h1>
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
                  'group flex items-center justify-between px-4 py-3 text-sm font-medium rounded-xl transition-all duration-300',
                  isActive
                    ? 'bg-gradient-to-r from-indigo-600 via-purple-600 to-pink-600 text-white shadow-colored scale-[1.02]'
                    : 'text-gray-700 hover:bg-gradient-to-r hover:from-indigo-50 hover:to-purple-50 hover:translate-x-1 hover:shadow-sm'
                )}
              >
                <div className="flex items-center">
                  <item.icon className={cn(
                    "mr-3 h-5 w-5 transition-transform duration-300",
                    isActive ? "text-white" : "text-gray-500 group-hover:text-indigo-600 group-hover:scale-110"
                  )} />
                  <span>{item.name}</span>
                </div>
                {isActive && (
                  <ChevronRight className="h-4 w-4 text-white animate-pulse" />
                )}
              </Link>
            );
          })}
        </nav>
      </div>
      
      {/* User section with modern card */}
      <div className="border-t border-gray-200/60 p-4 bg-white/80 backdrop-blur-sm">
        <div className="flex items-center gap-3 mb-3 px-3 py-2.5 bg-gradient-to-r from-indigo-50 via-purple-50 to-pink-50 rounded-xl border border-indigo-100/50">
          <div className="flex items-center justify-center w-10 h-10 bg-gradient-to-br from-indigo-500 via-purple-500 to-pink-500 rounded-full text-white font-semibold text-sm shadow-colored">
            {user?.first_name?.[0]}{user?.last_name?.[0]}
          </div>
          <div className="flex-1 min-w-0">
            <p className="text-sm font-semibold text-gray-900 truncate">
              {user?.first_name} {user?.last_name}
            </p>
            <p className="text-xs text-gray-500 truncate">{user?.email}</p>
          </div>
        </div>
        <button
          onClick={() => logout.mutate()}
          className="flex items-center justify-center w-full px-4 py-2.5 text-sm font-medium text-red-600 hover:text-white hover:bg-gradient-to-r hover:from-red-500 hover:to-red-600 rounded-xl transition-all duration-300 border border-red-200 hover:border-transparent hover:shadow-md hover:scale-[1.02]"
        >
          <LogOut className="mr-2 h-4 w-4" />
          Logout
        </button>
      </div>
    </div>
  );
}
