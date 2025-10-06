'use client';

import { useQuery } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Package, Warehouse, ShoppingCart, FileText } from 'lucide-react';
import { formatCurrency } from '@/lib/utils';

interface DashboardStats {
  totalProducts: number;
  totalInventoryValue: number;
  pendingOrders: number;
  unpaidInvoices: number;
  lowStockProducts: number;
  recentOrdersCount: number;
}

export default function DashboardPage() {
  const { data: stats, isLoading } = useQuery<DashboardStats>({
    queryKey: ['dashboard-stats'],
    queryFn: async () => {
      // In a real app, you'd have a dedicated dashboard stats endpoint
      // For now, we'll use placeholder data
      return {
        totalProducts: 0,
        totalInventoryValue: 0,
        pendingOrders: 0,
        unpaidInvoices: 0,
        lowStockProducts: 0,
        recentOrdersCount: 0,
      };
    },
  });

  const statCards = [
    {
      title: 'Total Products',
      value: stats?.totalProducts || 0,
      icon: Package,
      color: 'text-blue-600',
      bgColor: 'bg-blue-50',
    },
    {
      title: 'Inventory Value',
      value: formatCurrency(stats?.totalInventoryValue || 0),
      icon: Warehouse,
      color: 'text-green-600',
      bgColor: 'bg-green-50',
    },
    {
      title: 'Pending Orders',
      value: stats?.pendingOrders || 0,
      icon: ShoppingCart,
      color: 'text-orange-600',
      bgColor: 'bg-orange-50',
    },
    {
      title: 'Unpaid Invoices',
      value: stats?.unpaidInvoices || 0,
      icon: FileText,
      color: 'text-red-600',
      bgColor: 'bg-red-50',
    },
  ];

  return (
    <div className="space-y-8">
      {/* Header with gradient */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-4xl font-bold bg-gradient-to-r from-gray-900 to-gray-700 bg-clip-text text-transparent">
            Dashboard
          </h1>
          <p className="text-gray-600 mt-2 text-lg">Welcome back! Here's what's happening today.</p>
        </div>
      </div>

      {/* Stats Cards with modern design */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {statCards.map((stat, index) => (
          <Card 
            key={stat.title}
            className="card-hover border-0 shadow-md bg-white overflow-hidden"
            style={{ animationDelay: `${index * 100}ms` }}
          >
            <CardHeader className="flex flex-row items-center justify-between pb-3 pt-6">
              <CardTitle className="text-sm font-semibold text-gray-600 uppercase tracking-wide">
                {stat.title}
              </CardTitle>
              <div className={`p-3 rounded-xl ${stat.bgColor} shadow-sm`}>
                <stat.icon className={`h-6 w-6 ${stat.color}`} />
              </div>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold text-gray-900 mb-1">
                {isLoading ? (
                  <div className="h-9 w-24 bg-gray-200 rounded animate-pulse"></div>
                ) : (
                  stat.value
                )}
              </div>
              <p className="text-sm text-gray-500 flex items-center">
                <span className="text-green-600 font-medium mr-1">+12%</span>
                from last month
              </p>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Activity and Alerts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card className="border-0 shadow-md">
          <CardHeader className="border-b border-gray-100">
            <CardTitle className="text-lg font-bold text-gray-900">Recent Activity</CardTitle>
          </CardHeader>
          <CardContent className="pt-6">
            <div className="space-y-4">
              <p className="text-sm text-gray-500 text-center py-12">
                <Warehouse className="h-12 w-12 text-gray-300 mx-auto mb-3" />
                No recent activity yet
              </p>
            </div>
          </CardContent>
        </Card>

        <Card className="border-0 shadow-md">
          <CardHeader className="border-b border-gray-100">
            <CardTitle className="text-lg font-bold text-gray-900">Low Stock Alerts</CardTitle>
          </CardHeader>
          <CardContent className="pt-6">
            <div className="space-y-4">
              <p className="text-sm text-gray-500 text-center py-12">
                <Package className="h-12 w-12 text-gray-300 mx-auto mb-3" />
                All items are well stocked
              </p>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Quick Actions with modern cards */}
      <Card className="border-0 shadow-md bg-gradient-to-br from-white to-gray-50">
        <CardHeader className="border-b border-gray-100">
          <CardTitle className="text-lg font-bold text-gray-900">Quick Actions</CardTitle>
        </CardHeader>
        <CardContent className="pt-6">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <a
              href="/dashboard/products"
              className="group flex flex-col items-center justify-center p-6 bg-gradient-to-br from-blue-50 to-blue-100/50 rounded-2xl hover:shadow-lg hover:-translate-y-1 transition-all duration-200 border border-blue-100"
            >
              <div className="p-3 bg-blue-600 rounded-xl mb-3 group-hover:scale-110 transition-transform duration-200 shadow-md">
                <Package className="h-6 w-6 text-white" />
              </div>
              <span className="text-sm font-semibold text-blue-900">Add Product</span>
            </a>
            <a
              href="/dashboard/orders"
              className="group flex flex-col items-center justify-center p-6 bg-gradient-to-br from-green-50 to-green-100/50 rounded-2xl hover:shadow-lg hover:-translate-y-1 transition-all duration-200 border border-green-100"
            >
              <div className="p-3 bg-green-600 rounded-xl mb-3 group-hover:scale-110 transition-transform duration-200 shadow-md">
                <ShoppingCart className="h-6 w-6 text-white" />
              </div>
              <span className="text-sm font-semibold text-green-900">New Order</span>
            </a>
            <a
              href="/dashboard/inventory"
              className="group flex flex-col items-center justify-center p-6 bg-gradient-to-br from-purple-50 to-purple-100/50 rounded-2xl hover:shadow-lg hover:-translate-y-1 transition-all duration-200 border border-purple-100"
            >
              <div className="p-3 bg-purple-600 rounded-xl mb-3 group-hover:scale-110 transition-transform duration-200 shadow-md">
                <Warehouse className="h-6 w-6 text-white" />
              </div>
              <span className="text-sm font-semibold text-purple-900">Check Inventory</span>
            </a>
            <a
              href="/dashboard/invoices"
              className="group flex flex-col items-center justify-center p-6 bg-gradient-to-br from-orange-50 to-orange-100/50 rounded-2xl hover:shadow-lg hover:-translate-y-1 transition-all duration-200 border border-orange-100"
            >
              <div className="p-3 bg-orange-600 rounded-xl mb-3 group-hover:scale-110 transition-transform duration-200 shadow-md">
                <FileText className="h-6 w-6 text-white" />
              </div>
              <span className="text-sm font-semibold text-orange-900">View Invoices</span>
            </a>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
